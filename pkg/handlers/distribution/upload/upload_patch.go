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

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils/counter"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// PatchUpload handles the patch upload request
func (h *handler) PatchUpload(c echo.Context) error {
	host := c.Request().Host
	uri := c.Request().URL.Path
	protocol := c.Scheme()

	uploadID := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")
	c.Response().Header().Set(consts.UploadUUID, uploadID)
	c.Response().Header().Set("Location", fmt.Sprintf("%s://%s%s", protocol, host, uri))

	repository := h.getRepository(c)

	ctx := log.Logger.WithContext(c.Request().Context())
	blobUploadService := h.blobUploadServiceFactory.New()
	uploadObj, err := blobUploadService.GetLastPart(ctx, uploadID)
	if err != nil {
		log.Error().Err(err).Msg("Get blob upload record failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	sizeBefore, err := blobUploadService.TotalSizeByUploadID(ctx, uploadID)
	if err != nil {
		log.Error().Err(err).Msg("Get blob upload record failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	counterReader := counter.NewCounter(c.Request().Body)

	path := fmt.Sprintf("%s/%s", consts.BlobUploads, uploadObj.FileID)
	etag, err := storage.Driver.UploadPart(ctx, path, uploadObj.UploadID, int64(uploadObj.PartNumber+1), counterReader)
	if err != nil {
		log.Error().Err(err).Msg("Upload part failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	size := counterReader.Count()
	err = blobUploadService.Create(ctx, &models.BlobUpload{
		PartNumber: uploadObj.PartNumber + 1,
		UploadID:   uploadID,
		Etag:       strings.Trim(etag, "\""),
		Repository: repository,
		FileID:     uploadObj.FileID,
		Size:       size,
	})
	if err != nil {
		log.Error().Err(err).Msg("Save blob upload record failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	// Note that the HTTP Range header byte ranges are inclusive and that will be honored, even in non-standard use cases.
	// See: https://docs.docker.com/registry/spec/api/#pushing-a-layer
	c.Response().Header().Set("Range", fmt.Sprintf("%d-%d", sizeBefore, sizeBefore+size-1))

	return c.NoContent(http.StatusAccepted)
}
