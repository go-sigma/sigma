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

package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// func init() {
// 	utils.PanicIf(daemon.RegisterTask(enums.DaemonWebhook, webhookRunner))
// }

// nolint: unused
func webhookRunner(ctx context.Context, task *asynq.Task) error {
	var payload types.DaemonWebhookPayload
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("Unmarshal payload failed: %v", err)
	}
	w := webhook{
		namespaceServiceFactory: dao.NewNamespaceServiceFactory(),
		webhookServiceFactory:   dao.NewWebhookServiceFactory(),
	}
	if payload.Resend {
		return w.decorator(w.resend)(ctx, payload)
	}
	if payload.Ping {
		return w.decorator(w.ping)(ctx, payload)
	}
	return w.send(ctx, payload)
}

// nolint: unused
type webhook struct {
	namespaceServiceFactory dao.NamespaceServiceFactory
	webhookServiceFactory   dao.WebhookServiceFactory
}

// nolint: unused
type clientOption struct {
	SslVerify     bool
	RetryTimes    int
	RetryDuration int
}

// nolint: unused
func (w webhook) resend(ctx context.Context, payload types.DaemonWebhookPayload) (*models.WebhookLog, error) {
	webhookService := w.webhookServiceFactory.New()
	webhookLogObj, err := webhookService.GetLog(ctx, ptr.To(payload.WebhookLogID))
	if err != nil {
		return nil, err
	}
	var result = &models.WebhookLog{
		WebhookID: webhookLogObj.WebhookID,
		Event:     webhookLogObj.Event,
		ReqHeader: webhookLogObj.ReqHeader,
		ReqBody:   webhookLogObj.ReqBody,
	}
	var headers map[string]string
	err = json.Unmarshal(webhookLogObj.ReqHeader, &headers)
	if err != nil {
		return nil, err
	}
	headers, err = w.secretHeader(webhookLogObj.Webhook.Secret, webhookLogObj.ReqBody, headers)
	if err != nil {
		return nil, err
	}
	client := w.client(clientOption{
		SslVerify:     webhookLogObj.Webhook.SslVerify,
		RetryTimes:    webhookLogObj.Webhook.RetryTimes,
		RetryDuration: webhookLogObj.Webhook.RetryDuration,
	})
	resp, err := client.SetContext(ctx).
		SetDoNotParseResponse(true).
		SetHeaders(headers).
		SetBody(webhookLogObj.ReqBody).
		Execute(http.MethodPost, webhookLogObj.Webhook.Url)
	if err != nil {
		return nil, err
	}
	respBody, err := w.respBody(resp)
	if err != nil {
		return nil, err
	}
	result.RespHeader = utils.MustMarshal(resp.Header())
	result.RespBody = respBody
	return result, nil
}

// nolint: unused
func (w webhook) send(ctx context.Context, payload types.DaemonWebhookPayload) error {
	webhookService := w.webhookServiceFactory.New()
	filter := map[string]any{
		query.Webhook.Enable.ColumnName().String():      true,
		query.Webhook.NamespaceID.ColumnName().String(): payload.NamespaceID,
	}
	switch payload.Event {
	case enums.WebhookResourceTypeNamespace:
		filter[query.Webhook.EventNamespace.ColumnName().String()] = true
	case enums.WebhookResourceTypeRepository:
		filter[query.Webhook.EventRepository.ColumnName().String()] = true
	case enums.WebhookResourceTypeTag:
		filter[query.Webhook.EventTag.ColumnName().String()] = true
	case enums.WebhookResourceTypeArtifact:
		filter[query.Webhook.EventArtifact.ColumnName().String()] = true
	case enums.WebhookResourceTypeMember:
		filter[query.Webhook.EventMember.ColumnName().String()] = true
	}
	webhookObjs, err := webhookService.GetByFilter(ctx, filter)
	if err != nil {
		return err
	}
	body := utils.MustMarshal(types.DaemonWebhookPayloadPing{
		Event: string(enums.WebhookResourceTypePing),
	})
	headers := w.defaultHeaders()
	for _, webhookObj := range webhookObjs {
		headers, err = w.secretHeader(webhookObj.Secret, body, headers)
		if err != nil {
			log.Error().Err(err).Msg("Calculate secret header failed")
			continue
		}
		webhookLogObj := &models.WebhookLog{
			WebhookID: webhookObj.ID,
			Event:     payload.Event,
			Action:    payload.Action,
			ReqHeader: utils.MustMarshal(headers),
			ReqBody:   body,
		}
		client := w.client(clientOption{
			SslVerify:     webhookObj.SslVerify,
			RetryTimes:    webhookObj.RetryTimes,
			RetryDuration: webhookObj.RetryDuration,
		})
		resp, err := client.SetContext(ctx).
			SetDoNotParseResponse(true).
			SetHeaders(headers).
			SetBody(webhookLogObj.ReqBody).
			Execute(http.MethodPost, webhookObj.Url)
		if err != nil {
			log.Error().Err(err).Msg("Send webhook failed")
			continue
		}
		respBody, err := w.respBody(resp)
		if err != nil {
			log.Error().Err(err).Msg("Parse response body failed")
			continue
		}
		webhookLogObj.RespHeader = utils.MustMarshal(resp.Header())
		webhookLogObj.RespBody = respBody
		err = webhookService.CreateLog(ctx, webhookLogObj)
		if err != nil {
			log.Error().Err(err).Msg("Create webhook log failed") // must be database something wrong, webhook has been sent, so we can ignore this error
			continue
		}
	}
	return nil
}

// nolint: unused
func (w webhook) ping(ctx context.Context, payload types.DaemonWebhookPayload) (*models.WebhookLog, error) {
	webhookService := w.webhookServiceFactory.New()
	webhookObj, err := webhookService.Get(ctx, ptr.To(payload.WebhookID))
	if err != nil {
		return nil, err
	}
	headers := w.defaultHeaders()
	body := utils.MustMarshal(types.DaemonWebhookPayloadPing{
		Event: string(enums.WebhookResourceTypePing),
	})
	headers, err = w.secretHeader(webhookObj.Secret, body, headers)
	if err != nil {
		return nil, err
	}
	var result = &models.WebhookLog{
		WebhookID: ptr.To(payload.WebhookID),
		Event:     payload.Event,
		ReqHeader: utils.MustMarshal(headers),
		ReqBody:   body,
	}
	client := w.client(clientOption{
		SslVerify:     webhookObj.SslVerify,
		RetryTimes:    webhookObj.RetryTimes,
		RetryDuration: webhookObj.RetryDuration,
	})
	resp, err := client.SetContext(ctx).
		SetDoNotParseResponse(true).
		SetHeaders(headers).
		SetBody(body).
		Execute(http.MethodPost, webhookObj.Url)
	if err != nil {
		return nil, err
	}
	respBody, err := w.respBody(resp)
	if err != nil {
		return nil, err
	}
	result.RespHeader = utils.MustMarshal(resp.Header())
	result.RespBody = respBody
	return result, nil
}

// nolint: unused
func (w webhook) secretHeader(secret *string, body []byte, headers map[string]string) (map[string]string, error) {
	delete(headers, consts.WebhookSecretHeader)
	if secret == nil {
		return headers, nil
	}
	hash := hmac.New(sha256.New, []byte(ptr.To(secret)))
	_, err := hash.Write(body)
	if err != nil {
		return nil, err
	}
	headers[consts.WebhookSecretHeader] = hex.EncodeToString(hash.Sum(nil))
	return headers, nil
}

// nolint: unused
func (w webhook) client(opt clientOption) *resty.Request {
	client := resty.New()
	if !opt.SslVerify {
		client = resty.NewWithClient(&http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, // nolint: gosec
		})
	}
	client = client.SetRetryCount(opt.RetryTimes)
	client = client.SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
		return time.Duration(opt.RetryDuration) * time.Second, nil
	})
	client = client.AddRetryCondition(func(r *resty.Response, err error) bool {
		return err != nil || r.StatusCode() >= http.StatusInternalServerError || r.StatusCode() == http.StatusTooManyRequests
	})
	return client.R()
}

// nolint: unused
func (w webhook) decorator(runner func(context.Context, types.DaemonWebhookPayload) (*models.WebhookLog, error)) func(ctx context.Context, payload types.DaemonWebhookPayload) error {
	return func(ctx context.Context, payload types.DaemonWebhookPayload) error {
		webhookLogObj, err := runner(ctx, payload)
		if err != nil {
			return err
		}
		webhookService := w.webhookServiceFactory.New()
		err = webhookService.CreateLog(ctx, webhookLogObj)
		if err != nil {
			return err
		}
		return nil
	}
}

// nolint: unused
func (w webhook) respBody(resp *resty.Response) ([]byte, error) {
	contentLength, err := strconv.ParseInt(resp.Header().Get(echo.HeaderContentLength), 10, 0)
	if err != nil {
		return nil, err
	}
	var respBody []byte
	if contentLength > 0 && contentLength < 1024*100 {
		respBody, err = io.ReadAll(resp.RawBody())
		if err != nil {
			return nil, err
		}
	}
	return respBody, nil
}

// nolint: unused
func (w webhook) defaultHeaders() map[string]string {
	return map[string]string{
		"User-Agent":           consts.UserAgent,
		echo.HeaderContentType: "application/json",
	}
}
