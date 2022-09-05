package blob

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/services/blobs"
)

// HeadBlob returns the blob's size and digest.
func (h *handler) HeadBlob(c echo.Context) error {
	uri := c.Request().URL.Path

	dgest, err := digest.Parse(strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/"))
	if err != nil {
		log.Error().Err(err).Str("digest", c.QueryParam("digest")).Msg("Parse digest failed")
		return err
	}
	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/blobs"), "/v2/")
	log.Debug().Str("digest", dgest.String()).Str("repository", repository).Msg("Blob info")

	ctx := c.Request().Context()

	blobService := blobs.NewBlobService()
	blob, err := blobService.FindByDigest(ctx, dgest.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var err = dtspecv1.ErrorResponse{
				Errors: []dtspecv1.ErrorInfo{
					{
						Code:    "BLOB_UNKNOWN",
						Message: fmt.Sprintf("blob unknown to registry: %s", dgest.String()),
						Detail:  fmt.Sprintf("blob unknown to registry: %s", dgest.String()),
					},
				},
			}
			return c.JSON(http.StatusNotFound, err)
		}
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Check blob exist failed")
		return err
	}
	c.Request().Header.Set(consts.ContentDigest, dgest.String())
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", blob.Size))
	return c.NoContent(http.StatusOK)
}
