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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/manifest/manifestlist"
	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/distribution/reference"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svcanalytics "github.com/go-sigma/sigma/pkg/service/analytics"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

// Media types without a vendored constant definition.
const (
	// helmConfigMediaType is the reserved media type for the Helm chart manifest config.
	// Reference: https://github.com/helm/helm/blob/main/pkg/registry/constants.go
	helmConfigMediaType = "application/vnd.cncf.helm.config.v1+json"
	// helmChartContentMediaType is the reserved media type for Helm chart package content.
	helmChartContentMediaType = "application/vnd.cncf.helm.chart.content.v1.tar+gzip"
	// sifConfigMediaType is the media type for Sylabs SIF image config.
	sifConfigMediaType = "application/vnd.sylabs.sif.config.v1+json"
	// inTotoMediaType is the media type for in-toto attestations.
	inTotoMediaType = "application/vnd.in-toto+json"
	// cosignSimpleSigningMediaType is the media type for cosign simple signing attestations.
	cosignSimpleSigningMediaType = "application/vnd.dev.cosign.simplesigning.v1+json"
)

//go:generate mockgen -mock_names Service=MockDistributionManifestService -destination=manifest_mocks.go -package=manifest github.com/go-sigma/sigma/pkg/service/distribution/manifest Service

// Service encapsulates the OCI registry manifest business logic.
type Service interface {
	// GetNamespaceByName resolves a namespace by name. Used by handlers to resolve
	// the namespace id needed for auth checks before invoking manifest operations.
	GetNamespaceByName(ctx context.Context, name string) (*models.Namespace, error)
	// GetManifest gets a manifest by reference (tag or digest).
	// Returns: manifest bytes, content type, tag object (non-nil when reference was a tag).
	GetManifest(ctx context.Context, namespaceID string, repository string, reference string) ([]byte, string, *models.Tag, error)
	// HeadManifest gets manifest metadata (same as GetManifest, the handler won't write a body).
	HeadManifest(ctx context.Context, namespaceID string, repository string, reference string) ([]byte, string, *models.Tag, error)
	// PutManifest creates/updates a manifest (the most complex operation).
	// Includes: parse manifest, create/update artifact, tag, blob associations, trigger daemons.
	// Returns the manifest digest.
	PutManifest(ctx context.Context, userID string, namespaceID string, repository string, reference string, body []byte, mediaType string) (string, error)
	// DeleteManifest deletes a manifest by digest or tag.
	// Includes: check deletable, delete artifact + tag + associations.
	DeleteManifest(ctx context.Context, namespaceID string, repository string, reference string, userID string) error
	// GetReferrer gets the referrer manifest for a given digest.
	GetReferrer(ctx context.Context, repository string, digest string, artifactTypes []string) ([]byte, error)
}

type service struct {
	dig.In

	Config       *config.Configuration
	RepoArtifact reporegistry.ArtifactRepository
	RepoBlob     reporegistry.BlobRepository
	RepoTag      reporegistry.TagRepository
	RepoRegistry reporegistry.RepositoryRepository
	RepoNs       reponamespace.NamespaceRepository
	SvcAnalytics svcanalytics.Service `optional:"true"`
	Storage      storage.StorageDriver
	Producer     workq.Producer
}

func NewService(digCon *dig.Container) error {
	return digCon.Provide(func(params service) Service {
		return &params
	})
}

// parseRef splits a reference into either a tag or a digest.
func parseManifestRef(ref string) (tag string, dgst digest.Digest) {
	d, err := digest.Parse(ref)
	if err != nil {
		return ref, ""
	}
	return "", d
}

// GetNamespaceByName resolves a namespace by name.
func (s *service) GetNamespaceByName(ctx context.Context, name string) (*models.Namespace, error) {
	return s.RepoNs.GetByName(ctx, name)
}

// GetManifest gets a manifest by reference (tag or digest).
func (s *service) GetManifest(ctx context.Context, _ string, repository string, reference string) ([]byte, string, *models.Tag, error) {
	return s.getOrHeadManifest(ctx, repository, reference, true)
}

// HeadManifest gets manifest metadata (same as GetManifest).
func (s *service) HeadManifest(ctx context.Context, _ string, repository string, reference string) ([]byte, string, *models.Tag, error) {
	return s.getOrHeadManifest(ctx, repository, reference, false)
}

// getOrHeadManifest contains the shared lookup logic for GET and HEAD manifest.
func (s *service) getOrHeadManifest(ctx context.Context, repository string, reference string, recordPull bool) ([]byte, string, *models.Tag, error) {
	repositoryObj, err := s.RepoRegistry.GetByName(ctx, repository)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("cannot find repository", "err", err, "repository", repository)
			return nil, "", nil, errcode.DSErrCodeNameUnknown
		}
		slog.Error("get repository failed", "err", err, "repository", repository)
		return nil, "", nil, errcode.DSErrCodeUnknown
	}

	tagName, refDigest := parseManifestRef(reference)

	var tag *models.Tag
	var artifactDigest string
	if tagName != "" {
		tag, err = s.RepoTag.GetByName(ctx, repositoryObj.ID, tagName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("get artifact failed", "err", err, "ref", reference)
				return nil, "", nil, errcode.DSErrCodeManifestUnknown
			}
			slog.Error("get artifact failed", "err", err, "ref", reference)
			return nil, "", nil, errcode.DSErrCodeManifestUnknown
		}
		err = s.RepoTag.Incr(ctx, tag.ID)
		if err != nil {
			slog.Error("incr tag failed", "err", err, "ref", reference)
		}
		artifactDigest = tag.Artifact.Digest
	} else {
		artifactDigest = refDigest.String()
	}

	artifact, err := s.RepoArtifact.GetByDigest(ctx, repositoryObj.ID, artifactDigest)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get artifact failed", "err", err, "ref", reference)
			return nil, "", tag, errcode.DSErrCodeManifestUnknown
		}
		slog.Error("get artifact failed", "err", err, "ref", reference)
		return nil, "", tag, errcode.DSErrCodeUnknown
	}

	raw, err := s.readManifest(ctx, artifact.Digest)
	if err != nil {
		slog.Error("read manifest failed", "err", err, "digest", artifact.Digest)
		return nil, "", tag, errcode.DSErrCodeUnknown
	}
	if recordPull {
		s.recordPull(ctx, svcanalytics.PullEvent{
			NamespaceID:  repositoryObj.NamespaceID,
			RepositoryID: repositoryObj.ID,
			ArtifactID:   artifact.ID,
			Reference:    reference,
		})
	}

	return raw, artifact.ContentType, tag, nil
}

func (s *service) readManifest(ctx context.Context, digestStr string) ([]byte, error) {
	dgst, err := digest.Parse(digestStr)
	if err != nil {
		return nil, err
	}
	reader, err := s.Storage.Reader(ctx, utils.GenManifestPathByDigest(dgst))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func (s *service) recordPush(ctx context.Context, event svcanalytics.PushEvent) {
	if s.SvcAnalytics == nil {
		return
	}
	if err := s.SvcAnalytics.RecordPush(ctx, event); err != nil {
		slog.Warn("record analytics push failed", "err", err, "namespaceID", event.NamespaceID, "repositoryID", event.RepositoryID)
	}
}

func (s *service) recordPull(ctx context.Context, event svcanalytics.PullEvent) {
	if s.SvcAnalytics == nil {
		return
	}
	if err := s.SvcAnalytics.RecordPull(ctx, event); err != nil {
		slog.Warn("record analytics pull failed", "err", err, "namespaceID", event.NamespaceID, "repositoryID", event.RepositoryID)
	}
}

func (s *service) recordNamespaceSizeDelta(ctx context.Context, namespaceID string, delta int64) {
	if s.SvcAnalytics == nil {
		return
	}
	if err := s.SvcAnalytics.RecordNamespaceSizeDelta(ctx, namespaceID, delta); err != nil {
		slog.Warn("record analytics namespace size delta failed", "err", err, "namespaceID", namespaceID, "delta", delta)
	}
}

func (s *service) recordNamespaceTagDelta(ctx context.Context, namespaceID string, delta int64) {
	if s.SvcAnalytics == nil {
		return
	}
	if err := s.SvcAnalytics.RecordNamespaceTagDelta(ctx, namespaceID, delta); err != nil {
		slog.Warn("record analytics namespace tag delta failed", "err", err, "namespaceID", namespaceID, "delta", delta)
	}
}

func (s *service) ensureRepository(ctx context.Context, userID string, namespaceID string, repository string) (*models.Repository, error) {
	repositoryObj, err := s.RepoRegistry.GetByName(ctx, repository)
	if err == nil {
		return repositoryObj, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get repository failed", "err", err, "repository", repository)
		return nil, errcode.DSErrCodeUnknown
	}

	_, namespaceName, _, _, err := reference.Parse(repository)
	if err != nil {
		slog.Error("parse repository failed", "err", err, "repository", repository)
		return nil, errcode.DSErrCodeManifestWithNamespace
	}

	namespaceObj, createNamespace, err := s.resolveRepositoryNamespace(ctx, namespaceID, namespaceName)
	if err != nil {
		return nil, err
	}

	repositoryObj = &models.Repository{
		ID:          uuid.NewV7String(),
		NamespaceID: namespaceObj.ID,
		Name:        repository,
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		if createNamespace {
			namespaceRepository := reponamespace.NewNamespaceRepository(tx)
			err = namespaceRepository.Create(ctx, namespaceObj)
			if err != nil {
				slog.Error("create namespace failed", "err", err, "namespace", namespaceObj.Name)
				return errcode.DSErrCodeUnknown
			}
			namespaceMemberRepository := reponamespace.NewNamespaceMemberRepository(tx)
			_, err = namespaceMemberRepository.AddNamespaceMember(ctx, userID, *namespaceObj, enums.NamespaceRoleAdmin)
			if err != nil {
				slog.Error("add namespace member failed", "err", err, "namespace", namespaceObj.Name, "user_id", userID)
				return errcode.DSErrCodeUnknown
			}
			err = s.produceNamespaceCreateWebhook(ctx, namespaceObj)
			if err != nil {
				slog.Error("webhook event produce failed", "err", err)
				return errcode.DSErrCodeUnknown
			}
		}

		repoRegistry := reporegistry.NewRepositoryRepository(tx)
		err = repoRegistry.Create(ctx, repositoryObj)
		if err != nil {
			slog.Error("create repository failed", "err", err, "repository", repository)
			return errcode.DSErrCodeUnknown
		}
		err = s.produceRepositoryCreateWebhook(ctx, namespaceObj.ID, repositoryObj)
		if err != nil {
			slog.Error("webhook event produce failed", "err", err)
			return errcode.DSErrCodeUnknown
		}
		return nil
	})
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			return nil, e
		}
		return nil, errcode.DSErrCodeUnknown
	}

	if createNamespace {
		s.seedNamespaceCache(ctx, namespaceObj)
	}
	s.seedRepositoryCache(ctx, repositoryObj)
	return repositoryObj, nil
}

func (s *service) resolveRepositoryNamespace(ctx context.Context, namespaceID string, namespaceName string) (*models.Namespace, bool, error) {
	if namespaceID != "" {
		namespaceObj, err := s.RepoNs.Get(ctx, namespaceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, errcode.DSErrCodeNameUnknown
			}
			slog.Error("get namespace failed", "err", err, "namespace_id", namespaceID)
			return nil, false, errcode.DSErrCodeUnknown
		}
		if namespaceObj.Name != namespaceName {
			slog.Error("repository namespace does not match request namespace", "repository_namespace", namespaceName, "namespace_id", namespaceID)
			return nil, false, errcode.DSErrCodeNameUnknown
		}
		return namespaceObj, false, nil
	}

	namespaceObj, err := s.RepoNs.GetByName(ctx, namespaceName)
	if err == nil {
		return namespaceObj, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get namespace failed", "err", err, "namespace", namespaceName)
		return nil, false, errcode.DSErrCodeUnknown
	}
	if !s.Config.Namespace.AutoCreate {
		return nil, false, errcode.DSErrCodeNameUnknown
	}
	visibility := s.Config.Namespace.Visibility
	if !visibility.IsValid() {
		visibility = enums.VisibilityPrivate
	}
	return &models.Namespace{
		ID:         uuid.NewV7String(),
		Name:       namespaceName,
		Visibility: visibility,
	}, true, nil
}

func (s *service) produceNamespaceCreateWebhook(ctx context.Context, namespaceObj *models.Namespace) error {
	if s.Producer == nil {
		return nil
	}
	namespaceID := namespaceObj.ID
	return s.Producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
		NamespaceID:  &namespaceID,
		Action:       enums.WebhookActionCreate,
		ResourceType: enums.WebhookResourceTypeNamespace,
		Payload:      utils.MustMarshal(namespaceObj),
	})
}

func (s *service) produceRepositoryCreateWebhook(ctx context.Context, namespaceID string, repositoryObj *models.Repository) error {
	if s.Producer == nil {
		return nil
	}
	return s.Producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
		NamespaceID:  &namespaceID,
		Action:       enums.WebhookActionCreate,
		ResourceType: enums.WebhookResourceTypeRepository,
		Payload:      utils.MustMarshal(repositoryObj),
	})
}

// PutManifest creates/updates a manifest.
func (s *service) PutManifest(ctx context.Context, userID string, namespaceID string, repository string, reference string, body []byte, mediaType string) (string, error) {
	tagName, refDigest := parseManifestRef(reference)
	_ = refDigest // digest is computed from the body below

	repositoryObj, err := s.ensureRepository(ctx, userID, namespaceID, repository)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			return "", e
		}
		return "", errcode.DSErrCodeUnknown
	}

	dgst := digest.FromBytes(body)
	manifestPath := utils.GenManifestPathByDigest(dgst)

	manifest, descriptor, err := distribution.UnmarshalManifest(mediaType, body)
	if err != nil {
		slog.Error("unmarshal manifest failed", "err", err, "digest", dgst.String())
		return "", errcode.DSErrCodeManifestInvalid
	}
	err = s.Storage.Upload(ctx, manifestPath, bytes.NewReader(body))
	if err != nil {
		slog.Error("upload manifest failed", "err", err, "digest", dgst.String(), "path", manifestPath)
		return "", errcode.DSErrCodeUnknown
	}

	var blobsSize int64
	var digests = make([]string, 0, len(manifest.References())+1)
	for _, reference := range manifest.References() {
		digests = append(digests, reference.Digest.String())
		blobsSize += reference.Size
	}

	artifactObj := &models.Artifact{
		ID:           uuid.NewV7String(),
		NamespaceID:  repositoryObj.NamespaceID,
		RepositoryID: repositoryObj.ID,
		Digest:       dgst.String(),
		Size:         int64(len(body)),
		BlobsSize:    blobsSize,
		ContentType:  mediaType,
	}

	referrerID, err := s.getArtifactReferrer(ctx, repository, manifest)
	if err != nil {
		slog.Error("get artifact referrer failed", "err", err, "digest", dgst.String())
		return "", errcode.DSErrCodeUnknown
	}
	artifactObj.ReferrerID = referrerID

	tryFindArtifactObj, err := s.RepoArtifact.GetByDigest(ctx, repositoryObj.ID, dgst.String())
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("find artifact failed", "err", err, "repository", repositoryObj.Name, "digest", dgst.String(), "artifactObj", artifactObj)
			return "", errcode.DSErrCodeUnknown
		}
	}
	if tryFindArtifactObj != nil {
		artifactObj.ID = tryFindArtifactObj.ID
	}
	isNewArtifact := tryFindArtifactObj == nil
	isNewTag := false
	if tagName != "" {
		_, err = s.RepoTag.GetByName(ctx, repositoryObj.ID, tagName)
		isNewTag = errors.Is(err, gorm.ErrRecordNotFound)
		if err != nil && !isNewTag {
			slog.Warn("get tag before analytics failed", "err", err, "repository", repositoryObj.Name, "tag", tagName)
		}
	}

	refs := manifestRefs{Tag: tagName, Digest: dgst}

	if mediaType == manifestlist.MediaTypeManifestList ||
		mediaType == imgspecv1.MediaTypeImageIndex {
		artifactObj.Type = enums.ArtifactTypeImageIndex
		err = s.putManifestIndex(ctx, digests, repositoryObj, artifactObj, refs, manifest, descriptor, isNewArtifact)
		if err != nil {
			if e, ok := err.(errcode.ErrCode); ok {
				return "", e
			}
			return "", errcode.DSErrCodeUnknown
		}
	} else {
		for _, reference := range manifest.References() {
			if reference.MediaType == imgspecv1.MediaTypeImageConfig ||
				reference.MediaType == schema2.MediaTypeImageConfig ||
				reference.MediaType == helmConfigMediaType ||
				reference.MediaType == sifConfigMediaType {
				configRawReader, readErr := s.Storage.Reader(ctx, utils.GenBlobPathByDigest(reference.Digest))
				if readErr != nil {
					slog.Error("get image config raw layer failed", "err", readErr, "digest", reference.Digest.String())
					return "", errcode.DSErrCodeUnknown
				}
				defer configRawReader.Close()
				configRaw, readErr := io.ReadAll(configRawReader)
				if readErr != nil {
					slog.Error("get image config raw layer failed", "err", readErr, "digest", reference.Digest.String())
					return "", errcode.DSErrCodeUnknown
				}
				artifactObj.ConfigMediaType = new(reference.MediaType)
				artifactObj.ConfigRaw = configRaw
			}
			digests = append(digests, reference.Digest.String())
		}

		artifactObj.Type = s.getArtifactType(descriptor, manifest)
		err = s.putManifestManifest(ctx, digests, repositoryObj, artifactObj, refs, manifest, descriptor, isNewArtifact)
		if err != nil {
			if e, ok := err.(errcode.ErrCode); ok {
				return "", e
			}
			return "", errcode.DSErrCodeUnknown
		}
	}

	s.recordPush(ctx, svcanalytics.PushEvent{
		UserID:       userID,
		NamespaceID:  namespaceID,
		RepositoryID: repositoryObj.ID,
		ArtifactID:   artifactObj.ID,
		Digest:       dgst.String(),
	})
	if isNewArtifact {
		s.recordNamespaceSizeDelta(ctx, namespaceID, artifactObj.BlobsSize)
	}
	if isNewTag {
		s.recordNamespaceTagDelta(ctx, namespaceID, 1)
	}

	return dgst.String(), nil
}

// manifestRefs is the internal representation of a manifest reference (tag or digest).
type manifestRefs struct {
	Tag    string
	Digest digest.Digest
}

// putManifestManifest handles the single manifest (non-index) request.
// support media type:
// application/vnd.docker.distribution.manifest.v2+json
// application/vnd.oci.image.manifest.v1+json
func (s *service) putManifestManifest(ctx context.Context, digests []string, repositoryObj *models.Repository, artifactObj *models.Artifact, refs manifestRefs, manifest distribution.Manifest, descriptor distribution.Descriptor, isNewArtifact bool) error {
	blobObjs, err := s.RepoBlob.FindByDigests(ctx, digests)
	if err != nil {
		slog.Error("find blobs failed", "err", err, "digest", refs.Digest.String())
		return errcode.DSErrCodeUnknown
	}

	artifactObj.Blobs = blobObjs

	err = query.Q.Transaction(func(tx *query.Query) error {
		artifactRepository := reporegistry.NewArtifactRepository(tx)
		err = artifactRepository.Create(ctx, artifactObj)
		if err != nil {
			slog.Error("create artifact failed", "err", err, "repository", repositoryObj.Name, "digest", refs.Digest.String(),
				"artifactObj", artifactObj)
			if e, ok := err.(errcode.ErrCode); ok {
				return e
			}
			return errcode.DSErrCodeUnknown
		}
		if refs.Tag != "" {
			tagRepository := reporegistry.NewTagRepository(tx)
			err = tagRepository.Create(ctx, &models.Tag{
				ID:           uuid.NewV7String(),
				RepositoryID: repositoryObj.ID,
				ArtifactID:   artifactObj.ID,
				Name:         refs.Tag,
			})
			if err != nil {
				slog.Error("create tag failed", "err", err, "tag", refs.Tag, "digest", refs.Digest.String())
				if e, ok := err.(errcode.ErrCode); ok { // maybe got exceed tag quota error
					return e
				}
				return errcode.DSErrCodeUnknown
			}
		}
		if isNewArtifact {
			namespaceRepository := reponamespace.NewNamespaceRepository(tx)
			if err := namespaceRepository.IncrementSize(ctx, repositoryObj.NamespaceID, artifactObj.BlobsSize); err != nil {
				slog.Error("increment namespace size failed", "err", err, "namespace_id", repositoryObj.NamespaceID)
				if e, ok := err.(errcode.ErrCode); ok {
					return e
				}
				return errcode.DSErrCodeUnknown
			}
			repositoryRepository := reporegistry.NewRepositoryRepository(tx)
			if err := repositoryRepository.IncrementSize(ctx, repositoryObj.ID, artifactObj.BlobsSize); err != nil {
				slog.Error("increment repository size failed", "err", err, "repository_id", repositoryObj.ID)
				if e, ok := err.(errcode.ErrCode); ok {
					return e
				}
				return errcode.DSErrCodeUnknown
			}
		}
		return nil
	})
	if err != nil {
		if e, ok := err.(errcode.ErrCode); ok {
			return e
		}
		return errcode.DSErrCodeUnknown
	}
	if isNewArtifact {
		s.invalidateMetadataCache(ctx, repositoryObj)
	}

	if needScan(manifest, descriptor) {
		s.putManifestAsyncTask(ctx, artifactObj)
	}

	return nil
}

// needScan reports whether the manifest references a scannable image config.
func needScan(manifest distribution.Manifest, _ distribution.Descriptor) bool {
	if len(manifest.References()) > 0 {
		ref := manifest.References()[0]
		// only image can be scanned
		references := manifest.References()
		for _, descriptor := range references {
			if descriptor.MediaType == inTotoMediaType ||
				descriptor.MediaType == cosignSimpleSigningMediaType ||
				descriptor.MediaType == helmChartContentMediaType ||
				descriptor.MediaType == helmConfigMediaType {
				return false
			}
		}
		if ref.MediaType == schema2.MediaTypeImageConfig ||
			ref.MediaType == imgspecv1.MediaTypeImageConfig {
			return true
		}
	}
	return false
}

// putManifestIndex handles the manifest list / index request.
// support media type:
// application/vnd.docker.distribution.manifest.list.v2+json
// application/vnd.oci.image.index.v1+json
func (s *service) putManifestIndex(ctx context.Context, digests []string, repositoryObj *models.Repository, artifactObj *models.Artifact, refs manifestRefs, _ distribution.Manifest, _ distribution.Descriptor, isNewArtifact bool) error {
	artifactObjs, err := s.RepoArtifact.GetByDigests(ctx, repositoryObj.Name, digests)
	if err != nil {
		slog.Error("get artifacts failed", "err", err, "repository", repositoryObj.Name, "digests", digests)
		return errcode.DSErrCodeUnknown
	}

	artifactObj.ArtifactSubs = artifactObjs

	err = query.Q.Transaction(func(tx *query.Query) error {
		repoArtifact := reporegistry.NewArtifactRepository(tx)
		err = repoArtifact.Create(ctx, artifactObj)
		if err != nil {
			slog.Error("create artifact failed", "err", err, "repository", repositoryObj.Name, "digest", refs.Digest.String())
			if e, ok := err.(errcode.ErrCode); ok {
				return e
			}
			return errcode.DSErrCodeUnknown
		}
		if refs.Tag != "" {
			tagRepository := reporegistry.NewTagRepository(tx)
			err = tagRepository.Create(ctx, &models.Tag{
				ID:           uuid.NewV7String(),
				RepositoryID: repositoryObj.ID,
				ArtifactID:   artifactObj.ID,
				Name:         refs.Tag,
			})
			if err != nil {
				slog.Error("create tag failed", "err", err, "repository", repositoryObj.Name, "tag", refs.Tag)
				if e, ok := err.(errcode.ErrCode); ok {
					return e
				}
				return errcode.DSErrCodeUnknown
			}
		}
		if isNewArtifact {
			namespaceRepository := reponamespace.NewNamespaceRepository(tx)
			if err := namespaceRepository.IncrementSize(ctx, repositoryObj.NamespaceID, artifactObj.BlobsSize); err != nil {
				slog.Error("increment namespace size failed", "err", err, "namespace_id", repositoryObj.NamespaceID)
				if e, ok := err.(errcode.ErrCode); ok {
					return e
				}
				return errcode.DSErrCodeUnknown
			}
			repositoryRepository := reporegistry.NewRepositoryRepository(tx)
			if err := repositoryRepository.IncrementSize(ctx, repositoryObj.ID, artifactObj.BlobsSize); err != nil {
				slog.Error("increment repository size failed", "err", err, "repository_id", repositoryObj.ID)
				if e, ok := err.(errcode.ErrCode); ok {
					return e
				}
				return errcode.DSErrCodeUnknown
			}
		}
		err = s.Producer.Produce(ctx, enums.DaemonTagPushed, api.DaemonTagPushedPayload{
			RepositoryID: repositoryObj.ID,
			Tag:          refs.Tag,
		})
		if err != nil {
			slog.Error("enqueue tag pushed task failed", "err", err, "tag", refs.Tag, "digest", refs.Digest.String())
			return errcode.DSErrCodeUnknown
		}
		return nil
	})
	if err != nil {
		if e, ok := err.(errcode.ErrCode); ok {
			return e
		}
		return errcode.DSErrCodeUnknown
	}
	if isNewArtifact {
		s.invalidateMetadataCache(ctx, repositoryObj)
	}

	return nil
}

// putManifestAsyncTaskVulnerability enqueues the vulnerability scan task for the artifact.
func (s *service) putManifestAsyncTaskVulnerability(ctx context.Context, artifactObj *models.Artifact) {
	err := s.RepoArtifact.CreateVulnerability(ctx, &models.ArtifactVulnerability{
		ID:         uuid.NewV7String(),
		ArtifactID: artifactObj.ID,
		Status:     enums.TaskCommonStatusPending,
	})
	if err != nil {
		slog.Error("save vulnerability failed", "err", err)
		return
	}
}

// putManifestAsyncTask triggers the post-push async scan tasks.
func (s *service) putManifestAsyncTask(ctx context.Context, artifactObj *models.Artifact) {
	s.putManifestAsyncTaskVulnerability(ctx, artifactObj)
}

// getArtifactType determines the artifact type from the manifest descriptors.
func (s *service) getArtifactType(descriptor distribution.Descriptor, manifest distribution.Manifest) enums.ArtifactType {
	if descriptor.MediaType == manifestlist.MediaTypeManifestList ||
		descriptor.MediaType == imgspecv1.MediaTypeImageIndex {
		return enums.ArtifactTypeImage
	}
	references := manifest.References()
	for _, descriptor := range references {
		switch descriptor.MediaType {
		case inTotoMediaType:
			return enums.ArtifactTypeProvenance
		case cosignSimpleSigningMediaType:
			return enums.ArtifactTypeCosign
		}
	}
	var mediaType string
	if len(references) == 0 {
		return enums.ArtifactTypeUnknown
	}
	mediaType = references[0].MediaType
	switch mediaType {
	case imgspecv1.MediaTypeImageConfig, schema2.MediaTypeImageConfig:
		return enums.ArtifactTypeImage
	case "application/vnd.cnab.manifest.v1":
		return enums.ArtifactTypeCnab
	case "application/vnd.wasm.config.v1+json":
		return enums.ArtifactTypeWasm
	case helmConfigMediaType:
		return enums.ArtifactTypeChart
	case sifConfigMediaType:
		return enums.ArtifactTypeSif
	}
	return enums.ArtifactTypeUnknown
}

// getArtifactReferrer resolves the referrer artifact id (OCI Subject).
func (s *service) getArtifactReferrer(ctx context.Context, repository string, manifest distribution.Manifest) (*string, error) {
	mediaType, data, err := manifest.Payload()
	if err != nil {
		return nil, err
	}
	repositoryObj, err := s.RepoRegistry.GetByName(ctx, repository)
	if err != nil {
		return nil, err
	}

	var dgst string

	switch mediaType {
	case imgspecv1.MediaTypeImageManifest: // nolint: gocritic, staticcheck
		var decoded imgspecv1.Manifest
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			return nil, err
		}
		if decoded.Subject == nil {
			return nil, nil
		}
		dgst = decoded.Subject.Digest.String()
	case imgspecv1.MediaTypeImageIndex:
		var decoded imgspecv1.Index
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			return nil, err
		}
		if decoded.Subject == nil {
			return nil, nil
		}
		dgst = decoded.Subject.Digest.String()
	default:
		return nil, nil
	}

	artifactObj, err := s.RepoArtifact.GetByDigest(ctx, repositoryObj.ID, dgst)
	if err != nil {
		return nil, err
	}
	referrerID := artifactObj.ID
	return &referrerID, nil
}

// DeleteManifest deletes a manifest by digest or tag.
// if reference is a tag, just delete the tag
// if reference is a digest, delete the artifact and all of the tags that reference it
func (s *service) DeleteManifest(ctx context.Context, namespaceID string, repository string, reference string, userID string) error {
	_ = namespaceID // namespaceID is already auth-checked by the handler; repository owns the artifact

	repositoryObj, err := s.RepoRegistry.GetByName(ctx, repository)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("cannot find repository", "err", err, "repository", repository)
			return errcode.DSErrCodeNameUnknown
		}
		slog.Error("get repository failed", "err", err, "repository", repository)
		return errcode.DSErrCodeUnknown
	}

	tagName, refDigest := parseManifestRef(reference)

	if tagName != "" {
		_, err = s.RepoTag.GetByName(ctx, repositoryObj.ID, tagName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("cannot find tag", "err", err, "repository", repository, "tag", tagName)
				return errcode.DSErrCodeManifestUnknown
			}
			slog.Error("get tag failed", "err", err, "repository", repository, "tag", tagName)
			return errcode.DSErrCodeUnknown
		}
		err = s.RepoTag.DeleteByName(ctx, repositoryObj.ID, tagName)
		if err != nil {
			slog.Error("delete tag failed", "err", err, "Tag", tagName)
			return errcode.DSErrCodeUnknown
		}
		return nil
	}

	artifactObj, err := s.RepoArtifact.GetByDigest(ctx, repositoryObj.ID, refDigest.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("cannot find artifact", "err", err, "repository", repository, "artifact", refDigest.String())
			return errcode.DSErrCodeManifestUnknown
		}
		slog.Error("get artifact failed", "err", err, "repository", repository, "artifact", refDigest.String())
		return errcode.DSErrCodeUnknown
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		tagRepository := reporegistry.NewTagRepository(tx)
		err = tagRepository.DeleteByArtifactID(ctx, artifactObj.ID)
		if err != nil {
			slog.Error("delete tag by artifact id failed", "err", err, "ArtifactID", artifactObj.ID)
			return errcode.DSErrCodeUnknown
		}
		repoArtifact := reporegistry.NewArtifactRepository(tx)
		err = repoArtifact.DeleteByID(ctx, artifactObj.ID)
		if err != nil {
			slog.Error("delete artifact by id failed", "err", err, "ArtifactID", artifactObj.ID)
			return errcode.DSErrCodeUnknown
		}
		namespaceRepository := reponamespace.NewNamespaceRepository(tx)
		if err := namespaceRepository.DecrementSize(ctx, repositoryObj.NamespaceID, artifactObj.BlobsSize); err != nil {
			slog.Error("decrement namespace size failed", "err", err, "namespace_id", repositoryObj.NamespaceID)
			return errcode.DSErrCodeUnknown
		}
		repoRegistry := reporegistry.NewRepositoryRepository(tx)
		if err := repoRegistry.DecrementSize(ctx, repositoryObj.ID, artifactObj.BlobsSize); err != nil {
			slog.Error("decrement repository size failed", "err", err, "repository_id", repositoryObj.ID)
			return errcode.DSErrCodeUnknown
		}
		return nil
	})
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			return e
		}
		return errcode.DSErrCodeUnknown
	}
	s.invalidateMetadataCache(ctx, repositoryObj)
	err = s.Storage.Delete(ctx, utils.GenManifestPathByDigest(refDigest))
	if err != nil {
		slog.Error("delete manifest failed", "err", err, "digest", refDigest.String())
		return errcode.DSErrCodeUnknown
	}
	return nil
}

func (s *service) invalidateMetadataCache(ctx context.Context, repositoryObj *models.Repository) {
	if repositoryObj == nil {
		return
	}
	if invalidator, ok := s.RepoNs.(reponamespace.NamespaceCacheInvalidator); ok {
		if err := invalidator.InvalidateNamespaceID(ctx, repositoryObj.NamespaceID); err != nil {
			slog.Warn("invalidate namespace cache failed", "err", err, "namespace_id", repositoryObj.NamespaceID)
		}
	}
	if invalidator, ok := s.RepoRegistry.(reporegistry.RepositoryCacheInvalidator); ok {
		if err := invalidator.InvalidateRepository(ctx, repositoryObj); err != nil {
			slog.Warn("invalidate repository cache failed", "err", err, "repository_id", repositoryObj.ID)
		}
	}
}

func (s *service) seedNamespaceCache(ctx context.Context, namespaceObj *models.Namespace) {
	if seeder, ok := s.RepoNs.(reponamespace.NamespaceCacheSeeder); ok {
		if err := seeder.SeedNamespace(ctx, namespaceObj); err != nil {
			slog.Warn("seed namespace cache failed", "err", err, "namespace_id", namespaceObj.ID)
		}
	}
}

func (s *service) seedRepositoryCache(ctx context.Context, repositoryObj *models.Repository) {
	if seeder, ok := s.RepoRegistry.(reporegistry.RepositoryCacheSeeder); ok {
		if err := seeder.SeedRepository(ctx, repositoryObj); err != nil {
			slog.Warn("seed repository cache failed", "err", err, "repository_id", repositoryObj.ID)
		}
	}
}

// GetReferrer gets the referrer manifest for a given digest.
func (s *service) GetReferrer(ctx context.Context, repository string, dgst string, artifactTypes []string) ([]byte, error) {
	repositoryObj, err := s.RepoRegistry.GetByName(ctx, repository)
	if err != nil {
		slog.Error("get repository failed", "err", err, "repository", repository)
		return nil, errcode.DSErrCodeUnknown
	}

	artifactObjs, err := s.RepoArtifact.GetReferrers(ctx, repositoryObj.ID, dgst, artifactTypes)
	if err != nil {
		slog.Error("get referrers failed", "err", err)
		return nil, errcode.DSErrCodeUnknown
	}

	result := imgspecv1.Index{
		MediaType: imgspecv1.MediaTypeImageIndex,
		Manifests: make([]imgspecv1.Descriptor, 0, len(artifactObjs)),
	}
	result.SchemaVersion = 2
	for _, artifactObj := range artifactObjs {
		raw, readErr := s.readManifest(ctx, artifactObj.Digest)
		if readErr != nil {
			slog.Error("read artifact manifest failed", "err", readErr, "digest", artifactObj.Digest)
			return nil, errcode.DSErrCodeUnknown
		}
		var decoded imgspecv1.Manifest
		err = json.Unmarshal(raw, &decoded)
		if err != nil {
			slog.Error("unmarshal artifact failed", "err", err)
			return nil, errcode.DSErrCodeUnknown
		}
		result.Manifests = append(result.Manifests, imgspecv1.Descriptor{
			MediaType:    decoded.MediaType,
			Size:         artifactObj.Size,
			Digest:       digest.Digest(artifactObj.Digest),
			ArtifactType: decoded.Config.MediaType,
			Annotations:  decoded.Annotations,
		})
	}
	body, err := json.Marshal(result)
	if err != nil {
		slog.Error("marshal referrer index failed", "err", err)
		return nil, errcode.DSErrCodeUnknown
	}
	return body, nil
}
