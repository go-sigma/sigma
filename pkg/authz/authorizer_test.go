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

package authz

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
)

// ---------------------------------------------------------------------------
// TestResolveRequest: URL parsing
// ---------------------------------------------------------------------------

func TestResolveRequest(t *testing.T) {
	cases := []struct {
		name      string
		uri       string
		wantNS    string
		wantNSID  string
		wantRes   Resource
		wantRef   string
		wantError bool
	}{
		{name: "base", uri: "/v2/", wantRes: ResBase},
		{name: "base-no-slash", uri: "/v2", wantRes: ResBase},
		{name: "catalog", uri: "/v2/_catalog", wantRes: ResCatalog},
		{name: "tags", uri: "/v2/library/nginx/tags/list", wantNS: "library", wantRes: ResTags},
		{name: "blobs", uri: "/v2/library/nginx/blobs/sha256:abc", wantNS: "library", wantRes: ResBlobs, wantRef: "sha256:abc"},
		{name: "manifests", uri: "/v2/library/nginx/manifests/latest", wantNS: "library", wantRes: ResManifests, wantRef: "latest"},
		{name: "uploads-list", uri: "/v2/library/nginx/blobs/uploads/", wantNS: "library", wantRes: ResBlobUploads},
		{name: "uploads-list-no-slash", uri: "/v2/library/nginx/blobs/uploads", wantNS: "library", wantRes: ResBlobUploads},
		{name: "uploads-ref", uri: "/v2/library/nginx/blobs/uploads/abc-123", wantNS: "library", wantRes: ResBlobUploads, wantRef: "abc-123"},
		{name: "single-segment-repo", uri: "/v2/nginx/manifests/v1", wantNS: "nginx", wantRes: ResManifests, wantRef: "v1"},
		{name: "api-namespace", uri: "/api/v1/namespaces/123", wantNSID: "123", wantRes: ResNamespace},
		{name: "api-artifacts", uri: "/api/v1/namespaces/123/artifacts/456", wantNSID: "123", wantRes: ResArtifacts},
		{name: "api-repositories", uri: "/api/v1/namespaces/123/repositories", wantNSID: "123", wantRes: ResRepositories},
		{name: "with-query", uri: "/v2/library/nginx/tags/list?n=10&last=abc", wantNS: "library", wantRes: ResTags},
		{name: "unknown-path", uri: "/foo/bar", wantError: true},
		{name: "api-string-id", uri: "/api/v1/namespaces/abc", wantNSID: "abc", wantRes: ResNamespace},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			desc, err := ResolveRequest(tc.uri)
			if tc.wantError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantNS, desc.NamespaceName)
			assert.Equal(t, tc.wantNSID, desc.NamespaceID)
			assert.Equal(t, tc.wantRes, desc.Resource)
			assert.Equal(t, tc.wantRef, desc.Reference)
		})
	}
}

// ---------------------------------------------------------------------------
// TestMethodToAction
// ---------------------------------------------------------------------------

func TestMethodToAction(t *testing.T) {
	assert.Equal(t, ActionRead, MethodToAction(http.MethodGet))
	assert.Equal(t, ActionRead, MethodToAction(http.MethodHead))
	assert.Equal(t, ActionWrite, MethodToAction(http.MethodPost))
	assert.Equal(t, ActionWrite, MethodToAction(http.MethodPut))
	assert.Equal(t, ActionWrite, MethodToAction(http.MethodDelete))
	assert.Equal(t, ActionWrite, MethodToAction(http.MethodPatch))
}

// ---------------------------------------------------------------------------
// TestCheckPublicAccess
// ---------------------------------------------------------------------------

func TestCheckPublicAccess(t *testing.T) {
	// Global resources: read-only for everyone.
	assert.True(t, CheckPublicAccess(ResBase, ActionRead, enums.VisibilityPublic))
	assert.True(t, CheckPublicAccess(ResBase, ActionRead, enums.VisibilityPrivate))
	assert.False(t, CheckPublicAccess(ResBase, ActionWrite, enums.VisibilityPublic))
	assert.True(t, CheckPublicAccess(ResCatalog, ActionRead, enums.VisibilityPublic))
	assert.False(t, CheckPublicAccess(ResCatalog, ActionWrite, enums.VisibilityPublic))

	// Private namespace: anonymous/non-member cannot read anything.
	assert.False(t, CheckPublicAccess(ResBlobs, ActionRead, enums.VisibilityPrivate))
	assert.False(t, CheckPublicAccess(ResNamespace, ActionRead, enums.VisibilityPrivate))

	// Public namespace: read on readable resources allowed, write denied.
	assert.True(t, CheckPublicAccess(ResBlobs, ActionRead, enums.VisibilityPublic))
	assert.True(t, CheckPublicAccess(ResManifests, ActionRead, enums.VisibilityPublic))
	assert.True(t, CheckPublicAccess(ResTags, ActionRead, enums.VisibilityPublic))
	assert.True(t, CheckPublicAccess(ResNamespace, ActionRead, enums.VisibilityPublic))
	assert.True(t, CheckPublicAccess(ResArtifacts, ActionRead, enums.VisibilityPublic))
	assert.True(t, CheckPublicAccess(ResRepositories, ActionRead, enums.VisibilityPublic))
	assert.False(t, CheckPublicAccess(ResBlobUploads, ActionRead, enums.VisibilityPublic))
	assert.False(t, CheckPublicAccess(ResBlobs, ActionWrite, enums.VisibilityPublic))
}

// ---------------------------------------------------------------------------
// TestCheckRole: permission matrix
// ---------------------------------------------------------------------------

func TestCheckRole(t *testing.T) {
	allResources := []Resource{ResBlobs, ResBlobUploads, ResManifests, ResTags, ResNamespace, ResArtifacts, ResRepositories}

	// Reader: read-only on namespace-scoped resources (not blob_uploads).
	t.Run("Reader", func(t *testing.T) {
		for _, r := range allResources {
			got := CheckRole(enums.NamespaceRoleReader, r, ActionRead)
			if r == ResBlobUploads {
				assert.False(t, got, "Reader should not read %s", r)
			} else {
				assert.True(t, got, "Reader should read %s", r)
			}
			assert.False(t, CheckRole(enums.NamespaceRoleReader, r, ActionWrite), "Reader should never write %s", r)
		}
	})

	// Manager: read everything + write on image-layer resources.
	t.Run("Manager", func(t *testing.T) {
		for _, r := range allResources {
			assert.True(t, CheckRole(enums.NamespaceRoleManager, r, ActionRead), "Manager should read %s", r)
		}
		writeResources := map[Resource]bool{
			ResBlobs: true, ResBlobUploads: true, ResManifests: true, ResTags: true,
			ResRepositories: true, ResArtifacts: true,
		}
		for _, r := range allResources {
			got := CheckRole(enums.NamespaceRoleManager, r, ActionWrite)
			assert.Equal(t, writeResources[r], got, "Manager write on %s", r)
		}
		// Manager cannot write namespace meta.
		assert.False(t, CheckRole(enums.NamespaceRoleManager, ResNamespace, ActionWrite))
	})

	// Admin: everything allowed.
	t.Run("Admin", func(t *testing.T) {
		for _, r := range allResources {
			assert.True(t, CheckRole(enums.NamespaceRoleAdmin, r, ActionRead), "Admin should read %s", r)
			assert.True(t, CheckRole(enums.NamespaceRoleAdmin, r, ActionWrite), "Admin should write %s", r)
		}
	})

	// Unknown role: deny all.
	t.Run("Unknown", func(t *testing.T) {
		for _, r := range allResources {
			assert.False(t, CheckRole(enums.NamespaceRole("unknown"), r, ActionRead))
			assert.False(t, CheckRole(enums.NamespaceRole("unknown"), r, ActionWrite))
		}
	})
}

// ---------------------------------------------------------------------------
// NewAuthorizer construction tests
// ---------------------------------------------------------------------------

func TestNewAuthorizerInmemory(t *testing.T) {
	cfg := testAuthzConfig()
	a, err := NewAuthorizer(authorizer{
		Config:       cfg,
		RepoNs:       reponamespace.NewMockNamespaceRepository(gomock.NewController(t)),
		RepoNsMember: reponamespace.NewMockNamespaceMemberRepository(gomock.NewController(t)),
	})
	require.NoError(t, err)
	require.True(t, a != nil)
}

func TestNewAuthorizerRedisDisabled(t *testing.T) {
	cfg := &config.Configuration{}
	cfg.Cache.Type = enums.CacherTypeRedis
	cfg.Cache.WithDefaults()
	// Redis not enabled and no client factory: construction must fail.
	_, err := NewAuthorizer(authorizer{
		Config:       cfg,
		RepoNs:       reponamespace.NewMockNamespaceRepository(gomock.NewController(t)),
		RepoNsMember: reponamespace.NewMockNamespaceMemberRepository(gomock.NewController(t)),
	})
	require.Error(t, err)
}

func testAuthzConfig() *config.Configuration {
	cfg := &config.Configuration{}
	cfg.Cache.Type = enums.CacherTypeInmemory
	cfg.Cache.WithDefaults()
	return cfg
}

// ---------------------------------------------------------------------------
// TestAuthorizer: end-to-end with mock repositories
// ---------------------------------------------------------------------------

func TestAuthorizer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nsRepo := reponamespace.NewMockNamespaceRepository(ctrl)

	memberRepo := reponamespace.NewMockNamespaceMemberRepository(ctrl)
	ctx := context.Background()
	publicNS := &models.Namespace{ID: "1", Name: "library", Visibility: enums.VisibilityPublic}
	privateNS := &models.Namespace{ID: "2", Name: "private", Visibility: enums.VisibilityPrivate}

	// Setup namespace lookups.
	nsRepo.EXPECT().GetByName(gomock.Any(), "library").Return(publicNS, nil).AnyTimes()
	nsRepo.EXPECT().GetByName(gomock.Any(), "private").Return(privateNS, nil).AnyTimes()
	nsRepo.EXPECT().GetByName(gomock.Any(), "newns").Return(nil, gorm.ErrRecordNotFound).AnyTimes()
	nsRepo.EXPECT().Get(gomock.Any(), "1").Return(publicNS, nil).AnyTimes()
	nsRepo.EXPECT().Get(gomock.Any(), "2").Return(privateNS, nil).AnyTimes()

	// Setup membership lookups.
	memberRepo.EXPECT().GetNamespaceMember(gomock.Any(), "1", "100").Return(&models.NamespaceMember{Role: enums.NamespaceRoleReader}, nil).AnyTimes()
	memberRepo.EXPECT().GetNamespaceMember(gomock.Any(), "1", "200").Return(&models.NamespaceMember{Role: enums.NamespaceRoleManager}, nil).AnyTimes()
	memberRepo.EXPECT().GetNamespaceMember(gomock.Any(), "1", "300").Return(&models.NamespaceMember{Role: enums.NamespaceRoleAdmin}, nil).AnyTimes()
	memberRepo.EXPECT().GetNamespaceMember(gomock.Any(), "1", "400").Return(nil, gorm.ErrRecordNotFound).AnyTimes()
	memberRepo.EXPECT().GetNamespaceMember(gomock.Any(), "2", "100").Return(&models.NamespaceMember{Role: enums.NamespaceRoleReader}, nil).AnyTimes()
	memberRepo.EXPECT().GetNamespaceMember(gomock.Any(), "2", "200").Return(&models.NamespaceMember{Role: enums.NamespaceRoleManager}, nil).AnyTimes()
	memberRepo.EXPECT().GetNamespaceMember(gomock.Any(), "2", "400").Return(nil, gorm.ErrRecordNotFound).AnyTimes()

	cases := []struct {
		name        string
		userID      string
		isAnonymous bool
		uri         string
		method      string
		wantPass    bool
	}{
		// Global resources.
		{name: "anon-base-read", userID: "", isAnonymous: true, uri: "/v2/", method: http.MethodGet, wantPass: true},
		{name: "anon-catalog-read", userID: "", isAnonymous: true, uri: "/v2/_catalog", method: http.MethodGet, wantPass: true},
		{name: "anon-base-write", userID: "", isAnonymous: true, uri: "/v2/", method: http.MethodPost, wantPass: false},

		// Anonymous on public namespace.
		{name: "anon-public-blobs-read", userID: "", isAnonymous: true, uri: "/v2/library/nginx/blobs/sha256:abc", method: http.MethodGet, wantPass: true},
		{name: "anon-public-manifests-read", userID: "", isAnonymous: true, uri: "/v2/library/nginx/manifests/latest", method: http.MethodHead, wantPass: true},
		{name: "anon-public-blobs-write", userID: "", isAnonymous: true, uri: "/v2/library/nginx/blobs/sha256:abc", method: http.MethodPost, wantPass: false},
		{name: "anon-public-uploads", userID: "", isAnonymous: true, uri: "/v2/library/nginx/blobs/uploads/", method: http.MethodGet, wantPass: false},

		// Anonymous on private namespace.
		{name: "anon-private-read", userID: "", isAnonymous: true, uri: "/v2/private/nginx/manifests/v1", method: http.MethodGet, wantPass: false},

		// Anonymous on non-existent namespace.
		{name: "anon-newns", userID: "", isAnonymous: true, uri: "/v2/newns/nginx/manifests/v1", method: http.MethodGet, wantPass: false},

		// Authenticated non-member on private namespace.
		{name: "nonmember-private-read", userID: "400", isAnonymous: false, uri: "/v2/private/nginx/manifests/v1", method: http.MethodGet, wantPass: false},
		// Authenticated non-member on public namespace (public read allowed).
		{name: "nonmember-public-read", userID: "400", isAnonymous: false, uri: "/v2/library/nginx/manifests/v1", method: http.MethodGet, wantPass: true},

		// Reader: read ok, write denied.
		{name: "reader-read", userID: "100", isAnonymous: false, uri: "/v2/library/nginx/manifests/v1", method: http.MethodGet, wantPass: true},
		{name: "reader-write", userID: "100", isAnonymous: false, uri: "/v2/library/nginx/manifests/v1", method: http.MethodPut, wantPass: false},

		// Manager: read all + write image-layer, cannot write namespace meta.
		{name: "manager-write-manifest", userID: "200", isAnonymous: false, uri: "/v2/library/nginx/manifests/v1", method: http.MethodPut, wantPass: true},
		{name: "manager-write-blob", userID: "200", isAnonymous: false, uri: "/v2/library/nginx/blobs/sha256:abc", method: http.MethodPost, wantPass: true},
		{name: "manager-write-namespace-meta", userID: "200", isAnonymous: false, uri: "/api/v1/namespaces/1", method: http.MethodPut, wantPass: false},
		{name: "manager-read-private", userID: "200", isAnonymous: false, uri: "/v2/private/nginx/manifests/v1", method: http.MethodGet, wantPass: true},

		// Namespace admin: full access.
		{name: "nsadmin-write-namespace-meta", userID: "300", isAnonymous: false, uri: "/api/v1/namespaces/1", method: http.MethodPut, wantPass: true},
		{name: "nsadmin-write-manifest", userID: "300", isAnonymous: false, uri: "/v2/library/nginx/manifests/v1", method: http.MethodDelete, wantPass: true},

		// Authenticated on non-existent namespace: allowed (handler auto-creates).
		{name: "auth-newns", userID: "100", isAnonymous: false, uri: "/v2/newns/nginx/manifests/v1", method: http.MethodPut, wantPass: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Use a fresh authorizer per case to avoid cache cross-contamination.
			aFresh, err := NewAuthorizer(authorizer{
				Config:       testAuthzConfig(),
				RepoNs:       nsRepo,
				RepoNsMember: memberRepo,
			})
			require.NoError(t, err)
			passed, err := aFresh.Authorize(ctx, tc.userID, tc.isAnonymous, tc.uri, tc.method)
			require.NoError(t, err)
			assert.Equal(t, tc.wantPass, passed)
		})
	}
}

// ---------------------------------------------------------------------------
// TestAuthorizerCacheHit: verify cache is used on second call (no extra DB hit)
// ---------------------------------------------------------------------------

func TestAuthorizerCacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nsRepo := reponamespace.NewMockNamespaceRepository(ctrl)

	memberRepo := reponamespace.NewMockNamespaceMemberRepository(ctrl)

	publicNS := &models.Namespace{ID: "1", Name: "library", Visibility: enums.VisibilityPublic}

	// GetByName should be called exactly once (cached afterwards).
	nsRepo.EXPECT().GetByName(gomock.Any(), "library").Return(publicNS, nil).Times(1)
	// GetNamespaceMember should be called exactly once (cached afterwards).
	memberRepo.EXPECT().GetNamespaceMember(gomock.Any(), "1", "100").Return(&models.NamespaceMember{Role: enums.NamespaceRoleReader}, nil).Times(1)

	a, err := NewAuthorizer(authorizer{
		Config:       testAuthzConfig(),
		RepoNs:       nsRepo,
		RepoNsMember: memberRepo,
	})
	require.NoError(t, err)
	ctx := context.Background()

	// First call hits DB.
	passed, err := a.Authorize(ctx, "100", false, "/v2/library/nginx/manifests/v1", http.MethodGet)
	require.NoError(t, err)
	assert.True(t, passed)

	// Second call should use cache (no extra EXPECT calls).
	passed, err = a.Authorize(ctx, "100", false, "/v2/library/nginx/manifests/v1", http.MethodGet)
	require.NoError(t, err)
	assert.True(t, passed)
}
