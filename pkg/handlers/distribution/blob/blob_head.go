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

package blob

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/handlers/distribution/clients"
	"github.com/go-sigma/sigma/pkg/modules/cacher"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/imagerefs"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/validators"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// HeadBlob returns the blob's size and digest.
func (h *handler) HeadBlob(c echo.Context) error {
	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		log.Error().Msg("Get user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}
	user, ok := iuser.(*models.User)
	if !ok {
		log.Error().Msg("Convert user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}

	uri := c.Request().URL.Path

	dgest, err := digest.Parse(strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/"))
	if err != nil {
		log.Error().Err(err).Str("digest", c.QueryParam("digest")).Msg("Parse digest failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/blobs"), "/v2/")
	log.Debug().Str("digest", dgest.String()).Str("repository", repository).Msg("Blob info")

	ctx := log.Logger.WithContext(c.Request().Context())

	_, namespace, _, _, err := imagerefs.Parse(repository)
	if err != nil {
		log.Error().Err(err).Str("Repository", repository).Msg("Repository must container a valid namespace")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestWithNamespace)
	}
	if !(validators.ValidateNamespaceRaw(namespace) && validators.ValidateRepositoryRaw(repository)) {
		log.Error().Err(err).Str("Repository", repository).Msg("Repository must container a valid namespace")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestWithNamespace)
	}

	namespaceObj, err := h.namespaceServiceFactory.New().GetByName(ctx, namespace)
	if err != nil {
		log.Error().Err(err).Str("Name", repository).Msg("Get repository by name failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeBlobUnknown)
	}

	if !h.authServiceFactory.New().Namespace(c, namespaceObj.ID, enums.AuthRead) {
		log.Error().Int64("UserID", user.ID).Int64("NamespaceID", namespaceObj.ID).Msg("Auth check failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "No permission with this api")
	}

	c.Response().Header().Set(consts.ContentDigest, dgest.String())

	cache, err := cacher.New(consts.CacherBlob, func(key string) (*models.Blob, error) {
		dgest, err := digest.Parse(key)
		if err != nil {
			log.Error().Err(err).Str("digest", key).Msg("Parse digest failed")
			return nil, xerrors.DSErrCodeUnknown
		}
		blobService := h.blobServiceFactory.New()
		blob, err := blobService.FindByDigest(ctx, dgest.String())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if !h.config.Proxy.Enabled {
					log.Error().Err(err).Str("digest", dgest.String()).Msg("Blob not found")
					return nil, xerrors.DSErrCodeBlobUnknown
				}
				f := clients.NewClientsFactory()
				cli, err := f.New(ptr.To(h.config))
				if err != nil {
					log.Error().Err(err).Str("digest", dgest.String()).Msg("New proxy server failed")
					return nil, xerrors.DSErrCodeUnknown
				}
				statusCode, header, _, err := cli.DoRequest(ctx, c.Request().Method, c.Request().URL.Path, nil)
				if err != nil {
					log.Error().Err(err).Str("digest", dgest.String()).Msg("Request proxy server failed")
					return nil, xerrors.DSErrCodeUnknown
				}
				if statusCode != http.StatusOK {
					log.Error().Err(err).Str("digest", dgest.String()).Int("statusCode", statusCode).Msg("Request proxy server failed")
					return nil, xerrors.DSErrCodeUnknown
				}
				contentLength, err := strconv.ParseInt(header.Get(echo.HeaderContentLength), 10, 64)
				if err != nil {
					log.Error().Err(err).Str("digest", dgest.String()).Msg("Parse content length failed")
					return nil, xerrors.DSErrCodeUnknown
				}
				blob = &models.Blob{
					Digest:      dgest.String(),
					Size:        contentLength,
					ContentType: header.Get(echo.HeaderContentType),
				}
				c.Response().Header().Set("Content-Length", header.Get(echo.HeaderContentLength))
				return blob, nil
			}
			log.Error().Err(err).Str("digest", dgest.String()).Msg("Check blob exist failed")
			return nil, xerrors.DSErrCodeBlobUnknown
		}
		return blob, nil
	})
	if err != nil {
		log.Error().Err(err).Msg("Head blob failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	blobObj, err := cache.Get(ctx, dgest.String())
	if err != nil {
		if err, ok := err.(xerrors.ErrCode); ok {
			return xerrors.NewDSError(c, err)
		}
		log.Error().Err(err).Msg("Head blob failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", blobObj.Size))
	return c.NoContent(http.StatusOK)
}
