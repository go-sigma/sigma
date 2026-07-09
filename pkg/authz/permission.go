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
	"net/http"

	"github.com/go-sigma/sigma/pkg/api/enums"
)

// Resource represents an authorization resource type.
type Resource string

const (
	// ResBase corresponds to /v2/ version probe.
	ResBase Resource = "base"
	// ResCatalog corresponds to /v2/_catalog.
	ResCatalog Resource = "catalog"
	// ResTags corresponds to /v2/{repo}/tags/list.
	ResTags Resource = "tags"
	// ResBlobs corresponds to /v2/{repo}/blobs/{digest}.
	ResBlobs Resource = "blobs"
	// ResManifests corresponds to /v2/{repo}/manifests/{ref}.
	ResManifests Resource = "manifests"
	// ResBlobUploads corresponds to /v2/{repo}/blobs/uploads[/{uuid}].
	ResBlobUploads Resource = "blob_uploads"
	// ResNamespace corresponds to /api/v1/namespaces/{id}.
	ResNamespace Resource = "namespace"
	// ResArtifacts corresponds to /api/v1/namespaces/{id}/artifacts[/{id}].
	ResArtifacts Resource = "artifacts"
	// ResRepositories corresponds to /api/v1/namespaces/{id}/repositories[/{id}].
	ResRepositories Resource = "repositories"
)

// Action represents an authorization action category.
type Action string

const (
	// ActionRead covers GET, HEAD.
	ActionRead Action = "read"
	// ActionWrite covers POST, PUT, DELETE, PATCH.
	ActionWrite Action = "write"
)

// MethodToAction maps an HTTP method to an Action.
func MethodToAction(method string) Action {
	switch method {
	case http.MethodGet, http.MethodHead:
		return ActionRead
	default:
		return ActionWrite
	}
}

// publicReadableResources are resources readable by anonymous / non-member users
// on a public namespace.
var publicReadableResources = map[Resource]bool{
	ResBlobs:        true,
	ResManifests:    true,
	ResTags:         true,
	ResNamespace:    true,
	ResArtifacts:    true,
	ResRepositories: true,
}

// CheckPublicAccess checks whether an anonymous or non-member user can access
// the given resource. Only read actions on a public namespace's readable
// resources (plus the global base/catalog resources) are allowed.
func CheckPublicAccess(resource Resource, action Action, visibility enums.Visibility) bool {
	// Global resources (no namespace scope): read only.
	if resource == ResBase || resource == ResCatalog {
		return action == ActionRead
	}
	if action != ActionRead {
		return false
	}
	if visibility != enums.VisibilityPublic {
		return false
	}
	return publicReadableResources[resource]
}

// CheckRole checks whether a namespace member with the given role can perform
// the action on the resource. visibility is ignored for members since their
// role grants access regardless of namespace visibility.
func CheckRole(role enums.NamespaceRole, resource Resource, action Action) bool {
	switch role {
	case enums.NamespaceRoleReader:
		return checkReader(resource, action)
	case enums.NamespaceRoleManager:
		return checkManager(resource, action)
	case enums.NamespaceRoleAdmin:
		return true
	}
	return false
}

func checkNamespaceRole(role enums.NamespaceRole, auth enums.Auth) bool {
	switch role {
	case enums.NamespaceRoleReader:
		return auth == enums.AuthRead
	case enums.NamespaceRoleManager:
		return auth == enums.AuthManage || auth == enums.AuthRead
	case enums.NamespaceRoleAdmin:
		return auth == enums.AuthAdmin || auth == enums.AuthManage || auth == enums.AuthRead
	}
	return false
}

// checkReader: read-only on namespace-scoped resources.
func checkReader(resource Resource, action Action) bool {
	if action != ActionRead {
		return false
	}
	switch resource {
	case ResBlobs, ResManifests, ResTags, ResNamespace, ResArtifacts, ResRepositories:
		return true
	}
	return false
}

// checkManager: read everything + write on image-layer resources (not namespace meta).
func checkManager(resource Resource, action Action) bool {
	if action == ActionRead {
		return true
	}
	switch resource {
	case ResBlobs, ResBlobUploads, ResManifests, ResTags, ResRepositories, ResArtifacts:
		return true
	}
	return false
}
