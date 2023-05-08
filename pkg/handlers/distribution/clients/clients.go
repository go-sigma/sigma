// Copyright 2023 XImager
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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/distribution/distribution/v3/registry/client/auth/challenge"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/types"
)

// Clients is the interface of clients
type Clients interface {
	AuthToken() error
	DoRequest(method, path string) (int, http.Header, []byte, error)
}

// clients is the implementation of Clients
type clients struct {
	cli      *resty.Client
	endpoint string
}

// New returns a new Clients
func New() (Clients, error) {
	client := resty.New()
	if !viper.GetBool("proxy.tlsVerify") {
		client = resty.NewWithClient(&http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, // nolint: gosec
		})
	}
	if viper.GetString("log.level") == "debug" {
		client.SetDebug(true)
	}
	client.SetHeader("User-Agent", consts.UserAgent)
	client.SetRetryCount(3)
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		return err != nil || r.StatusCode() >= http.StatusInternalServerError || r.StatusCode() == http.StatusTooManyRequests
	})

	clients := &clients{
		cli:      client,
		endpoint: strings.TrimSuffix(viper.GetString("proxy.endpoint"), "/"),
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
	if cha.Scheme == "basic" {
		if viper.GetString("proxy.username") == "" || viper.GetString("proxy.password") == "" {
			return fmt.Errorf("no username or password")
		}
		c.cli.SetBasicAuth(viper.GetString("proxy.username"), viper.GetString("proxy.password"))
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
		token, err := c.token(cha)
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
func (c *clients) ping(headers ...map[string]string) (challenge.Challenge, error) {
	req := c.cli.R()
	if len(headers) > 0 {
		req.SetHeaders(headers[0])
	}
	resp, err := req.Get(fmt.Sprintf("%s/v2/", c.endpoint))
	if err != nil {
		return challenge.Challenge{}, err
	}
	if resp.StatusCode() == http.StatusOK {
		return challenge.Challenge{}, nil
	}
	if resp.StatusCode() != http.StatusUnauthorized {
		return challenge.Challenge{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
	authServerStr := resp.Header().Get("Www-Authenticate")
	if authServerStr == "" {
		return challenge.Challenge{}, fmt.Errorf("no auth server header")
	}
	challenges := challenge.ResponseChallenges(resp.RawResponse)
	if len(challenges) != 1 {
		return challenge.Challenge{}, fmt.Errorf("unexpected number of challenges: %d", len(challenges))
	}
	cha := challenges[0]
	return cha, nil
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
	if viper.GetString("proxy.username") != "" && viper.GetString("proxy.password") != "" {
		req.SetBasicAuth(viper.GetString("proxy.username"), viper.GetString("proxy.password"))
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
func (c *clients) DoRequest(method, path string) (int, http.Header, []byte, error) {
	req := c.cli.R()
	req.SetHeader("Content-Type", "application/json")
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
	req.Header.Add("Accept", "application/vnd.oci.image.index.v1+json")
	req.Header.Add("Accept", "application/vnd.oci.image.manifest.v1+json")
	req.Header.Add("Accept", "application/json")

	resp, err := req.Execute(method, fmt.Sprintf("%s/%s", c.endpoint, strings.TrimPrefix(path, "/")))
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
			return c.DoRequest(method, path)
		}
		return 0, nil, nil, fmt.Errorf("unsupported schema: %s", cha.Scheme)
	}

	return resp.StatusCode(), resp.RawResponse.Header, resp.Body(), nil
}
