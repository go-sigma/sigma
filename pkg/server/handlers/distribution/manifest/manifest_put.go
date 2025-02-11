// Copyright 2023 sigma
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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/distribution/distribution/v3"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/server/validators"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/counter"
	"github.com/go-sigma/sigma/pkg/utils/imagerefs"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// maxManifestBodySize ...
const maxManifestBodySize = 4 << 20

// PutManifest handles the put manifest request
func (h *handler) PutManifest(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	user, needRet, err := utils.GetUserFromCtxForDs(c)
	if err != nil {
		return err
	}
	if needRet {
		return nil
	}

	uri := c.Request().URL.Path

	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/manifests"), "/v2/")
	_, namespace, _, _, err := imagerefs.Parse(repository)
	if err != nil {
		log.Error().Err(err).Str("Repository", repository).Msg("Repository must container a valid namespace")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestWithNamespace)
	}
	if !(validators.ValidateNamespaceRaw(namespace) && validators.ValidateRepositoryRaw(repository)) {
		log.Error().Err(err).Str("Repository", repository).Msg("Repository must container a valid namespace")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestWithNamespace)
	}
	namespaceObj, err := h.NamespaceServiceFactory.New().GetByName(ctx, namespace)
	if err != nil {
		log.Error().Err(err).Str("Name", repository).Msg("Get repository by name failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeBlobUnknown)
	}

	authChecked, err := h.AuthServiceFactory.New().Namespace(ptr.To(user), namespaceObj.ID, enums.AuthRead)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Msg("Resource not found")
			return xerrors.GenDSErrCodeResourceNotFound(err)
		}
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	if !authChecked {
		log.Error().Int64("UserID", user.ID).Int64("NamespaceID", namespaceObj.ID).Msg("Auth check failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeDenied)
	}

	ref := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")
	if _, err := digest.Parse(ref); err != nil && !consts.TagRegexp.MatchString(ref) {
		log.Error().Err(err).Str("ref", ref).Msg("Invalid digest or tag")
		return xerrors.NewDSError(c, xerrors.DSErrCodeTagInvalid)
	}

	if !strings.Contains(repository, "/") {
		log.Error().Str("repository", repository).Msg("Invalid repository, repository name should have a namespace as prefix")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestWithNamespace)
	}

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

	repositoryService := h.RepositoryServiceFactory.New()
	repositoryObj := &models.Repository{Name: repository}
	err = repositoryService.Create(ctx, repositoryObj, dao.AutoCreateNamespace{
		AutoCreate: h.Config.Namespace.AutoCreate,
		Visibility: h.Config.Namespace.Visibility,
		UserID:     user.ID,
	})
	if err != nil {
		log.Error().Err(err).Str("repository", repository).Msg("Create repository failed")
		e, ok := err.(xerrors.ErrCode) // maybe got exceed tag quota limit error
		if ok {
			return xerrors.NewDSError(c, e)
		}
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	refs.Digest = digest.FromBytes(bodyBytes)

	c.Response().Header().Set(consts.ContentDigest, refs.Digest.String())
	contentType := c.Request().Header.Get("Content-Type")

	manifest, descriptor, err := distribution.UnmarshalManifest(contentType, bodyBytes)
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

	artifactObj := &models.Artifact{
		NamespaceID:  namespaceObj.ID,
		RepositoryID: repositoryObj.ID,
		Digest:       refs.Digest.String(),
		Size:         size,
		BlobsSize:    blobsSize,
		ContentType:  contentType,
		Raw:          bodyBytes,
	}

	referrerID, err := h.getArtifactReferrer(ctx, repository, manifest)
	if err != nil {
		log.Error().Err(err).Str("digest", refs.Digest.String()).Msg("Get artifact referrer failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	artifactObj.ReferrerID = referrerID

	artifactService := h.ArtifactServiceFactory.New()
	tryFindArtifactObj, err := artifactService.GetByDigest(ctx, repositoryObj.ID, refs.Digest.String())
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Str("repository", repositoryObj.Name).Str("digest", refs.Digest.String()).Interface("artifactObj", artifactObj).Msg("Find artifact failed")
			return ptr.Of(xerrors.DSErrCodeUnknown)
		}
	}
	if tryFindArtifactObj != nil {
		artifactObj.ID = tryFindArtifactObj.ID
	}

	if contentType == "application/vnd.docker.distribution.manifest.list.v2+json" ||
		contentType == "application/vnd.oci.image.index.v1+json" {
		artifactObj.Type = enums.ArtifactTypeImageIndex
		err := h.putManifestIndex(ctx, user, digests, repositoryObj, artifactObj, refs, manifest, descriptor)
		if err != nil {
			e, ok := err.(xerrors.ErrCode)
			if ok {
				return xerrors.NewDSError(c, e)
			}
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}
	} else {
		for _, reference := range manifest.References() {
			if reference.MediaType == "application/vnd.oci.image.config.v1+json" || reference.MediaType == "application/vnd.docker.container.image.v1+json" || reference.MediaType == "application/vnd.cncf.helm.config.v1+json" || reference.MediaType == "application/vnd.sylabs.sif.config.v1+json" {
				configRawReader, err := storage.Driver.Reader(ctx, path.Join(consts.Blobs, utils.GenPathByDigest(reference.Digest)))
				if err != nil {
					log.Error().Err(err).Str("digest", reference.Digest.String()).Msg("Get image config raw layer failed")
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
				}
				configRaw, err := io.ReadAll(configRawReader)
				if err != nil {
					log.Error().Err(err).Str("digest", reference.Digest.String()).Msg("Get image config raw layer failed")
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
				}
				artifactObj.ConfigMediaType = ptr.Of(reference.MediaType)
				artifactObj.ConfigRaw = configRaw
			}
			digests = append(digests, reference.Digest.String())
		}

		artifactObj.Type = h.getArtifactType(descriptor, manifest)
		err := h.putManifestManifest(ctx, user, digests, repositoryObj, artifactObj, refs, manifest, descriptor)
		if err != nil {
			e, ok := err.(xerrors.ErrCode)
			if ok {
				return xerrors.NewDSError(c, e)
			}
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}
	}

	return c.NoContent(http.StatusCreated)
}

// putManifestManifest handles the manifest manifest request
// support media type:
// application/vnd.docker.distribution.manifest.v2+json
// application/vnd.oci.image.manifest.v1+json
func (h *handler) putManifestManifest(ctx context.Context, user *models.User, digests []string, repositoryObj *models.Repository, artifactObj *models.Artifact, refs Refs, manifest distribution.Manifest, descriptor distribution.Descriptor) error {
	blobService := h.BlobServiceFactory.New()
	blobObjs, err := blobService.FindByDigests(ctx, digests)
	if err != nil {
		log.Error().Err(err).Str("digest", refs.Digest.String()).Msg("Find blobs failed")
		return xerrors.DSErrCodeUnknown
	}

	artifactObj.Blobs = blobObjs

	err = query.Q.Transaction(func(tx *query.Query) error {
		artifactService := h.ArtifactServiceFactory.New(tx)
		err = artifactService.Create(ctx, artifactObj)
		if err != nil {
			log.Error().Err(err).Str("repository", repositoryObj.Name).Str("digest", refs.Digest.String()).Interface("artifactObj", artifactObj).Msg("Create artifact failed")
			e, ok := err.(xerrors.ErrCode)
			if ok {
				return e
			}
			return xerrors.DSErrCodeUnknown
		}
		if refs.Tag != "" {
			tagService := h.TagServiceFactory.New(tx)
			err = tagService.Create(ctx, &models.Tag{
				RepositoryID: repositoryObj.ID,
				ArtifactID:   artifactObj.ID,
				Name:         refs.Tag,
			}, dao.WithAuditUser(user.ID))
			if err != nil {
				log.Error().Err(err).Str("tag", refs.Tag).Str("digest", refs.Digest.String()).Msg("Create tag failed")
				e, ok := err.(xerrors.ErrCode) // maybe got exceed tag quota error
				if ok {
					return e
				}
				return xerrors.DSErrCodeUnknown
			}
			if workq.ProducerClient != nil { // TODO: init in test
				err = workq.ProducerClient.Produce(ctx, enums.DaemonTagPushed, types.DaemonTagPushedPayload{
					RepositoryID: repositoryObj.ID,
					Tag:          refs.Tag,
				}, definition.ProducerOption{Tx: tx})
				if err != nil {
					log.Error().Err(err).Str("tag", refs.Tag).Str("digest", refs.Digest.String()).Msg("Enqueue tag pushed task failed")
					return xerrors.DSErrCodeUnknown
				}
			}
		}
		if workq.ProducerClient != nil {
			err = workq.ProducerClient.Produce(ctx, enums.DaemonArtifactPushed, types.DaemonArtifactPushedPayload{
				RepositoryID: repositoryObj.ID,
			}, definition.ProducerOption{Tx: tx})
			if err != nil {
				log.Error().Err(err).Str("tag", refs.Tag).Str("digest", refs.Digest.String()).Msg("Enqueue artifact pushed task failed")
				return xerrors.DSErrCodeUnknown
			}
		}
		return nil
	})
	if err != nil {
		e, ok := err.(xerrors.ErrCode)
		if ok {
			return e
		}
		return xerrors.DSErrCodeUnknown
	}

	if needScan(manifest, descriptor) {
		h.putManifestAsyncTask(ctx, artifactObj)
	}

	return nil
}

func needScan(manifest distribution.Manifest, _ distribution.Descriptor) bool {
	if len(manifest.References()) > 0 {
		ref := manifest.References()[0]
		// only image can be scanned
		references := manifest.References()
		for _, descriptor := range references {
			if descriptor.MediaType == "application/vnd.in-toto+json" ||
				descriptor.MediaType == "application/vnd.dev.cosign.simplesigning.v1+json" ||
				descriptor.MediaType == "application/vnd.cncf.helm.chart.content.v1.tar+gzip" ||
				descriptor.MediaType == "application/vnd.cncf.helm.config.v1+json" {
				return false
			}
		}
		if ref.MediaType == "application/vnd.docker.container.image.v1+json" ||
			ref.MediaType == "application/vnd.oci.image.config.v1+json" {
			return true
		}
	}
	return false
}

// putManifestIndex handles the manifest index request
// support media type:
// application/vnd.docker.distribution.manifest.list.v2+json
// application/vnd.oci.image.index.v1+json
func (h *handler) putManifestIndex(ctx context.Context, user *models.User, digests []string, repositoryObj *models.Repository, artifactObj *models.Artifact, refs Refs, _ distribution.Manifest, _ distribution.Descriptor) error {
	artifactService := h.ArtifactServiceFactory.New()
	artifactObjs, err := artifactService.GetByDigests(ctx, repositoryObj.Name, digests)
	if err != nil {
		log.Error().Err(err).Str("repository", repositoryObj.Name).Strs("digests", digests).Msg("Get artifacts failed")
		return xerrors.DSErrCodeUnknown
	}

	artifactObj.ArtifactSubs = artifactObjs

	err = query.Q.Transaction(func(tx *query.Query) error {
		artifactService := h.ArtifactServiceFactory.New(tx)
		err = artifactService.Create(ctx, artifactObj)
		if err != nil {
			log.Error().Err(err).Str("repository", repositoryObj.Name).Str("digest", refs.Digest.String()).Msg("Create artifact failed")
			e, ok := err.(xerrors.ErrCode)
			if ok {
				return e
			}
			return xerrors.DSErrCodeUnknown
		}
		if refs.Tag != "" {
			tagService := h.TagServiceFactory.New(tx)
			err = tagService.Create(ctx, &models.Tag{
				RepositoryID: repositoryObj.ID,
				ArtifactID:   artifactObj.ID,
				Name:         refs.Tag,
			}, dao.WithAuditUser(user.ID))
			if err != nil {
				log.Error().Err(err).Str("repository", repositoryObj.Name).Str("tag", refs.Tag).Msg("Create tag failed")
				e, ok := err.(xerrors.ErrCode)
				if ok {
					return e
				}
				return xerrors.DSErrCodeUnknown
			}
		}
		err = workq.ProducerClient.Produce(ctx, enums.DaemonTagPushed, types.DaemonTagPushedPayload{
			RepositoryID: repositoryObj.ID,
			Tag:          refs.Tag,
		}, definition.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Str("tag", refs.Tag).Str("digest", refs.Digest.String()).Msg("Enqueue tag pushed task failed")
			return xerrors.DSErrCodeUnknown
		}
		return nil
	})
	if err != nil {
		e, ok := err.(xerrors.ErrCode)
		if ok {
			return e
		}
		return xerrors.DSErrCodeUnknown
	}

	return nil
}

func (h *handler) putManifestAsyncTaskSbom(ctx context.Context, artifactObj *models.Artifact) {
	artifactService := h.ArtifactServiceFactory.New()
	err := artifactService.CreateSbom(ctx, &models.ArtifactSbom{
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
	err = workq.ProducerClient.Produce(ctx, enums.DaemonSbom, taskSbomPayload, definition.ProducerOption{})
	if err != nil {
		log.Error().Err(err).Interface("artifactObj", artifactObj).Msg("Enqueue task failed")
		return
	}
}

func (h *handler) putManifestAsyncTaskVulnerability(ctx context.Context, artifactObj *models.Artifact) {
	artifactService := h.ArtifactServiceFactory.New()
	err := artifactService.CreateVulnerability(ctx, &models.ArtifactVulnerability{
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
	err = workq.ProducerClient.Produce(ctx, enums.DaemonVulnerability, taskVulnerabilityPayload, definition.ProducerOption{})
	if err != nil {
		log.Error().Err(err).Interface("artifactObj", artifactObj).Msg("Enqueue task failed")
		return
	}
}

func (h *handler) putManifestAsyncTask(ctx context.Context, artifactObj *models.Artifact) {
	h.putManifestAsyncTaskSbom(ctx, artifactObj)
	h.putManifestAsyncTaskVulnerability(ctx, artifactObj)
}

func (h *handler) getArtifactType(descriptor distribution.Descriptor, manifest distribution.Manifest) enums.ArtifactType {
	if descriptor.MediaType == "application/vnd.docker.distribution.manifest.list.v2+json" ||
		descriptor.MediaType == "application/vnd.oci.image.index.v1+json" {
		return enums.ArtifactTypeImage
	}
	references := manifest.References()
	for _, descriptor := range references {
		if descriptor.MediaType == "application/vnd.in-toto+json" {
			return enums.ArtifactTypeProvenance
		} else if descriptor.MediaType == "application/vnd.dev.cosign.simplesigning.v1+json" {
			return enums.ArtifactTypeCosign
		}
	}
	var mediaType string
	if len(references) == 0 {
		return enums.ArtifactTypeUnknown
	}
	mediaType = references[0].MediaType
	switch mediaType {
	case "application/vnd.oci.image.config.v1+json", "application/vnd.docker.container.image.v1+json":
		return enums.ArtifactTypeImage
	case "application/vnd.cnab.manifest.v1":
		return enums.ArtifactTypeCnab
	case "application/vnd.wasm.config.v1+json":
		return enums.ArtifactTypeWasm
	case "application/vnd.cncf.helm.config.v1+json":
		return enums.ArtifactTypeChart
	case "application/vnd.sylabs.sif.config.v1+json":
		return enums.ArtifactTypeSif
	}
	return enums.ArtifactTypeUnknown
}

// getArtifactReferrer ...
func (h *handler) getArtifactReferrer(ctx context.Context, repository string, manifest distribution.Manifest) (*int64, error) {
	mediaType, data, err := manifest.Payload()
	if err != nil {
		return nil, err
	}
	repositoryService := h.RepositoryServiceFactory.New()
	repositoryObj, err := repositoryService.GetByName(ctx, repository)
	if err != nil {
		return nil, err
	}

	var digest string

	if mediaType == imgspecv1.MediaTypeImageManifest { // nolint: gocritic
		var decoded imgspecv1.Manifest
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			return nil, err
		}
		if decoded.Subject == nil {
			return nil, nil
		}
		digest = decoded.Subject.Digest.String()
	} else if mediaType == imgspecv1.MediaTypeImageIndex {
		var decoded imgspecv1.Index
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			return nil, err
		}
		if decoded.Subject == nil {
			return nil, nil
		}
		digest = decoded.Subject.Digest.String()
	} else {
		return nil, nil
	}

	artifactService := h.ArtifactServiceFactory.New()
	artifactObj, err := artifactService.GetByDigest(ctx, repositoryObj.ID, digest)
	if err != nil {
		return nil, err
	}
	return ptr.Of(artifactObj.ID), nil
}
