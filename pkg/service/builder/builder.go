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
	"bytes"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// BuildRunnerOption ...
type BuildRunnerOption struct {
	Tag       string
	ScmBranch *string
}

// BuildRunner ...
func BuildRunner(builder *models.Builder, option BuildRunnerOption) (*models.BuilderRunner, error) {
	runner := &models.BuilderRunner{
		BuilderID: builder.ID,
		Status:    enums.BuildStatusPending,

		Tag:       option.Tag,
		ScmBranch: option.ScmBranch,
	}
	return runner, nil
}

// BuildTagOption ...
type BuildTagOption struct {
	ScmBranch string
	ScmTag    string
	ScmRef    string
}

// BuildTag ...
func BuildTag(tmpl string, option BuildTagOption) (string, error) {
	t, err := template.New("tag").Funcs(sprig.FuncMap()).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("Template parse failed: %v", err)
	}
	var buffer bytes.Buffer
	err = t.Execute(&buffer, nil)
	if err != nil {
		return "", fmt.Errorf("Execute template failed: %v", err)
	}
	return buffer.String(), nil
}
