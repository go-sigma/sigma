package upload

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/services/blobuploads"
	"github.com/ximager/ximager/pkg/storage"
	"github.com/ximager/ximager/pkg/utils/counter"
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
	blobUploadService := blobuploads.NewBlobUploadService()
	upload, err := blobUploadService.GetLastPart(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Get blob upload record failed")
		return err
	}

	sizeBefore, err := blobUploadService.TotalSizeByUploadID(ctx, id)

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
