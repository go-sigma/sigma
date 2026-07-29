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
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
)

func TestGetNamespaceByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceRepository := reponamespace.NewMockNamespaceRepository(ctrl)
	namespaceRepository.EXPECT().
		GetByName(gomock.Any(), "sigma").
		Return(&models.Namespace{ID: "namespace-1", Name: "sigma"}, nil)

	namespace, err := (&distributionBlobService{namespaceRepository: namespaceRepository}).
		GetNamespaceByName(t.Context(), "sigma")
	require.NoError(t, err)
	require.Equal(t, "namespace-1", namespace.ID)
}

func TestDeleteBlob(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistry.NewMockBlobRepository(ctrl)
	service := &distributionBlobService{blobRepository: blobRepository}
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
