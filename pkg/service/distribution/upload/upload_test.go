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

package upload

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/storage"
)

func TestUploadQueries(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceRepository := reponamespace.NewMockNamespaceRepository(ctrl)
	blobRepository := reporegistry.NewMockBlobRepository(ctrl)
	service := &distributionUploadService{
		namespaceRepository: namespaceRepository,
		blobRepository:      blobRepository,
	}

	namespaceRepository.EXPECT().
		GetByName(gomock.Any(), "sigma").
		Return(&models.Namespace{ID: "namespace-1"}, nil)
	namespace, err := service.GetNamespaceByName(t.Context(), "sigma")
	require.NoError(t, err)
	require.Equal(t, "namespace-1", namespace.ID)

	blobRepository.EXPECT().
		UploadGetLastPart(gomock.Any(), "upload-1").
		Return(&models.BlobUpload{ID: "part-1"}, nil)
	upload, err := service.GetUpload(t.Context(), "upload-1")
	require.NoError(t, err)
	require.Equal(t, "part-1", upload.ID)
}

func TestDeleteUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistry.NewMockBlobRepository(ctrl)
	storageDriver := storage.NewMockStorageDriver(ctrl)
	service := &distributionUploadService{
		blobRepository: blobRepository,
		storageDriver:  storageDriver,
	}
	upload := &models.BlobUpload{UploadID: "storage-upload-1", FileID: "file-1"}
	blobRepository.EXPECT().
		UploadGetLastPart(gomock.Any(), "upload-1").
		Return(upload, nil)
	storageDriver.EXPECT().
		AbortUpload(gomock.Any(), "blob_uploads/file-1", "storage-upload-1").
		Return(nil)
	blobRepository.EXPECT().
		UploadDeleteByUploadID(gomock.Any(), "upload-1").
		Return(nil)

	require.NoError(t, service.DeleteUpload(t.Context(), "upload-1"))

	blobRepository.EXPECT().
		UploadGetLastPart(gomock.Any(), "missing").
		Return(nil, errors.New("missing"))
	require.Error(t, service.DeleteUpload(t.Context(), "missing"))
}
