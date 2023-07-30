package main

import (
	"fmt"
	"strings"

	"github.com/go-sigma/sigma/pkg/types/enums"
)

func (b *Builder) checker() error {
	if !b.ScmCredentialType.IsValid() {
		return fmt.Errorf("SCM_CREDENTIAL_TYPE should be one of 'ssh', 'token' or 'none', but got '%s'", b.ScmCredentialType.String())
	}
	if b.ScmCredentialType == enums.ScmCredentialTypeSsh && b.ScmSshKey == "" {
		return fmt.Errorf("SCM_SSH_KEY should be set, if SCM_CREDENTIAL_TYPE is 'ssh'")
	}
	if b.ScmCredentialType == enums.ScmCredentialTypeToken && b.ScmToken == "" {
		return fmt.Errorf("SCM_TOKEN should be set, if SCM_CREDENTIAL_TYPE is 'token'")
	}
	if b.ScmCredentialType == enums.ScmCredentialTypeToken && (!strings.HasPrefix(b.ScmRepository, "http://") && !strings.HasPrefix(b.ScmRepository, "https://")) {
		return fmt.Errorf("SCM_REPOSITORY should be started with 'http://' or 'https://', if SCM_CREDENTIAL_TYPE is 'token'")
	}
	if b.ScmCredentialType == enums.ScmCredentialTypeUsername && (b.ScmUsername == "" || b.ScmPassword == "") {
		return fmt.Errorf("SCM_USERNAME and SCM_PASSWORD should be set, if SCM_CREDENTIAL_TYPE is 'username'")
	}
	if !b.ScmProvider.IsValid() {
		return fmt.Errorf("SCM_PROVIDER should be one of 'github', 'gitlab' or 'bitbucket', but got '%s'", b.ScmProvider.String())
	}
	for _, platform := range b.BuildkitPlatforms {
		if !platform.IsValid() {
			return fmt.Errorf("BUILDKIT_PLATFORMS is invalid")
		}
	}
	return nil
}
