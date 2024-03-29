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

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/types/enums"
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

	validate := validator.New()
	register(validate)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := validate.Struct(test)
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

	validate := validator.New()
	register(validate)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := validate.Struct(test)

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

	validate := validator.New()
	register(validate)

	for _, test := range tests {
		t.Run(test.Digest, func(t *testing.T) {
			err := validate.Struct(test)

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

	validate := validator.New()
	register(validate)

	for _, test := range tests {
		t.Run(test.Namespace, func(t *testing.T) {
			err := validate.Struct(test)

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

	validate := validator.New()
	register(validate)

	for _, test := range tests {
		t.Run(test.Tag, func(t *testing.T) {
			err := validate.Struct(test)

			if !assert.Equal(t, test.Expected, err == nil) {
				t.Fatalf("expected %v but got %v", test.Expected, err == nil)
			}
		})
	}
}

func TestInitialize(t *testing.T) {
	e := echo.New()
	Initialize(e)
}

func TestValidate(t *testing.T) {
	e := echo.New()
	Initialize(e)

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
			err := e.Validator.Validate(test)

			if !assert.Equal(t, test.Expected, err == nil) {
				t.Fatalf("expected %v but got %v", test.Expected, err == nil)
			}
		})
	}
}
