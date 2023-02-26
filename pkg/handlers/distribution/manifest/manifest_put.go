// The MIT License (MIT)
//
// Copyright © 2023 Tosone <i@tosone.cn>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/storage"
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

	contentType := c.Request().Header.Get("Content-Type")
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

	var digests []string
	var manifest imgspecv1.Manifest
	err = json.Unmarshal(body, &manifest)
	if err != nil {
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Unmarshal manifest failed")
		return err
	}
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

	return nil
}

func (h *handler) getImageConfig(c echo.Context, dgest digest.Digest, configDescriptor imgspecv1.Descriptor) error {
	ctx := c.Request().Context()
	configReader, err := storage.Driver.Reader(ctx, path.Join(consts.Blobs, utils.GenPathByDigest(configDescriptor.Digest)), 0)
	if err != nil {
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Read config failed")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
	}
	defer configReader.Close() // nolint: errcheck
	configBytes, err := io.ReadAll(configReader)
	if err != nil {
		log.Error().Err(err).Msg("Read config failed")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
	}

	switch configDescriptor.MediaType {
	case schema2.MediaTypeImageConfig:
		var imageConfig imagev1.Image
		err = json.Unmarshal(configBytes, &imageConfig)
		if err != nil {
			log.Error().Err(err).Msg("Unmarshal config failed")
			return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
		}
		log.Info().Interface("config", imageConfig).Msg("config")
	case imgspecv1.MediaTypeImageConfig:
		var imageConfig imgspecv1.Image
		err = json.Unmarshal(configBytes, &imageConfig)
		if err != nil {
			log.Error().Err(err).Msg("Unmarshal config failed")
			return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
		}
	default:
		log.Error().Str("mediaType", configDescriptor.MediaType).Msg("Unsupported media type")
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnsupported)
	}
	log.Info().Interface("config", string(configBytes)).Msg("config")

	return nil
}
