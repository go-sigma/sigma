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
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespacemocks "github.com/go-sigma/sigma/pkg/dal/repository/namespace/mocks"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	reporegistrymocks "github.com/go-sigma/sigma/pkg/dal/repository/registry/mocks"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	storagemocks "github.com/go-sigma/sigma/pkg/storage/mocks"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestUploadQueries(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceRepository := reponamespacemocks.NewMockNamespaceRepository(ctrl)
	blobRepository := reporegistrymocks.NewMockBlobRepository(ctrl)
	service := &service{
		RepoNs:   namespaceRepository,
		RepoBlob: blobRepository,
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

func TestCanMountBlob(t *testing.T) {
	ctrl := gomock.NewController(t)
	repositoryRepository := reporegistrymocks.NewMockRepositoryRepository(ctrl)
	blobRepository := reporegistrymocks.NewMockBlobRepository(ctrl)
	service := &service{
		RepoBlob:     blobRepository,
		RepoRegistry: repositoryRepository,
	}

	blobDigest := "sha256:fac304336cb8d9bcd2cbf47dbd4b6543677ee9c14bbe691621e76a07ad500b40"

	repositoryRepository.EXPECT().
		GetByName(gomock.Any(), "library/alpine").
		Return(&models.Repository{ID: "repository-1"}, nil)
	blobRepository.EXPECT().
		ExistsInRepository(gomock.Any(), "repository-1", blobDigest).
		Return(true, nil)

	exists, err := service.CanMountBlob(t.Context(), "library/alpine", blobDigest)
	require.NoError(t, err)
	require.True(t, exists)

	repositoryRepository.EXPECT().
		GetByName(gomock.Any(), "library/missing").
		Return(nil, gorm.ErrRecordNotFound)

	exists, err = service.CanMountBlob(t.Context(), "library/missing", blobDigest)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestDeleteUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistrymocks.NewMockBlobRepository(ctrl)
	storageDriver := storagemocks.NewMockStorageDriver(ctrl)
	service := &service{
		RepoBlob: blobRepository,
		Storage:  storageDriver,
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

func TestPostUploadSingleShot(t *testing.T) {
	ctrl := gomock.NewController(t)
	testkit.InitRepository(t)
	blobRepository := reporegistry.NewBlobRepository()
	storageDriver := storagemocks.NewMockStorageDriver(ctrl)
	svc := &service{RepoBlob: blobRepository, Storage: storageDriver}

	content := []byte("hello world")
	dgest := digest.FromBytes(content).String()

	storageDriver.EXPECT().Upload(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, body io.Reader) error {
			_, _ = io.ReadAll(body)
			return nil
		})
	storageDriver.EXPECT().Move(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	storageDriver.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
	storageDriver.EXPECT().CreateUploadID(gomock.Any(), gomock.Any()).Return("upload-id-1", nil)

	uploadID, err := svc.PostUpload(t.Context(), "library/alpine", dgest, bytes.NewReader(content), "application/octet-stream")
	require.NoError(t, err)
	require.Equal(t, "upload-id-1", uploadID)

	blobObj, err := blobRepository.FindByDigest(t.Context(), dgest)
	require.NoError(t, err)
	require.Equal(t, int64(len(content)), blobObj.Size)
}

func TestPostUploadWithoutDigest(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistrymocks.NewMockBlobRepository(ctrl)
	storageDriver := storagemocks.NewMockStorageDriver(ctrl)
	svc := &service{RepoBlob: blobRepository, Storage: storageDriver}

	storageDriver.EXPECT().CreateUploadID(gomock.Any(), gomock.Any()).Return("upload-id-1", nil)
	blobRepository.EXPECT().UploadCreate(gomock.Any(), gomock.Any()).Return(nil)

	uploadID, err := svc.PostUpload(t.Context(), "library/alpine", "", nil, "application/octet-stream")
	require.NoError(t, err)
	require.Equal(t, "upload-id-1", uploadID)
}

func TestPostUploadInvalidDigest(t *testing.T) {
	svc := &service{}

	uploadID, err := svc.PostUpload(t.Context(), "library/alpine", "not-a-digest", bytes.NewReader([]byte("x")), "application/octet-stream")
	require.Empty(t, uploadID)
	require.Equal(t, errcode.DSErrCodeBlobUploadInvalid, err)
}

func TestPatchUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistrymocks.NewMockBlobRepository(ctrl)
	storageDriver := storagemocks.NewMockStorageDriver(ctrl)
	svc := &service{RepoBlob: blobRepository, Storage: storageDriver}

	upload := &models.BlobUpload{UploadID: "storage-upload-1", FileID: "file-1", PartNumber: 0}
	blobRepository.EXPECT().UploadGetLastPart(gomock.Any(), "upload-1").Return(upload, nil)
	blobRepository.EXPECT().UploadTotalSizeByUploadID(gomock.Any(), "upload-1").Return(int64(10), nil)
	storageDriver.EXPECT().
		UploadPart(gomock.Any(), gomock.Any(), "storage-upload-1", int32(1), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, _ int32, body io.Reader) (string, error) {
			_, _ = io.ReadAll(body)
			return "etag1", nil
		})
	blobRepository.EXPECT().UploadCreate(gomock.Any(), gomock.Any()).Return(nil)

	sizeBefore, sizeUploaded, err := svc.PatchUpload(t.Context(), "upload-1", bytes.NewReader([]byte("part")), "library/alpine")
	require.NoError(t, err)
	require.Equal(t, int64(10), sizeBefore)
	require.Equal(t, int64(4), sizeUploaded)
}

func TestPatchUploadGetLastPartError(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistrymocks.NewMockBlobRepository(ctrl)
	svc := &service{RepoBlob: blobRepository}

	blobRepository.EXPECT().UploadGetLastPart(gomock.Any(), "upload-1").Return(nil, errors.New("lookup failed"))

	sizeBefore, sizeUploaded, err := svc.PatchUpload(t.Context(), "upload-1", bytes.NewReader([]byte("part")), "library/alpine")
	require.Zero(t, sizeBefore)
	require.Zero(t, sizeUploaded)
	require.Equal(t, errcode.DSErrCodeUnknown, err)
}

func TestPutUploadInvalidDigest(t *testing.T) {
	svc := &service{}

	sizeBefore, sizeUploaded, err := svc.PutUpload(t.Context(), "upload-1", "not-a-digest", nil, "library/alpine", "application/octet-stream", 0)
	require.Zero(t, sizeBefore)
	require.Zero(t, sizeUploaded)
	require.Equal(t, errcode.DSErrCodeDigestInvalid, err)
}

func TestPutUploadBlobExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistrymocks.NewMockBlobRepository(ctrl)
	storageDriver := storagemocks.NewMockStorageDriver(ctrl)
	svc := &service{RepoBlob: blobRepository, Storage: storageDriver}

	dgest := digest.FromBytes([]byte("content")).String()
	upload := &models.BlobUpload{UploadID: "storage-upload-1", FileID: "file-1", PartNumber: 0}
	blobRepository.EXPECT().UploadGetLastPart(gomock.Any(), "upload-1").Return(upload, nil)
	blobRepository.EXPECT().Exists(gomock.Any(), dgest).Return(true, nil)
	storageDriver.EXPECT().AbortUpload(gomock.Any(), gomock.Any(), "storage-upload-1").Return(nil)

	sizeBefore, sizeUploaded, err := svc.PutUpload(t.Context(), "upload-1", dgest, nil, "library/alpine", "application/octet-stream", 0)
	require.NoError(t, err)
	require.Zero(t, sizeBefore)
	require.Zero(t, sizeUploaded)
}

func TestPutUploadInvalidRepositoryName(t *testing.T) {
	ctrl := gomock.NewController(t)
	blobRepository := reporegistrymocks.NewMockBlobRepository(ctrl)
	svc := &service{RepoBlob: blobRepository}

	dgest := digest.FromBytes([]byte("content")).String()
	upload := &models.BlobUpload{UploadID: "storage-upload-1", FileID: "file-1", PartNumber: 0}
	blobRepository.EXPECT().UploadGetLastPart(gomock.Any(), "upload-1").Return(upload, nil)
	blobRepository.EXPECT().Exists(gomock.Any(), dgest).Return(false, nil)

	_, _, err := svc.PutUpload(t.Context(), "upload-1", dgest, nil, "UPPER", "application/octet-stream", 0)
	require.Equal(t, errcode.DSErrCodeNameInvalid, err)
}

func TestPutUploadDigestMismatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	testkit.InitRepository(t)
	blobRepository := reporegistry.NewBlobRepository()
	storageDriver := storagemocks.NewMockStorageDriver(ctrl)
	svc := &service{RepoBlob: blobRepository, Storage: storageDriver}

	content := []byte("actual content")
	wrongDigest := digest.FromBytes([]byte("different content")).String()
	require.NoError(t, blobRepository.UploadCreate(t.Context(), &models.BlobUpload{
		ID:         uuid.NewV7String(),
		UploadID:   "upload-1",
		PartNumber: 0,
		Etag:       "etag1",
		Repository: "library/alpine",
		FileID:     "file-1",
		Size:       10,
	}))
	storageDriver.EXPECT().CommitUpload(gomock.Any(), gomock.Any(), "upload-1", []string{}).Return(nil)
	storageDriver.EXPECT().Reader(gomock.Any(), gomock.Any()).Return(io.NopCloser(bytes.NewReader(content)), nil)

	_, _, err := svc.PutUpload(t.Context(), "upload-1", wrongDigest, nil, "library/alpine", "application/octet-stream", 0)
	require.Equal(t, errcode.DSErrCodeBlobUploadDigestMismatch, err)

	_, err = blobRepository.FindByDigest(t.Context(), wrongDigest)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestPutUploadSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	testkit.InitRepository(t)
	blobRepository := reporegistry.NewBlobRepository()
	storageDriver := storagemocks.NewMockStorageDriver(ctrl)
	svc := &service{RepoBlob: blobRepository, Storage: storageDriver}

	content := []byte("hello world")
	dgest := digest.FromBytes(content).String()
	require.NoError(t, blobRepository.UploadCreate(t.Context(), &models.BlobUpload{
		ID:         uuid.NewV7String(),
		UploadID:   "upload-1",
		PartNumber: 0,
		Etag:       "etag1",
		Repository: "library/alpine",
		FileID:     "file-1",
		Size:       10,
	}))
	storageDriver.EXPECT().
		UploadPart(gomock.Any(), gomock.Any(), "upload-1", int32(1), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, _ string, _ int32, body io.Reader) (string, error) {
			_, _ = io.ReadAll(body)
			return "etag2", nil
		})
	storageDriver.EXPECT().CommitUpload(gomock.Any(), gomock.Any(), "upload-1", []string{"etag2"}).Return(nil)
	storageDriver.EXPECT().Reader(gomock.Any(), gomock.Any()).Return(io.NopCloser(bytes.NewReader(content)), nil)
	storageDriver.EXPECT().Move(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	storageDriver.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)

	sizeBefore, sizeUploaded, err := svc.PutUpload(t.Context(), "upload-1", dgest, bytes.NewReader(content), "library/alpine", "application/octet-stream", int64(len(content)))
	require.NoError(t, err)
	require.Equal(t, int64(10), sizeBefore)
	require.Equal(t, int64(len(content)), sizeUploaded)

	blobObj, err := blobRepository.FindByDigest(t.Context(), dgest)
	require.NoError(t, err)
	require.Equal(t, int64(10)+int64(len(content)), blobObj.Size)
}
