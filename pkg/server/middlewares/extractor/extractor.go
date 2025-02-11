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

// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package extractor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tidwall/gjson"
)

const (
	// extractorLimit is arbitrary number to limit values extractor can return. this limits possible resource exhaustion
	// attack vector
	extractorLimit = 60
)

var (
	errHeaderExtractorValueMissing = errors.New("missing value in request header")
	errHeaderExtractorValueInvalid = errors.New("invalid value in request header")
	errQueryExtractorValueMissing  = errors.New("missing value in the query string")
	errParamExtractorValueMissing  = errors.New("missing value in path params")
	errCookieExtractorValueMissing = errors.New("missing value in cookies")
	errFormExtractorValueMissing   = errors.New("missing value in the form")
	errJSONExtractorValueMissing   = errors.New("missing value in the JSON body")
)

// ValuesExtractor defines a function for extracting values (keys/tokens) from the given context.
type ValuesExtractor func(c echo.Context) ([]string, error)

// MustCreateExtractors creates ValuesExtractors from given lookups.
func MustCreateExtractors(lookups string) []ValuesExtractor {
	extractors, err := CreateExtractors(lookups)
	if err != nil {
		panic(err)
	}
	return extractors
}

// CreateExtractors creates ValuesExtractors from given lookups.
// Lookups is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
// to extract key from the request.
// Possible values:
//   - "header:<name>" or "header:<name>:<cut-prefix>"
//     `<cut-prefix>` is argument value to cut/trim prefix of the extracted value. This is useful if header
//     value has static prefix like `Authorization: <auth-scheme> <authorisation-parameters>` where part that we
//     want to cut is `<auth-scheme> ` note the space at the end.
//     In case of basic authentication `Authorization: Basic <credentials>` prefix we want to remove is `Basic `.
//   - "query:<name>"
//   - "param:<name>"
//   - "form:<name>"
//   - "json:<name>"
//   - "cookie:<name>"
//
// Multiple sources example:
// - "header:Authorization,header:X-Api-Key"
func CreateExtractors(lookups string) ([]ValuesExtractor, error) {
	return createExtractors(lookups, "")
}

func createExtractors(lookups string, authScheme string) ([]ValuesExtractor, error) {
	if lookups == "" {
		return nil, nil
	}
	sources := strings.Split(lookups, ",")
	var extractors = make([]ValuesExtractor, 0)
	for _, source := range sources {
		parts := strings.Split(source, ":")
		if len(parts) < 2 {
			return nil, fmt.Errorf("extractor source for lookup could not be split into needed parts: %v", source)
		}

		switch parts[0] {
		case "query":
			extractors = append(extractors, valuesFromQuery(parts[1]))
		case "param":
			extractors = append(extractors, valuesFromParam(parts[1]))
		case "cookie":
			extractors = append(extractors, valuesFromCookie(parts[1]))
		case "form":
			extractors = append(extractors, valuesFromForm(parts[1]))
		case "json":
			extractors = append(extractors, valuesFromJSONBody(parts[1]))
		case "header":
			prefix := ""
			if len(parts) > 2 {
				prefix = parts[2]
			} else if authScheme != "" && parts[1] == echo.HeaderAuthorization {
				// backwards compatibility for JWT and KeyAuth:
				// * we only apply this fix to Authorization as header we use and uses prefixes like "Bearer <token-value>" etc
				// * previously header extractor assumed that auth-scheme/prefix had a space as suffix we need to retain that
				//   behaviour for default values and Authorization header.
				prefix = authScheme
				if !strings.HasSuffix(prefix, " ") {
					prefix += " "
				}
			}
			extractors = append(extractors, valuesFromHeader(parts[1], prefix))
		}
	}
	return extractors, nil
}

// valuesFromHeader returns a functions that extracts values from the request header.
// valuePrefix is parameter to remove first part (prefix) of the extracted value. This is useful if header value has static
// prefix like `Authorization: <auth-scheme> <authorisation-parameters>` where part that we want to remove is `<auth-scheme> `
// note the space at the end. In case of basic authentication `Authorization: Basic <credentials>` prefix we want to remove
// is `Basic `. In case of JWT tokens `Authorization: Bearer <token>` prefix is `Bearer `.
// If prefix is left empty the whole value is returned.
func valuesFromHeader(header string, valuePrefix string) ValuesExtractor {
	prefixLen := len(valuePrefix)
	// standard library parses http.Request header keys in canonical form but we may provide something else so fix this
	header = textproto.CanonicalMIMEHeaderKey(header)
	return func(c echo.Context) ([]string, error) {
		values := c.Request().Header.Values(header)
		if len(values) == 0 {
			return nil, errHeaderExtractorValueMissing
		}

		result := make([]string, 0)
		for i, value := range values {
			if prefixLen == 0 {
				result = append(result, value)
				if i >= extractorLimit-1 {
					break
				}
				continue
			}
			if len(value) > prefixLen && strings.EqualFold(value[:prefixLen], valuePrefix) {
				result = append(result, value[prefixLen:])
				if i >= extractorLimit-1 {
					break
				}
			}
		}

		if len(result) == 0 {
			if prefixLen > 0 {
				return nil, errHeaderExtractorValueInvalid
			}
			return nil, errHeaderExtractorValueMissing
		}
		return result, nil
	}
}

// valuesFromQuery returns a function that extracts values from the query string.
func valuesFromQuery(param string) ValuesExtractor {
	return func(c echo.Context) ([]string, error) {
		result := c.QueryParams()[param]
		if len(result) == 0 {
			return nil, errQueryExtractorValueMissing
		} else if len(result) > extractorLimit-1 {
			result = result[:extractorLimit]
		}
		return result, nil
	}
}

// valuesFromParam returns a function that extracts values from the url param string.
func valuesFromParam(param string) ValuesExtractor {
	return func(c echo.Context) ([]string, error) {
		result := make([]string, 0)
		paramVales := c.ParamValues()
		for i, p := range c.ParamNames() {
			if param == p {
				result = append(result, paramVales[i])
				if i >= extractorLimit-1 {
					break
				}
			}
		}
		if len(result) == 0 {
			return nil, errParamExtractorValueMissing
		}
		return result, nil
	}
}

// valuesFromCookie returns a function that extracts values from the named cookie.
func valuesFromCookie(name string) ValuesExtractor {
	return func(c echo.Context) ([]string, error) {
		cookies := c.Cookies()
		if len(cookies) == 0 {
			return nil, errCookieExtractorValueMissing
		}

		result := make([]string, 0)
		for i, cookie := range cookies {
			if name == cookie.Name {
				result = append(result, cookie.Value)
				if i >= extractorLimit-1 {
					break
				}
			}
		}
		if len(result) == 0 {
			return nil, errCookieExtractorValueMissing
		}
		return result, nil
	}
}

// valuesFromForm returns a function that extracts values from the form field.
func valuesFromForm(name string) ValuesExtractor {
	return func(c echo.Context) ([]string, error) {
		if c.Request().Form == nil {
			_ = c.Request().ParseMultipartForm(32 << 20) // same what `c.Request().FormValue(name)` does
		}
		values := c.Request().Form[name]
		if len(values) == 0 {
			return nil, errFormExtractorValueMissing
		}
		if len(values) > extractorLimit-1 {
			values = values[:extractorLimit]
		}
		result := append([]string{}, values...)
		return result, nil
	}
}

func valuesFromJSONBody(name string) ValuesExtractor {
	return func(c echo.Context) ([]string, error) {
		reqBody := []byte{}
		if c.Request().Body != nil {
			reqBody, _ = io.ReadAll(c.Request().Body)
		}
		c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody))
		reqBodyStr := string(reqBody)
		result := gjson.Get(reqBodyStr, name)
		if result.IsArray() {
			var res []string
			for _, r := range result.Array() {
				res = append(res, r.String())
			}
			return res, nil
		}

		value := gjson.Get(string(reqBody), name).String()
		if value == "" {
			return nil, errJSONExtractorValueMissing
		}
		return []string{value}, nil
	}
}
