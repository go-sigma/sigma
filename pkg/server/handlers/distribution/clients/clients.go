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

package clients

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/distribution/distribution/v3"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils/challenge"

	_ "github.com/distribution/distribution/v3/manifest/manifestlist"
	_ "github.com/distribution/distribution/v3/manifest/ocischema"
	_ "github.com/distribution/distribution/v3/manifest/schema2"
)

//go:generate mockgen -destination=mocks/clients.go -package=mocks github.com/go-sigma/sigma/pkg/server/handlers/distribution/clients Clients
//go:generate mockgen -destination=mocks/clients_factory.go -package=mocks github.com/go-sigma/sigma/pkg/server/handlers/distribution/clients ClientsFactory

// Clients is the interface of clients
type Clients interface {
	// AuthToken auth the clients
	AuthToken() error
	// DoRequest request the target with auth
	DoRequest(ctx context.Context, method, path string, headers http.Header, bodyReaders ...io.Reader) (int, http.Header, io.ReadCloser, error)
	// GetBlob get blob from target
	GetBlob(ctx context.Context, repository string, digest digest.Digest) (distribution.Descriptor, io.ReadCloser, error)
	// HeadBlob get blob metadata from target
	HeadBlob(ctx context.Context, repository string, digest digest.Digest) (distribution.Descriptor, error)
	// PutBlob upload blob to target
	PutBlob(ctx context.Context, repository string, digest digest.Digest, content io.Reader) error
	// GetManifest ...
	GetManifest(ctx context.Context, repository, reference string) (distribution.Manifest, distribution.Descriptor, error)
	// HeadManifest ...
	HeadManifest(ctx context.Context, repository, reference string) (bool, error)
}

// clients is the implementation of Clients
type clients struct {
	config   configs.Configuration
	cli      *resty.Client
	endpoint string
}

// ClientsFactory ...
type ClientsFactory interface {
	New(config configs.Configuration) (Clients, error)
}

type clientsFactory struct{}

// NewClientsFactory ...
func NewClientsFactory() ClientsFactory {
	return &clientsFactory{}
}

// New returns a new Clients
func (c clientsFactory) New(config configs.Configuration) (Clients, error) {
	client := resty.New()
	if !config.Proxy.TlsVerify {
		client = resty.NewWithClient(&http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, // nolint: gosec
		})
	}
	if config.Log.ProxyLevel.String() == "debug" {
		client.SetDebug(true)
	}
	client.SetHeader("User-Agent", consts.UserAgent)
	client.SetRetryCount(3) // TODO: set in config
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		return err != nil || r.StatusCode() >= http.StatusInternalServerError || r.StatusCode() == http.StatusTooManyRequests
	})

	clients := &clients{
		config:   config,
		cli:      client,
		endpoint: strings.TrimSuffix(config.Proxy.Endpoint, "/"),
	}

	err := clients.AuthToken()
	if err != nil {
		return nil, err
	}

	return clients, nil
}

// AuthToken returns the auth token
func (c *clients) AuthToken() error {
	cha, err := c.ping()
	if err != nil {
		return err
	}
	if cha == nil {
		return nil
	}
	if cha.Scheme == "basic" {
		if c.config.Proxy.Username == "" || c.config.Proxy.Password == "" {
			return fmt.Errorf("no username or password")
		}
		c.cli.SetBasicAuth(c.config.Proxy.Username, c.config.Proxy.Password)
		_, err = c.ping()
		if err != nil {
			return err
		}
		return nil
	}
	if cha.Scheme == "bearer" {
		realm := cha.Parameters["realm"]
		if realm == "" {
			return fmt.Errorf("no realm parameter")
		}
		token, err := c.token(*cha)
		if err != nil {
			return err
		}
		if token == "" {
			return fmt.Errorf("no token")
		}
		_, err = c.ping(map[string]string{"Authorization": fmt.Sprintf("Bearer %s", token)})
		if err != nil {
			return err
		}
		c.cli.SetHeader("Authorization", fmt.Sprintf("Bearer %s", token))
		return nil
	}
	return fmt.Errorf("unsupported schema: %s", cha.Scheme)
}

// ping returns the ping
func (c *clients) ping(headers ...map[string]string) (*challenge.Challenge, error) {
	req := c.cli.R()
	if len(headers) > 0 {
		req.SetHeaders(headers[0])
	}
	resp, err := req.Get(fmt.Sprintf("%s/v2/", c.endpoint))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil, nil
	}
	if resp.StatusCode() != http.StatusUnauthorized {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	authServerStr := resp.Header().Get(echo.HeaderWWWAuthenticate)
	if authServerStr == "" {
		return nil, fmt.Errorf("no auth server header")
	}
	challenges := challenge.ResponseChallenges(resp.RawResponse)
	if len(challenges) != 1 {
		return nil, fmt.Errorf("unexpected number of challenges: %d", len(challenges))
	}
	cha := challenges[0]
	return &cha, nil
}

// token returns the token
func (c *clients) token(cha challenge.Challenge) (string, error) {
	c.cli.Header.Del("Authorization") // clear the authorization header
	req := c.cli.R()
	req.SetHeader("Content-Type", "application/json")
	if cha.Parameters["service"] != "" {
		req.SetQueryParam("service", cha.Parameters["service"])
	}
	if cha.Parameters["scope"] != "" {
		req.SetQueryParam("scope", cha.Parameters["scope"])
	}
	if c.config.Proxy.Username != "" && c.config.Proxy.Password != "" {
		req.SetBasicAuth(c.config.Proxy.Username, c.config.Proxy.Password)
	}
	if c.config.Proxy.Token != "" {
		req.SetHeader("Authorization", "Bearer "+c.config.Proxy.Token)
	}
	resp, err := req.Get(cha.Parameters["realm"])
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	var body types.PostUserTokenResponse
	err = json.Unmarshal(resp.Body(), &body)
	if err != nil {
		return "", err
	}
	return body.Token, nil
}

// DoRequest returns the response
func (c *clients) DoRequest(ctx context.Context, method, path string, headers http.Header, bodyReaders ...io.Reader) (int, http.Header, io.ReadCloser, error) {
	req := c.cli.R()
	if headers != nil {
		for k, vals := range headers {
			for _, val := range vals {
				req.Header.Add(k, val)
			}
		}
	} else {
		req.SetHeader("Content-Type", "application/json")
		req.Header.Add(echo.HeaderAccept, "application/vnd.docker.distribution.manifest.v2+json")
		req.Header.Add(echo.HeaderAccept, "application/vnd.docker.distribution.manifest.list.v2+json")
		req.Header.Add(echo.HeaderAccept, "application/vnd.oci.image.index.v1+json")
		req.Header.Add(echo.HeaderAccept, "application/vnd.oci.image.manifest.v1+json")
		req.Header.Add(echo.HeaderAccept, "application/json")
		req.Header.Add(echo.HeaderAccept, "application/octet-stream")
	}
	req.SetDoNotParseResponse(true)
	if len(bodyReaders) != 0 {
		req.SetBody(bodyReaders[0])
	}
	req.SetContext(ctx)
	url := fmt.Sprintf("%s/%s", c.endpoint, strings.TrimPrefix(path, "/"))
	log.Info().Str("url", url).Str("method", method).Interface("header", req.Header).Msg("clients do request")
	if strings.HasPrefix(path, "http") {
		url = path
	}
	resp, err := req.Execute(method, url)
	if err != nil {
		return 0, nil, nil, err
	}
	if resp.StatusCode() == http.StatusUnauthorized {
		challenges := challenge.ResponseChallenges(resp.RawResponse)
		if len(challenges) != 1 {
			return 0, nil, nil, fmt.Errorf("unexpected number of challenges: %d", len(challenges))
		}
		cha := challenges[0]
		if cha.Scheme == "bearer" {
			token, err := c.token(cha)
			if err != nil {
				return 0, nil, nil, err
			}
			if token == "" {
				return 0, nil, nil, fmt.Errorf("no token")
			}
			c.cli.SetHeader("Authorization", fmt.Sprintf("Bearer %s", token))
			return c.DoRequest(ctx, method, path, headers)
		}
		return 0, nil, nil, fmt.Errorf("unsupported schema: %s", cha.Scheme)
	}

	return resp.StatusCode(), resp.RawResponse.Header, resp.RawBody(), nil
}
