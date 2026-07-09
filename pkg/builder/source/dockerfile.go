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

package source

import (
	"encoding/base64"
	"os"
	"path"

	"github.com/go-sigma/sigma/pkg/utils/compress"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// DockerfileSource 通过写入 Dockerfile 内容准备源码
type DockerfileSource struct {
	Dockerfile *string
}

// Prepare 解码并写入 Dockerfile
func (s DockerfileSource) Prepare() error {
	base64Bytes, err := base64.StdEncoding.DecodeString(ptr.To(s.Dockerfile))
	if err != nil {
		return err
	}
	dockerfileStr, err := compress.Decompress(base64Bytes)
	if err != nil {
		return err
	}
	file, err := os.Create(path.Join(workspace, "Dockerfile"))
	if err != nil {
		return err
	}
	_, err = file.WriteString(dockerfileStr)
	if err != nil {
		return err
	}
	return nil
}
