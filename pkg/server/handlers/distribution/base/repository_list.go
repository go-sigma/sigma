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
	"net/http"
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// ListRepositories handles the list repositories request
func (h *handler) ListRepositories(c echo.Context) error {
	user, needRet, err := utils.GetUserFromCtxForDs(c)
	if err != nil {
		return err
	}
	if needRet {
		return nil
	}

	var n = 1000
	var nStr = c.QueryParam("n")
	if nStr != "" {
		n, err = strconv.Atoi(nStr)
		if err != nil {
			return xerrors.NewDSError(c, xerrors.DSErrCodePaginationNumberInvalid)
		}
	}

	ctx := log.Logger.WithContext(c.Request().Context())

	lastFound := false
	var lastID int64 = 0

	repositoryService := h.RepositoryServiceFactory.New()
	var last = c.QueryParam("last")
	if last != "" {
		tagObj, err := repositoryService.GetByName(ctx, last)
		if err != nil && err != gorm.ErrRecordNotFound {
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}
		lastFound = true
		lastID = tagObj.ID
	}

	var repositories []*models.Repository
	if !lastFound {
		repositories, err = repositoryService.ListWithScrollable(ctx, 0, user.ID, nil, n, 0)
	} else {
		repositories, err = repositoryService.ListWithScrollable(ctx, 0, user.ID, nil, n, lastID)
	}
	if err != nil {
		log.Error().Err(err).Msg("List repository failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	var names = make([]string, 0, len(repositories))
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
