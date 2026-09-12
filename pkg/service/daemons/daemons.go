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

package daemons

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repodaemon "github.com/go-sigma/sigma/pkg/dal/repository/daemon"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

//go:generate go tool mockgen -mock_names Service=MockDaemonService -destination=daemons_mocks.go -package=daemons github.com/go-sigma/sigma/pkg/service/daemons Service

// Service encapsulates daemon GC runner related business logic
type Service interface {
	// --- GcArtifact ---

	// UpdateGcArtifactRule updates or creates the GC artifact rule and emits a webhook when updating
	UpdateGcArtifactRule(ctx context.Context, req api.UpdateGcArtifactRuleRequest) error
	// GetGcArtifactRule gets the GC artifact rule by namespace ID
	GetGcArtifactRule(ctx context.Context, namespaceID string) (*models.DaemonGcArtifactRule, error)
	// GetGcArtifactLatestRunner gets the latest GC artifact runner by namespace ID
	GetGcArtifactLatestRunner(ctx context.Context, namespaceID string) (*models.DaemonGcArtifactRunner, error)
	// CreateGcArtifactRunner creates a GC artifact runner and queues the GC task and webhook event
	CreateGcArtifactRunner(ctx context.Context, req api.CreateGcArtifactRunnerRequest) error
	// ListGcArtifactRunners lists GC artifact runners by namespace ID
	ListGcArtifactRunners(ctx context.Context, namespaceID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcArtifactRunner, int64, error)
	// GetGcArtifactRunner gets a GC artifact runner by namespace ID and runner ID
	GetGcArtifactRunner(ctx context.Context, namespaceID, runnerID string) (*models.DaemonGcArtifactRunner, error)
	// ListGcArtifactRecords lists GC artifact records by runner ID
	ListGcArtifactRecords(ctx context.Context, runnerID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcArtifactRecord, int64, error)
	// GetGcArtifactRecord gets a GC artifact record by namespace ID, runner ID and record ID
	GetGcArtifactRecord(ctx context.Context, namespaceID, runnerID, recordID string) (*models.DaemonGcArtifactRecord, error)

	// --- GcRepository ---

	// UpdateGcRepositoryRule updates or creates the GC repository rule and emits a webhook when updating
	UpdateGcRepositoryRule(ctx context.Context, req api.UpdateGcRepositoryRuleRequest) error
	// GetGcRepositoryRule gets the GC repository rule by namespace ID
	GetGcRepositoryRule(ctx context.Context, namespaceID string) (*models.DaemonGcRepositoryRule, error)
	// GetGcRepositoryLatestRunner gets the latest GC repository runner by namespace ID
	GetGcRepositoryLatestRunner(ctx context.Context, namespaceID string) (*models.DaemonGcRepositoryRunner, error)
	// CreateGcRepositoryRunner creates a GC repository runner and queues the GC task and webhook event
	CreateGcRepositoryRunner(ctx context.Context, req api.CreateGcRepositoryRunnerRequest) error
	// ListGcRepositoryRunners lists GC repository runners by namespace ID
	ListGcRepositoryRunners(ctx context.Context, namespaceID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcRepositoryRunner, int64, error)
	// GetGcRepositoryRunner gets a GC repository runner by namespace ID and runner ID
	GetGcRepositoryRunner(ctx context.Context, namespaceID, runnerID string) (*models.DaemonGcRepositoryRunner, error)
	// ListGcRepositoryRecords lists GC repository records by runner ID
	ListGcRepositoryRecords(ctx context.Context, runnerID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcRepositoryRecord, int64, error)
	// GetGcRepositoryRecord gets a GC repository record by namespace ID, runner ID and record ID
	GetGcRepositoryRecord(ctx context.Context, namespaceID, runnerID, recordID string) (*models.DaemonGcRepositoryRecord, error)

	// --- GcTag ---

	// UpdateGcTagRule updates or creates the GC tag rule and emits a webhook when updating
	UpdateGcTagRule(ctx context.Context, req api.UpdateGcTagRuleRequest) error
	// GetGcTagRule gets the GC tag rule by namespace ID
	GetGcTagRule(ctx context.Context, namespaceID string) (*models.DaemonGcTagRule, error)
	// GetGcTagLatestRunner gets the latest GC tag runner by namespace ID
	GetGcTagLatestRunner(ctx context.Context, namespaceID string) (*models.DaemonGcTagRunner, error)
	// CreateGcTagRunner creates a GC tag runner and queues the GC task and webhook event
	CreateGcTagRunner(ctx context.Context, req api.CreateGcTagRunnerRequest) error
	// ListGcTagRunners lists GC tag runners by namespace ID
	ListGcTagRunners(ctx context.Context, namespaceID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcTagRunner, int64, error)
	// GetGcTagRunner gets a GC tag runner by namespace ID and runner ID
	GetGcTagRunner(ctx context.Context, namespaceID, runnerID string) (*models.DaemonGcTagRunner, error)
	// ListGcTagRecords lists GC tag records by runner ID
	ListGcTagRecords(ctx context.Context, runnerID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcTagRecord, int64, error)
	// GetGcTagRecord gets a GC tag record by namespace ID, runner ID and record ID
	GetGcTagRecord(ctx context.Context, namespaceID, runnerID, recordID string) (*models.DaemonGcTagRecord, error)

	// --- GcBlob ---

	// UpdateGcBlobRule updates or creates the global GC blob rule and emits a webhook when updating
	UpdateGcBlobRule(ctx context.Context, req api.UpdateGcBlobRuleRequest) error
	// GetGcBlobRule gets the global GC blob rule
	GetGcBlobRule(ctx context.Context) (*models.DaemonGcBlobRule, error)
	// GetGcBlobLatestRunner gets the latest global GC blob runner after validating namespace scope
	GetGcBlobLatestRunner(ctx context.Context, namespaceID string) (*models.DaemonGcBlobRunner, error)
	// CreateGcBlobRunner creates a global GC blob runner and queues the GC task and webhook event
	CreateGcBlobRunner(ctx context.Context, userID string, req api.CreateGcBlobRunnerRequest) error
	// ListGcBlobRunners lists global GC blob runners
	ListGcBlobRunners(ctx context.Context, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcBlobRunner, int64, error)
	// GetGcBlobRunner gets a global GC blob runner by runner ID
	GetGcBlobRunner(ctx context.Context, runnerID string) (*models.DaemonGcBlobRunner, error)
	// ListGcBlobRecords lists global GC blob records by runner ID
	ListGcBlobRecords(ctx context.Context, runnerID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcBlobRecord, int64, error)
	// GetGcBlobRecord gets a global GC blob record by runner ID and record ID
	GetGcBlobRecord(ctx context.Context, runnerID, recordID string) (*models.DaemonGcBlobRecord, error)
}

type service struct {
	dig.In

	RepoDaemon repodaemon.DaemonRepository
	Producer   workq.Producer
}

func NewService(params service) Service {
	return &params
}

// parseCronNextTrigger parses the cron rule and returns the next trigger timestamp in milliseconds
func parseCronNextTrigger(cronRule *string) *int64 {
	if cronRule == nil {
		return nil
	}
	schedule, _ := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(ptr.To(cronRule))
	next := schedule.Next(time.Now()).UnixMilli()
	return &next
}

// ===================== GcArtifact =====================

// UpdateGcArtifactRule updates or creates the GC artifact rule and emits a webhook when updating
func (s *service) UpdateGcArtifactRule(ctx context.Context, req api.UpdateGcArtifactRuleRequest) error {
	var namespaceID *string
	if req.NamespaceID != "" {
		namespaceID = &req.NamespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcArtifact, namespaceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get gc artifact rule failed", "err", err)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		slog.Error("the gc artifact rule is running", "NamespaceID", ptr.To(namespaceID))
		return errcode.HTTPErrCodeBadRequest.Detail("The gc artifact rule is running")
	}
	var nextTrigger *int64
	if req.CronRule != nil {
		nextTrigger = parseCronNextTrigger(req.CronRule)
	}
	updates := make(map[string]any, 5)
	updates[query.DaemonGcRule.RetentionDay.ColumnName().String()] = req.RetentionDay
	updates[query.DaemonGcRule.CronEnabled.ColumnName().String()] = req.CronEnabled
	if req.CronEnabled {
		updates[query.DaemonGcRule.CronRule.ColumnName().String()] = ptr.To(req.CronRule)
		updates[query.DaemonGcRule.CronNextTrigger.ColumnName().String()] = ptr.To(nextTrigger)
	}
	return query.Q.Transaction(func(tx *query.Query) error {
		repoDaemon := repodaemon.NewDaemonRepository(tx)
		if ruleObj == nil {
			err = repoDaemon.CreateGcRule(ctx, &models.DaemonGcArtifactRule{
				Type:            enums.DaemonGcArtifact,
				ID:              uuid.NewV7String(),
				NamespaceID:     namespaceID,
				RetentionDay:    req.RetentionDay,
				CronEnabled:     req.CronEnabled,
				CronRule:        req.CronRule,
				CronNextTrigger: nextTrigger,
			})
			if err != nil {
				slog.Error("create gc artifact rule failed", "err", err)
				return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc artifact rule failed: %v", err))
			}
			return nil
		}
		err = repoDaemon.UpdateGcRule(ctx, ruleObj.ID, updates)
		if err != nil {
			slog.Error("update gc artifact rule failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update gc artifact rule failed: %v", err))
		}
		err = s.Producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
			NamespaceID:  namespaceID,
			Action:       enums.WebhookActionUpdate,
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcArtifactRule,
			Payload:      utils.MustMarshal(req),
		})
		if err != nil {
			slog.Error("webhook event produce failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
}

// GetGcArtifactRule gets the GC artifact rule by namespace ID
func (s *service) GetGcArtifactRule(ctx context.Context, namespaceID string) (*models.DaemonGcArtifactRule, error) {
	var nsID *string
	if namespaceID != "" {
		nsID = &namespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcArtifact, nsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc artifact rule not found", "err", err, "NamespaceID", namespaceID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		slog.Error("get gc artifact rule failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	return ruleObj, nil
}

// GetGcArtifactLatestRunner gets the latest GC artifact runner by namespace ID
func (s *service) GetGcArtifactLatestRunner(ctx context.Context, namespaceID string) (*models.DaemonGcArtifactRunner, error) {
	var nsID *string
	if namespaceID != "" {
		nsID = &namespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcArtifact, nsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc artifact rule not found", "err", err, "NamespaceID", namespaceID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		slog.Error("get gc artifact rule failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	runnerObj, err := s.RepoDaemon.GetGcLatestRunner(ctx, ruleObj.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc artifact runner not found", "err", err, "NamespaceID", namespaceID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc artifact runner not found: %v", err))
		}
		slog.Error("get gc artifact runner failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc artifact runner failed: %v", err))
	}
	return runnerObj, nil
}

// CreateGcArtifactRunner creates a GC artifact runner and queues the GC task and webhook event
func (s *service) CreateGcArtifactRunner(ctx context.Context, req api.CreateGcArtifactRunnerRequest) error {
	var namespaceID *string
	if req.NamespaceID != "" {
		namespaceID = &req.NamespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcArtifact, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc artifact rule not found", "err", err, "namespaceID", req.NamespaceID)
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		slog.Error("get gc artifact rule failed", "err", err)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		slog.Error("the gc artifact rule is running", "NamespaceID", ptr.To(namespaceID))
		return errcode.HTTPErrCodeBadRequest.Detail("The gc artifact rule is running")
	}
	return query.Q.Transaction(func(tx *query.Query) error {
		runnerObj := &models.DaemonGcArtifactRunner{ID: uuid.NewV7String(), RuleID: ruleObj.ID, Status: enums.TaskCommonStatusPending, OperateType: enums.OperateTypeManual}
		err = s.RepoDaemon.CreateGcRunner(ctx, runnerObj)
		if err != nil {
			slog.Error(fmt.Sprintf("Create gc artifact runner failed: %v", err), "ruleID", ruleObj.ID)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc artifact runner failed: %v", err))
		}
		err = s.Producer.Produce(ctx, enums.DaemonGcArtifact,
			api.DaemonGcPayload{RunnerID: runnerObj.ID})
		if err != nil {
			slog.Error(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonGcArtifact.String()), "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonGcArtifact.String()))
		}
		err = s.Producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
			NamespaceID:  namespaceID,
			Action:       enums.WebhookActionCreate,
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcArtifactRunner,
			Payload:      utils.MustMarshal(req),
		})
		if err != nil {
			slog.Error("webhook event produce failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
}

// ListGcArtifactRunners lists GC artifact runners by namespace ID
func (s *service) ListGcArtifactRunners(ctx context.Context, namespaceID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcArtifactRunner, int64, error) {
	var nsID *string
	if namespaceID != "" {
		nsID = &namespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcArtifact, nsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc artifact rule not found", "err", err, "namespaceID", namespaceID)
			return nil, 0, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		slog.Error("get gc artifact rule failed", "err", err)
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	runnerObjs, total, err := s.RepoDaemon.ListGcRunners(ctx, ruleObj.ID, pagination, sort)
	if err != nil {
		slog.Error("list gc artifact rules failed", "err", err)
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List gc artifact rules failed: %v", err))
	}
	return runnerObjs, total, nil
}

// GetGcArtifactRunner gets a GC artifact runner by namespace ID and runner ID
func (s *service) GetGcArtifactRunner(ctx context.Context, namespaceID, runnerID string) (*models.DaemonGcArtifactRunner, error) {
	runnerObj, err := s.RepoDaemon.GetGcRunner(ctx, runnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc artifact runner not found", "err", err, "namespaceID", namespaceID, "runnerID", runnerID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc artifact runner not found: %v", err))
		}
		slog.Error("get gc artifact runner failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc artifact runner failed: %v", err))
	}
	if ptr.To(runnerObj.Rule.NamespaceID) != namespaceID {
		slog.Error("get gc artifact runner not found", "err", err, "namespaceID", namespaceID, "runnerID", runnerID)
		return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc artifact runner not found: %v", err))
	}
	return runnerObj, nil
}

// ListGcArtifactRecords lists GC artifact records by runner ID
func (s *service) ListGcArtifactRecords(ctx context.Context, runnerID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcArtifactRecord, int64, error) {
	recordObjs, total, err := s.RepoDaemon.ListGcRecords(ctx, runnerID, pagination, sort)
	if err != nil {
		slog.Error("list gc artifact records failed", "err", err, "ruleID", runnerID)
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List gc artifact records failed: %v", err))
	}
	return recordObjs, total, nil
}

// GetGcArtifactRecord gets a GC artifact record by namespace ID, runner ID and record ID
func (s *service) GetGcArtifactRecord(ctx context.Context, namespaceID, runnerID, recordID string) (*models.DaemonGcArtifactRecord, error) {
	var nsID *string
	if namespaceID != "" {
		nsID = &namespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcArtifact, nsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc artifact rule not found", "err", err, "namespaceID", namespaceID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		slog.Error("get gc artifact rule failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	recordObj, err := s.RepoDaemon.GetGcRecord(ctx, recordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc artifact record not found", "err", err, "namespaceID", namespaceID, "runnerID", runnerID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc artifact record not found: %v", err))
		}
		slog.Error("get gc artifact record failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc artifact record failed: %v", err))
	}
	if recordObj.Runner.ID != runnerID || recordObj.Runner.Rule.ID != ruleObj.ID {
		slog.Error("get gc artifact record not found", "err", err, "namespaceID", namespaceID, "runnerID", runnerID)
		return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc artifact record not found: %v", err))
	}
	return recordObj, nil
}

// ===================== GcRepository =====================

// UpdateGcRepositoryRule updates or creates the GC repository rule and emits a webhook when updating
func (s *service) UpdateGcRepositoryRule(ctx context.Context, req api.UpdateGcRepositoryRuleRequest) error {
	var namespaceID *string
	if req.NamespaceID != "" {
		namespaceID = &req.NamespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcRepository, namespaceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get gc repository rule failed", "err", err)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		slog.Error("the gc repository rule is running", "NamespaceID", ptr.To(namespaceID))
		return errcode.HTTPErrCodeBadRequest.Detail("The gc repository rule is running")
	}
	var nextTrigger *int64
	if req.CronRule != nil {
		nextTrigger = parseCronNextTrigger(req.CronRule)
	}
	updates := make(map[string]any, 5)
	updates[query.DaemonGcRule.RetentionDay.ColumnName().String()] = req.RetentionDay
	updates[query.DaemonGcRule.CronEnabled.ColumnName().String()] = req.CronEnabled
	if ptr.To(req.CronEnabled) {
		updates[query.DaemonGcRule.CronRule.ColumnName().String()] = ptr.To(req.CronRule)
		updates[query.DaemonGcRule.CronNextTrigger.ColumnName().String()] = ptr.To(nextTrigger)
	}
	return query.Q.Transaction(func(tx *query.Query) error {
		repoDaemon := repodaemon.NewDaemonRepository(tx)
		if ruleObj == nil {
			err = repoDaemon.CreateGcRule(ctx, &models.DaemonGcRepositoryRule{
				Type:            enums.DaemonGcRepository,
				ID:              uuid.NewV7String(),
				NamespaceID:     namespaceID,
				RetentionDay:    req.RetentionDay,
				CronEnabled:     ptr.To(req.CronEnabled),
				CronRule:        req.CronRule,
				CronNextTrigger: nextTrigger,
			})
			if err != nil {
				slog.Error("create gc repository rule failed", "err", err)
				return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc repository rule failed: %v", err))
			}
			return nil
		}
		err = repoDaemon.UpdateGcRule(ctx, ruleObj.ID, updates)
		if err != nil {
			slog.Error("update gc repository rule failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update gc repository rule failed: %v", err))
		}
		err = s.Producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
			NamespaceID:  namespaceID,
			Action:       enums.WebhookActionUpdate,
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcRepositoryRule,
			Payload:      utils.MustMarshal(req),
		})
		if err != nil {
			slog.Error("webhook event produce failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
}

// GetGcRepositoryRule gets the GC repository rule by namespace ID
func (s *service) GetGcRepositoryRule(ctx context.Context, namespaceID string) (*models.DaemonGcRepositoryRule, error) {
	var nsID *string
	if namespaceID != "" {
		nsID = &namespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcRepository, nsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc repository rule not found", "err", err, "NamespaceID", namespaceID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc repository rule not found: %v", err))
		}
		slog.Error("get gc repository rule failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	return ruleObj, nil
}

// GetGcRepositoryLatestRunner gets the latest GC repository runner by namespace ID
func (s *service) GetGcRepositoryLatestRunner(ctx context.Context, namespaceID string) (*models.DaemonGcRepositoryRunner, error) {
	var nsID *string
	if namespaceID != "" {
		nsID = &namespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcRepository, nsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc repository rule not found", "err", err, "NamespaceID", namespaceID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc repository rule not found: %v", err))
		}
		slog.Error("get gc repository rule failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	runnerObj, err := s.RepoDaemon.GetGcLatestRunner(ctx, ruleObj.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc repository rule not found", "err", err, "NamespaceID", namespaceID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc repository rule not found: %v", err))
		}
		slog.Error("get gc repository rule failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	return runnerObj, nil
}

// CreateGcRepositoryRunner creates a GC repository runner and queues the GC task and webhook event
func (s *service) CreateGcRepositoryRunner(ctx context.Context, req api.CreateGcRepositoryRunnerRequest) error {
	var namespaceID *string
	if req.NamespaceID != "" {
		namespaceID = &req.NamespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcRepository, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc repository rule not found", "err", err, "namespaceID", req.NamespaceID)
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc repository rule not found: %v", err))
		}
		slog.Error("get gc repository rule failed", "err", err)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		slog.Error("the gc repository rule is running", "NamespaceID", ptr.To(namespaceID))
		return errcode.HTTPErrCodeBadRequest.Detail("The gc repository rule is running")
	}
	return query.Q.Transaction(func(tx *query.Query) error {
		runnerObj := &models.DaemonGcRepositoryRunner{ID: uuid.NewV7String(), RuleID: ruleObj.ID, Status: enums.TaskCommonStatusPending, OperateType: enums.OperateTypeManual}
		err = s.RepoDaemon.CreateGcRunner(ctx, runnerObj)
		if err != nil {
			slog.Error(fmt.Sprintf("Create gc repository runner failed: %v", err), "ruleID", ruleObj.ID)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc repository runner failed: %v", err))
		}
		err = s.Producer.Produce(ctx, enums.DaemonGcRepository,
			api.DaemonGcPayload{RunnerID: runnerObj.ID})
		if err != nil {
			slog.Error(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonGcRepository.String()), "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonGcRepository.String()))
		}
		err = s.Producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
			NamespaceID:  namespaceID,
			Action:       enums.WebhookActionCreate,
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcRepositoryRunner,
			Payload:      utils.MustMarshal(req),
		})
		if err != nil {
			slog.Error("webhook event produce failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
}

// ListGcRepositoryRunners lists GC repository runners by namespace ID
func (s *service) ListGcRepositoryRunners(ctx context.Context, namespaceID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcRepositoryRunner, int64, error) {
	var nsID *string
	if namespaceID != "" {
		nsID = &namespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcRepository, nsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc repository rule not found", "err", err, "NamespaceID", namespaceID)
			return nil, 0, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc repository rule not found: %v", err))
		}
		slog.Error("get gc repository rule failed", "err", err)
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	runnerObjs, total, err := s.RepoDaemon.ListGcRunners(ctx, ruleObj.ID, pagination, sort)
	if err != nil {
		slog.Error("list gc repository rule failed", "err", err)
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List gc repository rule failed: %v", err))
	}
	return runnerObjs, total, nil
}

// GetGcRepositoryRunner gets a GC repository runner by namespace ID and runner ID
func (s *service) GetGcRepositoryRunner(ctx context.Context, namespaceID, runnerID string) (*models.DaemonGcRepositoryRunner, error) {
	runnerObj, err := s.RepoDaemon.GetGcRunner(ctx, runnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc repository runner not found", "err", err, "namespaceID", namespaceID, "runnerID", runnerID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc repository runner not found: %v", err))
		}
		slog.Error("get gc repository runner failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc repository runner failed: %v", err))
	}
	if ptr.To(runnerObj.Rule.NamespaceID) != namespaceID {
		slog.Error("get gc repository runner not found", "err", err, "namespaceID", namespaceID, "runnerID", runnerID)
		return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc repository runner not found: %v", err))
	}
	return runnerObj, nil
}

// ListGcRepositoryRecords lists GC repository records by runner ID
func (s *service) ListGcRepositoryRecords(ctx context.Context, runnerID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcRepositoryRecord, int64, error) {
	recordObjs, total, err := s.RepoDaemon.ListGcRecords(ctx, runnerID, pagination, sort)
	if err != nil {
		slog.Error("list gc repository records failed", "err", err, "ruleID", runnerID)
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List gc repository records failed: %v", err))
	}
	return recordObjs, total, nil
}

// GetGcRepositoryRecord gets a GC repository record by namespace ID, runner ID and record ID
func (s *service) GetGcRepositoryRecord(ctx context.Context, namespaceID, runnerID, recordID string) (*models.DaemonGcRepositoryRecord, error) {
	var nsID *string
	if namespaceID != "" {
		nsID = &namespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcRepository, nsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc repository rule not found", "err", err, "namespaceID", namespaceID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc repository rule not found: %v", err))
		}
		slog.Error("get gc repository rule failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	recordObj, err := s.RepoDaemon.GetGcRecord(ctx, recordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc repository record not found", "err", err, "namespaceID", namespaceID, "runnerID", runnerID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc repository record not found: %v", err))
		}
		slog.Error("get gc repository record failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc repository record failed: %v", err))
	}
	if recordObj.Runner.ID != runnerID || recordObj.Runner.Rule.ID != ruleObj.ID {
		slog.Error("get gc repository record not found", "err", err, "namespaceID", namespaceID, "runnerID", runnerID)
		return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc repository record not found: %v", err))
	}
	return recordObj, nil
}

// ===================== GcTag =====================

// UpdateGcTagRule updates or creates the GC tag rule and emits a webhook when updating
func (s *service) UpdateGcTagRule(ctx context.Context, req api.UpdateGcTagRuleRequest) error {
	var namespaceID *string
	if req.NamespaceID != "" {
		namespaceID = &req.NamespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcTag, namespaceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get gc tag rule failed", "err", err)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		slog.Error("the gc tag rule is running", "NamespaceID", ptr.To(namespaceID))
		return errcode.HTTPErrCodeBadRequest.Detail("The gc tag rule is running")
	}
	var nextTrigger *int64
	if req.CronRule != nil {
		nextTrigger = parseCronNextTrigger(req.CronRule)
	}
	updates := make(map[string]any, 6)
	updates[query.DaemonGcRule.CronEnabled.ColumnName().String()] = req.CronEnabled
	updates[query.DaemonGcRule.RetentionRuleType.ColumnName().String()] = req.RetentionRuleType
	updates[query.DaemonGcRule.RetentionRuleAmount.ColumnName().String()] = req.RetentionRuleAmount
	if req.CronEnabled {
		if req.CronRule != nil {
			updates[query.DaemonGcRule.CronRule.ColumnName().String()] = ptr.To(req.CronRule)
			updates[query.DaemonGcRule.CronNextTrigger.ColumnName().String()] = ptr.To(nextTrigger)
		}
	}
	if req.RetentionPattern != nil {
		updates[query.DaemonGcRule.RetentionPattern.ColumnName().String()] = ptr.To(req.RetentionPattern)
	}
	return query.Q.Transaction(func(tx *query.Query) error {
		repoDaemon := repodaemon.NewDaemonRepository(tx)
		if ruleObj == nil {
			err = repoDaemon.CreateGcRule(ctx, &models.DaemonGcTagRule{
				Type:                enums.DaemonGcTag,
				ID:                  uuid.NewV7String(),
				NamespaceID:         namespaceID,
				CronEnabled:         req.CronEnabled,
				CronRule:            req.CronRule,
				CronNextTrigger:     nextTrigger,
				RetentionPattern:    req.RetentionPattern,
				RetentionRuleType:   req.RetentionRuleType,
				RetentionRuleAmount: req.RetentionRuleAmount,
			})
			if err != nil {
				slog.Error("create gc tag rule failed", "err", err)
				return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc tag rule failed: %v", err))
			}
			return nil
		}
		err = repoDaemon.UpdateGcRule(ctx, ruleObj.ID, updates)
		if err != nil {
			slog.Error("update gc tag rule failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update gc tag rule failed: %v", err))
		}
		err = s.Producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
			NamespaceID:  namespaceID,
			Action:       enums.WebhookActionUpdate,
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcTagRule,
			Payload:      utils.MustMarshal(req),
		})
		if err != nil {
			slog.Error("webhook event produce failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
}

// GetGcTagRule gets the GC tag rule by namespace ID
func (s *service) GetGcTagRule(ctx context.Context, namespaceID string) (*models.DaemonGcTagRule, error) {
	var nsID *string
	if namespaceID != "" {
		nsID = &namespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcTag, nsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc tag rule not found", "err", err, "NamespaceID", namespaceID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc tag rule not found: %v", err))
		}
		slog.Error("get gc tag rule failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	return ruleObj, nil
}

// GetGcTagLatestRunner gets the latest GC tag runner by namespace ID
func (s *service) GetGcTagLatestRunner(ctx context.Context, namespaceID string) (*models.DaemonGcTagRunner, error) {
	var nsID *string
	if namespaceID != "" {
		nsID = &namespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcTag, nsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc tag rule not found", "err", err, "NamespaceID", namespaceID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc tag rule not found: %v", err))
		}
		slog.Error("get gc tag rule failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	runnerObj, err := s.RepoDaemon.GetGcLatestRunner(ctx, ruleObj.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc tag rule not found", "err", err, "NamespaceID", namespaceID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc tag rule not found: %v", err))
		}
		slog.Error("get gc tag rule failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	return runnerObj, nil
}

// CreateGcTagRunner creates a GC tag runner and queues the GC task and webhook event
func (s *service) CreateGcTagRunner(ctx context.Context, req api.CreateGcTagRunnerRequest) error {
	var namespaceID *string
	if req.NamespaceID != "" {
		namespaceID = &req.NamespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcTag, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc tag rule not found", "err", err, "NamespaceID", req.NamespaceID)
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc tag rule not found: %v", err))
		}
		slog.Error("get gc tag rule failed", "err", err)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		slog.Error("the gc tag rule is running", "NamespaceID", req.NamespaceID)
		return errcode.HTTPErrCodeBadRequest.Detail("The gc tag rule is running")
	}
	return query.Q.Transaction(func(tx *query.Query) error {
		runnerObj := &models.DaemonGcTagRunner{ID: uuid.NewV7String(), RuleID: ruleObj.ID, Status: enums.TaskCommonStatusPending, OperateType: enums.OperateTypeManual}
		err = s.RepoDaemon.CreateGcRunner(ctx, runnerObj)
		if err != nil {
			slog.Error(fmt.Sprintf("Create gc tag runner failed: %v", err), "RuleID", ruleObj.ID)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc tag runner failed: %v", err))
		}
		err = s.Producer.Produce(ctx, enums.DaemonGcTag,
			api.DaemonGcPayload{RunnerID: runnerObj.ID})
		if err != nil {
			slog.Error(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonGcTag.String()), "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonGcTag.String()))
		}
		err = s.Producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
			NamespaceID:  namespaceID,
			Action:       enums.WebhookActionCreate,
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcTagRule,
			Payload:      utils.MustMarshal(req),
		})
		if err != nil {
			slog.Error("webhook event produce failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
}

// ListGcTagRunners lists GC tag runners by namespace ID
func (s *service) ListGcTagRunners(ctx context.Context, namespaceID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcTagRunner, int64, error) {
	var nsID *string
	if namespaceID != "" {
		nsID = &namespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcTag, nsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc tag rule not found", "err", err, "NamespaceID", namespaceID)
			return nil, 0, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc tag rule not found: %v", err))
		}
		slog.Error("get gc tag rule failed", "err", err)
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	runnerObjs, total, err := s.RepoDaemon.ListGcRunners(ctx, ruleObj.ID, pagination, sort)
	if err != nil {
		slog.Error("list gc tag rule failed", "err", err)
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List gc tag rule failed: %v", err))
	}
	return runnerObjs, total, nil
}

// GetGcTagRunner gets a GC tag runner by namespace ID and runner ID
func (s *service) GetGcTagRunner(ctx context.Context, namespaceID, runnerID string) (*models.DaemonGcTagRunner, error) {
	runnerObj, err := s.RepoDaemon.GetGcRunner(ctx, runnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc tag runner not found", "err", err, "NamespaceID", namespaceID, "runnerID", runnerID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc tag runner not found: %v", err))
		}
		slog.Error("get gc tag runner failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc tag runner failed: %v", err))
	}
	if ptr.To(runnerObj.Rule.NamespaceID) != namespaceID {
		slog.Error("get gc tag runner not found", "err", err, "NamespaceID", namespaceID, "runnerID", runnerID)
		return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc tag runner not found: %v", err))
	}
	return runnerObj, nil
}

// ListGcTagRecords lists GC tag records by runner ID
func (s *service) ListGcTagRecords(ctx context.Context, runnerID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcTagRecord, int64, error) {
	recordObjs, total, err := s.RepoDaemon.ListGcRecords(ctx, runnerID, pagination, sort)
	if err != nil {
		slog.Error("list gc tag records failed", "err", err, "RuleID", runnerID)
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List gc tag records failed: %v", err))
	}
	return recordObjs, total, nil
}

// GetGcTagRecord gets a GC tag record by namespace ID, runner ID and record ID
func (s *service) GetGcTagRecord(ctx context.Context, namespaceID, runnerID, recordID string) (*models.DaemonGcTagRecord, error) {
	var nsID *string
	if namespaceID != "" {
		nsID = &namespaceID
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcTag, nsID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc tag rule not found", "err", err, "NamespaceID", namespaceID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc tag rule not found: %v", err))
		}
		slog.Error("get gc tag rule failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	recordObj, err := s.RepoDaemon.GetGcRecord(ctx, recordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc tag record not found", "err", err, "NamespaceID", namespaceID, "runnerID", runnerID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc tag record not found: %v", err))
		}
		slog.Error("get gc tag record failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc tag record failed: %v", err))
	}
	if recordObj.Runner.ID != runnerID || recordObj.Runner.Rule.ID != ruleObj.ID {
		slog.Error("get gc tag record not found", "err", err, "NamespaceID", namespaceID, "runnerID", runnerID)
		return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc tag record not found: %v", err))
	}
	return recordObj, nil
}

// ===================== GcBlob =====================

// UpdateGcBlobRule updates or creates the global GC blob rule and emits a webhook when updating
func (s *service) UpdateGcBlobRule(ctx context.Context, req api.UpdateGcBlobRuleRequest) error {
	if req.NamespaceID != "" {
		slog.Error("namespaceID should always be 0 in action UpdateGcBlobRule")
		return errcode.HTTPErrCodeUnauthorized.Detail("NamespaceID should always be 0 in action UpdateGcBlobRule")
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcBlob, nil)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get gc tag rule failed", "err", err)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		slog.Error("the gc blob rule is running")
		return errcode.HTTPErrCodeBadRequest.Detail("The gc tag rule is running")
	}
	var nextTrigger *int64
	if req.CronRule != nil {
		nextTrigger = parseCronNextTrigger(req.CronRule)
	}
	updates := make(map[string]any, 5)
	updates[query.DaemonGcRule.RetentionDay.ColumnName().String()] = req.RetentionDay
	updates[query.DaemonGcRule.CronEnabled.ColumnName().String()] = req.CronEnabled
	if req.CronEnabled {
		updates[query.DaemonGcRule.CronRule.ColumnName().String()] = ptr.To(req.CronRule)
		updates[query.DaemonGcRule.CronNextTrigger.ColumnName().String()] = ptr.To(nextTrigger)
	}
	return query.Q.Transaction(func(tx *query.Query) error {
		repoDaemon := repodaemon.NewDaemonRepository(tx)
		if ruleObj == nil {
			err = repoDaemon.CreateGcRule(ctx, &models.DaemonGcBlobRule{
				Type:            enums.DaemonGcBlob,
				ID:              uuid.NewV7String(),
				CronEnabled:     req.CronEnabled,
				RetentionDay:    req.RetentionDay,
				CronRule:        req.CronRule,
				CronNextTrigger: nextTrigger,
			})
			if err != nil {
				slog.Error("create gc blob rule failed", "err", err)
				return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc blob rule failed: %v", err))
			}
			return nil
		}
		err = repoDaemon.UpdateGcRule(ctx, ruleObj.ID, updates)
		if err != nil {
			slog.Error("update gc blob rule failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update gc blob rule failed: %v", err))
		}
		err = s.Producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
			Action:       enums.WebhookActionUpdate,
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcBlobRule,
			Payload:      utils.MustMarshal(req),
		})
		if err != nil {
			slog.Error("webhook event produce failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
}

// GetGcBlobRule gets the global GC blob rule
func (s *service) GetGcBlobRule(ctx context.Context) (*models.DaemonGcBlobRule, error) {
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcBlob, nil)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc blob rule not found", "err", err)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc blob rule not found: %v", err))
		}
		slog.Error("get gc blob rule failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc blob rule failed: %v", err))
	}
	return ruleObj, nil
}

// GetGcBlobLatestRunner gets the latest global GC blob runner after validating namespace scope
func (s *service) GetGcBlobLatestRunner(ctx context.Context, namespaceID string) (*models.DaemonGcBlobRunner, error) {
	if namespaceID != "" {
		slog.Error("namespaceID should always be 0 in action GetGcBlobLatestRunner")
		return nil, errcode.HTTPErrCodeUnauthorized.Detail("NamespaceID should always be 0 in action GetGcBlobLatestRunner")
	}
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcBlob, nil)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc blob rule not found", "err", err)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc blob rule not found: %v", err))
		}
		slog.Error("get gc blob rule failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc blob rule failed: %v", err))
	}
	runnerObj, err := s.RepoDaemon.GetGcLatestRunner(ctx, ruleObj.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc blob latest runner not found", "err", err)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc blob latest runner not found: %v", err))
		}
		slog.Error("get gc blob latest runner failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc blob latest runner failed: %v", err))
	}
	return runnerObj, nil
}

// CreateGcBlobRunner creates a global GC blob runner and queues the GC task and webhook event
func (s *service) CreateGcBlobRunner(ctx context.Context, userID string, req api.CreateGcBlobRunnerRequest) error {
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcBlob, nil)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc blob rule not found", "err", err)
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc blob rule not found: %v", err))
		}
		slog.Error("get gc blob rule failed", "err", err)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc blob rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		slog.Error("the gc blob rule is running")
		return errcode.HTTPErrCodeBadRequest.Detail("The gc blob rule is running")
	}
	return query.Q.Transaction(func(tx *query.Query) error {
		runnerObj := &models.DaemonGcBlobRunner{ID: uuid.NewV7String(), RuleID: ruleObj.ID, Status: enums.TaskCommonStatusPending,
			OperateType:   enums.OperateTypeManual,
			OperateUserID: new(userID)}
		err = s.RepoDaemon.CreateGcRunner(ctx, runnerObj)
		if err != nil {
			slog.Error(fmt.Sprintf("Create gc blob runner failed: %v", err), "RuleID", ruleObj.ID)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc blob runner failed: %v", err))
		}
		err = s.Producer.Produce(ctx, enums.DaemonGcBlob,
			api.DaemonGcPayload{RunnerID: runnerObj.ID})
		if err != nil {
			slog.Error(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonGcBlob.String()), "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonGcBlob.String()))
		}
		err = s.Producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
			Action:       enums.WebhookActionCreate,
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcBlobRunner,
			Payload:      utils.MustMarshal(req),
		})
		if err != nil {
			slog.Error("webhook event produce failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
}

// ListGcBlobRunners lists global GC blob runners
func (s *service) ListGcBlobRunners(ctx context.Context, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcBlobRunner, int64, error) {
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcBlob, nil)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc blob rule not found", "err", err)
			return nil, 0, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc blob rule not found: %v", err))
		}
		slog.Error("get gc blob rule failed", "err", err)
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc blob rule failed: %v", err))
	}
	runnerObjs, total, err := s.RepoDaemon.ListGcRunners(ctx, ruleObj.ID, pagination, sort)
	if err != nil {
		slog.Error("list gc blob rule failed", "err", err)
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List gc blob rule failed: %v", err))
	}
	return runnerObjs, total, nil
}

// GetGcBlobRunner gets a global GC blob runner by runner ID
func (s *service) GetGcBlobRunner(ctx context.Context, runnerID string) (*models.DaemonGcBlobRunner, error) {
	runnerObj, err := s.RepoDaemon.GetGcRunner(ctx, runnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc tag runner not found", "err", err, "runnerID", runnerID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc tag runner not found: %v", err))
		}
		slog.Error("get gc tag runner failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc tag runner failed: %v", err))
	}
	return runnerObj, nil
}

// ListGcBlobRecords lists global GC blob records by runner ID
func (s *service) ListGcBlobRecords(ctx context.Context, runnerID string, pagination api.Pagination, sort api.Sortable) ([]*models.DaemonGcBlobRecord, int64, error) {
	recordObjs, total, err := s.RepoDaemon.ListGcRecords(ctx, runnerID, pagination, sort)
	if err != nil {
		slog.Error("list gc blob records failed", "err", err, "RuleID", runnerID)
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List gc blob records failed: %v", err))
	}
	return recordObjs, total, nil
}

// GetGcBlobRecord gets a global GC blob record by runner ID and record ID
func (s *service) GetGcBlobRecord(ctx context.Context, runnerID, recordID string) (*models.DaemonGcBlobRecord, error) {
	ruleObj, err := s.RepoDaemon.GetGcRule(ctx, enums.DaemonGcBlob, nil)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc blob rule not found", "err", err)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc blob rule not found: %v", err))
		}
		slog.Error("get gc blob rule failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc blob rule failed: %v", err))
	}
	recordObj, err := s.RepoDaemon.GetGcRecord(ctx, recordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get gc blob record not found", "err", err, "runnerID", runnerID)
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc blob record not found: %v", err))
		}
		slog.Error("get gc blob record failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get gc blob record failed: %v", err))
	}
	if recordObj.Runner.ID != runnerID || recordObj.Runner.Rule.ID != ruleObj.ID {
		slog.Error("get gc blob record not found", "err", err, "runnerID", runnerID)
		return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get gc blob record not found: %v", err))
	}
	return recordObj, nil
}
