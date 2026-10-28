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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/go-sigma/sigma/pkg/api/enums"
	builderpkg "github.com/go-sigma/sigma/pkg/builder"
	"github.com/go-sigma/sigma/pkg/logger"
)

// NewCmdBuilder 返回 builder 子命令，该命令在构建容器内执行一次性镜像构建
func NewCmdBuilder() *cobra.Command {
	return &cobra.Command{
		Use:   "builder",
		Short: "Start the sigma builder",
		// The builder runs as a one-shot CLI inside the builder image: every input comes from the
		// environment and the image ships no sigma config file. Override the root command's
		// persistent pre-run hook so it does not try to read /etc/sigma/config.yaml.
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			logger.SetLevel(enums.LogLevelInfo.String())
		},
		Run: func(_ *cobra.Command, _ []string) {
			flow := &builderpkg.BuildFlow{}
			checkErr(flow.Run())
		},
	}
}

func checkErr(msg any) {
	if msg != nil {
		logger.Fatal(fmt.Sprintf("something error occurred: %v", msg))
	}
}
