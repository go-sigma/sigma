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
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/auth"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/handlers/distribution"
	"github.com/go-sigma/sigma/pkg/server/handlers/distribution/clients"
	"github.com/go-sigma/sigma/pkg/modules/cacher"
	"github.com/go-sigma/sigma/pkg/modules/cacher/definition"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Handler is the interface for the distribution blob handlers
type Handler interface {
	// DeleteBlob ...
	DeleteBlob(ctx echo.Context) error
	// HeadBlob ...
	HeadBlob(ctx echo.Context) error
	// GetBlob ...
	GetBlob(ctx echo.Context) error
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	Config                   configs.Configuration
	AuthServiceFactory       auth.AuthServiceFactory
	AuditServiceFactory      dao.AuditServiceFactory
	NamespaceServiceFactory  dao.NamespaceServiceFactory
	RepositoryServiceFactory dao.RepositoryServiceFactory
	BlobServiceFactory       dao.BlobServiceFactory
}

// handlerNew creates a new instance of the distribution blob handlers
func handlerNew(digCon *dig.Container) Handler {
	return ptr.Of(utils.MustGetObjFromDigCon[handler](digCon))
}

type factory struct{}

// Initialize initializes the distribution manifest handlers
func (f factory) Initialize(c echo.Context, digCon *dig.Container) error {
	method := c.Request().Method
	uri := c.Request().RequestURI
	urix := uri[:strings.LastIndex(uri, "/")]
	handler := handlerNew(digCon)

	if strings.HasSuffix(urix, "/blobs") {
		switch method {
		case http.MethodGet:
			return handler.GetBlob(c)
		case http.MethodHead:
			return handler.HeadBlob(c)
		default:
			return c.String(http.StatusMethodNotAllowed, "Method Not Allowed")
		}
	}
	return distribution.ErrNext
}

func init() {
	utils.PanicIf(distribution.RegisterRouterFactory(&factory{}, 3))
}

func (h *handler) BlobCacher(c echo.Context) (definition.Cacher[*models.Blob], error) {
	return cacher.New(nil, consts.CacherBlob, func(key string) (*models.Blob, error) {
		ctx := log.Logger.WithContext(c.Request().Context())

		dgest, err := digest.Parse(key)
		if err != nil {
			log.Error().Err(err).Str("digest", key).Msg("Parse digest failed")
			return nil, xerrors.DSErrCodeUnknown
		}
		blobService := h.BlobServiceFactory.New()
		blob, err := blobService.FindByDigest(ctx, dgest.String())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if !h.Config.Proxy.Enabled {
					log.Error().Err(err).Str("digest", dgest.String()).Msg("Blob not found")
					return nil, xerrors.DSErrCodeBlobUnknown
				}
				f := clients.NewClientsFactory()
				cli, err := f.New(h.Config)
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
}
