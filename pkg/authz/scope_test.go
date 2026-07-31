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
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// scopeFormatRe is the shape every defined scope constant must match:
// "resource:action" where resource is lowercase ascii letters, digits, or
// underscores (e.g. "oauth2", "namespace_member") and action is exactly
// "read" or "write".
var scopeFormatRe = regexp.MustCompile(`^[a-z0-9_]+:(read|write)$`)

func TestMake(t *testing.T) {
	cases := []struct {
		name     string
		resource ScopeResource
		action   ScopeAction
		want     Scope
	}{
		{name: "namespace read", resource: ScopeResourceNamespace, action: ScopeActionRead, want: ScopeNamespaceRead},
		{name: "namespace write", resource: ScopeResourceNamespace, action: ScopeActionWrite, want: ScopeNamespaceWrite},
		{name: "user read", resource: ScopeResourceUser, action: ScopeActionRead, want: ScopeUserRead},
		{name: "namespace member write", resource: ScopeResourceNamespaceMember, action: ScopeActionWrite, want: ScopeNamespaceMemberWrite},
		{name: "base read", resource: ScopeResourceBase, action: ScopeActionRead, want: ScopeBaseRead},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, Make(tc.resource, tc.action))
		})
	}
}

func TestSplit(t *testing.T) {
	for _, s := range AllScopes() {
		t.Run(string(s), func(t *testing.T) {
			resource, action, err := s.Split()
			require.NoError(t, err)
			require.NotEmpty(t, resource)
			require.NotEmpty(t, action)
			// Round-trip: Make(Split(s)) == s.
			require.Equal(t, s, Make(resource, action))
		})
	}
}

func TestSplitMalformed(t *testing.T) {
	cases := []struct {
		name  string
		scope Scope
	}{
		{name: "empty", scope: ""},
		{name: "no separator", scope: "noseparator"},
		{name: "resource only", scope: "namespace"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := tc.scope.Split()
			require.Error(t, err)
		})
	}
}

func TestParse(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, s := range AllScopes() {
			got, err := Parse(string(s))
			require.NoError(t, err)
			require.Equal(t, s, got)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		cases := []struct {
			name  string
			scope string
		}{
			{name: "empty", scope: ""},
			{name: "no separator", scope: "junk"},
			{name: "unknown action", scope: "namespace:foo"},
			{name: "unknown resource", scope: "unknown:read"},
			{name: "empty resource", scope: ":read"},
			{name: "empty action", scope: "namespace:"},
			{name: "extra separator", scope: "namespace:read:extra"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := Parse(tc.scope)
				require.Error(t, err)
			})
		}
	})
}

func TestValid(t *testing.T) {
	for _, s := range AllScopes() {
		require.True(t, Valid(s), "expected %q to be valid", s)
	}
	cases := []Scope{
		"",
		"namespace",
		"namespace:foo",
		"unknown:read",
		"Namespace:read", // case-sensitive
	}
	for _, s := range cases {
		require.False(t, Valid(s), "expected %q to be invalid", s)
	}
}

func TestAllScopes(t *testing.T) {
	scopes := AllScopes()

	// 20 resources x 2 actions = 40.
	require.Len(t, scopes, 40)

	// No duplicates.
	seen := make(map[Scope]struct{}, len(scopes))
	for _, s := range scopes {
		_, ok := seen[s]
		require.False(t, ok, "duplicate scope %q", s)
		seen[s] = struct{}{}
	}

	// Each distinct resource has exactly {read, write}.
	actionsByResource := make(map[ScopeResource]map[ScopeAction]struct{})
	for _, s := range scopes {
		resource, action, err := s.Split()
		require.NoError(t, err)
		if actionsByResource[resource] == nil {
			actionsByResource[resource] = make(map[ScopeAction]struct{})
		}
		actionsByResource[resource][action] = struct{}{}
	}
	require.Len(t, actionsByResource, 20, "expected 20 distinct resources")
	for resource, actions := range actionsByResource {
		require.Equal(t, map[ScopeAction]struct{}{
			ScopeActionRead:  {},
			ScopeActionWrite: {},
		}, actions, "resource %q must have exactly read+write", resource)
	}
}

func TestScopeFormat(t *testing.T) {
	for _, s := range AllScopes() {
		require.True(t, scopeFormatRe.MatchString(string(s)), "scope %q does not match format", s)
	}
}

func TestAllScopesImmutable(t *testing.T) {
	original := AllScopes()
	first := AllScopes()
	first[0] = Scope("tampered:read")
	// Mutating the returned slice must not affect subsequent calls.
	require.Equal(t, original, AllScopes())
}

func TestAllScopesSortedStable(t *testing.T) {
	// The order is a stable, curated grouping (distribution, namespaced,
	// global). Snapshot the first and last entry so reordering is intentional.
	scopes := AllScopes()
	require.NotEmpty(t, scopes)
	require.Equal(t, ScopeBaseRead, scopes[0])
	require.Equal(t, ScopeAnalyticsWrite, scopes[len(scopes)-1])
	// Sanity: action suffix alternates read then write within each pair.
	for i := 0; i+1 < len(scopes); i += 2 {
		require.True(t, strings.HasSuffix(string(scopes[i]), ":read"))
		require.True(t, strings.HasSuffix(string(scopes[i+1]), ":write"))
	}
	// slices.Sort is available as a reference comparator without mutating.
	sorted := slices.Clone(scopes)
	slices.Sort(sorted)
	require.NotEqual(t, sorted, scopes, "if this fails, allScopes is now sorted alphabetically; update the snapshot expectations")
}
