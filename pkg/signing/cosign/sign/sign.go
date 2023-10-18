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
	"github.com/go-sigma/sigma/pkg/handlers/distribution/clients"
	"github.com/go-sigma/sigma/pkg/signing/definition"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/hash"
	"github.com/go-sigma/sigma/pkg/utils/imagerefs"
)

// COSIGN_PASSWORD= ./cosign sign --tlog-upload=false --allow-http-registry --key cosign.key --registry-token eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzaWdtYSIsInN1YiI6IjIiLCJleHAiOjE2OTc2OTYxMDMsIm5iZiI6MTY5NzYwOTcwMywiaWF0IjoxNjk3NjA5NzAzLCJqdGkiOiIyYzIzNWNkOC0xYmJiLTRiNGYtODYyMi1iNjI2NDZiNTE3Y2EiLCJ1aWQiOiIyIn0.W_8F6Sh2Sj21lQm_ezwQHYvNTKSUdoYHru2hOrWqKtakx6s0arzW3NAyCo7ygSdjFpX18ZcSrsaOTDkccOYNmP-66ffYxLdV2PtfP8yGP-HkdvnO1cuTTUc5a25Xbn1nx2pJTDFpUKkehuBpSQ63pVnricRkQyxx-b7umg8i8MckMsSoX8eHM_IxKoFeKcMfmXqZb_RqDIzAXR2EBRTwPw16kZFD_ROtRTpeKihnXlUTcOz8t7ULoikBwNukwDG0nHXPwK4FQ64Rj-JyjGbRnSSJQTz0IhdnwFOW62_yHKic-4eK9MfWAlMvEXweV_uC1wbop-n5I84UtMPlQ0CT6Rz2iVHVkwPE4dZc9pLcBAv0UJJOnjPxQz5OUetBfIBgP9Kz7q7amA7qWf1fmq0Gf-GL7SkIqBp1ca45Fsqg2NOnNxOzxAPYUprQrQ9ObyooEJN85tTjQsIGmNMC853C9aPDN7pNBVWyba5I_xKFzaRNeO3BvdLjs8cq9K_SYgHA_yfZe-VBrfqc-jSx-t4HFxP429F9CBMKpz1JnH_ROg94Y5FAO9NgG9Jy0thZsMK6DjyiX_sG8sSeB4FULvmP7RV03ghMEu3nObDYeOxkKtUoNuGtL_BvVkOUF7zhvhQsjbeWOIOJVcJy-ljNk8AFnTZwpZQ0PAbWceRxFEEw-iw 10.3.198.80:3000/library/redis@sha256:b355b79a3117c4bfa2e8a05d5ed0496a19e542444d5ccf0328632091238a0136

type signing struct {
	Multiarch bool
	Http      bool
}

// New ...
func New(http, multiarch bool) definition.Signing {
	return &signing{
		Multiarch: multiarch,
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
	cmd.Args = append(cmd.Args, "--tlog-upload", "false")
	if s.Http {
		cmd.Args = append(cmd.Args, "--allow-http-registry")
	} else {
		cmd.Args = append(cmd.Args, "--allow-insecure-registry")
	}
	if s.Multiarch {
		cmd.Args = append(cmd.Args, "--recursive")
	}
	cmd.Args = append(cmd.Args, "--key", temp.Name())
	cmd.Args = append(cmd.Args, "--registry-token", token)
	cmd.Args = append(cmd.Args, imageRef)

	return nil
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
