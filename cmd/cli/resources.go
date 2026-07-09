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

package main

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
)

func newNamespaceCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:     "namespace",
		Aliases: []string{"ns", "namespaces"},
		Short:   "Manage namespaces",
	}
	command.AddCommand(
		newNamespaceListCmd(opts),
		newNamespaceGetCmd(opts),
		newNamespaceCreateCmd(opts),
		newNamespaceUpdateCmd(opts),
		newNamespaceDeleteCmd(opts),
	)
	return command
}

func newNamespaceListCmd(opts *options) *cobra.Command {
	var name, sort, method string
	var page, limit int
	command := &cobra.Command{
		Use:   "list",
		Short: "List namespaces",
		RunE: func(command *cobra.Command, _ []string) error {
			var response api.CommonList
			query := commonListQuery(name, page, limit, sort, method)
			if err := opts.client.do(command.Context(), http.MethodGet, "/api/v1/namespaces/", query, nil, &response, basicAuth{}); err != nil {
				return err
			}
			return printJSON(response)
		},
	}
	commonListFlags(command, &name, &page, &limit, &sort, &method)
	return command
}

func newNamespaceGetCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "get <namespace_id>",
		Short: "Get namespace detail",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			var response api.NamespaceItem
			if err := opts.client.do(command.Context(), http.MethodGet, endpoint("/api/v1/namespaces/%s", args[0]), nil, nil, &response, basicAuth{}); err != nil {
				return err
			}
			return printJSON(response)
		},
	}
	return command
}

func newNamespaceCreateCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "create <name>",
		Short: "Create namespace",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			visibility, err := optionalVisibility(command)
			if err != nil {
				return err
			}
			request := api.PostNamespaceRequest{
				Name:            args[0],
				Description:     optionalString(command, "description"),
				SizeLimit:       optionalInt64(command, "size-limit"),
				RepositoryLimit: optionalInt64(command, "repository-limit"),
				TagLimit:        optionalInt64(command, "tag-limit"),
				Visibility:      visibility,
			}
			var response api.PostNamespaceResponse
			if err := opts.client.do(command.Context(), http.MethodPost, "/api/v1/namespaces/", nil, request, &response, basicAuth{}); err != nil {
				return err
			}
			return printJSON(response)
		},
	}
	addNamespaceWriteFlags(command)
	return command
}

func newNamespaceUpdateCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "update <namespace_id>",
		Short: "Update namespace",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			visibility, err := optionalVisibility(command)
			if err != nil {
				return err
			}
			request := api.UpdateNamespaceRequest{
				Description:     optionalString(command, "description"),
				Overview:        optionalString(command, "overview"),
				SizeLimit:       optionalInt64(command, "size-limit"),
				RepositoryLimit: optionalInt64(command, "repository-limit"),
				TagLimit:        optionalInt64(command, "tag-limit"),
				Visibility:      visibility,
			}
			if err := opts.client.do(command.Context(), http.MethodPut, endpoint("/api/v1/namespaces/%s", args[0]), nil, request, nil, basicAuth{}); err != nil {
				return err
			}
			return printJSON(map[string]string{"status": "updated"})
		},
	}
	addNamespaceWriteFlags(command)
	command.Flags().String("overview", "", "namespace overview")
	return command
}

func newNamespaceDeleteCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "delete <namespace_id>",
		Short: "Delete namespace",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := opts.client.do(command.Context(), http.MethodDelete, endpoint("/api/v1/namespaces/%s", args[0]), nil, nil, nil, basicAuth{}); err != nil {
				return err
			}
			return printJSON(map[string]string{"status": "deleted"})
		},
	}
	return command
}

func addNamespaceWriteFlags(command *cobra.Command) {
	command.Flags().String("description", "", "namespace description")
	command.Flags().String("visibility", "", "namespace visibility")
	command.Flags().Int64("size-limit", 0, "namespace size limit")
	command.Flags().Int64("repository-limit", 0, "namespace repository limit")
	command.Flags().Int64("tag-limit", 0, "namespace tag limit")
}

func newRepositoryCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:     "repo",
		Aliases: []string{"repository", "repositories"},
		Short:   "Manage repositories",
	}
	command.AddCommand(
		newRepositoryListCmd(opts),
		newRepositoryGetCmd(opts),
		newRepositoryCreateCmd(opts),
		newRepositoryUpdateCmd(opts),
		newRepositoryDeleteCmd(opts),
	)
	return command
}

func newRepositoryListCmd(opts *options) *cobra.Command {
	var name, sort, method string
	var page, limit int
	command := &cobra.Command{
		Use:   "list <namespace_id>",
		Short: "List repositories",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			var response api.CommonList
			query := commonListQuery(name, page, limit, sort, method)
			if err := opts.client.do(command.Context(), http.MethodGet, endpoint("/api/v1/namespaces/%s/repositories/", args[0]), query, nil, &response, basicAuth{}); err != nil {
				return err
			}
			return printJSON(response)
		},
	}
	commonListFlags(command, &name, &page, &limit, &sort, &method)
	return command
}

func newRepositoryGetCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "get <namespace_id> <repository_id>",
		Short: "Get repository detail",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			var response api.RepositoryItem
			if err := opts.client.do(command.Context(), http.MethodGet, endpoint("/api/v1/namespaces/%s/repositories/%s", args[0], args[1]), nil, nil, &response, basicAuth{}); err != nil {
				return err
			}
			return printJSON(response)
		},
	}
	return command
}

func newRepositoryCreateCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "create <namespace_id> <name>",
		Short: "Create repository",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			visibility, err := optionalVisibility(command)
			if err != nil {
				return err
			}
			request := api.CreateRepositoryRequest{
				Name:        args[1],
				Description: optionalString(command, "description"),
				Overview:    optionalString(command, "overview"),
				SizeLimit:   optionalInt64(command, "size-limit"),
				TagLimit:    optionalInt64(command, "tag-limit"),
				Visibility:  visibility,
			}
			var response api.CreateRepositoryResponse
			if err := opts.client.do(command.Context(), http.MethodPost, endpoint("/api/v1/namespaces/%s/repositories/", args[0]), nil, request, &response, basicAuth{}); err != nil {
				return err
			}
			return printJSON(response)
		},
	}
	addRepositoryWriteFlags(command)
	command.Flags().String("visibility", "", "repository visibility")
	return command
}

func newRepositoryUpdateCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "update <namespace_id> <repository_id>",
		Short: "Update repository",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			request := api.UpdateRepositoryRequest{
				Description: optionalString(command, "description"),
				Overview:    optionalString(command, "overview"),
				SizeLimit:   optionalInt64(command, "size-limit"),
				TagLimit:    optionalInt64(command, "tag-limit"),
			}
			if err := opts.client.do(command.Context(), http.MethodPut, endpoint("/api/v1/namespaces/%s/repositories/%s", args[0], args[1]), nil, request, nil, basicAuth{}); err != nil {
				return err
			}
			return printJSON(map[string]string{"status": "updated"})
		},
	}
	addRepositoryWriteFlags(command)
	return command
}

func newRepositoryDeleteCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "delete <namespace_id> <repository_id>",
		Short: "Delete repository",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			if err := opts.client.do(command.Context(), http.MethodDelete, endpoint("/api/v1/namespaces/%s/repositories/%s", args[0], args[1]), nil, nil, nil, basicAuth{}); err != nil {
				return err
			}
			return printJSON(map[string]string{"status": "deleted"})
		},
	}
	return command
}

func addRepositoryWriteFlags(command *cobra.Command) {
	command.Flags().String("description", "", "repository description")
	command.Flags().String("overview", "", "repository overview")
	command.Flags().Int64("size-limit", 0, "repository size limit")
	command.Flags().Int64("tag-limit", 0, "repository tag limit")
}

func newTagCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "tag",
		Short: "Manage repository tags",
	}
	command.AddCommand(
		newTagListCmd(opts),
		newTagGetCmd(opts),
		newTagDeleteCmd(opts),
	)
	return command
}

func newTagListCmd(opts *options) *cobra.Command {
	var name, sort, method string
	var page, limit int
	var artifactTypes []string
	command := &cobra.Command{
		Use:   "list <namespace_id> <repository_id>",
		Short: "List tags",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			var response api.CommonList
			query := commonListQuery(name, page, limit, sort, method)
			for _, artifactType := range artifactTypes {
				query.Add("type", artifactType)
			}
			if err := opts.client.do(command.Context(), http.MethodGet, endpoint("/api/v1/namespaces/%s/repositories/%s/tags/", args[0], args[1]), query, nil, &response, basicAuth{}); err != nil {
				return err
			}
			return printJSON(response)
		},
	}
	commonListFlags(command, &name, &page, &limit, &sort, &method)
	command.Flags().StringSliceVar(&artifactTypes, "type", nil, "artifact type filter")
	return command
}

func newTagGetCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "get <namespace_id> <repository_id> <tag_id>",
		Short: "Get tag detail",
		Args:  cobra.ExactArgs(3),
		RunE: func(command *cobra.Command, args []string) error {
			var response api.TagItem
			if err := opts.client.do(command.Context(), http.MethodGet, endpoint("/api/v1/namespaces/%s/repositories/%s/tags/%s", args[0], args[1], args[2]), nil, nil, &response, basicAuth{}); err != nil {
				return err
			}
			return printJSON(response)
		},
	}
	return command
}

func newTagDeleteCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "delete <namespace_id> <repository_id> <tag_id>",
		Short: "Delete tag",
		Args:  cobra.ExactArgs(3),
		RunE: func(command *cobra.Command, args []string) error {
			if err := opts.client.do(command.Context(), http.MethodDelete, endpoint("/api/v1/namespaces/%s/repositories/%s/tags/%s", args[0], args[1], args[2]), nil, nil, nil, basicAuth{}); err != nil {
				return err
			}
			return printJSON(map[string]string{"status": "deleted"})
		},
	}
	return command
}

func newArtifactCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:     "artifact",
		Aliases: []string{"image"},
		Short:   "Manage image artifacts by namespace name, repository name, and digest",
	}
	command.AddCommand(
		newArtifactListCmd(opts),
		newArtifactGetCmd(opts),
		newArtifactDeleteCmd(opts),
	)
	return command
}

func newArtifactListCmd(opts *options) *cobra.Command {
	var page, limit int
	command := &cobra.Command{
		Use:   "list <namespace_name> <repository_name>",
		Short: "List artifacts",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			var response api.CommonList
			query := url.Values{}
			query.Set("repository", args[1])
			if page > 0 {
				query.Set("page", endpoint("%d", page))
			}
			if limit > 0 {
				query.Set("limit", endpoint("%d", limit))
			}
			if err := opts.client.do(command.Context(), http.MethodGet, endpoint("/api/v1/namespaces/%s/artifacts/", args[0]), query, nil, &response, basicAuth{}); err != nil {
				return err
			}
			return printJSON(response)
		},
	}
	command.Flags().IntVar(&page, "page", 1, "page number")
	command.Flags().IntVar(&limit, "limit", 10, "page size")
	return command
}

func newArtifactGetCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "get <namespace_name> <repository_name> <digest>",
		Short: "Get artifact detail",
		Args:  cobra.ExactArgs(3),
		RunE: func(command *cobra.Command, args []string) error {
			var response api.ArtifactItem
			query := url.Values{"repository": []string{args[1]}}
			if err := opts.client.do(command.Context(), http.MethodGet, endpoint("/api/v1/namespaces/%s/artifacts/%s", args[0], args[2]), query, nil, &response, basicAuth{}); err != nil {
				return err
			}
			return printJSON(response)
		},
	}
	return command
}

func newArtifactDeleteCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "delete <namespace_name> <repository_name> <digest>",
		Short: "Delete artifact by digest",
		Args:  cobra.ExactArgs(3),
		RunE: func(command *cobra.Command, args []string) error {
			query := url.Values{"repository": []string{args[1]}}
			if err := opts.client.do(command.Context(), http.MethodDelete, endpoint("/api/v1/namespaces/%s/artifacts/%s", args[0], args[2]), query, nil, nil, basicAuth{}); err != nil {
				return err
			}
			return printJSON(map[string]string{"status": "deleted"})
		},
	}
	return command
}

func newMemberCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "member",
		Short: "Manage namespace members",
	}
	command.AddCommand(
		newMemberListCmd(opts),
		newMemberAddCmd(opts),
		newMemberUpdateCmd(opts),
		newMemberDeleteCmd(opts),
	)
	return command
}

func newMemberListCmd(opts *options) *cobra.Command {
	var name, sort, method string
	var page, limit int
	command := &cobra.Command{
		Use:   "list <namespace_id>",
		Short: "List namespace members",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			var response api.CommonList
			query := commonListQuery(name, page, limit, sort, method)
			if err := opts.client.do(command.Context(), http.MethodGet, endpoint("/api/v1/namespaces/%s/members/", args[0]), query, nil, &response, basicAuth{}); err != nil {
				return err
			}
			return printJSON(response)
		},
	}
	commonListFlags(command, &name, &page, &limit, &sort, &method)
	return command
}

func newMemberAddCmd(opts *options) *cobra.Command {
	var role string
	command := &cobra.Command{
		Use:   "add <namespace_id> <user_id>",
		Short: "Add namespace member",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			parsedRole, err := enums.ParseNamespaceRole(role)
			if err != nil {
				return err
			}
			userID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}
			request := map[string]any{
				"user_id": userID,
				"role":    parsedRole,
			}
			var response api.AddNamespaceMemberResponse
			if err := opts.client.do(command.Context(), http.MethodPost, endpoint("/api/v1/namespaces/%s/members/", args[0]), nil, request, &response, basicAuth{}); err != nil {
				return err
			}
			return printJSON(response)
		},
	}
	command.Flags().StringVar(&role, "role", enums.NamespaceRoleReader.String(), "member role")
	return command
}

func newMemberUpdateCmd(opts *options) *cobra.Command {
	var role string
	command := &cobra.Command{
		Use:   "update <namespace_id> <user_id>",
		Short: "Update namespace member role",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			parsedRole, err := enums.ParseNamespaceRole(role)
			if err != nil {
				return err
			}
			request := api.UpdateNamespaceMemberRequest{Role: parsedRole}
			if err := opts.client.do(command.Context(), http.MethodPut, endpoint("/api/v1/namespaces/%s/members/%s", args[0], args[1]), nil, request, nil, basicAuth{}); err != nil {
				return err
			}
			return printJSON(map[string]string{"status": "updated"})
		},
	}
	command.Flags().StringVar(&role, "role", enums.NamespaceRoleReader.String(), "member role")
	return command
}

func newMemberDeleteCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "delete <namespace_id> <user_id>",
		Short: "Delete namespace member",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			if err := opts.client.do(command.Context(), http.MethodDelete, endpoint("/api/v1/namespaces/%s/members/%s", args[0], args[1]), nil, nil, nil, basicAuth{}); err != nil {
				return err
			}
			return printJSON(map[string]string{"status": "deleted"})
		},
	}
	return command
}

func newUserCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
	}
	command.AddCommand(
		newUserListCmd(opts),
		newUserCreateCmd(opts),
		newUserUpdateCmd(opts),
	)
	return command
}

func newUserListCmd(opts *options) *cobra.Command {
	var name, sort, method string
	var page, limit int
	var withoutAdmin bool
	command := &cobra.Command{
		Use:   "list",
		Short: "List users",
		RunE: func(command *cobra.Command, _ []string) error {
			var response api.CommonList
			query := commonListQuery(name, page, limit, sort, method)
			if withoutAdmin {
				query.Set("without_admin", "true")
			}
			if err := opts.client.do(command.Context(), http.MethodGet, "/api/v1/users/", query, nil, &response, basicAuth{}); err != nil {
				return err
			}
			return printJSON(response)
		},
	}
	commonListFlags(command, &name, &page, &limit, &sort, &method)
	command.Flags().BoolVar(&withoutAdmin, "without-admin", false, "exclude admin users")
	return command
}

func newUserCreateCmd(opts *options) *cobra.Command {
	var username, password, email, role string
	command := &cobra.Command{
		Use:   "create",
		Short: "Create user",
		RunE: func(command *cobra.Command, _ []string) error {
			parsedRole, err := enums.ParseUserRole(role)
			if err != nil {
				return err
			}
			request := api.PostUserRequest{
				Username:       username,
				Password:       password,
				Email:          email,
				NamespaceLimit: optionalInt64(command, "namespace-limit"),
				Role:           parsedRole,
			}
			if err := opts.client.do(command.Context(), http.MethodPost, "/api/v1/users/", nil, request, nil, basicAuth{}); err != nil {
				return err
			}
			return printJSON(map[string]string{"status": "created"})
		},
	}
	command.Flags().StringVar(&username, "username", "", "new username")
	command.Flags().StringVar(&password, "password", "", "new user password")
	command.Flags().StringVar(&email, "email", "", "new user email")
	command.Flags().StringVar(&role, "role", enums.UserRoleUser.String(), "user role")
	command.Flags().Int64("namespace-limit", 0, "namespace limit")
	_ = command.MarkFlagRequired("username")
	_ = command.MarkFlagRequired("password")
	_ = command.MarkFlagRequired("email")
	return command
}

func newUserUpdateCmd(opts *options) *cobra.Command {
	command := &cobra.Command{
		Use:   "update <user_id>",
		Short: "Update user",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			request := api.PutUserRequest{
				Username:       optionalString(command, "username"),
				Password:       optionalString(command, "password"),
				Email:          optionalString(command, "email"),
				NamespaceLimit: optionalInt64(command, "namespace-limit"),
			}
			if command.Flags().Changed("role") {
				role, _ := command.Flags().GetString("role")
				parsedRole, err := enums.ParseUserRole(role)
				if err != nil {
					return err
				}
				request.Role = &parsedRole
			}
			if command.Flags().Changed("status") {
				status, _ := command.Flags().GetString("status")
				parsedStatus, err := enums.ParseUserStatus(status)
				if err != nil {
					return err
				}
				request.Status = &parsedStatus
			}
			if err := opts.client.do(command.Context(), http.MethodPut, endpoint("/api/v1/users/%s", args[0]), nil, request, nil, basicAuth{}); err != nil {
				return err
			}
			return printJSON(map[string]string{"status": "updated"})
		},
	}
	command.Flags().String("username", "", "username")
	command.Flags().String("password", "", "password")
	command.Flags().String("email", "", "email")
	command.Flags().String("status", "", "user status")
	command.Flags().String("role", "", "user role")
	command.Flags().Int64("namespace-limit", 0, "namespace limit")
	return command
}
