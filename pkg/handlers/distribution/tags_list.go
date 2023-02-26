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
	"regexp"
	"strconv"
	"strings"

	"github.com/distribution/distribution/v3/reference"
	"github.com/labstack/echo/v4"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	services "github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/xerrors"
)

var listTagsReg = regexp.MustCompile(fmt.Sprintf(`^/v2/%s/tags/list$`, reference.NameRegexp.String()))

// ListTags handles the list tags request
func (h *handlers) ListTags(c echo.Context) error {
	var uri = c.Request().URL.Path
	if !listTagsReg.MatchString(uri) {
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeNameInvalid)
	}

	var nStr = c.QueryParam("n")
	n, err := strconv.Atoi(nStr)
	if err != nil {
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodePaginationNumberInvalid)
	}

	ctx := c.Request().Context()
	repository := strings.TrimSuffix(strings.TrimPrefix(uri, "/v2/"), "/tags/list")

	lastFound := false
	var lastID uint64 = 0

	tagService := services.NewTagService()
	var last = c.QueryParam("last")
	if last != "" {
		tagObj, err := tagService.GetByName(ctx, repository, last)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error().Err(err).Msg("get tag by name")
			return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
		}
		lastFound = true
		lastID = tagObj.ID
	}

	var tags []*models.Tag
	if !lastFound {
		tags, err = tagService.ListByDtPagination(ctx, repository, n)
	} else {
		tags, err = tagService.ListByDtPagination(ctx, repository, n, lastID)
	}
	if err != nil {
		return xerrors.GenDsResponseError(c, xerrors.ErrorCodeUnknown)
	}
	var names []string
	for _, tag := range tags {
		names = append(names, tag.Name)
	}

	var tagList = dtspecv1.TagList{
		Name: repository,
		Tags: names,
	}

	host := c.Request().Host
	protocol := c.Scheme()
	location := fmt.Sprintf("%s://%s%s", protocol, host, uri)
	values := url.Values{}
	values.Set("n", nStr)
	if len(tags) > 0 {
		values.Set("last", tags[len(tags)-1].Name)
		c.Response().Header().Set("Link", fmt.Sprintf("<%s?%s>; rel=\"next\"", location, values.Encode()))
	}

	return c.JSON(http.StatusOK, tagList)
}
