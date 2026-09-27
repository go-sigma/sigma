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

package validators

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
)

func TestValidateOCIPlatforms(t *testing.T) {
	type Test struct {
		Name      string
		Platforms []enums.OciPlatform `validate:"is_valid_oci_platforms"`
		Expected  bool
	}

	var tests = []Test{
		{"test-1", []enums.OciPlatform{"linux/amd64"}, true},
		{"test-2", []enums.OciPlatform{"linux/amd64", "linux/arm64"}, true},
		{"test-3", []enums.OciPlatform{"linux/amd64", "linux/arm641"}, false},
	}

	validator, err := newValidator()
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := validator.Struct(test)
			if !assert.Equal(t, test.Expected, err == nil) {
				t.Fatalf("expected %v but got %v", test.Expected, err == nil)
			}
		})
	}
}

func TestValidateRepository(t *testing.T) {
	type Test struct {
		Name     string `validate:"is_valid_repository"`
		Expected bool
	}

	var tests = []Test{
		{"my-repo", false},
		{"my/repo", true},
		{"my_repo", false},
		{"library/my_repo", true},
		{"%invalid:repo:latest$", false},
	}

	validator, err := newValidator()
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := validator.Struct(test)
			if !assert.Equal(t, test.Expected, err == nil) {
				t.Fatalf("expected %v but got %v", test.Expected, err == nil)
			}
		})
	}
}

func TestValidateDigest(t *testing.T) {
	type Test struct {
		Digest   string `validate:"is_valid_digest"`
		Expected bool
	}

	var tests = []Test{
		{"sha256:8699f120814ba2afc2e11630fcc75491a1ab95822cc842ed429cc10f71cc7d3c", true},
		{"sha256:1234", false},
		{"invalid-digest", false},
	}

	validator, err := newValidator()
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.Digest, func(t *testing.T) {
			err := validator.Struct(test)
			if !assert.Equal(t, test.Expected, err == nil) {
				t.Fatalf("expected %v but got %v", test.Expected, err == nil)
			}
		})
	}
}

func TestValidateNamespace(t *testing.T) {
	type Test struct {
		Namespace string `validate:"is_valid_namespace"`
		Expected  bool
	}

	var tests = []Test{
		{"my-namespace", true},
		{"-my-namespace", false},
		{"My-namespace", false},
		{"my-namespace-my-namespace-my-namespace-my-namespace", false},
	}

	validator, err := newValidator()
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.Namespace, func(t *testing.T) {
			err := validator.Struct(test)
			if !assert.Equal(t, test.Expected, err == nil) {
				t.Fatalf("expected %v but got %v", test.Expected, err == nil)
			}
		})
	}
}

func TestValidateTag(t *testing.T) {
	type Test struct {
		Tag      string `validate:"is_valid_tag"`
		Expected bool
	}

	var tests = []Test{
		{"valid.tag", true},
	}

	validator, err := newValidator()
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.Tag, func(t *testing.T) {
			err := validator.Struct(test)
			if !assert.Equal(t, test.Expected, err == nil) {
				t.Fatalf("expected %v but got %v", test.Expected, err == nil)
			}
		})
	}
}

func TestInitialize(t *testing.T) {
	require.NoError(t, Initialize())
}

func TestValidate(t *testing.T) {
	require.NoError(t, Initialize())

	type Test struct {
		Name     string `json:"name" validate:"required"`
		Expected bool
	}

	var tests = []Test{
		{"my-repo", true},
		{"", false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := binding.Validator.ValidateStruct(test)
			if !assert.Equal(t, test.Expected, err == nil) {
				t.Fatalf("expected %v but got %v", test.Expected, err == nil)
			}
		})
	}
}

func TestEngine(t *testing.T) {
	cv := &CustomValidator{validator: validator.New()}
	require.NotNil(t, cv.Engine())
}

// TestValidateFunctions exercises every registered validation function through the
// validator engine. ValidateNamespaceRole and ValidateBuilderSource are not
// reachable from newValidator (the former is shadowed by ValidateRetentionPattern
// and the latter is not registered at all), so they are registered locally.
func TestValidateFunctions(t *testing.T) {
	v, err := newValidator()
	require.NoError(t, err)
	require.NoError(t, v.RegisterValidation("test_namespace_role", ValidateNamespaceRole))
	require.NoError(t, v.RegisterValidation("test_builder_source", ValidateBuilderSource))

	tests := []struct {
		name  string
		tag   string
		value any
		want  bool
	}{
		{"namespace role ok", "test_namespace_role", string(enums.NamespaceRoleAdmin), true},
		{"namespace role bad", "test_namespace_role", "bad", false},

		{"retention pattern ok", "is_valid_retention_pattern", "a,b", true},
		{"retention pattern empty", "is_valid_retention_pattern", "", false},
		{"retention pattern invalid regexp", "is_valid_retention_pattern", "(", false},

		{"retention rule type ok", "is_valid_retention_rule_type", string(enums.RetentionRuleTypeDay), true},
		{"retention rule type bad", "is_valid_retention_rule_type", "bad", false},

		{"cron ok", "is_valid_cron_rule", "0 2 * * 6", true},
		{"cron bad", "is_valid_cron_rule", "bad", false},

		{"user role ok", "is_valid_user_role", string(enums.UserRoleRoot), true},
		{"user role bad", "is_valid_user_role", "bad", false},

		{"user status ok", "is_valid_user_status", string(enums.UserStatusActive), true},
		{"user status bad", "is_valid_user_status", "bad", false},

		{"password ok", "is_valid_password", "Admin@123", true},
		{"password weak", "is_valid_password", "123", false},

		{"email ok", "is_valid_email", "test@email.com", true},
		{"email bad", "is_valid_email", "not-an-email", false},

		{"username ok", "is_valid_username", "sigma_01", true},
		{"username bad", "is_valid_username", "si gma", false},

		{"namespace ok", "is_valid_namespace", "library", true},
		{"namespace bad", "is_valid_namespace", "Library", false},

		{"repository ok", "is_valid_repository", "library/nginx", true},
		{"repository bad", "is_valid_repository", "nginx", false},

		{"digest ok", "is_valid_digest", "sha256:8699f120814ba2afc2e11630fcc75491a1ab95822cc842ed429cc10f71cc7d3c", true},
		{"digest bad", "is_valid_digest", "bad", false},

		{"tag ok", "is_valid_tag", "v1.0.0", true},
		{"tag bad", "is_valid_tag", "bad/tag", false},

		{"visibility ok", "is_valid_visibility", string(enums.VisibilityPublic), true},
		{"visibility bad", "is_valid_visibility", "bad", false},

		{"provider ok", "is_valid_provider", string(enums.ProviderGithub), true},
		{"provider bad", "is_valid_provider", "bad", false},

		{"scm credential type ok", "is_valid_scm_credential_type", string(enums.ScmCredentialTypeSsh), true},
		{"scm credential type bad", "is_valid_scm_credential_type", "bad", false},

		{"builder source ok", "test_builder_source", string(enums.BuilderSourceDockerfile), true},
		{"builder source bad", "test_builder_source", "bad", false},

		{"oci platforms ok", "is_valid_oci_platforms", []string{"linux/amd64", "linux/arm64"}, true},
		{"oci platforms bad", "is_valid_oci_platforms", []string{"linux/amd64", "bad"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, v.Var(tt.value, tt.tag) == nil)
		})
	}
}
