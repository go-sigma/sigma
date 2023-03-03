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

package validators

import (
	"regexp"

	"github.com/distribution/distribution/v3/reference"
	"github.com/go-playground/validator"
	"github.com/opencontainers/go-digest"
)

const (
	maxNamespace = 20
)

var (
	namespaceRegex = regexp.MustCompile(`^[a-z][a-z-]{0,20}$`)
)

// Register registers the validators
func Register(v *validator.Validate) {
	v.RegisterValidation("is_valid_namespace", ValidateNamespace)
	v.RegisterValidation("is_valid_repository", ValidateRepository)
	v.RegisterValidation("is_valid_digest", ValidateDigest)
	v.RegisterValidation("is_valid_tag", ValidateTag)
}

// ValidateRepository validates the repository name
func ValidateRepository(field validator.FieldLevel) bool {
	_, err := reference.ParseNormalizedNamed(field.Field().String())
	return err == nil
}

// ValidateDigest validates the digest
func ValidateDigest(field validator.FieldLevel) bool {
	dgest := field.Field().String()
	_, err := digest.Parse(dgest)
	return err == nil
}

// ValidateNamespace validates the namespace name
func ValidateNamespace(field validator.FieldLevel) bool {
	namespace := field.Field().String()
	return namespaceRegex.MatchString(namespace) && len(namespace) <= maxNamespace
}

// ValidateTag validates the tag
func ValidateTag(field validator.FieldLevel) bool {
	return reference.TagRegexp.MatchString(field.Field().String())
}
