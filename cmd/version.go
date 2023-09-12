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

package cmd

import (
	"fmt"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// You can copy & paste this ascii graphic and use it e.g. as mail signature
// Font: doh   Reflection: no   Adjustment: left   Stretch: no      Width: 280	 Text: sigma
const banner = `
                   iiii
                  i::::i
                   iiii

    ssssssssss   iiiiiii    ggggggggg   ggggg   mmmmmmm    mmmmmmm     aaaaaaaaaaaaa
  ss::::::::::s  i:::::i   g:::::::::ggg::::g mm:::::::m  m:::::::mm   a::::::::::::a
ss:::::::::::::s  i::::i  g:::::::::::::::::gm::::::::::mm::::::::::m  aaaaaaaaa:::::a
s::::::ssss:::::s i::::i g::::::ggggg::::::ggm::::::::::::::::::::::m           a::::a
 s:::::s  ssssss  i::::i g:::::g     g:::::g m:::::mmm::::::mmm:::::m    aaaaaaa:::::a
   s::::::s       i::::i g:::::g     g:::::g m::::m   m::::m   m::::m  aa::::::::::::a
      s::::::s    i::::i g:::::g     g:::::g m::::m   m::::m   m::::m a::::aaaa::::::a
ssssss   s:::::s  i::::i g::::::g    g:::::g m::::m   m::::m   m::::ma::::a    a:::::a
s:::::ssss::::::si::::::ig:::::::ggggg:::::g m::::m   m::::m   m::::ma::::a    a:::::a
s::::::::::::::s i::::::i g::::::::::::::::g m::::m   m::::m   m::::ma:::::aaaa::::::a
 s:::::::::::ss  i::::::i  gg::::::::::::::g m::::m   m::::m   m::::m a::::::::::aa:::a
  sssssssssss    iiiiiiii    gggggggg::::::g mmmmmm   mmmmmm   mmmmmm  aaaaaaaaaa  aaaa
                                     g:::::g
                         gggggg      g:::::g
                         g:::::gg   gg:::::g
                          g::::::ggg:::::::g
                           gg:::::::::::::g
                             ggg::::::ggg
                                gggggg`

var (
	version   = ""
	gitHash   = ""
	buildDate = ""
)

// versionCmd represents the worker command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version of sigma",
	Run: func(_ *cobra.Command, _ []string) {
		color.Cyan(banner)
		fmt.Printf("Version:     %s\n", version)
		fmt.Printf("GoVersion:   %s\n", runtime.Version())
		fmt.Printf("Platform:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("BuildDate:   %s\n", buildDate)
		fmt.Printf("GitCommit:   %s\n", gitHash)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
