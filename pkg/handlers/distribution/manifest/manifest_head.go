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

// HeadManifest handles the head manifest request
func (h *handler) HeadManifest(c echo.Context) error {
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
	if reference.TagRegexp.MatchString(ref) {
		tagService := dao.NewTagService()
		tag, err := tagService.GetByName(ctx, repository, ref)
		if err != nil {
			log.Error().Err(err).Str("ref", ref).Msg("Get tag failed")
			return err
		}
		dgest = digest.Digest(tag.Digest)
	} else {
		dgest, err = digest.Parse(ref)
		if err != nil {
			log.Error().Err(err).Str("ref", ref).Msg("Parse digest failed")
			return err
		}
	}

	artifactService := dao.NewArtifactService()
	artifact, err := artifactService.GetByDigest(ctx, repository, dgest.String())
	if err != nil {
		log.Error().Err(err).Str("ref", ref).Msg("Get artifact failed")
		return err
	}

	contentType := artifact.ContentType
	if contentType == "" {
		contentType = "application/vnd.docker.distribution.manifest.v2+json"
	}
	c.Request().Header.Set("Content-Type", contentType)

	return c.NoContent(http.StatusOK)
}
