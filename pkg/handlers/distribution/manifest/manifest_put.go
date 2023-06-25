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

package manifest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/distribution/distribution/v3"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/daemon"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/types/enums"
	"github.com/ximager/ximager/pkg/utils/counter"
	"github.com/ximager/ximager/pkg/utils/ptr"
	"github.com/ximager/ximager/pkg/xerrors"
)

// maxManifestBodySize ...
const maxManifestBodySize = 4 << 20

// PutManifest handles the put manifest request
func (h *handler) PutManifest(c echo.Context) error {
	uri := c.Request().URL.Path
	ref := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")
	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/manifests"), "/v2/")

	if _, err := digest.Parse(ref); err != nil && !consts.TagRegexp.MatchString(ref) {
		log.Error().Err(err).Str("ref", ref).Msg("Invalid digest or tag")
		return xerrors.NewDSError(c, xerrors.DSErrCodeTagInvalid)
	}

	ctx := log.Logger.WithContext(c.Request().Context())

	countReader := counter.NewCounter(c.Request().Body)
	bodyBytes, err := io.ReadAll(countReader)
	if err != nil {
		log.Error().Err(err).Msg("Read the manifest failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestInvalid)
	}
	size := countReader.Count()
	if size > maxManifestBodySize {
		log.Error().Int64("size", size).Msg("Manifest size exceeds the limit")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestInvalid)
	}

	refs := h.parseRef(ref)

	repositoryService := h.repositoryServiceFactory.New()
	repositoryObj := &models.Repository{
		Name: repository,
	}
	err = repositoryService.Create(ctx, repositoryObj)
	if err != nil {
		log.Error().Err(err).Str("repository", repository).Msg("Create repository failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	refs.Digest = digest.FromBytes(bodyBytes)

	c.Response().Header().Set(consts.ContentDigest, refs.Digest.String())
	contentType := c.Request().Header.Get("Content-Type")

	manifest, _, err := distribution.UnmarshalManifest(contentType, bodyBytes)
	if err != nil {
		log.Error().Err(err).Str("digest", refs.Digest.String()).Msg("Unmarshal manifest failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestInvalid)
	}
	var blobsSize int64
	var digests = make([]string, 0, len(manifest.References())+1)
	for _, reference := range manifest.References() {
		digests = append(digests, reference.Digest.String())
		blobsSize += reference.Size
	}

	fmt.Printf("98: %+v\n", repositoryObj)

	artifactObj := &models.Artifact{
		RepositoryID: repositoryObj.ID,
		Digest:       refs.Digest.String(),
		Size:         size,
		BlobsSize:    blobsSize,
		ContentType:  contentType,
		Raw:          bodyBytes,
	}

	if contentType == "application/vnd.docker.distribution.manifest.list.v2+json" ||
		contentType == "application/vnd.oci.image.index.v1+json" {
		err := h.putManifestIndex(ctx, digests, repositoryObj, artifactObj, refs)
		if err != nil {
			return xerrors.NewDSError(c, ptr.To(err))
		}
	} else {
		err := h.putManifestManifest(ctx, digests, repositoryObj, artifactObj, refs)
		if err != nil {
			return xerrors.NewDSError(c, ptr.To(err))
		}
	}

	return c.NoContent(http.StatusCreated)
}

// putManifestManifest handles the manifest manifest request
// support media type:
// application/vnd.docker.distribution.manifest.v2+json
// application/vnd.oci.image.manifest.v1+json
func (h *handler) putManifestManifest(ctx context.Context, digests []string, repositoryObj *models.Repository, artifactObj *models.Artifact, refs Refs) *xerrors.ErrCode {
	blobService := h.blobServiceFactory.New()
	blobObjs, err := blobService.FindByDigests(ctx, digests)
	if err != nil {
		log.Error().Err(err).Str("digest", refs.Digest.String()).Msg("Find blobs failed")
		return ptr.Of(xerrors.DSErrCodeUnknown)
	}

	artifactObj.Blobs = blobObjs

	err = query.Q.Transaction(func(tx *query.Query) error {
		artifactService := h.artifactServiceFactory.New()
		err = artifactService.Create(ctx, artifactObj)
		if err != nil {
			log.Error().Err(err).Str("repository", repositoryObj.Name).Str("digest", refs.Digest.String()).Interface("artifactObj", artifactObj).Msg("Create artifact failed")
			return ptr.Of(xerrors.DSErrCodeUnknown)
		}
		if refs.Tag != "" {
			tagService := h.tagServiceFactory.New()
			err = tagService.Create(ctx, &models.Tag{
				RepositoryID: repositoryObj.ID,
				ArtifactID:   artifactObj.ID,
				Name:         refs.Tag,
			})
			if err != nil {
				log.Error().Err(err).Str("tag", refs.Tag).Str("digest", refs.Digest.String()).Msg("Create tag failed")
				return ptr.Of(xerrors.DSErrCodeUnknown)
			}
		}
		return nil
	})
	if err != nil {
		return err.(*xerrors.ErrCode)
	}

	h.putManifestAsyncTask(ctx, artifactObj)

	return nil
}

// putManifestIndex handles the manifest index request
// support media type:
// application/vnd.docker.distribution.manifest.list.v2+json
// application/vnd.oci.image.index.v1+json
func (h *handler) putManifestIndex(ctx context.Context, digests []string, repositoryObj *models.Repository, artifactObj *models.Artifact, refs Refs) *xerrors.ErrCode {
	artifactService := h.artifactServiceFactory.New()
	artifactObjs, err := artifactService.GetByDigests(ctx, repositoryObj.Name, digests)
	if err != nil {
		log.Error().Err(err).Str("repository", repositoryObj.Name).Strs("digests", digests).Msg("Get artifacts failed")
		return ptr.Of(xerrors.DSErrCodeUnknown)
	}

	artifactObj.ArtifactIndexes = artifactObjs

	err = query.Q.Transaction(func(tx *query.Query) error {
		err = artifactService.Create(ctx, artifactObj)
		if err != nil {
			log.Error().Err(err).Str("repository", repositoryObj.Name).Str("digest", refs.Digest.String()).Msg("Create artifact failed")
			return ptr.Of(xerrors.DSErrCodeUnknown)
		}
		if refs.Tag != "" {
			tagService := h.tagServiceFactory.New()
			err = tagService.Create(ctx, &models.Tag{
				RepositoryID: repositoryObj.ID,
				ArtifactID:   artifactObj.ID,
				Name:         refs.Tag,
			})
			if err != nil {
				log.Error().Err(err).Str("repository", repositoryObj.Name).Str("tag", refs.Tag).Msg("Create tag failed")
				return ptr.Of(xerrors.DSErrCodeUnknown)
			}
		}
		return nil
	})
	if err != nil {
		return err.(*xerrors.ErrCode)
	}

	return nil
}

func (h *handler) putManifestAsyncTaskSbom(ctx context.Context, artifactObj *models.Artifact) {
	artifactService := h.artifactServiceFactory.New()
	err := artifactService.SaveSbom(ctx, &models.ArtifactSbom{
		ArtifactID: artifactObj.ID,
		Status:     enums.TaskCommonStatusPending,
	})
	if err != nil {
		log.Error().Err(err).Msg("Save sbom failed")
		return
	}

	taskSbomPayload := types.TaskSbom{
		ArtifactID: artifactObj.ID,
	}
	taskSbomPayloadBytes, err := sonic.Marshal(taskSbomPayload)
	if err != nil {
		log.Error().Err(err).Interface("artifactObj", artifactObj).Msg("Marshal task payload failed")
		return
	}
	err = daemon.Enqueue(consts.TopicSbom, taskSbomPayloadBytes)
	if err != nil {
		log.Error().Err(err).Interface("artifactObj", artifactObj).Msg("Enqueue task failed")
		return
	}
}

func (h *handler) putManifestAsyncTaskVulnerability(ctx context.Context, artifactObj *models.Artifact) {
	artifactService := h.artifactServiceFactory.New()
	err := artifactService.SaveVulnerability(ctx, &models.ArtifactVulnerability{
		ArtifactID: artifactObj.ID,
		Status:     enums.TaskCommonStatusPending,
	})
	if err != nil {
		log.Error().Err(err).Msg("Save vulnerability failed")
		return
	}

	taskVulnerabilityPayload := types.TaskVulnerability{
		ArtifactID: artifactObj.ID,
	}
	taskVulnerabilityPayloadBytes, err := sonic.Marshal(taskVulnerabilityPayload)
	if err != nil {
		log.Error().Err(err).Interface("artifactObj", artifactObj).Msg("Marshal task payload failed")
		return
	}
	err = daemon.Enqueue(consts.TopicVulnerability, taskVulnerabilityPayloadBytes)
	if err != nil {
		log.Error().Err(err).Interface("artifactObj", artifactObj).Msg("Enqueue task failed")
		return
	}
}

func (h *handler) putManifestAsyncTask(ctx context.Context, artifactObj *models.Artifact) {
	h.putManifestAsyncTaskSbom(ctx, artifactObj)
	h.putManifestAsyncTaskVulnerability(ctx, artifactObj)
}
