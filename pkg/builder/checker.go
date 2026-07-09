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

package builder

import (
	"fmt"
	"strings"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/utils/crypt"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// checker 校验并解密构建参数
func (f *BuildFlow) checker() error {
	var err error

	if f.Authorization != "" {
		authorization, err := crypt.Decrypt(fmt.Sprintf("%s-%s", f.BuilderID, f.RunnerID), f.Authorization)
		if err != nil {
			return err
		}
		f.Authorization = authorization
	}

	if f.Source != enums.BuilderSourceDockerfile && (f.ScmCredentialType == nil || !f.ScmCredentialType.IsValid()) {
		return fmt.Errorf("SCM_CREDENTIAL_TYPE should be one of 'ssh', 'token' or 'none', but got '%s'", f.ScmCredentialType.String())
	}

	if f.ScmCredentialType != nil && ptr.To(f.ScmCredentialType) == enums.ScmCredentialTypeSsh && ptr.To(f.ScmSshKey) == "" {
		return fmt.Errorf("SCM_SSH_KEY should be set, if SCM_CREDENTIAL_TYPE is 'ssh'")
	}
	if ptr.To(f.ScmSshKey) != "" {
		scmSshKey, err := crypt.Decrypt(fmt.Sprintf("%s-%s", f.BuilderID, f.RunnerID), ptr.To(f.ScmSshKey))
		if err != nil {
			return fmt.Errorf("decrypt ssh key failed: %v", err)
		}
		f.ScmSshKey = new(scmSshKey)
	}

	if f.ScmCredentialType != nil && ptr.To(f.ScmCredentialType) == enums.ScmCredentialTypeToken && ptr.To(f.ScmToken) == "" {
		return fmt.Errorf("env SCM_TOKEN should be set, if SCM_CREDENTIAL_TYPE is 'token'")
	}
	if ptr.To(f.ScmToken) != "" {
		scmToken, err := crypt.Decrypt(fmt.Sprintf("%s-%s", f.BuilderID, f.RunnerID), ptr.To(f.ScmToken))
		if err != nil {
			return fmt.Errorf("decrypt scm token failed: %v", err)
		}
		f.ScmToken = new(scmToken)
	}

	if f.ScmCredentialType != nil && ptr.To(f.ScmCredentialType) == enums.ScmCredentialTypeToken &&
		(!strings.HasPrefix(ptr.To(f.ScmRepository), "http://") && !strings.HasPrefix(ptr.To(f.ScmRepository), "https://")) {
		return fmt.Errorf("env SCM_REPOSITORY should be started with 'http://' or 'https://', if SCM_CREDENTIAL_TYPE is 'token'")
	}
	if f.ScmCredentialType != nil && ptr.To(f.ScmCredentialType) == enums.ScmCredentialTypeUsername && (ptr.To(f.ScmUsername) == "" || ptr.To(f.ScmPassword) == "") {
		return fmt.Errorf("env SCM_USERNAME and SCM_PASSWORD should be set, if SCM_CREDENTIAL_TYPE is 'username'")
	}
	if ptr.To(f.ScmPassword) != "" {
		scmPassword, err := crypt.Decrypt(fmt.Sprintf("%s-%s", f.BuilderID, f.RunnerID), ptr.To(f.ScmPassword))
		if err != nil {
			return fmt.Errorf("decrypt scm password failed: %v", err)
		}
		f.ScmPassword = new(scmPassword)
	}

	if f.Source != enums.BuilderSourceDockerfile && (f.ScmProvider == nil || !f.ScmProvider.IsValid()) {
		return fmt.Errorf("env SCM_PROVIDER should be one of 'github', 'gitlab' or 'bitbucket', but got '%s'", f.ScmProvider.String())
	}
	for _, platform := range f.BuildkitPlatforms {
		if !platform.IsValid() {
			return fmt.Errorf("env BUILDKIT_PLATFORMS is invalid")
		}
	}

	if len(f.OciRegistryDomain) != len(f.OciRegistryUsername) || len(f.OciRegistryDomain) != len(f.OciRegistryPassword) {
		return fmt.Errorf("env OCI_REGISTRY_DOMAIN length should equal OCI_REGISTRY_USERNAME and OCI_REGISTRY_PASSWORD")
	}

	for index, password := range f.OciRegistryPassword {
		f.OciRegistryPassword[index], err = crypt.Decrypt(fmt.Sprintf("%s-%s", f.BuilderID, f.RunnerID), password)
		if err != nil {
			return fmt.Errorf("decrypt oci registry password failed: %v", err)
		}
	}

	signingPrivateKey, err := crypt.Decrypt(fmt.Sprintf("%s-%s", f.BuilderID, f.RunnerID), f.SigningPrivateKey)
	if err != nil {
		return fmt.Errorf("decrypt signing private key failed: %v", err)
	}
	f.SigningPrivateKey = signingPrivateKey

	return nil
}
