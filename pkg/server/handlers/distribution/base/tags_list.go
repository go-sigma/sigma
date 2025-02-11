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
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/distribution/reference"
	"github.com/labstack/echo/v4"
	dtspecv1 "github.com/opencontainers/distribution-spec/specs-go/v1"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

var listTagsReg = regexp.MustCompile(fmt.Sprintf(`^/v2/%s/tags/list$`, reference.NameRegexp.String()))

// ListTags handles the list tags request
func (h *handler) ListTags(c echo.Context) error {
	user, needRet, err := utils.GetUserFromCtxForDs(c)
	if err != nil {
		return err
	}
	if needRet {
		return nil
	}

	var uri = c.Request().URL.Path
	if !listTagsReg.MatchString(uri) {
		return xerrors.NewDSError(c, xerrors.DSErrCodeNameInvalid)
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
	repository := strings.TrimSuffix(strings.TrimPrefix(uri, "/v2/"), "/tags/list")

	repositoryService := h.RepositoryServiceFactory.New()
	repositoryObj, err := repositoryService.GetByName(ctx, repository)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Str("repository", repository).Msg("Cannot find repository")
			return xerrors.NewDSError(c, xerrors.DSErrCodeNameUnknown)
		}
		log.Error().Err(err).Str("repository", repository).Msg("Get repository failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	authChecked, err := h.AuthServiceFactory.New().Repository(ptr.To(user), repositoryObj.ID, enums.AuthRead)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Msg("Resource not found")
			return xerrors.GenDSErrCodeResourceNotFound(err)
		}
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	if !authChecked {
		log.Error().Int64("UserID", user.ID).Int64("RepositoryID", repositoryObj.ID).Msg("Auth check failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeDenied)
	}

	var lastFound bool
	var lastID int64 = 0

	tagService := h.TagServiceFactory.New()
	var last = c.QueryParam("last")
	if last != "" {
		tagObj, err := tagService.GetByName(ctx, repositoryObj.ID, last)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error().Err(err).Msg("get tag by name")
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
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
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	var names = make([]string, 0, len(tags))
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
