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

// Package authz implements a database-backed authorization system that
// replaces casbin. Authorization decisions are made by querying the
// `namespaces` and `namespace_members` tables directly, with a short-TTL
// in-memory cache to reduce database load. This keeps policy state
// eventually consistent across multiple instances without requiring Redis.
package authz

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
)

// Authorizer decides whether a user may perform an HTTP method on a request URI.
//
// Admin/Root users are bypassed by the caller (middleware) before invoking the
// authorizer. The authorizer itself handles anonymous users, authenticated
// non-members and namespace members.
type Authorizer interface {
	// Authorize returns true when the request is allowed.
	// userID is empty for anonymous users; isAnonymous reflects the user role.
	Authorize(ctx context.Context, userID string, isAnonymous bool, requestURI, method string) (bool, error)
	// Namespace checks namespace-scoped permissions for a concrete namespace id.
	Namespace(ctx context.Context, user models.User, namespaceID string, auth enums.Auth) (bool, error)
	// NamespaceRole returns the user's role in a namespace.
	NamespaceRole(ctx context.Context, user models.User, namespaceID string) (*enums.NamespaceRole, error)
	// NamespacesRole returns the user's roles in multiple namespaces.
	NamespacesRole(ctx context.Context, user models.User, namespaceIDs []string) (map[string]*enums.NamespaceRole, error)
	// Repository checks permissions for a concrete repository id.
	Repository(ctx context.Context, user models.User, repositoryID string, auth enums.Auth) (bool, error)
	// Tag checks permissions for a concrete tag id.
	Tag(ctx context.Context, user models.User, tagID string, auth enums.Auth) (bool, error)
	// Artifact checks permissions for a concrete artifact id.
	Artifact(ctx context.Context, user models.User, artifactID string, auth enums.Auth) (bool, error)
}

//go:generate mockgen -destination=authorizer_mocks.go -package=authz github.com/go-sigma/sigma/pkg/authz Authorizer

// authorizer declares dependencies needed to construct an Authorizer.
type authorizer struct {
	dig.In `ignore-unexported:"true"`

	RepoNs         reponamespace.NamespaceRepository
	RepoNsMember   reponamespace.NamespaceMemberRepository
	RepoRepository reporegistry.RepositoryRepository
	RepoTag        reporegistry.TagRepository
	RepoArtifact   reporegistry.ArtifactRepository

	cacheRole *roleCache
	cacheNs   *namespaceCache
}

// NewAuthorizer constructs an Authorizer backed by the given repositories.
// It is intended to be provided via the dig container.
func NewAuthorizer(params authorizer) Authorizer {
	params.cacheRole = newRoleCache()
	params.cacheNs = newNamespaceCache()
	return &params
}

// Authorize implements the Authorizer interface.
func (a *authorizer) Authorize(ctx context.Context, userID string, isAnonymous bool, requestURI, method string) (bool, error) {
	desc, err := ResolveRequest(requestURI)
	if err != nil {
		return false, err
	}
	action := MethodToAction(method)

	// Global resources (base/catalog): read-only for everyone.
	if desc.Resource == ResBase || desc.Resource == ResCatalog {
		return CheckPublicAccess(desc.Resource, action, enums.VisibilityPublic), nil
	}

	// Resolve the namespace (by name for /v2/, by id for /api/).
	ns, found, err := a.lookupNamespace(ctx, desc)
	if err != nil {
		return false, err
	}
	if !found {
		// Namespace does not exist (typical on first push to a new repo).
		// Anonymous users are denied; authenticated users are allowed to
		// proceed so the handler can auto-create the namespace or return 404.
		return !isAnonymous, nil
	}

	visibility := ns.Visibility

	// Anonymous users: only public read access.
	if isAnonymous {
		return CheckPublicAccess(desc.Resource, action, visibility), nil
	}

	// Authenticated users: look up their role in the namespace.
	role, isMember, err := a.lookupRole(ctx, userID, ns.ID)
	if err != nil {
		return false, err
	}
	if !isMember {
		// Non-member: same as anonymous (public read only).
		return CheckPublicAccess(desc.Resource, action, visibility), nil
	}
	return CheckRole(role, desc.Resource, action), nil
}

// Namespace checks whether user has the requested namespace permission.
func (a *authorizer) Namespace(ctx context.Context, user models.User, namespaceID string, auth enums.Auth) (bool, error) {
	if user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleRoot {
		return true, nil
	}

	ns, err := a.RepoNs.Get(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.Join(err, fmt.Errorf("get namespace by id(%s) not found", namespaceID))
		}
		return false, errors.Join(err, fmt.Errorf("get namespace by id(%s) failed", namespaceID))
	}
	if ns.Visibility == enums.VisibilityPublic && auth == enums.AuthRead {
		return true, nil
	}

	role, isMember, err := a.lookupRole(ctx, user.ID, namespaceID)
	if err != nil {
		return false, err
	}
	if !isMember {
		return false, nil
	}
	return checkNamespaceRole(role, auth), nil
}

// NamespaceRole returns the user's namespace role.
func (a *authorizer) NamespaceRole(ctx context.Context, user models.User, namespaceID string) (*enums.NamespaceRole, error) {
	member, err := a.RepoNsMember.GetNamespaceMember(ctx, namespaceID, user.ID)
	if err != nil {
		return nil, err
	}
	return &member.Role, nil
}

// NamespacesRole returns the user's roles in the given namespaces.
func (a *authorizer) NamespacesRole(ctx context.Context, user models.User, namespaceIDs []string) (map[string]*enums.NamespaceRole, error) {
	members, err := a.RepoNsMember.GetNamespacesMember(ctx, namespaceIDs, user.ID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*enums.NamespaceRole, len(namespaceIDs))
	for _, member := range members {
		result[member.NamespaceID] = &member.Role
	}
	return result, nil
}

// Repository checks whether user has the requested repository permission.
func (a *authorizer) Repository(ctx context.Context, user models.User, repositoryID string, auth enums.Auth) (bool, error) {
	repo, err := a.RepoRepository.Get(ctx, repositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.Join(err, fmt.Errorf("get repository by id(%s) not found", repositoryID))
		}
		return false, errors.Join(err, fmt.Errorf("get repository by id(%s) failed", repositoryID))
	}
	return a.Namespace(ctx, user, repo.NamespaceID, auth)
}

// Tag checks whether user has the requested tag permission.
func (a *authorizer) Tag(ctx context.Context, user models.User, tagID string, auth enums.Auth) (bool, error) {
	tag, err := a.RepoTag.GetByID(ctx, tagID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.Join(err, fmt.Errorf("get tag by id(%s) not found", tagID))
		}
		return false, errors.Join(err, fmt.Errorf("get tag by id(%s) failed", tagID))
	}
	return a.Repository(ctx, user, tag.RepositoryID, auth)
}

// Artifact checks whether user has the requested artifact permission.
func (a *authorizer) Artifact(ctx context.Context, user models.User, artifactID string, auth enums.Auth) (bool, error) {
	artifact, err := a.RepoArtifact.Get(ctx, artifactID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.Join(err, fmt.Errorf("get artifact by id(%s) not found", artifactID))
		}
		return false, errors.Join(err, fmt.Errorf("get artifact by id(%s) failed", artifactID))
	}
	return a.Repository(ctx, user, artifact.RepositoryID, auth)
}

// lookupNamespace returns the namespace for the descriptor. When the namespace
// does not exist, found is false and ns is nil.
func (a *authorizer) lookupNamespace(ctx context.Context, desc ResourceDescriptor) (ns *models.Namespace, found bool, err error) {
	if desc.NamespaceID != "" {
		if cached, ok := a.cacheNs.getByID(desc.NamespaceID); ok {
			return cached, true, nil
		}
		ns, err = a.RepoNs.Get(ctx, desc.NamespaceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, nil
			}
			return nil, false, err
		}
		a.cacheNs.set(ns)
		return ns, true, nil
	}
	if desc.NamespaceName != "" {
		if cached, ok := a.cacheNs.getByName(desc.NamespaceName); ok {
			return cached, true, nil
		}
		ns, err = a.RepoNs.GetByName(ctx, desc.NamespaceName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, nil
			}
			return nil, false, err
		}
		a.cacheNs.set(ns)
		return ns, true, nil
	}
	return nil, false, errors.New("namespace not resolvable from descriptor")
}

// lookupRole returns the user's role in the namespace. isMember is false when
// the user has no membership record.
func (a *authorizer) lookupRole(ctx context.Context, userID, namespaceID string) (enums.NamespaceRole, bool, error) {
	if cached, ok := a.cacheRole.get(userID, namespaceID); ok {
		if cached == "" {
			return "", false, nil
		}
		return cached, true, nil
	}
	member, err := a.RepoNsMember.GetNamespaceMember(ctx, namespaceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			a.cacheRole.set(userID, namespaceID, "")
			return "", false, nil
		}
		return "", false, err
	}
	a.cacheRole.set(userID, namespaceID, member.Role)
	return member.Role, true, nil
}

// Compile-time interface check.
var _ Authorizer = (*authorizer)(nil)
