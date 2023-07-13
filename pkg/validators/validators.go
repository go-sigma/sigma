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
	"net/http"
	"regexp"
	"strings"

	"github.com/distribution/distribution/v3/reference"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

const (
	maxNamespace = 20
)

var (
	namespaceRegex = regexp.MustCompile(`^[a-z][0-9a-z-]{0,20}$`)
)

// CustomValidator is a custom validator for echo
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates the input
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// Initialize initializes the validator
func Initialize(e *echo.Echo) {
	validate := validator.New()
	register(validate)
	e.Validator = &CustomValidator{validator: validate}
}

// register registers the validators
func register(v *validator.Validate) {
	v.RegisterValidation("is_valid_namespace", ValidateNamespace)   // nolint:errcheck
	v.RegisterValidation("is_valid_repository", ValidateRepository) // nolint:errcheck
	v.RegisterValidation("is_valid_digest", ValidateDigest)         // nolint:errcheck
	v.RegisterValidation("is_valid_tag", ValidateTag)               // nolint:errcheck
	v.RegisterValidation("is_valid_visibility", ValidateVisibility) // nolint:errcheck
}

// ValidateRepository validates the repository name
func ValidateRepository(field validator.FieldLevel) bool {
	repository := field.Field().String()
	if len(strings.Split(repository, "/")) < 2 {
		return false
	}
	_, err := reference.ParseNormalizedNamed(repository)
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
	return consts.TagRegexp.MatchString(field.Field().String())
}

// ValidateVisibility validates the visibility
func ValidateVisibility(field validator.FieldLevel) bool {
	v := field.Field().String()
	_, err := enums.ParseVisibility(v)
	return err == nil
}
