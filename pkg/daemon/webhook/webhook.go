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
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	workq.TopicHandlers[enums.DaemonWebhook.String()] = definition.Consumer{
		Handler:     webhookRunner,
		MaxRetry:    6,
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}
}

func webhookRunner(ctx context.Context, data []byte) error {
	ctx = log.Logger.WithContext(ctx)

	var payload types.DaemonWebhookPayload
	err := json.Unmarshal(data, &payload)
	if err != nil {
		return fmt.Errorf("unmarshal payload failed: %v", err)
	}
	w := webhook{
		namespaceServiceFactory:  dao.NewNamespaceServiceFactory(),
		repositoryServiceFactory: dao.NewRepositoryServiceFactory(),
		tagServiceFactory:        dao.NewTagServiceFactory(),
		webhookServiceFactory:    dao.NewWebhookServiceFactory(),
	}
	switch payload.Type {
	case enums.WebhookTypeResend:
		return w.decorator(w.resend)(ctx, payload)
	case enums.WebhookTypePing:
		return w.decorator(w.ping)(ctx, payload)
	case enums.WebhookTypeSend:
		return w.send(ctx, payload)
	}
	return nil
}

type webhook struct {
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	webhookServiceFactory    dao.WebhookServiceFactory
}

type clientOption struct {
	SslVerify     bool
	RetryTimes    int
	RetryDuration int
}

func (w webhook) resend(ctx context.Context, payload types.DaemonWebhookPayload) (*models.WebhookLog, error) {
	webhookService := w.webhookServiceFactory.New()
	webhookLogObj, err := webhookService.GetLog(ctx, ptr.To(payload.WebhookLogID))
	if err != nil {
		return nil, err
	}
	var result = &models.WebhookLog{
		WebhookID:    webhookLogObj.WebhookID,
		ResourceType: webhookLogObj.ResourceType,
		ReqHeader:    webhookLogObj.ReqHeader,
		ReqBody:      webhookLogObj.ReqBody,
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
		Execute(http.MethodPost, webhookLogObj.Webhook.URL)
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

func (w webhook) send(ctx context.Context, payload types.DaemonWebhookPayload) error {
	webhookService := w.webhookServiceFactory.New()
	filter := map[string]any{
		query.Webhook.Enable.ColumnName().String():      true,
		query.Webhook.NamespaceID.ColumnName().String(): payload.NamespaceID,
	}
	switch payload.ResourceType {
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
		ResourceType: enums.WebhookResourceTypeNamespace,
	})
	headers := w.defaultHeaders()
	for _, webhookObj := range webhookObjs {
		headers, err = w.secretHeader(webhookObj.Secret, body, headers)
		if err != nil {
			log.Error().Err(err).Msg("Calculate secret header failed")
			continue
		}
		webhookLogObj := &models.WebhookLog{
			WebhookID:    &webhookObj.ID,
			ResourceType: payload.ResourceType,
			Action:       payload.Action,
			ReqHeader:    utils.MustMarshal(headers),
			ReqBody:      body,
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
			Execute(http.MethodPost, webhookObj.URL)
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

func (w webhook) ping(ctx context.Context, payload types.DaemonWebhookPayload) (*models.WebhookLog, error) {
	webhookService := w.webhookServiceFactory.New()
	webhookObj, err := webhookService.Get(ctx, ptr.To(payload.WebhookID))
	if err != nil {
		return nil, err
	}
	headers := w.defaultHeaders()
	pingObj := types.DaemonWebhookPayloadPing{
		ResourceType: enums.WebhookResourceTypeWebhook,
		Action:       enums.WebhookActionPing,
	}
	pingObj.Namespace, err = w.getNamespace(ctx, webhookObj.NamespaceID)
	if err != nil {
		return nil, err
	}
	body := utils.MustMarshal(pingObj)
	headers, err = w.secretHeader(webhookObj.Secret, body, headers)
	if err != nil {
		return nil, err
	}
	var result = &models.WebhookLog{
		WebhookID:    payload.WebhookID,
		ResourceType: payload.ResourceType,
		Action:       payload.Action,
		ReqHeader:    utils.MustMarshal(headers),
		ReqBody:      body,
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
		Execute(http.MethodPost, webhookObj.URL)
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

func (w webhook) respBody(resp *resty.Response) ([]byte, error) {
	reader := resp.RawBody()
	defer reader.Close() // nolint: errcheck
	return io.ReadAll(&io.LimitedReader{R: reader, N: 10 * 1024})
}

func (w webhook) defaultHeaders() map[string]string {
	return map[string]string{
		"User-Agent":           consts.UserAgent,
		echo.HeaderContentType: "application/json",
	}
}
