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
	"net/url"
	"strings"
)

// ResourceDescriptor describes a parsed request resource.
type ResourceDescriptor struct {
	// NamespaceName is the namespace name resolved from a /v2/ request path
	// (the first segment of the repository path). Empty for /api/ requests.
	NamespaceName string
	// NamespaceID is the namespace id parsed from a /api/v1/namespaces/{id}
	// request path. Empty when not applicable.
	NamespaceID string
	// Resource is the resource kind.
	Resource Resource
	// Reference is the digest / ref / uuid when present; empty for list ops.
	Reference string
}

// ResolveRequest parses an HTTP request URI into a ResourceDescriptor.
//
// Supported patterns:
//   - /v2/                                  -> {Resource: base}
//   - /v2/_catalog                          -> {Resource: catalog}
//   - /v2/{repo}/tags/list                  -> {NamespaceName, Resource: tags}
//   - /v2/{repo}/blobs/uploads/             -> {NamespaceName, Resource: blob_uploads}
//   - /v2/{repo}/blobs/uploads/{uuid}       -> {NamespaceName, Resource: blob_uploads, Reference: uuid}
//   - /v2/{repo}/manifests/{ref}            -> {NamespaceName, Resource: manifests, Reference: ref}
//   - /v2/{repo}/blobs/{digest}             -> {NamespaceName, Resource: blobs, Reference: digest}
//   - /api/v1/namespaces/{id}               -> {NamespaceID, Resource: namespace}
//   - /api/v1/namespaces/{id}/artifacts[...]  -> {NamespaceID, Resource: artifacts}
//   - /api/v1/namespaces/{id}/repositories[...] -> {NamespaceID, Resource: repositories}
func ResolveRequest(requestURI string) (ResourceDescriptor, error) {
	// Strip query string.
	rawURL := requestURI
	if i := strings.IndexByte(rawURL, '?'); i >= 0 {
		rawURL = rawURL[:i]
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ResourceDescriptor{}, fmt.Errorf("parse request uri: %w", err)
	}
	path := parsed.Path

	switch {
	case path == "/v2/" || path == "/v2":
		return ResourceDescriptor{Resource: ResBase}, nil
	case path == "/v2/_catalog":
		return ResourceDescriptor{Resource: ResCatalog}, nil
	case strings.HasPrefix(path, "/v2/"):
		return resolveV2(path)
	case strings.HasPrefix(path, "/api/v1/namespaces/"):
		return resolveAPI(path)
	}
	return ResourceDescriptor{}, fmt.Errorf("unrecognized request path: %s", path)
}

// resolveV2 parses /v2/{repo}/... distribution paths.
func resolveV2(path string) (ResourceDescriptor, error) {
	rest := strings.TrimPrefix(path, "/v2/")

	switch {
	case strings.HasSuffix(rest, "/tags/list"):
		repo := strings.TrimSuffix(rest, "/tags/list")
		return ResourceDescriptor{NamespaceName: firstSegment(repo), Resource: ResTags}, nil
	case strings.HasSuffix(rest, "/blobs/uploads/"):
		repo := strings.TrimSuffix(rest, "/blobs/uploads/")
		return ResourceDescriptor{NamespaceName: firstSegment(repo), Resource: ResBlobUploads}, nil
	case strings.HasSuffix(rest, "/blobs/uploads"):
		repo := strings.TrimSuffix(rest, "/blobs/uploads")
		return ResourceDescriptor{NamespaceName: firstSegment(repo), Resource: ResBlobUploads}, nil
	}

	// Split into [repo...suffix, lastSegment] using the last "/".
	idx := strings.LastIndex(rest, "/")
	if idx < 0 {
		return ResourceDescriptor{}, fmt.Errorf("unrecognized /v2 path: %s", path)
	}
	ref := rest[idx+1:]
	prefix := rest[:idx] // repo + suffix marker

	switch {
	case strings.HasSuffix(prefix, "/manifests"):
		repo := strings.TrimSuffix(prefix, "/manifests")
		return ResourceDescriptor{NamespaceName: firstSegment(repo), Resource: ResManifests, Reference: ref}, nil
	case strings.HasSuffix(prefix, "/blobs"):
		repo := strings.TrimSuffix(prefix, "/blobs")
		return ResourceDescriptor{NamespaceName: firstSegment(repo), Resource: ResBlobs, Reference: ref}, nil
	case strings.HasSuffix(prefix, "/blobs/uploads"):
		repo := strings.TrimSuffix(prefix, "/blobs/uploads")
		return ResourceDescriptor{NamespaceName: firstSegment(repo), Resource: ResBlobUploads, Reference: ref}, nil
	}
	return ResourceDescriptor{}, fmt.Errorf("unrecognized /v2 path: %s", path)
}

// resolveAPI parses /api/v1/namespaces/{id}... management paths.
func resolveAPI(path string) (ResourceDescriptor, error) {
	rest := strings.TrimPrefix(path, "/api/v1/namespaces/")
	// rest = "{id}" or "{id}/artifacts..." or "{id}/repositories..."
	segments := strings.SplitN(rest, "/", 2)
	if len(segments) == 0 || segments[0] == "" {
		return ResourceDescriptor{}, fmt.Errorf("missing namespace id in path: %s", path)
	}
	desc := ResourceDescriptor{NamespaceID: segments[0], Resource: ResNamespace}
	if len(segments) == 2 {
		sub := segments[1]
		switch {
		case strings.HasPrefix(sub, "artifacts"):
			desc.Resource = ResArtifacts
		case strings.HasPrefix(sub, "repositories"):
			desc.Resource = ResRepositories
		}
	}
	return desc, nil
}

// firstSegment returns the first "/"-separated segment of a repo path, which is
// the namespace name. e.g. "library/nginx" -> "library"; "nginx" -> "nginx".
func firstSegment(repo string) string {
	if before, _, ok := strings.Cut(repo, "/"); ok {
		return before
	}
	return repo
}
