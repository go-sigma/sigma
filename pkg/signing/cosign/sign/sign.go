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

package sign

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/aquasecurity/trivy/pkg/digest"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/handlers/distribution/clients"
	"github.com/go-sigma/sigma/pkg/signing/definition"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/hash"
	"github.com/go-sigma/sigma/pkg/utils/imagerefs"
)

type signing struct {
	MultiArch bool
	Http      bool
}

// New ...
func New(http, multiArch bool) definition.Signing {
	return &signing{
		MultiArch: multiArch,
		Http:      http,
	}
}

// Sign ...
func (s *signing) Sign(ctx context.Context, token, priKey, ref string) error {
	imageRef, err := s.GetImageRef(ctx, token, ref)
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp("", consts.AppName)
	if err != nil {
		return err
	}
	defer func() {
		err := os.Remove(temp.Name())
		if err != nil {
			log.Error().Err(err).Msg("Remove temp file failed")
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
	log.Info().Str("command", cmd.String()).Strs("env", cmd.Env).Msg("Signing image")

	return cmd.Run()
}

// GetDigest ...
func (s *signing) GetImageRef(ctx context.Context, token, ref string) (string, error) {
	domain, _, repo, tag, err := imagerefs.Parse(ref)
	if err != nil {
		return "", err
	}
	if s.Http {
		domain = fmt.Sprintf("http://%s", domain)
	} else {
		domain = fmt.Sprintf("https://%s", domain)
	}
	clientsFactory := clients.NewClientsFactory()
	client, err := clientsFactory.New(configs.Configuration{
		Proxy: configs.ConfigurationProxy{
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
	d := digest.NewDigestFromString(digest.SHA256, hash.MustString(string(manifestBytes)))
	return fmt.Sprintf("%s/%s@%s", utils.TrimHTTP(domain), repo, d.String()), nil
}
