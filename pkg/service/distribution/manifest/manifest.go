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
	"encoding/json/v2"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/manifest/manifestlist"
	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.uber.org/dig"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
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

//go:generate go tool mockgen -mock_names Service=MockDistributionManifestService -destination=mocks/manifest_mocks.go -package=mocks github.com/go-sigma/sigma/pkg/service/distribution/manifest Service

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

	RepoArtifact reporegistry.ArtifactRepository
	RepoBlob     reporegistry.BlobRepository
	RepoTag      reporegistry.TagRepository
	RepoRegistry reporegistry.RepositoryRepository
	RepoNs       reponamespace.NamespaceRepository
	SvcAnalytics svcanalytics.Service `optional:"true"`
	Storage      storage.StorageDriver
	Producer     workq.Producer
}

func NewService(params service) Service {
	return &params
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

func (s *service) ensureRepository(ctx context.Context, namespaceID string, repository string) (*models.Repository, error) {
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

	namespaceObj, err := s.resolveRepositoryNamespace(ctx, namespaceID, namespaceName)
	if err != nil {
		return nil, err
	}

	repositoryObj = &models.Repository{
		ID:          uuid.NewV7String(),
		NamespaceID: namespaceObj.ID,
		Name:        repository,
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		repoRegistry := reporegistry.NewRepositoryRepository(tx)
		if err := s.checkRepositoryQuota(ctx, tx, namespaceObj.ID); err != nil {
			logWriteError(ctx, err, "check repository quota failed", "repository", repository)
			return err
		}
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

	s.invalidateNamespaceCache(ctx, namespaceObj.ID)
	s.seedRepositoryCache(ctx, repositoryObj)
	return repositoryObj, nil
}

func (s *service) resolveRepositoryNamespace(ctx context.Context, namespaceID string, namespaceName string) (*models.Namespace, error) {
	if namespaceID != "" {
		namespaceObj, err := s.RepoNs.Get(ctx, namespaceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errcode.DSErrCodeNameUnknown
			}
			slog.Error("get namespace failed", "err", err, "namespace_id", namespaceID)
			return nil, errcode.DSErrCodeUnknown
		}
		if namespaceObj.Name != namespaceName {
			slog.Error("repository namespace does not match request namespace", "repository_namespace", namespaceName, "namespace_id", namespaceID)
			return nil, errcode.DSErrCodeNameUnknown
		}
		return namespaceObj, nil
	}

	namespaceObj, err := s.RepoNs.GetByName(ctx, namespaceName)
	if err == nil {
		return namespaceObj, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get namespace failed", "err", err, "namespace", namespaceName)
		return nil, errcode.DSErrCodeUnknown
	}
	slog.Error("cannot find namespace", "namespace", namespaceName)
	return nil, errcode.DSErrCodeNameUnknown
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

	repositoryObj, err := s.ensureRepository(ctx, namespaceID, repository)
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
	digests := make([]string, 0, len(manifest.References()))
	for _, reference := range manifest.References() {
		digests = append(digests, reference.Digest.String())
		blobsSize += reference.Size
	}
	digests = dedupeDigests(digests)

	artifactObj := &models.Artifact{
		ID:           uuid.NewV7String(),
		NamespaceID:  repositoryObj.NamespaceID,
		RepositoryID: repositoryObj.ID,
		Digest:       dgst.String(),
		Size:         int64(len(body)),
		BlobsSize:    blobsSize,
		ContentType:  mediaType,
	}

	subject, err := s.getArtifactSubject(ctx, repositoryObj, manifest)
	if err != nil {
		slog.Error("get artifact subject failed", "err", err, "digest", dgst.String())
		return "", errcode.DSErrCodeUnknown
	}
	if subject != nil {
		artifactObj.ReferrerDigest = &subject.Digest
		artifactObj.ReferrerID = subject.ID
	}

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

	var namespaceSizeDelta int64
	if mediaType == manifestlist.MediaTypeManifestList ||
		mediaType == imgspecv1.MediaTypeImageIndex {
		artifactObj.Type = enums.ArtifactTypeImageIndex
		err = s.putManifestIndex(ctx, digests, repositoryObj, artifactObj, refs, manifest, descriptor, isNewArtifact, isNewTag)
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
		}

		artifactObj.Type = s.getArtifactType(descriptor, manifest)
		namespaceSizeDelta, err = s.putManifestManifest(ctx, digests, repositoryObj, artifactObj, refs, manifest, descriptor, isNewArtifact, isNewTag)
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
	if namespaceSizeDelta > 0 {
		s.recordNamespaceSizeDelta(ctx, namespaceID, namespaceSizeDelta)
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
func (s *service) putManifestManifest(ctx context.Context, digests []string, repositoryObj *models.Repository, artifactObj *models.Artifact, refs manifestRefs, manifest distribution.Manifest, descriptor distribution.Descriptor, isNewArtifact bool, isNewTag bool) (int64, error) {
	blobObjs, err := s.RepoBlob.FindByDigests(ctx, digests)
	if err != nil {
		logWriteError(ctx, err, "find blobs failed", "digest", refs.Digest.String())
		return 0, errcode.DSErrCodeUnknown
	}

	artifactObj.Blobs = blobObjs
	artifactObj.BlobsSize = attributedBlobSize(blobObjs)

	var namespaceSizeDelta int64
	err = query.Q.Transaction(func(tx *query.Query) error {
		artifactRepository := reporegistry.NewArtifactRepository(tx)
		err = artifactRepository.Create(ctx, artifactObj)
		if err != nil {
			logWriteError(ctx, err, "create artifact failed", "repository", repositoryObj.Name, "digest", refs.Digest.String(),
				"artifactObj", artifactObj)
			if e, ok := err.(errcode.ErrCode); ok {
				return e
			}
			return errcode.DSErrCodeUnknown
		}
		// The subject may have been pushed after this referrer: link the
		// referrers that are still waiting for it.
		s.resolveReferrers(ctx, artifactRepository, repositoryObj, artifactObj)
		if refs.Tag != "" {
			if err := s.checkTagQuota(ctx, tx, repositoryObj, isNewTag); err != nil {
				logWriteError(ctx, err, "check tag quota failed", "tag", refs.Tag)
				return err
			}
			tagRepository := reporegistry.NewTagRepository(tx)
			err = tagRepository.Create(ctx, &models.Tag{
				ID:           uuid.NewV7String(),
				RepositoryID: repositoryObj.ID,
				ArtifactID:   artifactObj.ID,
				Name:         refs.Tag,
			})
			if err != nil {
				logWriteError(ctx, err, "create tag failed", "tag", refs.Tag, "digest", refs.Digest.String())
				if e, ok := err.(errcode.ErrCode); ok { // maybe got exceed tag quota error
					return e
				}
				return errcode.DSErrCodeUnknown
			}
		}
		namespaceSizeDelta, err = s.applyScopeSizeDelta(ctx, tx, repositoryObj, blobObjs)
		if err != nil {
			logWriteError(ctx, err, "update size quota failed", "repository", repositoryObj.Name, "digest", refs.Digest.String())
			if e, ok := err.(errcode.ErrCode); ok {
				return e
			}
			return errcode.DSErrCodeUnknown
		}
		return nil
	})
	if err != nil {
		if e, ok := err.(errcode.ErrCode); ok {
			return 0, e
		}
		return 0, errcode.DSErrCodeUnknown
	}
	if isNewArtifact || isNewTag {
		s.invalidateMetadataCache(ctx, repositoryObj)
	}

	if needScan(manifest, descriptor) {
		s.putManifestAsyncTask(ctx, artifactObj, isNewArtifact)
	}

	return namespaceSizeDelta, nil
}

// attributedBlobSize returns the deduplicated size of the given blobs.
func attributedBlobSize(blobs []*models.Blob) int64 {
	seen := make(map[string]struct{}, len(blobs))
	var total int64
	for _, blob := range blobs {
		if blob == nil {
			continue
		}
		if _, ok := seen[blob.ID]; ok {
			continue
		}
		seen[blob.ID] = struct{}{}
		total += blob.Size
	}
	return total
}

// applyScopeSizeDelta records the blobs attributed to the repository and
// namespace, then applies only the newly inserted bytes to their size
// counters. The (scope_type, scope_id, blob_id) primary key makes the write
// idempotent, so concurrent pushes cannot double count a shared blob.
func (s *service) applyScopeSizeDelta(ctx context.Context, tx *query.Query, repositoryObj *models.Repository, blobs []*models.Blob) (int64, error) {
	if len(blobs) == 0 {
		return 0, nil
	}

	// Lock the scope rows before touching blob_attributions. ReconcileScopeSize
	// takes the same rows first and then rewrites attributions; doing the
	// opposite order here can deadlock on MySQL when both transactions touch an
	// existing attribution row.
	if _, err := tx.Namespace.WithContext(ctx).
		Where(tx.Namespace.ID.Eq(repositoryObj.NamespaceID)).
		UpdateColumns(map[string]any{"size_dirty": true}); err != nil {
		return 0, err
	}
	if _, err := tx.Repository.WithContext(ctx).
		Where(tx.Repository.ID.Eq(repositoryObj.ID)).
		UpdateColumns(map[string]any{"size_dirty": true}); err != nil {
		return 0, err
	}

	batchID := uuid.NewV7String()
	attributionRepository := reporegistry.NewBlobAttributionRepository(tx)

	scopes := []struct {
		scopeType string
		scopeID   string
	}{
		{reporegistry.ScopeTypeNamespace, repositoryObj.NamespaceID},
		{reporegistry.ScopeTypeRepository, repositoryObj.ID},
	}

	var namespaceDelta int64
	for _, scope := range scopes {
		if err := attributionRepository.UpsertBatch(ctx, scope.scopeType, scope.scopeID, blobs, batchID); err != nil {
			return 0, err
		}
		delta, err := attributionRepository.SumBatch(ctx, scope.scopeType, scope.scopeID, batchID)
		if err != nil {
			return 0, err
		}
		if delta == 0 {
			continue
		}
		switch scope.scopeType {
		case reporegistry.ScopeTypeNamespace:
			if err := reponamespace.NewNamespaceRepository(tx).IncrementSize(ctx, scope.scopeID, delta); err != nil {
				return 0, err
			}
			namespaceDelta = delta
		case reporegistry.ScopeTypeRepository:
			if err := reporegistry.NewRepositoryRepository(tx).IncrementSize(ctx, scope.scopeID, delta); err != nil {
				return 0, err
			}
		}
	}
	return namespaceDelta, nil
}

// checkTagQuota enforces the repository and namespace tag count limits before
// a new tag is created. The parent rows are locked so concurrent pushes cannot
// both pass the count check.
func (s *service) checkTagQuota(ctx context.Context, tx *query.Query, repositoryObj *models.Repository, isNewTag bool) error {
	if !isNewTag {
		return nil
	}

	// Lock the namespace first, then the repository: this keeps the lock order
	// consistent with applyScopeSizeDelta (namespace -> repository).
	namespaceObj, err := tx.Namespace.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(tx.Namespace.ID.Eq(repositoryObj.NamespaceID)).
		First()
	if err != nil {
		return err
	}
	repo, err := tx.Repository.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(tx.Repository.ID.Eq(repositoryObj.ID)).
		First()
	if err != nil {
		return err
	}

	if repo.TagLimit > 0 {
		if repo.TagCount >= repo.TagLimit {
			return errcode.GenDSErrCodeResourceCountQuotaExceedRepository(repo.Name, repo.TagLimit)
		}
	}

	if namespaceObj.TagLimit <= 0 {
		return nil
	}
	if namespaceObj.TagCount >= namespaceObj.TagLimit {
		return errcode.GenDSErrCodeResourceCountQuotaExceedNamespaceTag(namespaceObj.Name, namespaceObj.TagLimit)
	}
	return nil
}

// checkRepositoryQuota enforces the namespace repository count limit before a
// new repository is created.
func (s *service) checkRepositoryQuota(ctx context.Context, tx *query.Query, namespaceID string) error {
	namespaceObj, err := tx.Namespace.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(tx.Namespace.ID.Eq(namespaceID)).
		First()
	if err != nil {
		return err
	}
	if namespaceObj.RepositoryLimit <= 0 {
		return nil
	}
	if namespaceObj.RepositoryCount >= namespaceObj.RepositoryLimit {
		return errcode.GenDSErrCodeResourceCountQuotaExceedNamespaceRepository(namespaceObj.Name, namespaceObj.RepositoryLimit)
	}
	return nil
}

// reconcileScopeSizeAsync repairs and refreshes the scope size counters in the
// background. Deletion only marks scopes dirty; this gives immediate release
// instead of waiting for the periodic reconciler.
func (s *service) reconcileScopeSizeAsync(ctx context.Context, namespaceID, repositoryID string) {
	if namespaceID == "" && repositoryID == "" {
		return
	}
	asyncCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	go func() {
		defer cancel()
		if namespaceID != "" {
			if _, err := reporegistry.ReconcileScopeSize(asyncCtx, reporegistry.ScopeTypeNamespace, namespaceID); err != nil {
				slog.Error("reconcile namespace size failed", "err", err, "namespace_id", namespaceID)
			} else if invalidator, ok := s.RepoNs.(reponamespace.NamespaceCacheInvalidator); ok {
				if err = invalidator.InvalidateNamespaceID(asyncCtx, namespaceID); err != nil {
					slog.Warn("invalidate namespace cache failed", "err", err, "namespace_id", namespaceID)
				}
			}
		}
		if repositoryID != "" {
			if _, err := reporegistry.ReconcileScopeSize(asyncCtx, reporegistry.ScopeTypeRepository, repositoryID); err != nil {
				slog.Error("reconcile repository size failed", "err", err, "repository_id", repositoryID)
			} else if invalidator, ok := s.RepoRegistry.(reporegistry.RepositoryCacheInvalidator); ok {
				if err = invalidator.InvalidateRepositoryID(asyncCtx, repositoryID); err != nil {
					slog.Warn("invalidate repository cache failed", "err", err, "repository_id", repositoryID)
				}
			}
		}
	}()
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
func (s *service) putManifestIndex(ctx context.Context, digests []string, repositoryObj *models.Repository, artifactObj *models.Artifact, refs manifestRefs, _ distribution.Manifest, _ distribution.Descriptor, isNewArtifact bool, isNewTag bool) error {
	artifactObjs, err := s.RepoArtifact.GetByDigests(ctx, repositoryObj.Name, digests)
	if err != nil {
		logWriteError(ctx, err, "get artifacts failed", "repository", repositoryObj.Name, "digests", digests)
		return errcode.DSErrCodeUnknown
	}

	artifactObj.ArtifactSubs = artifactObjs

	err = query.Q.Transaction(func(tx *query.Query) error {
		repoArtifact := reporegistry.NewArtifactRepository(tx)
		err = repoArtifact.Create(ctx, artifactObj)
		if err != nil {
			logWriteError(ctx, err, "create artifact failed", "repository", repositoryObj.Name, "digest", refs.Digest.String())
			if e, ok := err.(errcode.ErrCode); ok {
				return e
			}
			return errcode.DSErrCodeUnknown
		}
		// The subject may have been pushed after this referrer: link the
		// referrers that are still waiting for it.
		s.resolveReferrers(ctx, repoArtifact, repositoryObj, artifactObj)
		if refs.Tag != "" {
			if err := s.checkTagQuota(ctx, tx, repositoryObj, isNewTag); err != nil {
				logWriteError(ctx, err, "check tag quota failed", "tag", refs.Tag)
				return err
			}
			tagRepository := reporegistry.NewTagRepository(tx)
			err = tagRepository.Create(ctx, &models.Tag{
				ID:           uuid.NewV7String(),
				RepositoryID: repositoryObj.ID,
				ArtifactID:   artifactObj.ID,
				Name:         refs.Tag,
			})
			if err != nil {
				logWriteError(ctx, err, "create tag failed", "repository", repositoryObj.Name, "tag", refs.Tag)
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
			logWriteError(ctx, err, "enqueue tag pushed task failed", "tag", refs.Tag, "digest", refs.Digest.String())
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
	if isNewArtifact || isNewTag {
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

// putManifestAsyncTaskSbom enqueues the sbom scan task for the artifact.
func (s *service) putManifestAsyncTaskSbom(ctx context.Context, artifactObj *models.Artifact) {
	err := s.RepoArtifact.CreateSbom(ctx, &models.ArtifactSbom{
		ID:         uuid.NewV7String(),
		ArtifactID: artifactObj.ID,
		Status:     enums.TaskCommonStatusPending,
	})
	if err != nil {
		slog.Error("save sbom failed", "err", err, "artifact_id", artifactObj.ID)
		return
	}
	err = s.Producer.Produce(ctx, enums.DaemonSbom, api.TaskSbom{ArtifactID: artifactObj.ID})
	if err != nil {
		slog.Error("enqueue sbom task failed", "err", err, "artifact_id", artifactObj.ID)
	}
}

// putManifestAsyncTask triggers the post-push async scan tasks.
func (s *service) putManifestAsyncTask(ctx context.Context, artifactObj *models.Artifact, isNewArtifact bool) {
	s.putManifestAsyncTaskVulnerability(ctx, artifactObj)
	// The sbom scan is a one-shot task, so it is only queued the first time an artifact is
	// stored: re-pushing an existing digest must not scan the same image again.
	if isNewArtifact {
		s.putManifestAsyncTaskSbom(ctx, artifactObj)
	}
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

// artifactSubject is the OCI subject referenced by a manifest. Digest is always
// known, while ID is only set once the subject artifact is stored.
type artifactSubject struct {
	Digest string
	ID     *string
}

// getArtifactSubject extracts the OCI subject from the manifest. A subject that
// is not stored yet is not an error: a client pushes the referrer and its
// subject in parallel (buildkit pushes the image manifest together with its
// provenance attestation), and the registry must accept either order. The link
// is completed by ResolveReferrers when the subject is stored later.
func (s *service) getArtifactSubject(ctx context.Context, repositoryObj *models.Repository, manifest distribution.Manifest) (*artifactSubject, error) {
	mediaType, data, err := manifest.Payload()
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

	subject := &artifactSubject{Digest: dgst}
	artifactObj, err := s.RepoArtifact.GetByDigest(ctx, repositoryObj.ID, dgst)
	switch {
	case err == nil:
		subject.ID = &artifactObj.ID
	case errors.Is(err, gorm.ErrRecordNotFound):
		slog.Debug("subject artifact not found, it will be linked once pushed",
			"digest", dgst, "repository", repositoryObj.Name)
	default:
		return nil, err
	}
	return subject, nil
}

// dedupeDigests removes repeated digests while keeping their order. A manifest
// may reference the same digest more than once, and the digest list feeds both
// the blob lookup and the blob associations, so it must not contain duplicates.
func dedupeDigests(digests []string) []string {
	seen := make(map[string]struct{}, len(digests))
	result := make([]string, 0, len(digests))
	for _, dgst := range digests {
		if _, ok := seen[dgst]; ok {
			continue
		}
		seen[dgst] = struct{}{}
		result = append(result, dgst)
	}
	return result
}

// resolveReferrers links the referrers that were pushed before this artifact,
// so a referrer stored ahead of its subject is treated as referenced by the GC.
// A failure here is not fatal: the referrers API is queried by digest, and a
// later push of the same artifact retries the link.
func (s *service) resolveReferrers(ctx context.Context, repo reporegistry.ArtifactRepository, repositoryObj *models.Repository, artifactObj *models.Artifact) {
	if err := repo.ResolveReferrers(ctx, repositoryObj.ID, artifactObj.Digest, artifactObj.ID); err != nil {
		logWriteError(ctx, err, "resolve referrers failed", "repository", repositoryObj.Name, "digest", artifactObj.Digest)
	}
}

// logWriteError reports a manifest write failure. A canceled context means the
// client closed the connection (docker push abandons the in-flight requests
// when it retries), which is not a server failure: it is logged at warning level
// because the client retry usually succeeds.
func logWriteError(ctx context.Context, err error, msg string, args ...any) {
	args = append([]any{"err", err}, args...)
	if ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		slog.Warn(msg, append(args, "reason", "client closed the request")...)
		return
	}
	slog.Error(msg, args...)
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
		s.invalidateMetadataCache(ctx, repositoryObj)
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
		// A child manifest that is still referenced by an index must not be
		// deleted: doing so would leave a broken index and release its size
		// attribution early.
		assocErr := repoArtifact.IsArtifactAssociatedWithArtifact(ctx, artifactObj.ID)
		if assocErr == nil {
			return errcode.GenDSErrCodeManifestReferenced(artifactObj.Digest)
		}
		if !errors.Is(assocErr, gorm.ErrRecordNotFound) {
			slog.Error("check artifact association failed", "err", assocErr, "ArtifactID", artifactObj.ID)
			return errcode.DSErrCodeUnknown
		}
		err = repoArtifact.DeleteByID(ctx, artifactObj.ID)
		if err != nil {
			slog.Error("delete artifact by id failed", "err", err, "ArtifactID", artifactObj.ID)
			return errcode.DSErrCodeUnknown
		}
		// Deletion does not subtract synchronously: mark both scopes dirty and
		// let the size reconciler remove the now unreachable attributions.
		if err := reponamespace.NewNamespaceRepository(tx).UpdateByID(ctx, repositoryObj.NamespaceID, map[string]any{
			"size_dirty": true,
		}); err != nil {
			slog.Error("mark namespace size dirty failed", "err", err, "namespace_id", repositoryObj.NamespaceID)
			return errcode.DSErrCodeUnknown
		}
		if err := reporegistry.NewRepositoryRepository(tx).UpdateRepository(ctx, repositoryObj.ID, map[string]any{
			"size_dirty": true,
		}); err != nil {
			slog.Error("mark repository size dirty failed", "err", err, "repository_id", repositoryObj.ID)
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
	s.reconcileScopeSizeAsync(ctx, repositoryObj.NamespaceID, repositoryObj.ID)
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
	s.invalidateNamespaceCache(ctx, repositoryObj.NamespaceID)
	if invalidator, ok := s.RepoRegistry.(reporegistry.RepositoryCacheInvalidator); ok {
		if err := invalidator.InvalidateRepository(ctx, repositoryObj); err != nil {
			slog.Warn("invalidate repository cache failed", "err", err, "repository_id", repositoryObj.ID)
		}
	}
}

func (s *service) invalidateNamespaceCache(ctx context.Context, namespaceID string) {
	if invalidator, ok := s.RepoNs.(reponamespace.NamespaceCacheInvalidator); ok {
		if err := invalidator.InvalidateNamespaceID(ctx, namespaceID); err != nil {
			slog.Warn("invalidate namespace cache failed", "err", err, "namespace_id", namespaceID)
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
		MediaType:     imgspecv1.MediaTypeImageIndex,
		Manifests:     make([]imgspecv1.Descriptor, 0, len(artifactObjs)),
		SchemaVersion: 2,
	}
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
