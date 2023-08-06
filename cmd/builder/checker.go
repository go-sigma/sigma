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

package main

import (
	"fmt"
	"strings"

	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/crypt"
)

func (b *Builder) checker() error {
	var err error

	if !b.ScmCredentialType.IsValid() {
		return fmt.Errorf("SCM_CREDENTIAL_TYPE should be one of 'ssh', 'token' or 'none', but got '%s'", b.ScmCredentialType.String())
	}

	if b.ScmCredentialType == enums.ScmCredentialTypeSsh && b.ScmSshKey == "" {
		return fmt.Errorf("SCM_SSH_KEY should be set, if SCM_CREDENTIAL_TYPE is 'ssh'")
	}
	if b.ScmSshKey != "" {
		b.ScmSshKey, err = crypt.Decrypt(fmt.Sprintf("%d-%d", b.ID, b.RunnerID), b.ScmSshKey)
		if err != nil {
			return fmt.Errorf("Decrypt ssh key failed: %v", err)
		}
	}

	if b.ScmCredentialType == enums.ScmCredentialTypeToken && b.ScmToken == "" {
		return fmt.Errorf("SCM_TOKEN should be set, if SCM_CREDENTIAL_TYPE is 'token'")
	}
	if b.ScmToken != "" {
		b.ScmToken, err = crypt.Decrypt(fmt.Sprintf("%d-%d", b.ID, b.RunnerID), b.ScmToken)
		if err != nil {
			return fmt.Errorf("Decrypt scm token failed: %v", err)
		}
	}

	if b.ScmCredentialType == enums.ScmCredentialTypeToken && (!strings.HasPrefix(b.ScmRepository, "http://") && !strings.HasPrefix(b.ScmRepository, "https://")) {
		return fmt.Errorf("SCM_REPOSITORY should be started with 'http://' or 'https://', if SCM_CREDENTIAL_TYPE is 'token'")
	}
	if b.ScmCredentialType == enums.ScmCredentialTypeUsername && (b.ScmUsername == "" || b.ScmPassword == "") {
		return fmt.Errorf("SCM_USERNAME and SCM_PASSWORD should be set, if SCM_CREDENTIAL_TYPE is 'username'")
	}
	if b.ScmPassword != "" {
		b.ScmPassword, err = crypt.Decrypt(fmt.Sprintf("%d-%d", b.ID, b.RunnerID), b.ScmPassword)
		if err != nil {
			return fmt.Errorf("Decrypt scm password failed: %v", err)
		}
	}

	if !b.ScmProvider.IsValid() {
		return fmt.Errorf("SCM_PROVIDER should be one of 'github', 'gitlab' or 'bitbucket', but got '%s'", b.ScmProvider.String())
	}
	for _, platform := range b.BuildkitPlatforms {
		if !platform.IsValid() {
			return fmt.Errorf("BUILDKIT_PLATFORMS is invalid")
		}
	}

	if len(b.OciRegistryDomain) != len(b.OciRegistryUsername) || len(b.OciRegistryDomain) != len(b.OciRegistryPassword) {
		return fmt.Errorf("OCI_REGISTRY_DOMAIN length should equal OCI_REGISTRY_USERNAME and OCI_REGISTRY_PASSWORD")
	}

	for index, password := range b.OciRegistryPassword {
		b.OciRegistryPassword[index], err = crypt.Decrypt(fmt.Sprintf("%d-%d", b.ID, b.RunnerID), password)
		if err != nil {
			return fmt.Errorf("Decrypt oci registry password failed: %v", err)
		}
	}

	return nil
}
