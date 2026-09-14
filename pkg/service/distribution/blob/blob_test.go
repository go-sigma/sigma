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

package blob

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/storage"
)

func TestGetNamespaceByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceRepository := reponamespace.NewMockNamespaceRepository(ctrl)
	namespaceRepository.EXPECT().
		GetByName(gomock.Any(), "sigma").
		Return(&models.Namespace{ID: "namespace-1", Name: "sigma"}, nil)

	namespace, err := (&service{RepoNs: namespaceRepository}).
		GetNamespaceByName(t.Context(), "sigma")
	require.NoError(t, err)
	require.Equal(t, "namespace-1", namespace.ID)
}

func TestDeleteBlob(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistry.NewMockBlobRepository(ctrl)
	service := &service{RepoBlob: blobRepository}
	blobRepository.EXPECT().
		FindByDigest(gomock.Any(), "sha256:digest").
		Return(&models.Blob{ID: "blob-1"}, nil)
	blobRepository.EXPECT().
		FindAssociateWithArtifact(gomock.Any(), []string{"blob-1"}).
		Return(nil, nil)
	blobRepository.EXPECT().DeleteByID(gomock.Any(), "blob-1").Return(nil)

	require.NoError(t, service.DeleteBlob(t.Context(), "namespace-1", "sha256:digest", "user-1"))

	blobRepository.EXPECT().
		FindByDigest(gomock.Any(), "sha256:failed").
		Return(nil, errors.New("lookup failed"))
	require.Error(t, service.DeleteBlob(t.Context(), "namespace-1", "sha256:failed", "user-1"))
}

func TestProxyReadCloser(t *testing.T) {
	body := &trackingReadCloser{Reader: strings.NewReader("content")}
	reader := &proxyReadCloser{Reader: body, body: body}

	require.NoError(t, reader.Close())
	require.True(t, body.closed)
}

type trackingReadCloser struct {
	*strings.Reader
	closed bool
}

func (r *trackingReadCloser) Close() error {
	r.closed = true
	return nil
}

func TestDeleteBlobNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistry.NewMockBlobRepository(ctrl)
	svc := &service{RepoBlob: blobRepository}

	blobRepository.EXPECT().FindByDigest(gomock.Any(), "sha256:missing").Return(nil, gorm.ErrRecordNotFound)

	err := svc.DeleteBlob(t.Context(), "namespace-1", "sha256:missing", "user-1")
	require.Equal(t, errcode.DSErrCodeBlobUnknown, err)
}

func TestDeleteBlobAssociated(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistry.NewMockBlobRepository(ctrl)
	svc := &service{RepoBlob: blobRepository}

	blobRepository.EXPECT().FindByDigest(gomock.Any(), "sha256:digest").Return(&models.Blob{ID: "blob-1"}, nil)
	blobRepository.EXPECT().FindAssociateWithArtifact(gomock.Any(), []string{"blob-1"}).Return([]string{"artifact-1"}, nil)

	err := svc.DeleteBlob(t.Context(), "namespace-1", "sha256:digest", "user-1")
	require.Equal(t, errcode.DSErrCodeBlobAssociated, err)
}

func TestHeadBlobFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistry.NewMockBlobRepository(ctrl)
	svc := &service{Config: &config.Configuration{}, RepoBlob: blobRepository}

	blobRepository.EXPECT().FindByDigest(gomock.Any(), "sha256:digest").Return(&models.Blob{ID: "blob-1"}, nil)

	blob, err := svc.HeadBlob(t.Context(), "namespace-1", "sha256:digest", "HEAD", "/v2/library/alpine/blobs/sha256:digest")
	require.NoError(t, err)
	require.Equal(t, "blob-1", blob.ID)
}

func TestHeadBlobNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistry.NewMockBlobRepository(ctrl)
	svc := &service{Config: &config.Configuration{}, RepoBlob: blobRepository}

	blobRepository.EXPECT().FindByDigest(gomock.Any(), "sha256:missing").Return(nil, gorm.ErrRecordNotFound)

	blob, err := svc.HeadBlob(t.Context(), "namespace-1", "sha256:missing", "HEAD", "/v2/library/alpine/blobs/sha256:missing")
	require.Nil(t, blob)
	require.Equal(t, errcode.DSErrCodeBlobUnknown, err)
}

func TestGetBlobInvalidDigest(t *testing.T) {
	svc := &service{Config: &config.Configuration{}}

	reader, blob, redirect, err := svc.GetBlob(t.Context(), "namespace-1", "not-a-digest", "GET", "/v2/library/alpine/blobs/not-a-digest")
	require.Nil(t, reader)
	require.Nil(t, blob)
	require.Empty(t, redirect)
	require.Equal(t, errcode.DSErrCodeDigestInvalid, err)
}

func TestGetBlobNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistry.NewMockBlobRepository(ctrl)
	svc := &service{Config: &config.Configuration{}, RepoBlob: blobRepository}

	dgest := digest.FromBytes([]byte("content")).String()
	blobRepository.EXPECT().FindByDigest(gomock.Any(), dgest).Return(nil, gorm.ErrRecordNotFound)

	reader, blob, redirect, err := svc.GetBlob(t.Context(), "namespace-1", dgest, "GET", "/v2/library/alpine/blobs/"+dgest)
	require.Nil(t, reader)
	require.Nil(t, blob)
	require.Empty(t, redirect)
	require.Equal(t, errcode.DSErrCodeBlobUnknown, err)
}

func TestGetBlobRedirect(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistry.NewMockBlobRepository(ctrl)
	storageDriver := storage.NewMockStorageDriver(ctrl)
	svc := &service{
		Config: &config.Configuration{
			Storage: config.ConfigurationStorage{Redirect: true, Type: enums.StorageTypeS3},
		},
		RepoBlob: blobRepository,
		Storage:  storageDriver,
	}

	dgest := digest.FromBytes([]byte("content")).String()
	blobRepository.EXPECT().FindByDigest(gomock.Any(), dgest).Return(&models.Blob{ID: "blob-1", Digest: dgest}, nil)
	storageDriver.EXPECT().Redirect(gomock.Any(), gomock.Any()).Return("https://example.com/blob", nil)

	reader, blob, redirect, err := svc.GetBlob(t.Context(), "namespace-1", dgest, "GET", "/v2/library/alpine/blobs/"+dgest)
	require.NoError(t, err)
	require.Nil(t, reader)
	require.Equal(t, "blob-1", blob.ID)
	require.Equal(t, "https://example.com/blob", redirect)
}

func TestGetBlobSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistry.NewMockBlobRepository(ctrl)
	storageDriver := storage.NewMockStorageDriver(ctrl)
	svc := &service{Config: &config.Configuration{}, RepoBlob: blobRepository, Storage: storageDriver}

	content := []byte("blob content")
	dgest := digest.FromBytes(content).String()
	blobRepository.EXPECT().FindByDigest(gomock.Any(), dgest).Return(&models.Blob{ID: "blob-1", Digest: dgest}, nil)
	storageDriver.EXPECT().Reader(gomock.Any(), gomock.Any()).Return(io.NopCloser(bytes.NewReader(content)), nil)

	reader, blob, redirect, err := svc.GetBlob(t.Context(), "namespace-1", dgest, "GET", "/v2/library/alpine/blobs/"+dgest)
	require.NoError(t, err)
	require.NotNil(t, reader)
	require.Equal(t, "blob-1", blob.ID)
	require.Empty(t, redirect)
}

func TestGetBlobReaderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistry.NewMockBlobRepository(ctrl)
	storageDriver := storage.NewMockStorageDriver(ctrl)
	svc := &service{Config: &config.Configuration{}, RepoBlob: blobRepository, Storage: storageDriver}

	dgest := digest.FromBytes([]byte("content")).String()
	blobRepository.EXPECT().FindByDigest(gomock.Any(), dgest).Return(&models.Blob{ID: "blob-1", Digest: dgest}, nil)
	storageDriver.EXPECT().Reader(gomock.Any(), gomock.Any()).Return(nil, errors.New("read failed"))

	reader, blob, redirect, err := svc.GetBlob(t.Context(), "namespace-1", dgest, "GET", "/v2/library/alpine/blobs/"+dgest)
	require.Nil(t, reader)
	require.Nil(t, blob)
	require.Empty(t, redirect)
	require.Equal(t, errcode.DSErrCodeUnknown, err)
}
