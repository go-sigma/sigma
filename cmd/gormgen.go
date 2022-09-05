/*
Copyright © 2023 Tosone <i@tosone.cn>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"github.com/spf13/cobra"
	"gorm.io/gen"

	"github.com/ximager/ximager/pkg/dal/models"
)

// gormGenCmd represents the gormGen command
var gormGenCmd = &cobra.Command{
	Use:   "gorm-gen",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(_ *cobra.Command, _ []string) {
		g := gen.NewGenerator(gen.Config{
			OutPath: "pkg/dal/query",
			Mode:    gen.WithDefaultQuery,
		})

		g.ApplyBasic(
			models.Namespace{},
			models.Repository{},
			models.Artifact{},
			models.Tag{},
			models.Blob{},
			models.BlobUpload{},
		)

		g.ApplyInterface(func(models.TagQuerier) {}, models.Tag{})

		g.Execute()
	},
}

func init() {
	rootCmd.AddCommand(gormGenCmd)
}
