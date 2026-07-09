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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"resty.dev/v3"

	"github.com/go-sigma/sigma/pkg/api/enums"
)

const defaultTimeout = 30 * time.Second

type options struct {
	server          string
	credentialsFile string
	timeout         time.Duration
	debug           bool
	client          *client
}

type client struct {
	baseURL  string
	username string
	password string
	cli      *resty.Client
}

type credentials struct {
	Server   string `json:"server"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func newRootCmd() *cobra.Command {
	opts := &options{}
	command := &cobra.Command{
		Use:          "sigma-cli",
		Short:        "Operate a remote sigma server through HTTP APIs",
		SilenceUsage: true,
		PersistentPreRunE: func(command *cobra.Command, _ []string) error {
			if command.Name() == "login" {
				return nil
			}
			return opts.initialize(command.Context(), command)
		},
	}

	command.PersistentFlags().StringVar(&opts.server, "server", envOrDefault("SIGMA_SERVER", "http://localhost:3000"), "sigma server address")
	command.PersistentFlags().StringVar(&opts.credentialsFile, "credentials-file", defaultCredentialsFile(), "credentials file")
	command.PersistentFlags().DurationVar(&opts.timeout, "timeout", defaultTimeout, "HTTP request timeout")
	command.PersistentFlags().BoolVar(&opts.debug, "debug", false, "enable HTTP request and response debug logging")

	command.AddCommand(
		newLoginCmd(opts),
		newNamespaceCmd(opts),
		newRepositoryCmd(opts),
		newTagCmd(opts),
		newArtifactCmd(opts),
		newMemberCmd(opts),
		newUserCmd(opts),
	)
	return command
}

func (opts *options) initialize(_ context.Context, command *cobra.Command) error {
	cred, err := loadCredentials(opts.credentialsFile)
	if err != nil {
		return err
	}

	server := strings.TrimRight(strings.TrimSpace(opts.server), "/")
	if !command.Root().PersistentFlags().Changed("server") && strings.TrimSpace(os.Getenv("SIGMA_SERVER")) == "" && cred.Server != "" {
		server = strings.TrimRight(cred.Server, "/")
	}
	username := strings.TrimSpace(cred.Username)
	password := cred.Password
	if server == "" {
		return fmt.Errorf("server is required")
	}
	if username == "" || password == "" {
		return fmt.Errorf("basic auth credentials are required; run `sigma-cli login --server <addr> -u <username> -p <password>` first")
	}
	opts.client = newClient(server, username, password, opts.timeout, opts.debug)
	return nil
}

func newLoginCmd(opts *options) *cobra.Command {
	var username string
	var password string
	command := &cobra.Command{
		Use:   "login",
		Short: "Login and save credentials locally",
		RunE: func(command *cobra.Command, _ []string) error {
			server := strings.TrimRight(strings.TrimSpace(opts.server), "/")
			if server == "" {
				return fmt.Errorf("server is required")
			}
			loginClient := newClient(server, "", "", opts.timeout, opts.debug)
			if err := loginClient.do(command.Context(), http.MethodPost, "/api/v1/users/login", nil, nil, nil, basicAuth{
				username: username,
				password: password,
			}); err != nil {
				return fmt.Errorf("login failed: %w", err)
			}
			cred := credentials{
				Server:   server,
				Username: username,
				Password: password,
			}
			if err := saveCredentials(opts.credentialsFile, cred); err != nil {
				return err
			}
			return printJSON(map[string]string{
				"status":           "logged_in",
				"credentials_file": opts.credentialsFile,
				"server":           server,
			})
		},
	}
	command.Flags().StringVarP(&username, "username", "u", os.Getenv("SIGMA_USERNAME"), "login username")
	command.Flags().StringVarP(&password, "password", "p", os.Getenv("SIGMA_PASSWORD"), "login password")
	_ = command.MarkFlagRequired("username")
	_ = command.MarkFlagRequired("password")
	return command
}

func newClient(baseURL, username, password string, timeout time.Duration, debug bool) *client {
	cli := resty.New().
		SetTimeout(timeout).
		SetDebug(debug)
	return &client{
		baseURL:  baseURL,
		username: username,
		password: password,
		cli:      cli,
	}
}

func (cli *client) do(ctx context.Context, method, endpoint string, query url.Values, body any, out any, auth basicAuth) error {
	requestURL := cli.baseURL + endpoint
	request := cli.cli.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json")
	if len(query) > 0 {
		request.SetQueryParamsFromValues(query)
	}
	if body != nil {
		request.SetHeader("Content-Type", "application/json").
			SetBody(body)
	}
	if auth.username != "" || auth.password != "" {
		request.SetBasicAuth(auth.username, auth.password)
	} else if cli.username != "" || cli.password != "" {
		request.SetBasicAuth(cli.username, cli.password)
	}

	response, err := request.Execute(method, requestURL)
	if err != nil {
		return err
	}
	if response.StatusCode() < http.StatusOK || response.StatusCode() >= http.StatusMultipleChoices {
		data := response.Bytes()
		return fmt.Errorf("%s %s: %s", method, endpoint, formatErrorBody(data))
	}
	data := response.Bytes()
	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

type basicAuth struct {
	username string
	password string
}

func formatErrorBody(data []byte) string {
	var pretty bytes.Buffer
	if json.Indent(&pretty, data, "", "  ") == nil {
		return strings.TrimSpace(pretty.String())
	}
	return strings.TrimSpace(string(data))
}

func printJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func defaultCredentialsFile() string {
	if value := strings.TrimSpace(os.Getenv("SIGMA_CREDENTIALS_FILE")); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ".sigma-cli-credentials.json"
	}
	return filepath.Join(home, ".sigma", "cli-credentials.json")
}

func loadCredentials(path string) (credentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return credentials{}, nil
		}
		return credentials{}, fmt.Errorf("read credentials file: %w", err)
	}
	var cred credentials
	if err := json.Unmarshal(data, &cred); err != nil {
		return credentials{}, fmt.Errorf("decode credentials file: %w", err)
	}
	return cred, nil
}

func saveCredentials(path string, cred credentials) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create credentials directory: %w", err)
	}
	// #nosec G117 -- The CLI intentionally persists Basic Auth credentials with 0600 permissions.
	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write credentials file: %w", err)
	}
	return nil
}

func commonListFlags(command *cobra.Command, name *string, page, limit *int, sort, method *string) {
	command.Flags().StringVar(name, "name", "", "filter by name")
	command.Flags().IntVar(page, "page", 1, "page number")
	command.Flags().IntVar(limit, "limit", 10, "page size")
	command.Flags().StringVar(sort, "sort", "", "sort field")
	command.Flags().StringVar(method, "method", "", "sort method: asc or desc")
}

func commonListQuery(name string, page, limit int, sort, method string) url.Values {
	query := url.Values{}
	if name != "" {
		query.Set("name", name)
	}
	if page > 0 {
		query.Set("page", fmt.Sprintf("%d", page))
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if sort != "" {
		query.Set("sort", sort)
	}
	if method != "" {
		query.Set("method", method)
	}
	return query
}

func optionalString(command *cobra.Command, name string) *string {
	if !command.Flags().Changed(name) {
		return nil
	}
	value, _ := command.Flags().GetString(name)
	return &value
}

func optionalInt64(command *cobra.Command, name string) *int64 {
	if !command.Flags().Changed(name) {
		return nil
	}
	value, _ := command.Flags().GetInt64(name)
	return &value
}

func optionalVisibility(command *cobra.Command) (*enums.Visibility, error) {
	if !command.Flags().Changed("visibility") {
		return nil, nil
	}
	value, _ := command.Flags().GetString("visibility")
	visibility, err := enums.ParseVisibility(value)
	if err != nil {
		return nil, err
	}
	return &visibility, nil
}

func endpoint(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
