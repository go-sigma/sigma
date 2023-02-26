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
package main

import (
	"gorm.io/gen"

	"github.com/ximager/ximager/pkg/dal/models"
)

func main() {
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
}
