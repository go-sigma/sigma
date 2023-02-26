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

package distribution

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
	"gorm.io/gorm"

	services "github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/xerrors"
)

// ListRepositories handles the list repositories request
func (h *handlers) ListRepositories(c echo.Context) error {
	var nStr = c.QueryParam("n")
	n, err := strconv.Atoi(nStr)
	if err != nil {
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodePaginationNumberInvalid)
	}

	ctx := c.Request().Context()

	lastFound := false
	var lastID uint64 = 0

	repositoryService := services.NewRepositoryService()
	var last = c.QueryParam("last")
	if last != "" {
		tagObj, err := repositoryService.GetByName(ctx, last)
		if err != nil && err != gorm.ErrRecordNotFound {
			return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
		}
		lastFound = true
		lastID = tagObj.ID
	}

	var repositories []*models.Repository
	if !lastFound {
		repositories, err = repositoryService.ListByDtPagination(ctx, n)
	} else {
		repositories, err = repositoryService.ListByDtPagination(ctx, n, lastID)
	}
	if err != nil {
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
	}
	var names []string
	for _, repository := range repositories {
		names = append(names, repository.Name)
	}

	var repositoryList = dtspecv1.RepositoryList{
		Repositories: names,
	}

	location := fmt.Sprintf("%s://%s%s", c.Scheme(), c.Request().Host, c.Request().URL.Path)
	values := url.Values{}
	values.Set("n", nStr)
	if len(repositories) > 0 {
		values.Set("last", repositories[len(repositories)-1].Name)
		c.Response().Header().Set("Link", fmt.Sprintf("<%s?%s>; rel=\"next\"", location, values.Encode()))
	}

	return c.JSON(http.StatusOK, repositoryList)
}
