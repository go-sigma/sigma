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

package distribution

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/distribution/reference"
	"github.com/gin-gonic/gin"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

var listTagsReg = regexp.MustCompile(fmt.Sprintf(`^/v2/%s/tags/list$`, reference.NameRegexp.String()))

// ListTags handles the list tags request
func (h *handler) ListTags(c *gin.Context) {
	user, needRet := utils.GetUserFromCtx(c, utils.UserCtxErrorDistribution)
	if needRet {
		return
	}

	var uri = c.Request.URL.Path
	if !listTagsReg.MatchString(uri) {
		errcode.NewDSError(c, errcode.DSErrCodeNameInvalid)
		return
	}

	var n = 1000
	var nStr = c.Query("n")
	if nStr != "" {
		var err error
		n, err = strconv.Atoi(nStr)
		if err != nil {
			errcode.NewDSError(c, errcode.DSErrCodePaginationNumberInvalid)
			return
		}
	}
	if n > 100 {
		n = 100
	}

	ctx := c.Request.Context()
	repository := strings.TrimSuffix(strings.TrimPrefix(uri, "/v2/"), "/tags/list")

	repositoryObj, err := h.RepoSvc.GetRepositoryByName(ctx, repository)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok && e.Code == errcode.HTTPErrCodeNotFound.Code {
			slog.Error("cannot find repository", "err", err, "repository", repository)
			errcode.NewDSError(c, errcode.DSErrCodeNameUnknown)
			return
		}
		slog.Error("get repository failed", "err", err, "repository", repository)
		errcode.NewDSError(c, errcode.DSErrCodeUnknown)
		return
	}

	authChecked, err := h.Authorizer.Repository(ctx, *user, repositoryObj.ID, enums.AuthRead)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("resource not found", "err", errors.New(utils.UnwrapJoinedErrors(err)))
			errcode.NewDSError(c, errcode.GenDSErrCodeResourceNotFound(err))
			return
		}
		errcode.NewDSError(c, errcode.DSErrCodeUnknown)
		return
	}
	if !authChecked {
		slog.Error("auth check failed", "UserID", user.ID, "RepositoryID", repositoryObj.ID)
		errcode.NewDSError(c, errcode.DSErrCodeDenied)
		return
	}

	page := 1
	limit := n
	sortField := "id"
	sortMethod := enums.SortMethodAsc
	tags, _, err := h.TagSvc.ListTags(ctx, repositoryObj.NamespaceID, repositoryObj.ID, nil, nil,
		api.Pagination{Page: &page, Limit: &limit},
		api.Sortable{Sort: &sortField, Method: &sortMethod})
	if err != nil {
		slog.Error("list tags failed", "err", err)
		errcode.NewDSError(c, errcode.DSErrCodeUnknown)
		return
	}

	var names = make([]string, 0, len(tags))
	for _, tag := range tags {
		names = append(names, tag.Name)
	}

	var tagList = dtspecv1.TagList{
		Name: repository,
		Tags: names,
	}

	host := c.Request.Host
	protocol := scheme(c)
	location := fmt.Sprintf("%s://%s%s", protocol, host, uri)
	values := url.Values{}
	values.Set("n", nStr)
	if len(tags) > 0 {
		values.Set("last", tags[len(tags)-1].Name)
		c.Writer.Header().Set("Link", fmt.Sprintf("<%s?%s>; rel=\"next\"", location, values.Encode()))
	}

	c.JSON(http.StatusOK, tagList)
}
