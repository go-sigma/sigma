// Copyright 2026 sigma
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
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/manifest/manifestlist"
	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svcanalytics "github.com/go-sigma/sigma/pkg/service/analytics"
	mockstorage "github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils"
)

func TestParseManifestRef(t *testing.T) {
	tag, dgst := parseManifestRef("latest")
	require.Equal(t, "latest", tag)
	require.Empty(t, dgst)

	expected := digest.FromString("manifest")
	tag, dgst = parseManifestRef(expected.String())
	require.Empty(t, tag)
	require.Equal(t, expected, dgst)
}

func TestManifestClassification(t *testing.T) {
	service := &service{}

	require.Equal(t, enums.ArtifactTypeImage, service.getArtifactType(
		distribution.Descriptor{MediaType: manifestlist.MediaTypeManifestList},
		fakeManifest{},
	))
	require.Equal(t, enums.ArtifactTypeImage, service.getArtifactType(
		distribution.Descriptor{},
		fakeManifest{references: []distribution.Descriptor{{MediaType: schema2.MediaTypeImageConfig}}},
	))
	require.Equal(t, enums.ArtifactTypeProvenance, service.getArtifactType(
		distribution.Descriptor{},
		fakeManifest{references: []distribution.Descriptor{{MediaType: inTotoMediaType}}},
	))
	require.Equal(t, enums.ArtifactTypeUnknown, service.getArtifactType(
		distribution.Descriptor{},
		fakeManifest{},
	))

	require.True(t, needScan(
		fakeManifest{references: []distribution.Descriptor{{MediaType: imgspecv1.MediaTypeImageConfig}}},
		distribution.Descriptor{},
	))
	require.False(t, needScan(
		fakeManifest{references: []distribution.Descriptor{{MediaType: cosignSimpleSigningMediaType}}},
		distribution.Descriptor{},
	))
}

func TestGetManifestByTagRecordsPull(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := t.Context()

	repository := &models.Repository{ID: "repo-id", NamespaceID: "namespace-id", Name: "library/alpine"}
	artifactDigest := digest.FromString("manifest").String()
	artifact := &models.Artifact{
		ID:           "artifact-id",
		RepositoryID: repository.ID,
		Digest:       artifactDigest,
		ContentType:  imgspecv1.MediaTypeImageManifest,
	}
	tag := &models.Tag{
		ID:         "tag-id",
		ArtifactID: artifact.ID,
		Name:       "latest",
		Artifact:   artifact,
	}
	raw := []byte(`{"schemaVersion":2}`)

	repoRegistry := reporegistry.NewMockRepositoryRepository(ctrl)
	repoTag := reporegistry.NewMockTagRepository(ctrl)
	repoArtifact := reporegistry.NewMockArtifactRepository(ctrl)
	storageDriver := mockstorage.NewMockStorageDriver(ctrl)
	analyticsSvc := svcanalytics.NewMockService(ctrl)

	repoRegistry.EXPECT().GetByName(ctx, repository.Name).Return(repository, nil)
	repoTag.EXPECT().GetByName(ctx, repository.ID, tag.Name).Return(tag, nil)
	repoTag.EXPECT().Incr(ctx, tag.ID).Return(nil)
	repoArtifact.EXPECT().GetByDigest(ctx, repository.ID, artifactDigest).Return(artifact, nil)
	storageDriver.EXPECT().Reader(ctx, utils.GenManifestPathByDigest(digest.Digest(artifactDigest))).Return(io.NopCloser(bytes.NewReader(raw)), nil)
	analyticsSvc.EXPECT().RecordPull(ctx, svcanalytics.PullEvent{
		NamespaceID:  repository.NamespaceID,
		RepositoryID: repository.ID,
		ArtifactID:   artifact.ID,
		Reference:    tag.Name,
	}).Return(nil)

	svc := &service{
		RepoRegistry: repoRegistry,
		RepoTag:      repoTag,
		RepoArtifact: repoArtifact,
		Storage:      storageDriver,
		SvcAnalytics: analyticsSvc,
	}
	got, contentType, gotTag, err := svc.GetManifest(ctx, repository.NamespaceID, repository.Name, tag.Name)
	require.NoError(t, err)
	require.Equal(t, raw, got)
	require.Equal(t, artifact.ContentType, contentType)
	require.Same(t, tag, gotTag)
}

func TestHeadManifestDoesNotRecordPull(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := t.Context()

	repository := &models.Repository{ID: "repo-id", NamespaceID: "namespace-id", Name: "library/alpine"}
	artifactDigest := digest.FromString("manifest").String()
	artifact := &models.Artifact{
		ID:           "artifact-id",
		RepositoryID: repository.ID,
		Digest:       artifactDigest,
		ContentType:  imgspecv1.MediaTypeImageManifest,
	}
	raw := []byte(`{"schemaVersion":2}`)

	repoRegistry := reporegistry.NewMockRepositoryRepository(ctrl)
	repoArtifact := reporegistry.NewMockArtifactRepository(ctrl)
	storageDriver := mockstorage.NewMockStorageDriver(ctrl)

	repoRegistry.EXPECT().GetByName(ctx, repository.Name).Return(repository, nil)
	repoArtifact.EXPECT().GetByDigest(ctx, repository.ID, artifactDigest).Return(artifact, nil)
	storageDriver.EXPECT().Reader(ctx, utils.GenManifestPathByDigest(digest.Digest(artifactDigest))).Return(io.NopCloser(bytes.NewReader(raw)), nil)

	svc := &service{
		RepoRegistry: repoRegistry,
		RepoArtifact: repoArtifact,
		Storage:      storageDriver,
	}
	got, contentType, tag, err := svc.HeadManifest(ctx, repository.NamespaceID, repository.Name, artifactDigest)
	require.NoError(t, err)
	require.Equal(t, raw, got)
	require.Equal(t, artifact.ContentType, contentType)
	require.Nil(t, tag)
}

func TestGetManifestErrorMapping(t *testing.T) {
	ctx := t.Context()
	repository := &models.Repository{ID: "repo-id", NamespaceID: "namespace-id", Name: "library/alpine"}
	artifactDigest := digest.FromString("manifest").String()

	tests := []struct {
		name    string
		ref     string
		setup   func(*gomock.Controller) *service
		wantErr errcode.ErrCode
	}{
		{
			name: "repository not found",
			ref:  artifactDigest,
			setup: func(ctrl *gomock.Controller) *service {
				repoRegistry := reporegistry.NewMockRepositoryRepository(ctrl)
				repoRegistry.EXPECT().GetByName(ctx, repository.Name).Return(nil, gorm.ErrRecordNotFound)
				return &service{RepoRegistry: repoRegistry}
			},
			wantErr: errcode.DSErrCodeNameUnknown,
		},
		{
			name: "tag not found",
			ref:  "missing",
			setup: func(ctrl *gomock.Controller) *service {
				repoRegistry := reporegistry.NewMockRepositoryRepository(ctrl)
				repoTag := reporegistry.NewMockTagRepository(ctrl)
				repoRegistry.EXPECT().GetByName(ctx, repository.Name).Return(repository, nil)
				repoTag.EXPECT().GetByName(ctx, repository.ID, "missing").Return(nil, gorm.ErrRecordNotFound)
				return &service{RepoRegistry: repoRegistry, RepoTag: repoTag}
			},
			wantErr: errcode.DSErrCodeManifestUnknown,
		},
		{
			name: "artifact not found",
			ref:  artifactDigest,
			setup: func(ctrl *gomock.Controller) *service {
				repoRegistry := reporegistry.NewMockRepositoryRepository(ctrl)
				repoArtifact := reporegistry.NewMockArtifactRepository(ctrl)
				repoRegistry.EXPECT().GetByName(ctx, repository.Name).Return(repository, nil)
				repoArtifact.EXPECT().GetByDigest(ctx, repository.ID, artifactDigest).Return(nil, gorm.ErrRecordNotFound)
				return &service{RepoRegistry: repoRegistry, RepoArtifact: repoArtifact}
			},
			wantErr: errcode.DSErrCodeManifestUnknown,
		},
		{
			name: "storage read failed",
			ref:  artifactDigest,
			setup: func(ctrl *gomock.Controller) *service {
				repoRegistry := reporegistry.NewMockRepositoryRepository(ctrl)
				repoArtifact := reporegistry.NewMockArtifactRepository(ctrl)
				storageDriver := mockstorage.NewMockStorageDriver(ctrl)
				artifact := &models.Artifact{ID: "artifact-id", RepositoryID: repository.ID, Digest: artifactDigest}
				repoRegistry.EXPECT().GetByName(ctx, repository.Name).Return(repository, nil)
				repoArtifact.EXPECT().GetByDigest(ctx, repository.ID, artifactDigest).Return(artifact, nil)
				storageDriver.EXPECT().Reader(ctx, utils.GenManifestPathByDigest(digest.Digest(artifactDigest))).Return(nil, errors.New("read failed"))
				return &service{RepoRegistry: repoRegistry, RepoArtifact: repoArtifact, Storage: storageDriver}
			},
			wantErr: errcode.DSErrCodeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			_, _, _, err := tt.setup(ctrl).GetManifest(ctx, repository.NamespaceID, repository.Name, tt.ref)
			require.Equal(t, tt.wantErr, err)
		})
	}
}

func TestPutManifestRejectsInvalidPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := t.Context()
	repository := &models.Repository{ID: "repo-id", NamespaceID: "namespace-id", Name: "library/alpine"}
	repoRegistry := reporegistry.NewMockRepositoryRepository(ctrl)
	repoRegistry.EXPECT().GetByName(ctx, repository.Name).Return(repository, nil)

	svc := &service{RepoRegistry: repoRegistry}
	dgst, err := svc.PutManifest(ctx, "user-id", repository.NamespaceID, repository.Name, "latest", []byte("not-json"), imgspecv1.MediaTypeImageManifest)
	require.Empty(t, dgst)
	require.Equal(t, errcode.DSErrCodeManifestInvalid, err)
}

func TestDeleteManifestByTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := t.Context()
	repository := &models.Repository{ID: "repo-id", NamespaceID: "namespace-id", Name: "library/alpine"}
	tag := &models.Tag{ID: "tag-id", RepositoryID: repository.ID, Name: "latest"}

	repoRegistry := reporegistry.NewMockRepositoryRepository(ctrl)
	repoTag := reporegistry.NewMockTagRepository(ctrl)
	repoRegistry.EXPECT().GetByName(ctx, repository.Name).Return(repository, nil)
	repoTag.EXPECT().GetByName(ctx, repository.ID, tag.Name).Return(tag, nil)
	repoTag.EXPECT().DeleteByName(ctx, repository.ID, tag.Name).Return(nil)

	svc := &service{RepoRegistry: repoRegistry, RepoTag: repoTag}
	require.NoError(t, svc.DeleteManifest(ctx, repository.NamespaceID, repository.Name, tag.Name, "user-id"))
}

func TestDeleteManifestErrorMapping(t *testing.T) {
	ctx := t.Context()
	repository := &models.Repository{ID: "repo-id", NamespaceID: "namespace-id", Name: "library/alpine"}

	tests := []struct {
		name    string
		ref     string
		setup   func(*gomock.Controller) *service
		wantErr errcode.ErrCode
	}{
		{
			name: "repository not found",
			ref:  "latest",
			setup: func(ctrl *gomock.Controller) *service {
				repoRegistry := reporegistry.NewMockRepositoryRepository(ctrl)
				repoRegistry.EXPECT().GetByName(ctx, repository.Name).Return(nil, gorm.ErrRecordNotFound)
				return &service{RepoRegistry: repoRegistry}
			},
			wantErr: errcode.DSErrCodeNameUnknown,
		},
		{
			name: "tag not found",
			ref:  "missing",
			setup: func(ctrl *gomock.Controller) *service {
				repoRegistry := reporegistry.NewMockRepositoryRepository(ctrl)
				repoTag := reporegistry.NewMockTagRepository(ctrl)
				repoRegistry.EXPECT().GetByName(ctx, repository.Name).Return(repository, nil)
				repoTag.EXPECT().GetByName(ctx, repository.ID, "missing").Return(nil, gorm.ErrRecordNotFound)
				return &service{RepoRegistry: repoRegistry, RepoTag: repoTag}
			},
			wantErr: errcode.DSErrCodeManifestUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			err := tt.setup(ctrl).DeleteManifest(ctx, repository.NamespaceID, repository.Name, tt.ref, "user-id")
			require.Equal(t, tt.wantErr, err)
		})
	}
}

func TestGetArtifactReferrer(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := t.Context()
	repository := &models.Repository{ID: "repo-id", NamespaceID: "namespace-id", Name: "library/alpine"}
	subjectDigest := digest.FromString("subject")
	referrerArtifact := &models.Artifact{ID: "referrer-id", RepositoryID: repository.ID, Digest: subjectDigest.String()}
	payload, err := json.Marshal(imgspecv1.Manifest{
		MediaType: imgspecv1.MediaTypeImageManifest,
		Subject:   &imgspecv1.Descriptor{Digest: subjectDigest},
	})
	require.NoError(t, err)

	repoRegistry := reporegistry.NewMockRepositoryRepository(ctrl)
	repoArtifact := reporegistry.NewMockArtifactRepository(ctrl)
	repoRegistry.EXPECT().GetByName(ctx, repository.Name).Return(repository, nil)
	repoArtifact.EXPECT().GetByDigest(ctx, repository.ID, subjectDigest.String()).Return(referrerArtifact, nil)

	svc := &service{RepoRegistry: repoRegistry, RepoArtifact: repoArtifact}
	got, err := svc.getArtifactReferrer(ctx, repository.Name, fakeManifest{
		mediaType: imgspecv1.MediaTypeImageManifest,
		payload:   payload,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, referrerArtifact.ID, *got)
}

func TestGetArtifactReferrerWithoutSubject(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := t.Context()
	repository := &models.Repository{ID: "repo-id", NamespaceID: "namespace-id", Name: "library/alpine"}
	payload, err := json.Marshal(imgspecv1.Manifest{MediaType: imgspecv1.MediaTypeImageManifest})
	require.NoError(t, err)

	repoRegistry := reporegistry.NewMockRepositoryRepository(ctrl)
	repoRegistry.EXPECT().GetByName(ctx, repository.Name).Return(repository, nil)

	svc := &service{RepoRegistry: repoRegistry}
	got, err := svc.getArtifactReferrer(ctx, repository.Name, fakeManifest{
		mediaType: imgspecv1.MediaTypeImageManifest,
		payload:   payload,
	})
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestGetReferrerBuildsOCIIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := t.Context()
	repository := &models.Repository{ID: "repo-id", NamespaceID: "namespace-id", Name: "library/alpine"}
	referrerDigest := digest.FromString("referrer")
	subjectDigest := digest.FromString("subject")
	artifact := &models.Artifact{
		ID:     "artifact-id",
		Digest: referrerDigest.String(),
		Size:   321,
	}
	rawManifest, err := json.Marshal(imgspecv1.Manifest{
		MediaType: imgspecv1.MediaTypeImageManifest,
		Config: imgspecv1.Descriptor{
			MediaType: cosignSimpleSigningMediaType,
		},
		Annotations: map[string]string{"key": "value"},
	})
	require.NoError(t, err)

	repoRegistry := reporegistry.NewMockRepositoryRepository(ctrl)
	repoArtifact := reporegistry.NewMockArtifactRepository(ctrl)
	storageDriver := mockstorage.NewMockStorageDriver(ctrl)
	repoRegistry.EXPECT().GetByName(ctx, repository.Name).Return(repository, nil)
	repoArtifact.EXPECT().GetReferrers(ctx, repository.ID, subjectDigest.String(), []string{"signature"}).Return([]*models.Artifact{artifact}, nil)
	storageDriver.EXPECT().Reader(ctx, utils.GenManifestPathByDigest(referrerDigest)).Return(io.NopCloser(bytes.NewReader(rawManifest)), nil)

	svc := &service{RepoRegistry: repoRegistry, RepoArtifact: repoArtifact, Storage: storageDriver}
	body, err := svc.GetReferrer(ctx, repository.Name, subjectDigest.String(), []string{"signature"})
	require.NoError(t, err)

	var index imgspecv1.Index
	require.NoError(t, json.Unmarshal(body, &index))
	require.Equal(t, imgspecv1.MediaTypeImageIndex, index.MediaType)
	require.Len(t, index.Manifests, 1)
	require.Equal(t, imgspecv1.MediaTypeImageManifest, index.Manifests[0].MediaType)
	require.Equal(t, artifact.Size, index.Manifests[0].Size)
	require.Equal(t, referrerDigest, index.Manifests[0].Digest)
	require.Equal(t, cosignSimpleSigningMediaType, index.Manifests[0].ArtifactType)
	require.Equal(t, map[string]string{"key": "value"}, index.Manifests[0].Annotations)
}

func TestReadManifestRejectsInvalidDigest(t *testing.T) {
	svc := &service{}
	raw, err := svc.readManifest(t.Context(), "not-a-digest")
	require.Nil(t, raw)
	require.Error(t, err)
}

type fakeManifest struct {
	references []distribution.Descriptor
	mediaType  string
	payload    []byte
}

func (f fakeManifest) References() []distribution.Descriptor {
	return f.references
}

func (f fakeManifest) Payload() (string, []byte, error) {
	return f.mediaType, f.payload, nil
}
