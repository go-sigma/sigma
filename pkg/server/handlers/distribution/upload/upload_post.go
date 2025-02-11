// Copyright 2023 sigma
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package upload

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/validators"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/counter"
	"github.com/go-sigma/sigma/pkg/utils/imagerefs"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// PostUpload creates a new upload.
func (h *handler) PostUpload(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	user, needRet, err := utils.GetUserFromCtxForDs(c)
	if err != nil {
		return err
	}
	if needRet {
		return nil
	}

	host := c.Request().Host
	uri := c.Request().URL.Path
	protocol := c.Scheme()

	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/blobs"), "/v2/")
	_, namespace, _, _, err := imagerefs.Parse(repository)
	if err != nil {
		log.Error().Err(err).Str("Repository", repository).Msg("Repository must container a valid namespace")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestWithNamespace)
	}
	if !(validators.ValidateNamespaceRaw(namespace) && validators.ValidateRepositoryRaw(repository)) {
		log.Error().Err(err).Str("Repository", repository).Msg("Repository must container a valid namespace")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestWithNamespace)
	}
	namespaceObj, err := h.NamespaceServiceFactory.New().GetByName(ctx, namespace)
	if err != nil {
		log.Error().Err(err).Str("Name", repository).Msg("Get repository by name failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeBlobUnknown)
	}

	authChecked, err := h.AuthServiceFactory.New().Namespace(ptr.To(user), namespaceObj.ID, enums.AuthManage)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Msg("Resource not found")
			return xerrors.GenDSErrCodeResourceNotFound(err)
		}
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	if !authChecked {
		log.Error().Int64("UserID", user.ID).Int64("NamespaceID", namespaceObj.ID).Msg("Auth check failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeDenied)
	}

	// fileID is the filename that upload to the blob_uploads
	fileID := gonanoid.MustGenerate(consts.Alphanum, 64)

	// according to the docker registry api, if the digest is provided, the upload is complete
	if c.QueryParam("digest") != "" {
		dgest, err := digest.Parse(c.QueryParam("digest"))
		if err != nil {
			log.Error().Err(err).Str("digest", c.QueryParam("digest")).Msg("Parse digest failed")
			return xerrors.NewDSError(c, xerrors.DSErrCodeBlobUploadInvalid)
		}
		c.Response().Header().Set(consts.ContentDigest, dgest.String())

		countReader := counter.NewCounter(c.Request().Body)

		srcPath := fmt.Sprintf("%s/%s", consts.BlobUploads, fileID)
		err = storage.Driver.Upload(ctx, srcPath, countReader)
		if err != nil {
			log.Error().Err(err).Msg("Upload blob failed")
			return xerrors.NewDSError(c, xerrors.DSErrCodeBlobUploadInvalid)
		}
		destPath := path.Join(consts.Blobs, utils.GenPathByDigest(dgest))
		err = storage.Driver.Move(ctx, srcPath, destPath)
		if err != nil {
			log.Error().Err(err).Msg("Move blob failed")
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}

		err = storage.Driver.Delete(ctx, srcPath)
		if err != nil {
			log.Error().Err(err).Msg("Delete blob upload failed")
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}

		size := countReader.Count()

		contentType := c.Request().Header.Get("Content-Type")
		blobService := h.BlobServiceFactory.New()
		err = blobService.Create(ctx, &models.Blob{
			Digest:      dgest.String(),
			Size:        size,
			ContentType: contentType,
		})
		if err != nil {
			log.Error().Err(err).Msg("Save blob record failed")
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}
	}

	uploadID, err := storage.Driver.CreateUploadID(ctx, fmt.Sprintf("%s/%s", consts.BlobUploads, fileID))
	if err != nil {
		log.Error().Err(err).Msg("Create blob upload id failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	blobUploadService := h.BlobUploadServiceFactory.New()
	err = blobUploadService.Create(ctx, &models.BlobUpload{
		PartNumber: 0,
		UploadID:   uploadID,
		Etag:       "fake",
		Repository: repository,
		FileID:     fileID,
	})
	if err != nil {
		log.Error().Err(err).Msg("Save blob upload record failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	c.Response().Header().Set("Docker-Upload-UUID", uploadID)
	c.Response().Header().Set("Location", fmt.Sprintf("%s://%s%s%s", protocol, host, uri, uploadID))
	c.Response().Header().Set("Range", "0-0")

	return c.NoContent(http.StatusAccepted)
}
