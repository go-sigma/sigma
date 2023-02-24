package manifest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/distribution/distribution/v3/reference"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"github.com/ximager/ximager/pkg/dal/dao"
)

// DeleteManifest handles the delete manifest request
func (h *handler) DeleteManifest(c echo.Context) error {
	uri := c.Request().URL.Path
	ref := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")

	if _, err := digest.Parse(ref); err != nil && !reference.TagRegexp.MatchString(ref) {
		log.Debug().Err(err).Str("ref", ref).Msg("not valid digest or tag")
		return fmt.Errorf("not valid digest or tag")
	}

	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/manifests"), "/v2/")

	ctx := c.Request().Context()

	var err error
	var dgest digest.Digest
	if dgest, err = digest.Parse(ref); err == nil {
		dgest, err = digest.Parse(ref)
		if err != nil {
			log.Error().Err(err).Str("ref", ref).Msg("Parse digest failed")
			return err
		}
		artifactService := dao.NewArtifactService()
		err = artifactService.DeleteByDigest(ctx, repository, dgest.String())
		if err != nil {
			log.Error().Err(err).Str("ref", ref).Msg("Delete artifact failed")
			return err
		}
	} else {
		tagService := dao.NewTagService()
		err = tagService.DeleteByName(ctx, repository, ref)
		if err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusAccepted)
}
