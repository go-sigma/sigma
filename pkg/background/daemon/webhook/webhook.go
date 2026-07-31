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
	"log/slog"
	"maps"
	"net/http"
	"path"
	"reflect"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/dig"
	"resty.dev/v3"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/background/daemon"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	repowebhook "github.com/go-sigma/sigma/pkg/dal/repository/webhook"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	sigmatelemetry "github.com/go-sigma/sigma/pkg/telemetry"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func init() {
	utils.PanicIf(daemon.Daemons.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}

type factory struct{}

type params struct {
	dig.In

	HandlerRegistry      workq.HandlerRegistry
	NamespaceRepository  reponamespace.NamespaceRepository
	RepositoryRepository reporegistry.RepositoryRepository
	TagRepository        reporegistry.TagRepository
	WebhookRepository    repowebhook.WebhookRepository
}

// Initialize initializes the webhook daemon, which is used to send webhooks
// when resources are pushed or updated.
func (f factory) Initialize(digCon *dig.Container) error {
	var p params
	if err := digCon.Invoke(func(deps params) { p = deps }); err != nil {
		return err
	}
	return p.HandlerRegistry.Register(enums.DaemonWebhook, workq.Consumer{
		Handler:     webhookRunner(p),
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	})
}

func webhookRunner(p params) func(ctx context.Context, data []byte) error {
	return func(ctx context.Context, data []byte) error {
		var payload api.DaemonWebhookPayload
		err := json.Unmarshal(data, &payload)
		if err != nil {
			return fmt.Errorf("unmarshal payload failed: %v", err)
		}
		w := webhook{
			namespaceRepository:  p.NamespaceRepository,
			repositoryRepository: p.RepositoryRepository,
			tagRepository:        p.TagRepository,
			webhookRepository:    p.WebhookRepository,
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
}

type webhook struct {
	namespaceRepository  reponamespace.NamespaceRepository
	repositoryRepository reporegistry.RepositoryRepository
	tagRepository        reporegistry.TagRepository
	webhookRepository    repowebhook.WebhookRepository
}

type clientOption struct {
	SslVerify     bool
	RetryTimes    int
	RetryDuration int
}

func (w webhook) resend(ctx context.Context, payload api.DaemonWebhookPayload) (*models.WebhookLog, error) {
	webhookRepository := w.webhookRepository
	webhookLogObj, err := webhookRepository.GetLog(ctx, ptr.To(payload.WebhookLogID))
	if err != nil {
		return nil, err
	}
	var result = &models.WebhookLog{
		ID:           uuid.NewV7String(),
		WebhookID:    webhookLogObj.WebhookID,
		ResourceType: webhookLogObj.ResourceType,
		Action:       webhookLogObj.Action,
		ReqBody:      webhookLogObj.ReqBody,
	}
	var headers map[string]string
	err = json.Unmarshal(webhookLogObj.ReqHeader, &headers)
	if err != nil {
		return nil, err
	}
	headers = sigmatelemetry.InjectCarrier(ctx, headers)
	headers, err = w.secretHeader(webhookLogObj.Webhook.Secret, webhookLogObj.ReqBody, headers)
	if err != nil {
		return nil, err
	}
	traceContext := sigmatelemetry.CarrierFromContext(ctx)
	result.TraceContext = utils.MustMarshal(traceContext)
	result.ReqHeader = utils.MustMarshal(headers)
	client := w.client(clientOption{
		SslVerify:     webhookLogObj.Webhook.SslVerify,
		RetryTimes:    webhookLogObj.Webhook.RetryTimes,
		RetryDuration: webhookLogObj.Webhook.RetryDuration,
	})
	resp, err := client.SetContext(ctx).
		SetResponseDoNotParse(true).
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
	result.StatusCode = resp.StatusCode()
	result.RespHeader = utils.MustMarshal(resp.Header())
	result.RespBody = respBody
	return result, nil
}

func (w webhook) send(ctx context.Context, payload api.DaemonWebhookPayload) error {
	webhookRepository := w.webhookRepository
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
	case enums.WebhookResourceTypeDaemonTaskGcArtifactRule, enums.WebhookResourceTypeDaemonTaskGcArtifactRunner,
		enums.WebhookResourceTypeDaemonTaskGcBlobRule, enums.WebhookResourceTypeDaemonTaskGcBlobRunner,
		enums.WebhookResourceTypeDaemonTaskGcRepositoryRule, enums.WebhookResourceTypeDaemonTaskGcRepositoryRunner,
		enums.WebhookResourceTypeDaemonTaskGcTagRule, enums.WebhookResourceTypeDaemonTaskGcTagRunner:
		filter[query.Webhook.EventDaemonTaskGc.ColumnName().String()] = true
	}
	webhookObjs, err := webhookRepository.GetByFilter(ctx, filter)
	if err != nil {
		return err
	}
	headers := w.defaultHeaders()
	for _, webhookObj := range webhookObjs {
		requestHeaders := make(map[string]string, len(headers)+2)
		maps.Copy(requestHeaders, headers)
		requestHeaders = sigmatelemetry.InjectCarrier(ctx, requestHeaders)
		requestHeaders, err = w.secretHeader(webhookObj.Secret, payload.Payload, requestHeaders)
		if err != nil {
			slog.Error("calculate secret header failed", "err", err)
			continue
		}
		traceContext := sigmatelemetry.CarrierFromContext(ctx)
		webhookLogObj := &models.WebhookLog{
			ID:           uuid.NewV7String(),
			WebhookID:    webhookObj.ID,
			ResourceType: payload.ResourceType,
			Action:       payload.Action,
			TraceContext: utils.MustMarshal(traceContext),
			ReqHeader:    utils.MustMarshal(requestHeaders),
			ReqBody:      payload.Payload,
		}
		client := w.client(clientOption{
			SslVerify:     webhookObj.SslVerify,
			RetryTimes:    webhookObj.RetryTimes,
			RetryDuration: webhookObj.RetryDuration,
		})
		resp, err := client.SetContext(ctx).
			SetResponseDoNotParse(true).
			SetHeaders(requestHeaders).
			SetBody(webhookLogObj.ReqBody).
			Execute(http.MethodPost, webhookObj.URL)
		if err != nil {
			slog.Error("send webhook failed", "err", err)
			continue
		}
		respBody, err := w.respBody(resp)
		if err != nil {
			slog.Error("parse response body failed", "err", err)
			continue
		}
		webhookLogObj.StatusCode = resp.StatusCode()
		webhookLogObj.RespHeader = utils.MustMarshal(resp.Header())
		webhookLogObj.RespBody = respBody
		err = webhookRepository.CreateLog(ctx, webhookLogObj)
		if err != nil {
			// must be database something wrong, webhook has been sent, so we can ignore this error
			slog.Error("create webhook log failed", "err", err)
			continue
		}
	}
	return nil
}

func (w webhook) ping(ctx context.Context, payload api.DaemonWebhookPayload) (*models.WebhookLog, error) {
	webhookRepository := w.webhookRepository
	webhookObj, err := webhookRepository.Get(ctx, payload.WebhookID)
	if err != nil {
		return nil, err
	}
	headers := w.defaultHeaders()
	pingObj := api.DaemonWebhookPayloadPing{
		ResourceType: enums.WebhookResourceTypeWebhook,
		Action:       enums.WebhookActionPing,
	}
	pingObj.Namespace, err = w.getNamespace(ctx, webhookObj.NamespaceID)
	if err != nil {
		return nil, err
	}
	body := utils.MustMarshal(pingObj)
	headers = sigmatelemetry.InjectCarrier(ctx, headers)
	headers, err = w.secretHeader(webhookObj.Secret, body, headers)
	if err != nil {
		return nil, err
	}
	traceContext := sigmatelemetry.CarrierFromContext(ctx)
	var result = &models.WebhookLog{
		ID:           uuid.NewV7String(),
		WebhookID:    payload.WebhookID,
		ResourceType: payload.ResourceType,
		Action:       payload.Action,
		TraceContext: utils.MustMarshal(traceContext),
		ReqHeader:    utils.MustMarshal(headers),
		ReqBody:      body,
	}
	client := w.client(clientOption{
		SslVerify:     webhookObj.SslVerify,
		RetryTimes:    webhookObj.RetryTimes,
		RetryDuration: webhookObj.RetryDuration,
	})
	resp, err := client.SetContext(ctx).
		SetResponseDoNotParse(true).
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
	result.StatusCode = resp.StatusCode()
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
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if !opt.SslVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // nolint: gosec
	}
	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(transport),
	}
	client := resty.NewWithClient(httpClient)
	client = client.SetRetryCount(opt.RetryTimes)
	client = client.SetRetryWaitTime(time.Duration(opt.RetryDuration) * time.Second)
	client = client.AddRetryConditions(func(r *resty.Response, err error) bool {
		return err != nil || r.StatusCode() >= http.StatusInternalServerError || r.StatusCode() == http.StatusTooManyRequests
	})
	return client.R()
}

func (w webhook) decorator(runner func(context.Context, api.DaemonWebhookPayload) (*models.WebhookLog, error)) func(ctx context.Context, payload api.DaemonWebhookPayload) error {
	return func(ctx context.Context, payload api.DaemonWebhookPayload) error {
		webhookLogObj, err := runner(ctx, payload)
		if err != nil {
			return err
		}
		webhookRepository := w.webhookRepository
		err = webhookRepository.CreateLog(ctx, webhookLogObj)
		if err != nil {
			return err
		}
		return nil
	}
}

func (w webhook) respBody(resp *resty.Response) ([]byte, error) {
	reader := resp.Body
	defer reader.Close() // nolint: errcheck
	return io.ReadAll(&io.LimitedReader{R: reader, N: 10 * 1024})
}

func (w webhook) defaultHeaders() map[string]string {
	return map[string]string{
		consts.HeaderUserAgent:   consts.UserAgent,
		consts.HeaderContentType: "application/json",
	}
}
