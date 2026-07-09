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
	"encoding/base64"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/utils"
)

// buildTagOption tag 模板渲染参数
type buildTagOption struct {
	ScmBranch string
	ScmTag    string
	ScmRef    string
}

const workspace = "/code"

// buildTag 使用 Go template 渲染镜像 tag
func buildTag(tmpl string, option buildTagOption) (string, error) {
	t, err := template.New("tag").Funcs(sprig.FuncMap()).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("template parse failed: %v", err)
	}
	var buffer bytes.Buffer
	err = t.Execute(&buffer, option)
	if err != nil {
		return "", fmt.Errorf("execute template failed: %v", err)
	}
	return buffer.String(), nil
}

// genTag 生成完整的镜像名称（含 tag）
func (f *BuildFlow) genTag() (string, error) {
	var tagOption = buildTagOption{}
	if f.Source != enums.BuilderSourceDockerfile {
		r, err := git.PlainOpen(workspace)
		if err != nil {
			return "", err
		}
		tagRefs, err := r.Tags()
		if err != nil {
			return "", err
		}
		var latestTag = struct {
			ref  *plumbing.Reference
			when time.Time
		}{}
		err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
			commitObj, err := r.CommitObject(tagRef.Hash())
			if err != nil {
				return err
			}
			if latestTag.ref == nil || commitObj.Committer.When.After(latestTag.when) {
				latestTag.ref = tagRef
				latestTag.when = commitObj.Committer.When
			}
			return nil
		})
		if err != nil {
			return "", err
		}

		ref, err := r.Head()
		if err != nil {
			return "", err
		}

		branchName := ref.Name().Short()

		tagOption = buildTagOption{
			ScmBranch: branchName,
			ScmRef:    ref.Hash().String(),
		}
		if latestTag.ref != nil {
			tagOption.ScmTag = strings.TrimPrefix(latestTag.ref.String(), "refs/tags/")
		}
	}

	tagBytes, err := base64.StdEncoding.DecodeString(f.Tag)
	if err != nil {
		return "", err
	}
	tag, err := buildTag(string(tagBytes), tagOption)
	if err != nil {
		return "", err
	}
	repositoryBytes, err := base64.StdEncoding.DecodeString(f.Repository)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s:%s", utils.TrimHTTP(f.Endpoint), string(repositoryBytes), tag), nil
}
