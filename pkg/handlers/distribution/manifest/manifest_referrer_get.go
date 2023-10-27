package manifest

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetReferrer ...
func (h *handler) GetReferrer(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())
	uri := c.Request().URL.Path
	ref := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")
	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/referrers"), "/v2/")

	_, err := digest.Parse(ref)
	if err != nil {
		log.Error().Err(err).Str("ref", ref).Msg("Digest is invalid")
		return xerrors.NewDSError(c, xerrors.DSErrCodeDigestInvalid)
	}

	artifactType := c.QueryParam("artifactType")
	c.Response().Header().Set("OCI-Filters-Applied", artifactType)

	var result = imgspecv1.Index{
		MediaType: "application/vnd.oci.image.index.v1+json",
	}
	result.SchemaVersion = 2

	repositoryService := h.repositoryServiceFactory.New()
	repositoryObj, err := repositoryService.GetByName(ctx, repository)
	if err != nil {
		log.Error().Err(err).Str("repository", repository).Msg("Get repository failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	artifactService := h.artifactServiceFactory.New()
	artifactObjs, err := artifactService.GetReferrers(ctx, repositoryObj.ID, ref, strings.Split(artifactType, ","))
	if err != nil {
		log.Error().Err(err).Msg("Get referrers failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	if len(artifactObjs) == 0 {
		return c.JSON(http.StatusOK, result)
	}
	result.Manifests = make([]imgspecv1.Descriptor, 0, len(artifactObjs))
	for _, artifactObj := range artifactObjs {
		var decoded imgspecv1.Manifest
		err = json.Unmarshal(artifactObj.Raw, &decoded)
		if err != nil {
			log.Error().Err(err).Msg("Unmarshal artifact failed")
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}
		result.Manifests = append(result.Manifests, imgspecv1.Descriptor{
			MediaType:    decoded.MediaType,
			Size:         artifactObj.Size,
			Digest:       digest.Digest(artifactObj.Digest),
			ArtifactType: decoded.Config.MediaType,
			Annotations:  decoded.Annotations,
		})
	}
	return c.JSON(http.StatusOK, result)
}
