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

package builders

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/websocket"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/builder"
	"github.com/go-sigma/sigma/pkg/builder/logger"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetRunnerLog ...
func (h *handler) GetRunnerLog(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetRunnerLog
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	builderService := h.BuilderServiceFactory.New()
	builderObj, err := builderService.GetByRepositoryID(ctx, req.RepositoryID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Get builder by repository id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get builder by repository id failed: %v", err))
	}
	if builderObj.ID != req.BuilderID {
		log.Error().Int64("builder_id", req.BuilderID).Int64("builder_id", builderObj.ID).Msg("Get builder by id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, "Get builder by id failed")
	}

	runnerObj, err := builderService.GetRunner(ctx, req.RunnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msgf("Builder runner not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, "Builder runner not found")
		}
		log.Error().Err(err).Msgf("Builder runner find failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Builder runner find failed: %v", err))
	}

	websocket.Handler(func(ws *websocket.Conn) {
		defer func() {
			err = ws.Close()
			if err != nil {
				log.Error().Err(err).Msg("Close the ws failed")
			}
		}()
		if runnerObj.Status == enums.BuildStatusFailed || runnerObj.Status == enums.BuildStatusSuccess { // already built
			var reader io.Reader
			if logger.Driver == nil {
				reader = strings.NewReader("")
			} else {
				reader, err = logger.Driver.Read(ctx, runnerObj.ID)
				if err != nil {
					log.Error().Err(err).Msg("Read log failed")
					return
				}
			}
			err = h.sendLogWithGzip(ws, reader)
			if err != nil {
				log.Error().Err(err).Msg("Send log failed")
				return
			}
		} else if runnerObj.Status == enums.BuildStatusBuilding { // still building
			for {
				reader, writer := io.Pipe()
				go func() {
					err = builder.Driver.LogStream(ctx, req.BuilderID, req.RunnerID, writer)
					if err != nil {
						log.Error().Err(err).Msg("Read log failed")
						_, err = writer.Write([]byte{10})
						if err != nil {
							log.Error().Err(err).Msg("Write log failed")
						}
					}
				}()
				err = h.sendLogWithoutGzip(ws, reader)
				if err != nil {
					log.Error().Err(err).Msg("Send log failed")
					return
				}
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) sendLogWithGzip(ws *websocket.Conn, reader io.Reader) error {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	return h.sendLog(ws, gzipReader)
}

func (h *handler) sendLogWithoutGzip(ws *websocket.Conn, reader io.Reader) error {
	return h.sendLog(ws, reader)
}

func (h *handler) sendLog(ws *websocket.Conn, reader io.Reader) error {
	for {
		var data = make([]byte, 512)
		_, err := reader.Read(data)
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil
		}
		if err != nil {
			log.Error().Err(err).Msg("Read builder runner log failed")
			return err
		}
		err = websocket.Message.Send(ws, data)
		if err != nil {
			log.Error().Err(err).Msg("Send builder runner log failed")
			return err
		}
	}
}
