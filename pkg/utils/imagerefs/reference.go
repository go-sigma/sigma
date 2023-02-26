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

// Package imagerefs ...
package imagerefs

import (
	"fmt"
	"strings"

	"github.com/distribution/distribution/v3/reference"
)

// Parse ...
func Parse(name string) (string, string, string, string, error) {
	if !strings.Contains(name, "/") {
		return "", "", "", "", fmt.Errorf("invalid reference: %s", name)
	}

	named, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to parse reference: %v, %s", err, name)
	}
	named = reference.TagNameOnly(named)
	domain := reference.Domain(named)
	path := reference.Path(named)
	tagged, ok := named.(reference.Tagged)
	if !ok {
		return "", "", "", "", fmt.Errorf("reference is not tagged: %v, %s", named, name)
	}
	tag := tagged.Tag()
	if !strings.Contains(path, "/") {
		return "", "", "", "", fmt.Errorf("invalid reference: %s", name)
	}
	parts := strings.Split(path, "/")
	ns := parts[0]
	repo := path
	return domain, ns, repo, tag, nil
}
