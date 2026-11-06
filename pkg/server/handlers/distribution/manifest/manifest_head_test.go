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
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestHeadManifest(t *testing.T) {
	tests := []struct {
		name      string
		reference string
		tag       *models.Tag
	}{
		{
			name:      "by tag",
			reference: testTagName,
			tag:       &models.Tag{Artifact: &models.Artifact{Digest: testManifestDigest}},
		},
		{
			name:      "by digest",
			reference: testManifestDigest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := manifestmocks.NewMockDistributionManifestService(ctrl)
			authorizer := authz.NewMockAuthorizer(ctrl)

			body := []byte(`{"schemaVersion":2}`)
			svc.EXPECT().GetNamespaceByName(gomock.Any(), testNamespaceName).
				Return(&models.Namespace{ID: testNamespaceID, Name: testNamespaceName}, nil)
			authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), testNamespaceID, enums.AuthRead).Return(true, nil)
			svc.EXPECT().HeadManifest(gomock.Any(), testNamespaceID, testRepositoryName, tt.reference).
				Return(body, testManifestContentType, tt.tag, nil)

			recorder, c := newManifestContext(t, http.MethodHead, "/v2/"+testRepositoryName+"/manifests/"+tt.reference, nil)
			h := &handler{Config: &config.Configuration{}, ManifestSvc: svc, Authorizer: authorizer}
			h.HeadManifest(c)
			c.Writer.WriteHeaderNow()

			require.Equal(t, http.StatusOK, recorder.Code)
			require.Equal(t, testManifestContentType, recorder.Header().Get(consts.HeaderContentType))
			require.Equal(t, strconv.Itoa(len(body)), recorder.Header().Get(consts.HeaderContentLength))
			require.Equal(t, testManifestDigest, recorder.Header().Get(consts.ContentDigest))
			require.Empty(t, recorder.Body.String())
		})
	}
}

func TestHeadManifestErrors(t *testing.T) {
	namespace := &models.Namespace{ID: testNamespaceID, Name: testNamespaceName}

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
			name:   "authorization lookup error",
			target: "/v2/" + testRepositoryName + "/manifests/" + testTagName,
			setup: func(svc *manifestmocks.MockDistributionManifestService, authorizer *authz.MockAuthorizer) {
				svc.EXPECT().GetNamespaceByName(gomock.Any(), testNamespaceName).Return(namespace, nil)
				authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), testNamespaceID, enums.AuthRead).Return(false, gorm.ErrRecordNotFound)
			},
			wantStatus: http.StatusNotFound,
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
			name:   "manifest unknown",
			target: "/v2/" + testRepositoryName + "/manifests/" + testTagName,
			setup: func(svc *manifestmocks.MockDistributionManifestService, authorizer *authz.MockAuthorizer) {
				svc.EXPECT().GetNamespaceByName(gomock.Any(), testNamespaceName).Return(namespace, nil)
				authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), testNamespaceID, enums.AuthRead).Return(true, nil)
				svc.EXPECT().HeadManifest(gomock.Any(), testNamespaceID, testRepositoryName, testTagName).
					Return(nil, "", nil, errcode.DSErrCodeManifestUnknown)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "service error",
			target: "/v2/" + testRepositoryName + "/manifests/" + testTagName,
			setup: func(svc *manifestmocks.MockDistributionManifestService, authorizer *authz.MockAuthorizer) {
				svc.EXPECT().GetNamespaceByName(gomock.Any(), testNamespaceName).Return(namespace, nil)
				authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), testNamespaceID, enums.AuthRead).Return(true, nil)
				svc.EXPECT().HeadManifest(gomock.Any(), testNamespaceID, testRepositoryName, testTagName).
					Return(nil, "", nil, errors.New("db error"))
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

			recorder, c := newManifestContext(t, http.MethodHead, tt.target, nil)
			h := &handler{Config: &config.Configuration{}, ManifestSvc: svc, Authorizer: authorizer}
			h.HeadManifest(c)
			c.Writer.WriteHeaderNow()

			require.Equal(t, tt.wantStatus, recorder.Code)
		})
	}
}

func TestHeadManifestFallbackProxy(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v2/"+testRepositoryName+"/manifests/"+testTagName, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(consts.HeaderContentType, testManifestContentType)
		w.WriteHeader(http.StatusOK)
	})

	upstream := httptest.NewServer(mux)
	defer upstream.Close()

	ctrl := gomock.NewController(t)
	svc := manifestmocks.NewMockDistributionManifestService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	svc.EXPECT().GetNamespaceByName(gomock.Any(), testNamespaceName).
		Return(&models.Namespace{ID: testNamespaceID, Name: testNamespaceName}, nil)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), testNamespaceID, enums.AuthRead).Return(true, nil)
	svc.EXPECT().HeadManifest(gomock.Any(), testNamespaceID, testRepositoryName, testTagName).
		Return(nil, "", nil, errcode.DSErrCodeManifestUnknown)

	recorder, c := newManifestContext(t, http.MethodHead, "/v2/"+testRepositoryName+"/manifests/"+testTagName, nil)
	h := &handler{
		Config: &config.Configuration{
			Proxy: config.ConfigurationProxy{Enabled: true, Endpoint: upstream.URL},
		},
		ManifestSvc: svc,
		Authorizer:  authorizer,
	}
	h.HeadManifest(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, testManifestContentType, recorder.Header().Get(consts.HeaderContentType))
}

func TestHeadManifestFallbackProxyError(t *testing.T) {
	// The upstream registry answers the /v2/ ping with an error, so building the
	// proxy client fails and the handler reports an unknown distribution error.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer upstream.Close()

	ctrl := gomock.NewController(t)
	svc := manifestmocks.NewMockDistributionManifestService(ctrl)
	authorizer := authz.NewMockAuthorizer(ctrl)
	svc.EXPECT().GetNamespaceByName(gomock.Any(), testNamespaceName).
		Return(&models.Namespace{ID: testNamespaceID, Name: testNamespaceName}, nil)
	authorizer.EXPECT().Namespace(gomock.Any(), gomock.Any(), testNamespaceID, enums.AuthRead).Return(true, nil)
	svc.EXPECT().HeadManifest(gomock.Any(), testNamespaceID, testRepositoryName, testTagName).
		Return(nil, "", nil, errcode.DSErrCodeManifestUnknown)

	recorder, c := newManifestContext(t, http.MethodHead, "/v2/"+testRepositoryName+"/manifests/"+testTagName, nil)
	h := &handler{
		Config: &config.Configuration{
			Proxy: config.ConfigurationProxy{Enabled: true, Endpoint: upstream.URL},
		},
		ManifestSvc: svc,
		Authorizer:  authorizer,
	}
	h.HeadManifest(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}
