// Copyright 2023 XImager
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
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/handlers/distribution/clients"
	"github.com/ximager/ximager/pkg/xerrors"
)

// HeadBlob returns the blob's size and digest.
func (h *handler) HeadBlob(c echo.Context) error {
	uri := c.Request().URL.Path

	dgest, err := digest.Parse(strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/"))
	if err != nil {
		log.Error().Err(err).Str("digest", c.QueryParam("digest")).Msg("Parse digest failed")
		return err
	}
	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/blobs"), "/v2/")
	log.Debug().Str("digest", dgest.String()).Str("repository", repository).Msg("Blob info")

	ctx := log.Logger.WithContext(c.Request().Context())

	c.Response().Header().Set(consts.ContentDigest, dgest.String())

	blobService := dao.NewBlobService()
	blob, err := blobService.FindByDigest(ctx, dgest.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) && viper.GetBool("proxy.enabled") {
			cli, err := clients.New()
			if err != nil {
				log.Error().Err(err).Str("digest", dgest.String()).Msg("New proxy server failed")
				return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
			}
			statusCode, header, _, err := cli.DoRequest(c.Request().Method, c.Request().URL.Path)
			if err != nil {
				log.Error().Err(err).Str("digest", dgest.String()).Msg("Request proxy server failed")
				return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
			}
			if statusCode != http.StatusOK {
				log.Error().Err(err).Str("digest", dgest.String()).Int("statusCode", statusCode).Msg("Request proxy server failed")
				return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
			}
			c.Response().Header().Set("Content-Length", header.Get("Content-Length"))
			return c.NoContent(http.StatusOK)
		}
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Check blob exist failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeBlobUnknown)
	}
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", blob.Size))
	return c.NoContent(http.StatusOK)
}
