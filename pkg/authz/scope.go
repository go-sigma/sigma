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
	"fmt"
	"slices"
	"strings"
)

// Scope is a permission scope that can be bound to an API key. It is expressed
// as "resource:action", e.g. "namespace:read", "user:write".
//
// Scopes describe grantable resource areas for API keys and are intentionally
// separate from the URI-resolvable Resource type in permission.go: a scope
// covers a whole resource area (e.g. "user") rather than a single request path.
type Scope string

// ScopeResource is a top-level platform resource that scopes can be granted on.
type ScopeResource string

// ScopeAction is the operation granularity of a scope.
type ScopeAction string

const (
	// ScopeActionRead grants read-only operations (GET, HEAD).
	ScopeActionRead ScopeAction = "read"
	// ScopeActionWrite grants mutating operations (POST, PUT, DELETE, PATCH).
	ScopeActionWrite ScopeAction = "write"
)

// ScopeResource constants enumerate every platform resource that an API key
// scope can target. Names use the singular form so scopes read naturally
// (e.g. "repository:read"); this differs from the plural Res* constants in
// permission.go, which represent URI-resolvable collections.
const (
	// --- Distribution protocol resources ---

	// ScopeResourceBase corresponds to the /v2/ version probe.
	ScopeResourceBase ScopeResource = "base"
	// ScopeResourceCatalog corresponds to /v2/_catalog.
	ScopeResourceCatalog ScopeResource = "catalog"
	// ScopeResourceBlob corresponds to /v2/{repo}/blobs/{digest}.
	ScopeResourceBlob ScopeResource = "blob"
	// ScopeResourceManifest corresponds to /v2/{repo}/manifests/{ref}.
	ScopeResourceManifest ScopeResource = "manifest"
	// ScopeResourceUpload corresponds to /v2/{repo}/blobs/uploads.
	ScopeResourceUpload ScopeResource = "upload"
	// ScopeResourceTag corresponds to /v2/{repo}/tags/list.
	ScopeResourceTag ScopeResource = "tag"

	// --- Platform API (namespaced) resources ---

	// ScopeResourceNamespace corresponds to /api/v1/namespaces.
	ScopeResourceNamespace ScopeResource = "namespace"
	// ScopeResourceNamespaceMember corresponds to namespace members.
	ScopeResourceNamespaceMember ScopeResource = "namespace_member"
	// ScopeResourceRepository corresponds to /api/v1/namespaces/{id}/repositories.
	ScopeResourceRepository ScopeResource = "repository"
	// ScopeResourceArtifact corresponds to /api/v1/namespaces/{id}/artifacts.
	ScopeResourceArtifact ScopeResource = "artifact"
	// ScopeResourceWebhook corresponds to webhooks.
	ScopeResourceWebhook ScopeResource = "webhook"
	// ScopeResourceBuilder corresponds to builders and their runners.
	ScopeResourceBuilder ScopeResource = "builder"
	// ScopeResourceCoderepo corresponds to source code repositories.
	ScopeResourceCoderepo ScopeResource = "coderepo"

	// --- Platform API (global) resources ---

	// ScopeResourceUser corresponds to /api/v1/users.
	ScopeResourceUser ScopeResource = "user"
	// ScopeResourceSystem corresponds to /api/v1/systems.
	ScopeResourceSystem ScopeResource = "system"
	// ScopeResourceDaemon corresponds to /api/v1/daemons (gc jobs).
	ScopeResourceDaemon ScopeResource = "daemon"
	// ScopeResourceOauth2 corresponds to /api/v1/oauth2.
	ScopeResourceOauth2 ScopeResource = "oauth2"
	// ScopeResourceToken corresponds to /api/v1/tokens.
	ScopeResourceToken ScopeResource = "token"
	// ScopeResourceValidator corresponds to /api/v1/validators.
	ScopeResourceValidator ScopeResource = "validator"
	// ScopeResourceAnalytics corresponds to /api/v1/analytics.
	ScopeResourceAnalytics ScopeResource = "analytics"
)

// Scope constants enumerate every valid (resource, action) pair. Some resources
// (e.g. base, catalog) are read-only in practice; their :write scope is defined
// for uniformity and may simply never be granted by the API-key issuance layer.
const (
	// Distribution protocol
	ScopeBaseRead      Scope = "base:read"
	ScopeBaseWrite     Scope = "base:write"
	ScopeCatalogRead   Scope = "catalog:read"
	ScopeCatalogWrite  Scope = "catalog:write"
	ScopeBlobRead      Scope = "blob:read"
	ScopeBlobWrite     Scope = "blob:write"
	ScopeManifestRead  Scope = "manifest:read"
	ScopeManifestWrite Scope = "manifest:write"
	ScopeUploadRead    Scope = "upload:read"
	ScopeUploadWrite   Scope = "upload:write"
	ScopeTagRead       Scope = "tag:read"
	ScopeTagWrite      Scope = "tag:write"

	// Platform API (namespaced)
	ScopeNamespaceRead        Scope = "namespace:read"
	ScopeNamespaceWrite       Scope = "namespace:write"
	ScopeNamespaceMemberRead  Scope = "namespace_member:read"
	ScopeNamespaceMemberWrite Scope = "namespace_member:write"
	ScopeRepositoryRead       Scope = "repository:read"
	ScopeRepositoryWrite      Scope = "repository:write"
	ScopeArtifactRead         Scope = "artifact:read"
	ScopeArtifactWrite        Scope = "artifact:write"
	ScopeWebhookRead          Scope = "webhook:read"
	ScopeWebhookWrite         Scope = "webhook:write"
	ScopeBuilderRead          Scope = "builder:read"
	ScopeBuilderWrite         Scope = "builder:write"
	ScopeCoderepoRead         Scope = "coderepo:read"
	ScopeCoderepoWrite        Scope = "coderepo:write"

	// Platform API (global)
	ScopeUserRead       Scope = "user:read"
	ScopeUserWrite      Scope = "user:write"
	ScopeSystemRead     Scope = "system:read"
	ScopeSystemWrite    Scope = "system:write"
	ScopeDaemonRead     Scope = "daemon:read"
	ScopeDaemonWrite    Scope = "daemon:write"
	ScopeOauth2Read     Scope = "oauth2:read"
	ScopeOauth2Write    Scope = "oauth2:write"
	ScopeTokenRead      Scope = "token:read"
	ScopeTokenWrite     Scope = "token:write"
	ScopeValidatorRead  Scope = "validator:read"
	ScopeValidatorWrite Scope = "validator:write"
	ScopeAnalyticsRead  Scope = "analytics:read"
	ScopeAnalyticsWrite Scope = "analytics:write"
)

// allScopes is the complete, ordered list of valid scope constants. It is the
// single source of truth for enumeration and validation.
var allScopes = []Scope{
	// Distribution protocol
	ScopeBaseRead, ScopeBaseWrite,
	ScopeCatalogRead, ScopeCatalogWrite,
	ScopeBlobRead, ScopeBlobWrite,
	ScopeManifestRead, ScopeManifestWrite,
	ScopeUploadRead, ScopeUploadWrite,
	ScopeTagRead, ScopeTagWrite,

	// Platform API (namespaced)
	ScopeNamespaceRead, ScopeNamespaceWrite,
	ScopeNamespaceMemberRead, ScopeNamespaceMemberWrite,
	ScopeRepositoryRead, ScopeRepositoryWrite,
	ScopeArtifactRead, ScopeArtifactWrite,
	ScopeWebhookRead, ScopeWebhookWrite,
	ScopeBuilderRead, ScopeBuilderWrite,
	ScopeCoderepoRead, ScopeCoderepoWrite,

	// Platform API (global)
	ScopeUserRead, ScopeUserWrite,
	ScopeSystemRead, ScopeSystemWrite,
	ScopeDaemonRead, ScopeDaemonWrite,
	ScopeOauth2Read, ScopeOauth2Write,
	ScopeTokenRead, ScopeTokenWrite,
	ScopeValidatorRead, ScopeValidatorWrite,
	ScopeAnalyticsRead, ScopeAnalyticsWrite,
}

// validScopes is the lookup set built once from allScopes.
var validScopes = func() map[Scope]struct{} {
	m := make(map[Scope]struct{}, len(allScopes))
	for _, s := range allScopes {
		m[s] = struct{}{}
	}
	return m
}()

// Make constructs a Scope from a resource and an action.
func Make(resource ScopeResource, action ScopeAction) Scope {
	return Scope(string(resource) + ":" + string(action))
}

// Split decomposes a Scope into its resource and action parts. It returns an
// error only when the scope string is malformed (missing ":"); it does not
// validate that the parts are known. Use Valid for full validation.
func (s Scope) Split() (ScopeResource, ScopeAction, error) {
	before, after, ok := strings.Cut(string(s), ":")
	if !ok {
		return "", "", fmt.Errorf("malformed scope %q: missing ':'", s)
	}
	return ScopeResource(before), ScopeAction(after), nil
}

// Parse parses a raw string into a Scope, validating that it is one of the
// defined scope constants. Used when loading scopes from API-key storage.
func Parse(s string) (Scope, error) {
	scope := Scope(s)
	if !Valid(scope) {
		return "", fmt.Errorf("invalid scope %q", s)
	}
	return scope, nil
}

// Valid reports whether s is one of the defined scope constants.
func Valid(s Scope) bool {
	_, ok := validScopes[s]
	return ok
}

// AllScopes returns a copy of every defined scope constant, for enumeration
// and UI rendering. Callers may safely mutate the returned slice.
func AllScopes() []Scope {
	return slices.Clone(allScopes)
}
