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
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	manifestmocks "github.com/go-sigma/sigma/pkg/service/distribution/manifest/mocks"
)

func TestPutManifest(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := manifestmocks.NewMockDistributionManifestService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	body := []byte(`{"schemaVersion":2}`)

	svc.EXPECT().GetNamespaceByName(gomock.Any(), testNamespaceName).
		Return(&models.Namespace{ID: testNamespaceID, Name: testNamespaceName}, nil)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), testNamespaceID, enums.AuthRead).Return(true, nil)
	svc.EXPECT().PutManifest(gomock.Any(), testUserID, testNamespaceID, testRepositoryName, testTagName, body, testManifestContentType).
		Return(testManifestDigest, nil)

	recorder, c := newManifestContext(t, http.MethodPut, "/v2/"+testRepositoryName+"/manifests/"+testTagName, bytes.NewReader(body))
	c.Request.Header.Set(consts.HeaderContentType, testManifestContentType)

	h := &handler{Config: &config.Configuration{}, ManifestSvc: svc, Authorizer: authorizer}
	h.PutManifest(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Equal(t, "/v2/"+testRepositoryName+"/manifests/"+testManifestDigest, recorder.Header().Get(consts.HeaderLocation))
	require.Equal(t, testManifestDigest, recorder.Header().Get(consts.ContentDigest))
}

func TestPutManifestErrors(t *testing.T) {
	namespace := &models.Namespace{ID: testNamespaceID, Name: testNamespaceName}
	body := []byte(`{"schemaVersion":2}`)

	tests := []struct {
		name       string
		target     string
		setup      func(*manifestmocks.MockDistributionManifestService, *authz.MockAuthorizer)
		wantStatus int
	}{
		{
			name:   "namespace not found",
			target: "/v2/" + testRepositoryName + "/manifests/" + testTagName,
			setup: func(svc *manifestmocks.MockDistributionManifestService, _ *authz.MockAuthorizer) {
				svc.EXPECT().GetNamespaceByName(gomock.Any(), testNamespaceName).Return(nil, gorm.ErrRecordNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "authorization denied",
			target: "/v2/" + testRepositoryName + "/manifests/" + testTagName,
			setup: func(svc *manifestmocks.MockDistributionManifestService, authorizer *authz.MockAuthorizer) {
				svc.EXPECT().GetNamespaceByName(gomock.Any(), testNamespaceName).Return(namespace, nil)
				authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), testNamespaceID, enums.AuthRead).Return(false, nil)
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "invalid reference",
			target: "/v2/" + testRepositoryName + "/manifests/*&invalid-tag",
			setup: func(svc *manifestmocks.MockDistributionManifestService, authorizer *authz.MockAuthorizer) {
				svc.EXPECT().GetNamespaceByName(gomock.Any(), testNamespaceName).Return(namespace, nil)
				authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), testNamespaceID, enums.AuthRead).Return(true, nil)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "manifest invalid",
			target: "/v2/" + testRepositoryName + "/manifests/" + testTagName,
			setup: func(svc *manifestmocks.MockDistributionManifestService, authorizer *authz.MockAuthorizer) {
				svc.EXPECT().GetNamespaceByName(gomock.Any(), testNamespaceName).Return(namespace, nil)
				authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), testNamespaceID, enums.AuthRead).Return(true, nil)
				svc.EXPECT().PutManifest(gomock.Any(), testUserID, testNamespaceID, testRepositoryName, testTagName, body, testManifestContentType).
					Return("", errcode.DSErrCodeManifestInvalid)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "service error",
			target: "/v2/" + testRepositoryName + "/manifests/" + testTagName,
			setup: func(svc *manifestmocks.MockDistributionManifestService, authorizer *authz.MockAuthorizer) {
				svc.EXPECT().GetNamespaceByName(gomock.Any(), testNamespaceName).Return(namespace, nil)
				authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), testNamespaceID, enums.AuthRead).Return(true, nil)
				svc.EXPECT().PutManifest(gomock.Any(), testUserID, testNamespaceID, testRepositoryName, testTagName, body, testManifestContentType).
					Return("", errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := manifestmocks.NewMockDistributionManifestService(ctrl)
			authorizer := authz.NewMockAuthorizer(ctrl)
			tt.setup(svc, authorizer)

			recorder, c := newManifestContext(t, http.MethodPut, tt.target, bytes.NewReader(body))
			c.Request.Header.Set(consts.HeaderContentType, testManifestContentType)

			h := &handler{Config: &config.Configuration{}, ManifestSvc: svc, Authorizer: authorizer}
			h.PutManifest(c)
			c.Writer.WriteHeaderNow()

			require.Equal(t, tt.wantStatus, recorder.Code)
		})
	}
}
