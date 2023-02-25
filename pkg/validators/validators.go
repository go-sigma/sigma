// The MIT License (MIT)
//
// Copyright © 2023 Tosone <i@tosone.cn>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package validators

import (
	"regexp"

	"github.com/distribution/distribution/v3/reference"
	"github.com/go-playground/validator"
	"github.com/opencontainers/go-digest"
)

var (
	namespaceRegex = regexp.MustCompile(`^[a-z]+$`)
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
	return reference.NameRegexp.MatchString(field.Field().String())
}

// ValidateDigest validates the digest
func ValidateDigest(field validator.FieldLevel) bool {
	dgest := field.Field().String()
	_, err := digest.Parse(dgest)
	return err == nil
}

// ValidateNamespace validates the namespace name
func ValidateNamespace(field validator.FieldLevel) bool {
	return namespaceRegex.MatchString(field.Field().String())
}

// ValidateTag validates the tag
func ValidateTag(field validator.FieldLevel) bool {
	return reference.TagRegexp.MatchString(field.Field().String())
}
