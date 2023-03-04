// Copyright 2023 XImager
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
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/storage"
	"github.com/ximager/ximager/pkg/utils/counter"
	"github.com/ximager/ximager/pkg/xerrors"
)

// PatchUpload handles the patch upload request
func (h *handler) PatchUpload(c echo.Context) error {
	host := c.Request().Host
	uri := c.Request().URL.Path
	protocol := c.Scheme()

	id := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")
	c.Response().Header().Set(consts.UploadUUID, id)
	location := fmt.Sprintf("%s://%s%s", protocol, host, uri)
	c.Response().Header().Set("Location", location)

	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/blobs"), "/v2/")

	ctx := c.Request().Context()
	blobUploadService := dao.NewBlobUploadService()
	upload, err := blobUploadService.GetLastPart(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Get blob upload record failed")
		return err
	}

	sizeBefore, err := blobUploadService.TotalSizeByUploadID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Get blob upload record failed")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
	}

	counterReader := counter.NewCounter(c.Request().Body)

	path := fmt.Sprintf("%s/%s", consts.BlobUploads, upload.FileID)
	etag, err := storage.Driver.UploadPart(ctx, path, upload.UploadID, int64(upload.PartNumber+1), counterReader)
	if err != nil {
		log.Error().Err(err).Msg("Upload part failed")
		return err
	}

	size := counterReader.Count()
	_, err = blobUploadService.Create(ctx, &models.BlobUpload{
		PartNumber: upload.PartNumber + 1,
		UploadID:   id,
		Etag:       strings.Trim(etag, "\""),
		Repository: repository,
		FileID:     upload.FileID,
		Size:       size,
	})
	if err != nil {
		log.Error().Err(err).Msg("Save blob upload record failed")
		return err
	}
	c.Response().Header().Set("Range", fmt.Sprintf("%d-%d", sizeBefore, sizeBefore+size))

	return c.NoContent(http.StatusAccepted)
}
