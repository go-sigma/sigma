// Copyright 2025 sigma
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

package cosign

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/distribution/clients"
	"github.com/go-sigma/sigma/pkg/distribution/reference"
	"github.com/go-sigma/sigma/pkg/distribution/signing"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/hash"
)

func init() {
	utils.PanicIf(signing.RegisterSigning(enums.SigningTypeCosign, &factory{}))
}

type client struct {
	MultiArch bool
	Http      bool
}

type factory struct{}

var _ signing.SigningFactory = factory{}

// New returns a cosign signing client; http selects plain-HTTP registries instead of TLS, and multiArch enables recursive signing of a multi-architecture index.
func (factory) New(http, multiArch bool) (signing.Signing, error) {
	return &client{
		MultiArch: multiArch,
		Http:      http,
	}, nil
}

// Sign signs the image identified by ref with the PEM private key priKey, resolving ref to a digest-pinned reference and shelling out to the cosign executable with signature transparency logging disabled. The key is written to a temporary file that is removed afterwards, and the command's exit error is returned.
func (s *client) Sign(ctx context.Context, token, priKey, ref string) error {
	imageRef, err := s.GetImageRef(ctx, token, ref)
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp("", consts.AppName)
	if err != nil {
		return err
	}
	defer func() {
		e := os.Remove(temp.Name())
		if e != nil {
			slog.Error("remove temp file failed", "err", e)
		}
	}()
	_, err = temp.WriteString(priKey)
	if err != nil {
		return err
	}

	cmd := exec.Command("cosign", "sign")
	cmd.Args = append(cmd.Args, "--tlog-upload=false")
	if s.Http {
		cmd.Args = append(cmd.Args, "--allow-http-registry")
	} else {
		cmd.Args = append(cmd.Args, "--allow-insecure-registry")
	}
	if s.MultiArch {
		cmd.Args = append(cmd.Args, "--recursive")
	}
	cmd.Args = append(cmd.Args, "--key", temp.Name())
	cmd.Args = append(cmd.Args, "--registry-token", token)
	cmd.Args = append(cmd.Args, "--registry-referrers-mode", "oci-1-1")
	cmd.Args = append(cmd.Args, imageRef)
	cmd.Env = append(cmd.Env, "COSIGN_PASSWORD=")
	cmd.Env = append(cmd.Env, "COSIGN_EXPERIMENTAL=1")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	slog.Info("signing image", "command", cmd.String(), "env", cmd.Env)

	return cmd.Run()
}

// GetImageRef resolves ref against its registry and returns a digest-pinned reference of the form <domain>/<repository>@sha256:<digest>, where the digest is computed over the raw manifest payload. It honors the client's Http flag by contacting the registry over http instead of https.
func (s *client) GetImageRef(ctx context.Context, token, ref string) (string, error) {
	domain, _, repo, tag, err := reference.Parse(ref)
	if err != nil {
		return "", err
	}
	if s.Http {
		domain = fmt.Sprintf("http://%s", domain)
	} else {
		domain = fmt.Sprintf("https://%s", domain)
	}
	clientsFactory := clients.NewClientsFactory()
	client, err := clientsFactory.New(&config.Configuration{
		Proxy: config.ConfigurationProxy{
			Endpoint:  domain,
			TlsVerify: !s.Http,
			Token:     token,
		},
	})
	if err != nil {
		return "", err
	}
	manifest, _, err := client.GetManifest(ctx, repo, tag)
	if err != nil {
		return "", err
	}
	_, manifestBytes, err := manifest.Payload()
	if err != nil {
		return "", err
	}
	d := fmt.Sprintf("sha256:%s", hash.MustString(string(manifestBytes)))
	return fmt.Sprintf("%s/%s@%s", utils.TrimHTTP(domain), repo, d), nil
}
