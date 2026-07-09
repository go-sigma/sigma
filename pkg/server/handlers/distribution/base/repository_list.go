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
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// scheme returns the request URL scheme, honoring the X-Forwarded-Proto header.
func scheme(c *gin.Context) string { return utils.Scheme(c) }

// ListRepositories handles the list repositories request
func (h *handler) ListRepositories(c *gin.Context) {
	user, needRet := utils.GetUserFromCtx(c, utils.UserCtxErrorDistribution)
	if needRet {
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

	page := 1
	limit := n
	sortField := "id"
	sortMethod := enums.SortMethodAsc
	repositories, _, _, err := h.RepoSvc.ListRepositories(ctx, user.ID, "", nil,
		api.Pagination{Page: &page, Limit: &limit},
		api.Sortable{Sort: &sortField, Method: &sortMethod})
	if err != nil {
		slog.Error("list repository failed", "err", err)
		errcode.NewDSError(c, errcode.DSErrCodeUnknown)
		return
	}

	var names = make([]string, 0, len(repositories))
	for _, repository := range repositories {
		names = append(names, repository.Name)
	}

	var repositoryList = dtspecv1.RepositoryList{
		Repositories: names,
	}

	location := fmt.Sprintf("%s://%s%s", scheme(c), c.Request.Host, c.Request.URL.Path)
	values := url.Values{}
	values.Set("n", nStr)
	if len(repositories) > 0 {
		values.Set("last", repositories[len(repositories)-1].Name)
		c.Writer.Header().Set("Link", fmt.Sprintf("<%s?%s>; rel=\"next\"", location, values.Encode()))
	}

	c.JSON(http.StatusOK, repositoryList)
}
