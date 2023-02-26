// The MIT License (MIT)
//
// Copyright © 2023 Tosone <i@tosone.cn>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package upload

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/distribution/distribution/v3/reference"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/storage"
	"github.com/ximager/ximager/pkg/utils"
	"github.com/ximager/ximager/pkg/utils/counter"
	"github.com/ximager/ximager/pkg/xerrors"
)

// PutUpload handles the put upload request
func (h *handler) PutUpload(c echo.Context) error {
	dgest, err := digest.Parse(c.QueryParam("digest"))
	if err != nil {
		log.Error().Err(err).Str("digest", c.QueryParam("digest")).Msg("Parse digest failed")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeDigestInvalid)
	}
	c.Response().Header().Set(consts.ContentDigest, dgest.String())

	uri := c.Request().URL.Path
	id := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")
	location := fmt.Sprintf("%s://%s%s", c.Scheme(), c.Request().Host, uri)
	c.Response().Header().Set("Location", location)

	ctx := c.Request().Context()

	blobUploadService := dao.NewBlobUploadService()
	upload, err := blobUploadService.GetLastPart(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Get blob upload record failed")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
	}
	srcPath := fmt.Sprintf("%s/%s", consts.BlobUploads, upload.FileID)

	blobService := dao.NewBlobService()
	exist, err := blobService.Exists(ctx, dgest.String())
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Check blob exist failed")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
	}
	if exist {
		err = storage.Driver.AbortUpload(ctx, srcPath, upload.UploadID)
		if err != nil {
			log.Error().Err(err).Msg("Abort upload failed")
			return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
		}
	}

	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/blobs"), "/v2/")
	if !reference.NameRegexp.MatchString(repository) {
		log.Error().Str("repository", repository).Msg("Invalid repository name")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeNameInvalid)
	}

	etags, err := blobUploadService.TotalEtagsByUploadID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Get blob upload etags failed")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
	}

	sizeBefore, err := blobUploadService.TotalSizeByUploadID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Get blob upload size failed")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
	}

	length, err := utils.GetContentLength(c.Request())
	if err != nil {
		log.Error().Err(err).Msg("Get content length failed")
		return err
	}
	if length != 0 {
		counterReader := counter.NewCounter(c.Request().Body)
		etag, err := storage.Driver.UploadPart(ctx, srcPath, upload.UploadID, int64(upload.PartNumber+1), counterReader)
		if err != nil {
			log.Error().Err(err).Msg("Upload part failed")
			return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
		}
		size := counterReader.Count()
		etags = append(etags, etag)
		_, err = blobUploadService.Create(ctx, &models.BlobUpload{
			PartNumber: upload.PartNumber + 1,
			UploadID:   id,
			Etag:       strings.Trim(etag, "\""),
			Repository: repository,
			FileID:     upload.FileID,
			Size:       size,
		})
		if err != nil {
			log.Error().Err(err).Msg("Create blob upload record failed")
			return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
		}
		c.Response().Header().Set("Content-Range", fmt.Sprintf("%d-%d", sizeBefore, sizeBefore+size))
	}

	err = storage.Driver.CommitUpload(ctx, srcPath, id, etags)
	if err != nil {
		log.Error().Err(err).Str("id", id).Strs("etags", etags).Msg("Commit upload failed")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
	}

	destPath := path.Join(consts.Blobs, utils.GenPathByDigest(dgest))
	err = storage.Driver.Move(ctx, srcPath, destPath)
	if err != nil {
		log.Error().Err(err).Str("path", srcPath).Str("digest", dgest.String()).Str("dest", destPath).Msg("Move blob failed")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
	}

	err = storage.Driver.Delete(ctx, srcPath)
	if err != nil {
		log.Error().Err(err).Msg("Delete blob upload failed")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
	}

	err = blobUploadService.DeleteByUploadID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Delete blob upload record failed")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
	}

	contentType := c.Request().Header.Get("Content-Type")
	_, err = blobService.Create(ctx, &models.Blob{
		Digest:      dgest.String(),
		Size:        sizeBefore + length,
		ContentType: contentType,
		PushedAt:    time.Now(),
	})
	if err != nil {
		log.Error().Err(err).Msg("Create blob record failed")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
	}

	return c.NoContent(http.StatusCreated)
}
