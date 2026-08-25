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
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	buildlogger "github.com/go-sigma/sigma/pkg/background/build/logger"
	"github.com/go-sigma/sigma/pkg/background/build/runtime"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repobuilder "github.com/go-sigma/sigma/pkg/dal/repository/builder"
	repocoderepo "github.com/go-sigma/sigma/pkg/dal/repository/coderepo"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/compress"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

//go:generate mockgen -mock_names Service=MockBuilderService -destination=builders_mocks.go -package=builders github.com/go-sigma/sigma/pkg/service/builders Service

// Service encapsulates builder-related business logic.
type Service interface {
	// CreateBuilder creates a builder (includes: validate + check existing + compress dockerfile + resolve credentials + create).
	CreateBuilder(ctx context.Context, userID string, req api.CreateBuilderRequest) error
	// UpdateBuilder updates a builder (includes: compress dockerfile + resolve credentials + update).
	UpdateBuilder(ctx context.Context, userID string, builderID string, req api.UpdateBuilderRequest) error
	// GetBuilderByRepositoryID gets a builder by repository ID.
	GetBuilderByRepositoryID(ctx context.Context, repositoryID string) (*models.Builder, error)
	// GetRunner gets a runner by ID.
	GetRunner(ctx context.Context, runnerID string) (*models.BuilderRunner, error)
	// ListRunners lists builder runners with pagination.
	ListRunners(ctx context.Context, builderID string, pagination api.Pagination, sort api.Sortable) ([]*models.BuilderRunner, int64, error)
	// RunRunner creates a new runner and produces a build start event.
	RunRunner(ctx context.Context, req api.PostRunnerRun) (string, error)
	// RerunRunner creates a new runner from an existing one and produces a build start event.
	RerunRunner(ctx context.Context, req api.GetRunnerStop) (string, error)
	// StopRunner updates runner status to stopping and produces a build stop event.
	StopRunner(ctx context.Context, req api.GetRunnerStop) error
	// RunnerLogReader returns a log reader for the runner and whether the stream is gzip-compressed.
	RunnerLogReader(ctx context.Context, builderID string, runnerID string, status enums.BuildStatus) (io.Reader, bool, error)
}

type service struct {
	dig.In

	RepoBuilder  repobuilder.BuilderRepository
	RepoRegistry reporegistry.RepositoryRepository
	RepoCode     repocoderepo.CodeRepositoryRepository
	Producer     workq.Producer
}

func NewService(params service) Service {
	return &params
}

// createBuilderValidator validates the create builder request.
func createBuilderValidator(req api.CreateBuilderRequest) error {
	switch req.Source {
	case enums.BuilderSourceSelfCodeRepository:
		if req.ScmCredentialType == nil {
			slog.Error("scmCredentialType cannot be nil", "ScmCredentialType", ptr.To(req.ScmCredentialType))
			return errcode.HTTPErrCodeBadRequest.Detail("parameter 'scm credential_type' is invalid")
		}
		switch ptr.To(req.ScmCredentialType) {
		case enums.ScmCredentialTypeNone:
		case enums.ScmCredentialTypeUsername:
			if len(ptr.To(req.ScmUsername)) == 0 || len(ptr.To(req.ScmPassword)) == 0 {
				slog.Error("scmUsername and ScmPassword cannot be nil", "ScmUsername", ptr.To(req.ScmUsername), "ScmPassword", ptr.To(req.ScmPassword))
				return errcode.HTTPErrCodeBadRequest.Detail("parameter 'scm_username' or 'scm_password' is invalid")
			}
		case enums.ScmCredentialTypeToken:
			if len(ptr.To(req.ScmToken)) == 0 {
				slog.Error("scmToken cannot be nil", "ScmToken", ptr.To(req.ScmToken))
				return errcode.HTTPErrCodeBadRequest.Detail("parameter 'scm_token' is invalid")
			}
		case enums.ScmCredentialTypeSsh:
			if len(ptr.To(req.ScmSshKey)) == 0 {
				slog.Error("scmSshKey cannot be nil", "ScmSshKey", ptr.To(req.ScmSshKey))
				return errcode.HTTPErrCodeBadRequest.Detail("parameter 'scm_ssh_key' is invalid")
			}
		default:
			slog.Error("scmCredentialType cannot be nil", "ScmCredentialType", ptr.To(req.ScmCredentialType))
			return errcode.HTTPErrCodeBadRequest.Detail("parameter 'scm credential_type' is invalid")
		}
	case enums.BuilderSourceCodeRepository:
		if req.CodeRepositoryID == nil {
			slog.Error("codeRepositoryID cannot be nil")
			return errcode.HTTPErrCodeBadRequest.Detail("parameter 'code_repository_id' is invalid")
		}
	case enums.BuilderSourceDockerfile:
		if req.Dockerfile == nil || len(ptr.To(req.Dockerfile)) == 0 {
			slog.Error("dockerfile cannot be nil", "Dockerfile", ptr.To(req.Dockerfile))
			return errcode.HTTPErrCodeBadRequest.Detail("parameter 'dockerfile' is invalid")
		}
	default:
		slog.Error("source is invalid", "Source", string(req.Source))
		return errcode.HTTPErrCodeBadRequest.Detail("parameter 'source' is invalid")
	}
	return nil
}

func (s *service) RunnerLogReader(ctx context.Context, builderID string, runnerID string, status enums.BuildStatus) (io.Reader, bool, error) {
	if status == enums.BuildStatusFailed || status == enums.BuildStatusSuccess {
		if buildlogger.LogStoreDriver == nil {
			return strings.NewReader(""), true, nil
		}
		reader, err := buildlogger.LogStoreDriver.Read(ctx, runnerID)
		return reader, true, err
	}

	if status == enums.BuildStatusBuilding {
		reader, writer := io.Pipe()
		go func() {
			defer func() {
				if err := writer.Close(); err != nil {
					slog.Error("close log pipe writer failed", "err", err)
				}
			}()
			if err := runtime.Driver.LogStream(ctx, builderID, runnerID, writer); err != nil {
				slog.Error("read log failed", "err", err)
				if _, writeErr := writer.Write([]byte{10}); writeErr != nil {
					slog.Error("write log failed", "err", writeErr)
				}
			}
		}()
		return reader, false, nil
	}

	return strings.NewReader(""), false, nil
}

// compressDockerfile decodes a base64 dockerfile and compresses it.
func compressDockerfile(str *string) ([]byte, error) {
	if str == nil {
		return nil, nil
	}
	data, err := base64.StdEncoding.DecodeString(ptr.To(str))
	if err != nil {
		return nil, err
	}
	return compress.CompressBytes(data)
}

func (s *service) CreateBuilder(ctx context.Context, userID string, req api.CreateBuilderRequest) error {
	if err := createBuilderValidator(req); err != nil {
		return err
	}

	_, err := s.RepoRegistry.Get(ctx, req.RepositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("repository not found", "err", err, "id", req.RepositoryID)
			return errcode.HTTPErrCodeNotFound.Detail("Repository not found")
		}
		slog.Error("repository find failed", "err", err, "id", req.RepositoryID)
		return errcode.HTTPErrCodeInternalError.Detail("Repository find failed")
	}

	_, err = s.RepoBuilder.GetByRepositoryID(ctx, req.RepositoryID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get builder by repository id failed", "err", err, "id", req.RepositoryID)
		return errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}
	if err == nil {
		slog.Error("repository has been already create builder", "err", err, "id", req.RepositoryID)
		return errcode.HTTPErrCodeConflict.Detail("Repository has been already create builder")
	}

	compressedDockerfile, err := compressDockerfile(req.Dockerfile)
	if err != nil {
		slog.Error("dockerfile base64 decode failed", "err", err)
		return errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("Dockerfile base64 decode failed: %v", err))
	}

	builderObj := &models.Builder{
		ID:           uuid.NewV7String(),
		RepositoryID: req.RepositoryID,

		Source: req.Source,

		CodeRepositoryID: req.CodeRepositoryID,

		Dockerfile: compressedDockerfile,

		ScmRepository:     req.ScmRepository,
		ScmCredentialType: req.ScmCredentialType,
		ScmToken:          req.ScmToken,
		ScmSshKey:         req.ScmSshKey,
		ScmUsername:       req.ScmUsername,
		ScmPassword:       req.ScmPassword,

		ScmBranch: req.ScmBranch,

		ScmDepth:     req.ScmDepth,
		ScmSubmodule: req.ScmSubmodule,

		CronRule:        req.CronRule,
		CronBranch:      req.CronBranch,
		CronTagTemplate: req.CronTagTemplate,

		WebhookBranchName:        req.WebhookBranchName,
		WebhookBranchTagTemplate: req.WebhookBranchTagTemplate,
		WebhookTagTagTemplate:    req.WebhookTagTagTemplate,

		BuildkitInsecureRegistries: strings.Join(req.BuildkitInsecureRegistries, ","),
		BuildkitContext:            req.BuildkitContext,
		BuildkitDockerfile:         req.BuildkitDockerfile,
		BuildkitPlatforms:          utils.StringsJoin(req.BuildkitPlatforms, ","),
	}
	if builderObj.Source == enums.BuilderSourceCodeRepository && req.ScmCredentialType == nil {
		codeRepositoryObj, err := s.RepoCode.Get(ctx, ptr.To(req.CodeRepositoryID))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("get code repository by id not found", "err", err, "CodeRepositoryID", ptr.To(req.CodeRepositoryID))
				return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get code repository by id(%s) not found: %v", ptr.To(req.CodeRepositoryID), err))
			}
			slog.Error("get code repository by id failed", "err", err, "CodeRepositoryID", ptr.To(req.CodeRepositoryID))
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get code repository by id(%s) failed: %v", ptr.To(req.CodeRepositoryID), err))
		}
		cloneCredentialObj, err := s.RepoCode.GetCloneCredential(ctx, codeRepositoryObj.User3rdPartyID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("get code repository clone credential failed", "err", err)
			} else {
				scmCredentialType := enums.ScmCredentialTypeToken
				builderObj.ScmCredentialType = &scmCredentialType
				builderObj.ScmToken = codeRepositoryObj.User3rdParty.Token
			}
		} else {
			builderObj.ScmCredentialType = new(cloneCredentialObj.Type)
			builderObj.ScmUsername = cloneCredentialObj.Username
			builderObj.ScmPassword = cloneCredentialObj.Password
			builderObj.ScmSshKey = cloneCredentialObj.SshKey
			builderObj.ScmToken = cloneCredentialObj.Token
		}
	}
	return query.Q.Transaction(func(tx *query.Query) error {
		repoBuilder := repobuilder.NewBuilderRepository(tx)
		err = repoBuilder.Create(ctx, builderObj)
		if err != nil {
			slog.Error("create builder for repository failed", "err", err, "id", req.RepositoryID)
			return errcode.HTTPErrCodeInternalError.Detail("Create builder for repository failed")
		}
		return nil
	})
}

func (s *service) UpdateBuilder(ctx context.Context, userID string, builderID string, req api.UpdateBuilderRequest) error {
	compressedDockerfile, err := compressDockerfile(req.Dockerfile)
	if err != nil {
		slog.Error("dockerfile base64 decode failed", "err", err)
		return errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("Dockerfile base64 decode failed: %v", err))
	}

	updates := map[string]any{
		query.Builder.Source.ColumnName().String():                     req.Source,
		query.Builder.CodeRepositoryID.ColumnName().String():           req.CodeRepositoryID,
		query.Builder.Dockerfile.ColumnName().String():                 compressedDockerfile,
		query.Builder.ScmRepository.ColumnName().String():              req.ScmRepository,
		query.Builder.ScmCredentialType.ColumnName().String():          req.ScmCredentialType,
		query.Builder.ScmSshKey.ColumnName().String():                  req.ScmSshKey,
		query.Builder.ScmToken.ColumnName().String():                   req.ScmToken,
		query.Builder.ScmUsername.ColumnName().String():                req.ScmUsername,
		query.Builder.ScmPassword.ColumnName().String():                req.ScmPassword,
		query.Builder.ScmBranch.ColumnName().String():                  req.ScmBranch,
		query.Builder.ScmDepth.ColumnName().String():                   req.ScmDepth,
		query.Builder.ScmSubmodule.ColumnName().String():               req.ScmSubmodule,
		query.Builder.CronRule.ColumnName().String():                   req.CronRule,
		query.Builder.CronBranch.ColumnName().String():                 req.CronBranch,
		query.Builder.CronTagTemplate.ColumnName().String():            req.CronTagTemplate,
		query.Builder.WebhookBranchName.ColumnName().String():          req.WebhookBranchName,
		query.Builder.WebhookBranchTagTemplate.ColumnName().String():   req.WebhookBranchTagTemplate,
		query.Builder.WebhookTagTagTemplate.ColumnName().String():      req.WebhookTagTagTemplate,
		query.Builder.BuildkitInsecureRegistries.ColumnName().String(): strings.Join(req.BuildkitInsecureRegistries, ","),
		query.Builder.BuildkitContext.ColumnName().String():            req.BuildkitContext,
		query.Builder.BuildkitDockerfile.ColumnName().String():         req.BuildkitDockerfile,
		query.Builder.BuildkitPlatforms.ColumnName().String():          utils.StringsJoin(req.BuildkitPlatforms, ","),
	}
	if req.Source == enums.BuilderSourceCodeRepository && req.ScmCredentialType == nil {
		codeRepositoryObj, err := s.RepoCode.Get(ctx, ptr.To(req.CodeRepositoryID))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("get code repository by id not found", "err", err, "CodeRepositoryID", ptr.To(req.CodeRepositoryID))
				return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get code repository by id(%s) not found: %v", ptr.To(req.CodeRepositoryID), err))
			}
			slog.Error("get code repository by id failed", "err", err, "CodeRepositoryID", ptr.To(req.CodeRepositoryID))
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get code repository by id(%s) failed: %v", ptr.To(req.CodeRepositoryID), err))
		}
		cloneCredentialObj, err := s.RepoCode.GetCloneCredential(ctx, codeRepositoryObj.User3rdPartyID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("get code repository clone credential failed", "err", err)
			} else {
				scmCredentialType := enums.ScmCredentialTypeToken
				updates[query.Builder.ScmCredentialType.ColumnName().String()] = &scmCredentialType
				updates[query.Builder.ScmToken.ColumnName().String()] = codeRepositoryObj.User3rdParty.Token
			}
		} else {
			updates[query.Builder.ScmCredentialType.ColumnName().String()] = new(cloneCredentialObj.Type)
			updates[query.Builder.ScmToken.ColumnName().String()] = cloneCredentialObj.Token
			updates[query.Builder.ScmUsername.ColumnName().String()] = cloneCredentialObj.Username
			updates[query.Builder.ScmPassword.ColumnName().String()] = cloneCredentialObj.Password
			updates[query.Builder.ScmSshKey.ColumnName().String()] = cloneCredentialObj.SshKey
		}
	}
	return query.Q.Transaction(func(tx *query.Query) error {
		builderRepository := repobuilder.NewBuilderRepository(tx)
		err = builderRepository.Update(ctx, builderID, updates)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("builder id not found", "err", err, "builder_id", builderID)
				return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Builder id(%s) not found", builderID))
			}
			slog.Error("builder find failed", "err", err, "builder_id", builderID)
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Builder id(%s) find failed: %v", builderID, err))
		}
		return nil
	})
}

func (s *service) GetBuilderByRepositoryID(ctx context.Context, repositoryID string) (*models.Builder, error) {
	builderObj, err := s.RepoBuilder.GetByRepositoryID(ctx, repositoryID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get builder by repository id failed", "err", err, "id", repositoryID)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get builder by repository id failed: %v", err))
	}
	return builderObj, err
}

func (s *service) GetRunner(ctx context.Context, runnerID string) (*models.BuilderRunner, error) {
	runnerObj, err := s.RepoBuilder.GetRunner(ctx, runnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("builder runner not found", "err", err)
			return nil, errcode.HTTPErrCodeNotFound.Detail("Builder runner not found")
		}
		slog.Error("builder runner find failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Builder runner find failed: %v", err))
	}
	return runnerObj, nil
}

func (s *service) ListRunners(ctx context.Context, builderID string, pagination api.Pagination, sort api.Sortable) ([]*models.BuilderRunner, int64, error) {
	return s.RepoBuilder.ListRunners(ctx, builderID, pagination, sort)
}

func (s *service) RunRunner(ctx context.Context, req api.PostRunnerRun) (string, error) {
	builderObj, err := s.RepoBuilder.GetByRepositoryID(ctx, req.RepositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get builder by repository id not found", "err", err, "id", req.RepositoryID)
			return "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get builder by repository id not found: %v", err))
		}
		slog.Error("get builder by repository id failed", "err", err, "id", req.RepositoryID)
		return "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get builder by repository id failed: %v", err))
	}
	if builderObj.ID != req.BuilderID {
		slog.Error("get builder by id failed", "builder_id", req.BuilderID, "builder_id", builderObj.ID)
		return "", errcode.HTTPErrCodeInternalError.Detail("Get builder by id failed")
	}

	var runnerObj *models.BuilderRunner
	err = query.Q.Transaction(func(tx *query.Query) error {
		repoBuilder := repobuilder.NewBuilderRepository(tx)
		runnerObj = &models.BuilderRunner{
			ID:        uuid.NewV7String(),
			BuilderID: req.BuilderID,
			RawTag:    req.RawTag,
			ScmBranch: req.ScmBranch,
		}
		err = repoBuilder.CreateRunner(ctx, runnerObj)
		if err != nil {
			slog.Error("create builder runner failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create builder runner failed: %v", err))
		}
		err = s.Producer.Produce(ctx, enums.DaemonBuilder, api.DaemonBuilderPayload{
			Action:       enums.DaemonBuilderActionStart,
			RepositoryID: req.RepositoryID,
			BuilderID:    req.BuilderID,
			RunnerID:     runnerObj.ID,
		})
		if err != nil {
			slog.Error(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonBuilder.String()), "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonBuilder.String()))
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return runnerObj.ID, nil
}

func (s *service) RerunRunner(ctx context.Context, req api.GetRunnerStop) (string, error) {
	builderObj, err := s.RepoBuilder.GetByRepositoryID(ctx, req.RepositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("get builder by repository id not found", "err", err, "id", req.RepositoryID)
			return "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get builder by repository id not found: %v", err))
		}
		slog.Error("get builder by repository id failed", "err", err, "id", req.RepositoryID)
		return "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get builder by repository id failed: %v", err))
	}
	if builderObj.ID != req.BuilderID {
		slog.Error("get builder by id failed", "builder_id", req.BuilderID, "builder_id", builderObj.ID)
		return "", errcode.HTTPErrCodeInternalError.Detail("Get builder by id failed")
	}

	runnerObj, err := s.RepoBuilder.GetRunner(ctx, req.RunnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("builder runner not found", "err", err)
			return "", errcode.HTTPErrCodeNotFound.Detail("Builder runner not found")
		}
		slog.Error("builder runner find failed", "err", err)
		return "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Builder runner find failed: %v", err))
	}

	if runnerObj.Status != enums.BuildStatusSuccess && runnerObj.Status != enums.BuildStatusFailed && runnerObj.Status != enums.BuildStatusStopped {
		slog.Error(fmt.Sprintf("Builder runner status %s not support rerun", runnerObj.Status.String()), "status", runnerObj.Status.String())
		return "", errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("Builder runner status %s not support rerun", runnerObj.Status.String()))
	}

	var newRunnerObj *models.BuilderRunner
	err = query.Q.Transaction(func(tx *query.Query) error {
		repoBuilder := repobuilder.NewBuilderRepository(tx)
		newRunnerObj = &models.BuilderRunner{
			ID:        uuid.NewV7String(),
			BuilderID: req.BuilderID,
			RawTag:    runnerObj.RawTag,
			ScmBranch: runnerObj.ScmBranch,
		}
		err = repoBuilder.CreateRunner(ctx, newRunnerObj)
		if err != nil {
			slog.Error("create builder runner failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create builder runner failed: %v", err))
		}
		err = s.Producer.Produce(ctx, enums.DaemonBuilder, api.DaemonBuilderPayload{
			Action:       enums.DaemonBuilderActionStart,
			RepositoryID: req.RepositoryID,
			BuilderID:    req.BuilderID,
			RunnerID:     newRunnerObj.ID,
		})
		if err != nil {
			slog.Error(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonBuilder.String()), "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonBuilder.String()))
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return newRunnerObj.ID, nil
}

func (s *service) StopRunner(ctx context.Context, req api.GetRunnerStop) error {
	builderObj, err := s.RepoBuilder.GetByRepositoryID(ctx, req.RepositoryID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get builder by repository id failed", "err", err, "id", req.RepositoryID)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get builder by repository id failed: %v", err))
	}
	if builderObj.ID != req.BuilderID {
		slog.Error("get builder by id failed", "builder_id", req.BuilderID, "builder_id", builderObj.ID)
		return errcode.HTTPErrCodeInternalError.Detail("Get builder by id failed")
	}

	runnerObj, err := s.RepoBuilder.GetRunner(ctx, req.RunnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("builder runner not found", "err", err)
			return errcode.HTTPErrCodeNotFound.Detail("Builder runner not found")
		}
		slog.Error("builder runner find failed", "err", err)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Builder runner find failed: %v", err))
	}

	if runnerObj.Status != enums.BuildStatusBuilding {
		slog.Error(fmt.Sprintf("Builder runner status %s not support stop", runnerObj.Status.String()), "status", runnerObj.Status.String())
		return errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("Builder runner status %s not support stop", runnerObj.Status.String()))
	}

	return query.Q.Transaction(func(tx *query.Query) error {
		repoBuilder := repobuilder.NewBuilderRepository(tx)
		err = repoBuilder.UpdateRunner(ctx, req.BuilderID, req.RunnerID, map[string]any{
			query.BuilderRunner.Status.ColumnName().String(): enums.BuildStatusStopping,
		})
		if err != nil {
			slog.Error("update runner status failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update runner status failed: %v", err))
		}
		err = s.Producer.Produce(ctx, enums.DaemonBuilder, api.DaemonBuilderPayload{
			Action:       enums.DaemonBuilderActionStop,
			RepositoryID: req.RepositoryID,
			BuilderID:    req.BuilderID,
			RunnerID:     req.RunnerID,
		})
		if err != nil {
			slog.Error(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonBuilder.String()), "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonBuilder.String()))
		}
		return nil
	})
}
