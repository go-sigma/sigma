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
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// GetRunnerLog ...
func (h *handler) GetRunnerLog(c *gin.Context, req *api.GetRunnerLog) {
	ctx := c.Request.Context()

	builderSvc := h.BuilderSvc
	builderObj, err := builderSvc.GetBuilderByRepositoryID(ctx, req.RepositoryID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get builder by repository id failed: %v", err))
		return
	}
	if builderObj == nil || builderObj.ID != req.BuilderID {
		slog.Error("get builder by id failed", "builder_id", req.BuilderID)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, "Get builder by id failed")
		return
	}

	runnerObj, err := builderSvc.GetRunner(ctx, req.RunnerID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Builder runner find failed: %v", err))
		return
	}

	websocket.Handler(func(ws *websocket.Conn) {
		defer func() {
			err = ws.Close()
			if err != nil {
				slog.Error("close the ws failed", "err", err)
			}
		}()
		switch runnerObj.Status {
		case enums.BuildStatusFailed, enums.BuildStatusSuccess: // already built
			reader, compressed, err := builderSvc.RunnerLogReader(ctx, req.BuilderID, runnerObj.ID, runnerObj.Status)
			if err != nil {
				slog.Error("read log failed", "err", err)
				return
			}
			if err = h.sendRunnerLog(ws, reader, compressed); err != nil {
				slog.Error("send log failed", "err", err)
				return
			}
		case enums.BuildStatusBuilding: // still building
			for {
				reader, compressed, err := builderSvc.RunnerLogReader(ctx, req.BuilderID, req.RunnerID, runnerObj.Status)
				if err != nil {
					slog.Error("read log failed", "err", err)
					return
				}
				if err = h.sendRunnerLog(ws, reader, compressed); err != nil {
					slog.Error("send log failed", "err", err)
					return
				}
			}
		}
	}).ServeHTTP(c.Writer, c.Request)
	c.Status(http.StatusNoContent)
}

func (h *handler) sendRunnerLog(ws *websocket.Conn, reader io.Reader, compressed bool) error {
	if compressed {
		return h.sendLogWithGzip(ws, reader)
	}
	return h.sendLogWithoutGzip(ws, reader)
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
			slog.Error("read builder runner log failed", "err", err)
			return err
		}
		err = websocket.Message.Send(ws, data)
		if err != nil {
			slog.Error("send builder runner log failed", "err", err)
			return err
		}
	}
}
