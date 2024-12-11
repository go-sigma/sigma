// Copyright 2024 sigma
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

	"github.com/distribution/reference"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/robfig/cron/v3"
	pwdvalidate "github.com/wagslane/go-password-validator"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
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
func Initialize(digCon *dig.Container) error {
	e := utils.MustGetObjFromDigCon[*echo.Echo](digCon)
	validator, err := newValidator()
	if err != nil {
		return err
	}
	e.Validator = &CustomValidator{validator: validator}
	return nil
}

// newValidator new validator
func newValidator() (*validator.Validate, error) {
	v := validator.New()
	err := v.RegisterValidation("is_valid_namespace_role", ValidateRetentionPattern)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_retention_pattern", ValidateRetentionPattern)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_retention_rule_type", ValidateRetentionRuleType)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_cron_rule", ValidateCronRule)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_user_role", ValidateUserRole)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_user_status", ValidateUserStatue)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_email", ValidateEmail)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_username", ValidateUsername)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_password", ValidatePassword)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_namespace", ValidateNamespace)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_repository", ValidateRepository)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_digest", ValidateDigest)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_tag", ValidateTag)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_visibility", ValidateVisibility)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_provider", ValidateProvider)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_scm_credential_type", ValidateScmCredentialType)
	if err != nil {
		return nil, err
	}
	err = v.RegisterValidation("is_valid_oci_platforms", ValidateOciPlatforms)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// ValidateNamespaceRole ...
func ValidateNamespaceRole(field validator.FieldLevel) bool {
	v := field.Field().String()
	_, err := enums.ParseNamespaceRole(v)
	return err == nil
}

// ValidateRetentionPattern ...
func ValidateRetentionPattern(field validator.FieldLevel) bool {
	patterns := strings.Split(field.Field().String(), ",")
	for _, pattern := range patterns {
		if pattern == "" {
			return false
		}
		_, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
	}
	return true
}

// ValidateRetentionRuleType ...
func ValidateRetentionRuleType(field validator.FieldLevel) bool {
	v := field.Field().String()
	_, err := enums.ParseRetentionRuleType(v)
	return err == nil
}

// ValidateCronRule ...
func ValidateCronRule(field validator.FieldLevel) bool {
	v := field.Field().String()
	_, err := cron.ParseStandard(v)
	return err == nil
}

// ValidateUserRole validates the user role
func ValidateUserRole(field validator.FieldLevel) bool {
	v := field.Field().String()
	_, err := enums.ParseUserRole(v)
	return err == nil
}

// ValidateUserStatue validates the user status
func ValidateUserStatue(field validator.FieldLevel) bool {
	v := field.Field().String()
	_, err := enums.ParseUserStatus(v)
	return err == nil
}

// ValidatePassword validates the password
func ValidatePassword(field validator.FieldLevel) bool {
	password := field.Field().String()
	return pwdvalidate.Validate(password, consts.PwdStrength) == nil
}

// ValidateEmail validates the email
func ValidateEmail(field validator.FieldLevel) bool {
	email := field.Field().String()
	return consts.RegexEmail.MatchString(email)
}

// ValidateUsername validates the username
func ValidateUsername(field validator.FieldLevel) bool {
	username := field.Field().String()
	return consts.RegexUsername.MatchString(username)
}

// ValidateRepository validates the repository name
func ValidateRepository(field validator.FieldLevel) bool {
	return ValidateRepositoryRaw(field.Field().String())
}

// ValidateRepositoryRaw ...
func ValidateRepositoryRaw(repository string) bool {
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
	return ValidateNamespaceRaw(field.Field().String())
}

// ValidateNamespaceRaw ...
func ValidateNamespaceRaw(namespace string) bool {
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

// ValidateProvider validates the provider
func ValidateProvider(field validator.FieldLevel) bool {
	v := field.Field().String()
	_, err := enums.ParseProvider(v)
	return err == nil
}

// ValidateScmCredentialType validates the ScmCredentialType
func ValidateScmCredentialType(field validator.FieldLevel) bool {
	v := field.Field().String()
	_, err := enums.ParseScmCredentialType(v)
	return err == nil
}

// ValidateBuilderSource validates the builder source
func ValidateBuilderSource(field validator.FieldLevel) bool {
	v := field.Field().String()
	_, err := enums.ParseBuilderSource(v)
	return err == nil
}

// ValidateOciPlatforms validates oci platforms
func ValidateOciPlatforms(field validator.FieldLevel) bool {
	for i := 0; i < field.Field().Len(); i++ {
		v := field.Field().Index(i).String()
		_, err := enums.ParseOciPlatform(v)
		if err != nil {
			return false
		}
	}
	return true
}
