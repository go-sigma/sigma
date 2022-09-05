package upload

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/services/blobs"
	"github.com/ximager/ximager/pkg/services/blobuploads"
	"github.com/ximager/ximager/pkg/storage"
	"github.com/ximager/ximager/pkg/utils"
	"github.com/ximager/ximager/pkg/utils/counter"
)

// PostUpload creates a new upload.
func (h *handler) PostUpload(c echo.Context) error {
	host := c.Request().Host
	uri := c.Request().URL.Path
	protocol := c.Scheme()

	ctx := c.Request().Context()
	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/blobs"), "/v2/")

	fileID := gonanoid.MustGenerate("0123456789abcdefghijklmnopqrstuvwxyz", 64)

	// according to the docker registry api, if the digest is provided, the upload is complete
	if c.QueryParam("digest") != "" {
		dgest, err := digest.Parse(c.QueryParam("digest"))
		if err != nil {
			log.Error().Err(err).Str("digest", c.QueryParam("digest")).Msg("Parse digest failed")
			return err
		}
		c.Response().Header().Set(consts.ContentDigest, dgest.String())

		countReader := counter.NewCounter(c.Request().Body)

		srcPath := fmt.Sprintf("%s/%s", consts.BlobUploads, fileID)
		err = storage.Driver.Upload(ctx, srcPath, countReader)
		if err != nil {
			log.Error().Err(err).Msg("Upload blob failed")
			return err
		}
		destPath := path.Join(consts.Blobs, utils.GenPathByDigest(dgest))
		err = storage.Driver.Move(ctx, srcPath, destPath)
		if err != nil {
			log.Error().Err(err).Msg("Move blob failed")
			return err
		}

		err = storage.Driver.Delete(ctx, srcPath)
		if err != nil {
			log.Error().Err(err).Msg("Delete blob upload failed")
			return err
		}

		size := countReader.Count()

		contentType := c.Request().Header.Get("Content-Type")
		blobService := blobs.NewBlobService()
		_, err = blobService.Create(ctx, &models.Blob{
			Digest:      dgest.String(),
			Size:        size,
			ContentType: contentType,
		})
		if err != nil {
			log.Error().Err(err).Msg("Save blob record failed")
			return err
		}
	}

	id, err := storage.Driver.CreateUploadID(ctx, fmt.Sprintf("%s/%s", consts.BlobUploads, fileID))
	if err != nil {
		log.Error().Err(err).Msg("Create blob upload id failed")
		return err
	}
	c.Response().Header().Set("Docker-Upload-UUID:", id)

	location := fmt.Sprintf("%s://%s%s%s", protocol, host, uri, id)
	c.Response().Header().Set("Location", location)

	blobUploadService := blobuploads.NewBlobUploadService()
	_, err = blobUploadService.Create(ctx, &models.BlobUpload{
		PartNumber: 0,
		UploadID:   id,
		Etag:       "fake",
		Repository: repository,
		FileID:     fileID,
	})
	if err != nil {
		log.Error().Err(err).Msg("Save blob upload record failed")
		return err
	}

	c.Response().Header().Set("Range", "0-0")

	return c.NoContent(http.StatusAccepted)
}
