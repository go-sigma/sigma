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
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/handlers/distribution/clients"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/imagerefs"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/reader"
	"github.com/go-sigma/sigma/pkg/validators"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetBlob returns the blob's size and digest.
func (h *handler) GetBlob(c echo.Context) error {
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
		return err
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

	blobService := h.blobServiceFactory.New()
	blob, err := blobService.FindByDigest(ctx, dgest.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) && h.config.Proxy.Enabled {
			f := clients.NewClientsFactory()
			cli, err := f.New(ptr.To(h.config)) // TODO: config param
			if err != nil {
				if err != nil {
					log.Error().Err(err).Str("digest", dgest.String()).Msg("New proxy server failed")
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
				}
			}
			statusCode, header, bodyReader, err := cli.DoRequest(ctx, c.Request().Method, c.Request().URL.Path, nil)
			if err != nil {
				log.Error().Err(err).Str("digest", dgest.String()).Msg("Request proxy server failed")
				return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
			}
			if statusCode != http.StatusOK {
				log.Error().Err(err).Str("digest", dgest.String()).Int("statusCode", statusCode).Msg("Request proxy server failed")
				return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
			}
			contentType := header.Get("Content-Type")
			pipeReader, pipeWriter := io.Pipe()
			newBodyReader := io.TeeReader(bodyReader, pipeWriter)
			go func() {
				blobSize, err := strconv.ParseInt(header.Get(echo.HeaderContentLength), 10, 64)
				if err != nil {
					log.Error().Err(err).Str("digest", dgest.String()).Msg("Parse content length failed")
					return
				}
				log.Info().Str("digest", dgest.String()).Int64("length", blobSize).Msg("Proxy blob")
				ctx := context.Background()
				err = storage.Driver.Upload(ctx, path.Join(consts.Blobs, utils.GenPathByDigest(dgest)), reader.LimitReader(pipeReader, blobSize))
				if err != nil {
					log.Error().Err(err).Str("digest", dgest.String()).Msg("Upload blob failed")
					return
				}
				// Note: the blob exist in the storage, but not in the database,
				// so gc should delete the file directly.
				err = blobService.Create(ctx, &models.Blob{Digest: dgest.String(), Size: blobSize, ContentType: contentType, PushedAt: time.Now()})
				if err != nil {
					log.Error().Err(err).Str("digest", dgest.String()).Msg("Create blob failed")
					return
				}
			}()
			c.Response().Header().Set(echo.HeaderContentLength, header.Get(echo.HeaderContentLength))
			return c.Stream(http.StatusOK, contentType, newBodyReader)
		}
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Check blob exist failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeBlobUnknown)
	}
	c.Request().Header.Set(consts.ContentDigest, dgest.String())
	c.Response().Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", blob.Size))

	reader, err := storage.Driver.Reader(ctx, path.Join(consts.Blobs, utils.GenPathByDigest(dgest)), 0)
	if err != nil {
		log.Error().Err(err).Str("digest", dgest.String()).Msg("Get blob reader failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	return c.Stream(http.StatusOK, blob.ContentType, reader)
}
