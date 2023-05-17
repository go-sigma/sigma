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
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/distribution/distribution/v3/reference"
	"github.com/labstack/echo/v4"
	imagev1 "github.com/moby/moby/image"
	"github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/daemon"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/storage"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/types/enums"
	"github.com/ximager/ximager/pkg/utils"
	"github.com/ximager/ximager/pkg/utils/counter"
	"github.com/ximager/ximager/pkg/xerrors"
)

// PutManifest handles the put manifest request
func (h *handler) PutManifest(c echo.Context) error {
	ctx := c.Request().Context()
	uri := c.Request().URL.Path
	ref := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")
	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/manifests"), "/v2/")

	if _, err := digest.Parse(ref); err != nil && !reference.TagRegexp.MatchString(ref) {
		log.Debug().Err(err).Str("ref", ref).Msg("not valid digest or tag")
		return fmt.Errorf("not valid digest or tag")
	}

	contentType := c.Request().Header.Get("Content-Type")
	if contentType == "application/vnd.docker.distribution.manifest.list.v2+json" ||
		contentType == "application/vnd.oci.image.index.v1+json" {
		return h.manifestList(c, repository, ref)
	}

	countReader := counter.NewCounter(c.Request().Body)
	body, err := io.ReadAll(countReader)
	if err != nil {
		log.Error().Err(err).Msg("Read the manifest failed")
		return err
	}
	size := countReader.Count()

	var dgest digest.Digest
	isTag := false
	if dgest, err = digest.Parse(ref); err == nil {
	} else {
		isTag = true
		dgest = digest.FromBytes(body)
		c.Response().Header().Set(consts.ContentDigest, dgest.String())
	}

	repositoryService := dao.NewRepositoryService()
	repoObj, err := repositoryService.Save(ctx, &models.Repository{
		Name: repository,
	})
	if err != nil {
		log.Error().Err(err).Str("repository", repository).Msg("Create repository failed")
		return err
	}

	artifactService := dao.NewArtifactService()
	artifactObj, err := artifactService.Save(ctx, &models.Artifact{
		RepositoryID: repoObj.ID,
		Digest:       dgest.String(),
		Size:         size,
		ContentType:  contentType,
		Raw:          string(body),
		PushedAt:     time.Now(),
		PullTimes:    0,
		LastPull:     sql.NullTime{},
	})
	if err != nil {
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Create artifact failed")
		return err
	}

	if isTag {
		tag := ref
		tagService := dao.NewTagService()
		_, err = tagService.Save(ctx, &models.Tag{
			RepositoryID: repoObj.ID,
			ArtifactID:   artifactObj.ID,
			Name:         tag,
			PushedAt:     time.Now(),
			LastPull:     sql.NullTime{},
			PullTimes:    0,
		})
		if err != nil {
			log.Error().Err(err).Str("tag", tag).Str("digest", dgest.String()).Msg("Create tag failed")
			return err
		}
	}

	var manifest imgspecv1.Manifest
	err = json.Unmarshal(body, &manifest)
	if err != nil {
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Unmarshal manifest failed")
		return err
	}
	var digests = make([]string, 0, len(manifest.Layers)+1)
	digests = append(digests, manifest.Config.Digest.String())
	for _, layer := range manifest.Layers {
		digests = append(digests, layer.Digest.String())
	}

	blobService := dao.NewBlobService()
	bs, err := blobService.FindByDigests(ctx, digests)
	if err != nil {
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Find blobs failed")
		return err
	}

	err = artifactService.AssociateBlobs(ctx, artifactObj, bs)
	if err != nil {
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Associate blobs failed")
		return err
	}

	_, err = artifactService.SaveSbom(ctx, &models.ArtifactSbom{
		ArtifactID: artifactObj.ID,
		Status:     enums.TaskCommonStatusPending,
	})
	if err != nil {
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Save sbom failed")
		return err
	}

	taskSbomPayload := types.TaskSbom{
		ArtifactID: artifactObj.ID,
	}
	taskSbomPayloadBytes, err := json.Marshal(taskSbomPayload)
	if err != nil {
		log.Error().Err(err).Interface("artifactObj", artifactObj).Msg("Marshal task payload failed")
		return err
	}
	err = daemon.Enqueue(consts.TopicSbom, taskSbomPayloadBytes)
	if err != nil {
		log.Error().Err(err).Interface("artifactObj", artifactObj).Msg("Enqueue task failed")
		return err
	}

	_, err = artifactService.SaveVulnerability(ctx, &models.ArtifactVulnerability{
		ArtifactID: artifactObj.ID,
		Status:     enums.TaskCommonStatusPending,
	})
	if err != nil {
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Save vulnerability failed")
		return err
	}

	taskVulnerabilityPayload := types.TaskVulnerability{
		ArtifactID: artifactObj.ID,
	}
	taskVulnerabilityPayloadBytes, err := json.Marshal(taskVulnerabilityPayload)
	if err != nil {
		log.Error().Err(err).Interface("artifactObj", artifactObj).Msg("Marshal task payload failed")
		return err
	}
	err = daemon.Enqueue(consts.TopicVulnerability, taskVulnerabilityPayloadBytes)
	if err != nil {
		log.Error().Err(err).Interface("artifactObj", artifactObj).Msg("Enqueue task failed")
		return err
	}

	return nil
}

// nolint: unused
func (h *handler) getImageConfig(c echo.Context, dgest digest.Digest, configDescriptor imgspecv1.Descriptor) error {
	ctx := c.Request().Context()
	configReader, err := storage.Driver.Reader(ctx, path.Join(consts.Blobs, utils.GenPathByDigest(configDescriptor.Digest)), 0)
	if err != nil {
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Read config failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	defer configReader.Close() // nolint: errcheck
	configBytes, err := io.ReadAll(configReader)
	if err != nil {
		log.Error().Err(err).Msg("Read config failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	switch configDescriptor.MediaType {
	case schema2.MediaTypeImageConfig:
		var imageConfig imagev1.Image
		err = json.Unmarshal(configBytes, &imageConfig)
		if err != nil {
			log.Error().Err(err).Msg("Unmarshal config failed")
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}
		log.Info().Interface("config", imageConfig).Msg("config")
	case imgspecv1.MediaTypeImageConfig:
		var imageConfig imgspecv1.Image
		err = json.Unmarshal(configBytes, &imageConfig)
		if err != nil {
			log.Error().Err(err).Msg("Unmarshal config failed")
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}
	default:
		log.Error().Str("mediaType", configDescriptor.MediaType).Msg("Unsupported media type")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnsupported)
	}
	log.Info().Interface("config", string(configBytes)).Msg("config")

	return nil
}

// manifestList handles the manifest list request
// support media type:
// application/vnd.docker.distribution.manifest.list.v2+json
// application/vnd.oci.image.index.v1+json
func (h *handler) manifestList(c echo.Context, repository, ref string) error {
	ctx := c.Request().Context()
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Error().Err(err).Msg("Read body failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	var imageIndex imgspecv1.Index
	err = json.Unmarshal(bodyBytes, &imageIndex)
	if err != nil {
		log.Error().Err(err).Str("body", string(bodyBytes)).Msg("Decode manifest list failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestInvalid)
	}
	var dgests = make([]string, 0, len(imageIndex.Manifests))
	for _, manifest := range imageIndex.Manifests {
		dgests = append(dgests, manifest.Digest.String())
	}
	artifactService := dao.NewArtifactService()
	artifacts, err := artifactService.GetByDigests(ctx, repository, dgests)
	if err != nil {
		log.Error().Err(err).Str("repository", repository).Interface("digests", dgests).Msg("Get artifacts failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	// ensure all of the artifacts exist
	for _, dgest := range dgests {
		var exist bool
		for _, artifact := range artifacts {
			if artifact.Digest == dgest {
				exist = true
				break
			}
		}
		if !exist {
			log.Error().Str("digest", dgest).Msg("Artifact not found")
			return xerrors.NewDSError(c, xerrors.DSErrCodeManifestUnknown)
		}
	}

	dgest := digest.FromBytes(bodyBytes)

	err = query.Q.Transaction(func(tx *query.Query) error {
		// Save the repository
		repositoryService := dao.NewRepositoryService(tx)
		repoObj, err := repositoryService.Save(ctx, &models.Repository{
			Name: repository,
		})
		if err != nil {
			log.Error().Err(err).Str("repository", repository).Msg("Save repository failed")
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}
		// Save the artifact
		artifactService := dao.NewArtifactService(tx)
		artifactObj, err := artifactService.Save(ctx, &models.Artifact{
			RepositoryID: repoObj.ID,
			Digest:       dgest.String(),
			Size:         uint64(len(bodyBytes)),
			ContentType:  imageIndex.MediaType,
			Raw:          string(bodyBytes),
			PushedAt:     time.Now(),
			PullTimes:    0,
			LastPull:     sql.NullTime{},
		})
		if err != nil {
			log.Error().Err(err).Str("repository", repository).Str("digest", dgest.String()).Msg("Save artifact failed")
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}

		// Save the tag if it is a tag
		if reference.TagRegexp.MatchString(ref) {
			tagService := dao.NewTagService(tx)
			_, err = tagService.Save(ctx, &models.Tag{
				RepositoryID: repoObj.ID,
				ArtifactID:   artifactObj.ID,
				Name:         ref,
				PushedAt:     time.Now(),
				LastPull:     sql.NullTime{},
				PullTimes:    0,
			})
			if err != nil {
				log.Error().Err(err).Str("repository", repository).Str("tag", ref).Msg("Save tag failed")
				return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("Save artifact failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	return nil
}
