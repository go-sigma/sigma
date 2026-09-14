// Copyright 2026 sigma
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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repodaemon "github.com/go-sigma/sigma/pkg/dal/repository/daemon"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

var testNamespace = "namespace-1"

func requireErrCode(t *testing.T, err error, want string) {
	t.Helper()
	code, ok := errcode.AsType[errcode.ErrCode](err)
	require.True(t, ok)
	require.Equal(t, want, code.Code)
}

func TestParseCronNextTrigger(t *testing.T) {
	require.Nil(t, parseCronNextTrigger(nil))

	rule := "* * * * *"
	before := time.Now().UnixMilli()
	next := parseCronNextTrigger(&rule)
	require.NotNil(t, next)
	require.Greater(t, *next, before)
	require.LessOrEqual(t, *next, time.Now().Add(61*time.Second).UnixMilli())
}

func TestGetGcRule(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		name   string
		daemon enums.Daemon
		call   func(*service) (*models.DaemonGcRule, error)
	}{
		{"artifact", enums.DaemonGcArtifact, func(s *service) (*models.DaemonGcRule, error) { return s.GetGcArtifactRule(ctx, testNamespace) }},
		{"repository", enums.DaemonGcRepository, func(s *service) (*models.DaemonGcRule, error) { return s.GetGcRepositoryRule(ctx, testNamespace) }},
		{"tag", enums.DaemonGcTag, func(s *service) (*models.DaemonGcRule, error) { return s.GetGcTagRule(ctx, testNamespace) }},
		{"blob", enums.DaemonGcBlob, func(s *service) (*models.DaemonGcRule, error) { return s.GetGcBlobRule(ctx) }},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_success", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(&models.DaemonGcRule{ID: "rule-1"}, nil)

			got, err := tt.call(&service{RepoDaemon: repo})
			require.NoError(t, err)
			require.Equal(t, "rule-1", got.ID)
		})
		t.Run(tt.name+"_not_found", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(nil, gorm.ErrRecordNotFound)

			_, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeNotFound.Code)
		})
		t.Run(tt.name+"_internal_error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(nil, errors.New("db error"))

			_, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeInternalError.Code)
		})
	}
}

func TestGetGcLatestRunner(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		name   string
		daemon enums.Daemon
		call   func(*service) (*models.DaemonGcRunner, error)
	}{
		{"artifact", enums.DaemonGcArtifact, func(s *service) (*models.DaemonGcRunner, error) {
			return s.GetGcArtifactLatestRunner(ctx, testNamespace)
		}},
		{"repository", enums.DaemonGcRepository, func(s *service) (*models.DaemonGcRunner, error) {
			return s.GetGcRepositoryLatestRunner(ctx, testNamespace)
		}},
		{"tag", enums.DaemonGcTag, func(s *service) (*models.DaemonGcRunner, error) { return s.GetGcTagLatestRunner(ctx, testNamespace) }},
		{"blob", enums.DaemonGcBlob, func(s *service) (*models.DaemonGcRunner, error) { return s.GetGcBlobLatestRunner(ctx, "") }},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_success", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(&models.DaemonGcRule{ID: "rule-1"}, nil)
			repo.EXPECT().GetGcLatestRunner(gomock.Any(), "rule-1").Return(&models.DaemonGcRunner{ID: "runner-1"}, nil)

			got, err := tt.call(&service{RepoDaemon: repo})
			require.NoError(t, err)
			require.Equal(t, "runner-1", got.ID)
		})
		t.Run(tt.name+"_rule_not_found", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(nil, gorm.ErrRecordNotFound)

			_, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeNotFound.Code)
		})
		t.Run(tt.name+"_rule_internal_error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(nil, errors.New("db error"))

			_, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeInternalError.Code)
		})
		t.Run(tt.name+"_runner_not_found", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(&models.DaemonGcRule{ID: "rule-1"}, nil)
			repo.EXPECT().GetGcLatestRunner(gomock.Any(), "rule-1").Return(nil, gorm.ErrRecordNotFound)

			_, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeNotFound.Code)
		})
		t.Run(tt.name+"_runner_internal_error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(&models.DaemonGcRule{ID: "rule-1"}, nil)
			repo.EXPECT().GetGcLatestRunner(gomock.Any(), "rule-1").Return(nil, errors.New("db error"))

			_, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeInternalError.Code)
		})
	}
}

func TestGetGcRunner(t *testing.T) {
	ctx := t.Context()
	otherNS := "other-namespace"
	tests := []struct {
		name  string
		call  func(*service) (*models.DaemonGcRunner, error)
		hasNS bool
	}{
		{"artifact", func(s *service) (*models.DaemonGcRunner, error) {
			return s.GetGcArtifactRunner(ctx, testNamespace, "runner-1")
		}, true},
		{"repository", func(s *service) (*models.DaemonGcRunner, error) {
			return s.GetGcRepositoryRunner(ctx, testNamespace, "runner-1")
		}, true},
		{"tag", func(s *service) (*models.DaemonGcRunner, error) {
			return s.GetGcTagRunner(ctx, testNamespace, "runner-1")
		}, true},
		{"blob", func(s *service) (*models.DaemonGcRunner, error) { return s.GetGcBlobRunner(ctx, "runner-1") }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_success", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			runner := &models.DaemonGcRunner{ID: "runner-1"}
			if tt.hasNS {
				runner.Rule = models.DaemonGcRule{NamespaceID: &testNamespace}
			}
			repo.EXPECT().GetGcRunner(gomock.Any(), "runner-1").Return(runner, nil)

			got, err := tt.call(&service{RepoDaemon: repo})
			require.NoError(t, err)
			require.Equal(t, "runner-1", got.ID)
		})
		t.Run(tt.name+"_not_found", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRunner(gomock.Any(), "runner-1").Return(nil, gorm.ErrRecordNotFound)

			_, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeNotFound.Code)
		})
		t.Run(tt.name+"_internal_error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRunner(gomock.Any(), "runner-1").Return(nil, errors.New("db error"))

			_, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeInternalError.Code)
		})
		if tt.hasNS {
			t.Run(tt.name+"_namespace_mismatch", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				repo := repodaemon.NewMockDaemonRepository(ctrl)
				repo.EXPECT().GetGcRunner(gomock.Any(), "runner-1").Return(&models.DaemonGcRunner{ID: "runner-1", Rule: models.DaemonGcRule{NamespaceID: &otherNS}}, nil)

				_, err := tt.call(&service{RepoDaemon: repo})
				requireErrCode(t, err, errcode.HTTPErrCodeNotFound.Code)
			})
		}
	}
}

func TestListGcRunners(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		name   string
		daemon enums.Daemon
		call   func(*service) ([]*models.DaemonGcRunner, int64, error)
	}{
		{"artifact", enums.DaemonGcArtifact, func(s *service) ([]*models.DaemonGcRunner, int64, error) {
			return s.ListGcArtifactRunners(ctx, testNamespace, api.Pagination{}, api.Sortable{})
		}},
		{"repository", enums.DaemonGcRepository, func(s *service) ([]*models.DaemonGcRunner, int64, error) {
			return s.ListGcRepositoryRunners(ctx, testNamespace, api.Pagination{}, api.Sortable{})
		}},
		{"tag", enums.DaemonGcTag, func(s *service) ([]*models.DaemonGcRunner, int64, error) {
			return s.ListGcTagRunners(ctx, testNamespace, api.Pagination{}, api.Sortable{})
		}},
		{"blob", enums.DaemonGcBlob, func(s *service) ([]*models.DaemonGcRunner, int64, error) {
			return s.ListGcBlobRunners(ctx, api.Pagination{}, api.Sortable{})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_success", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(&models.DaemonGcRule{ID: "rule-1"}, nil)
			repo.EXPECT().ListGcRunners(gomock.Any(), "rule-1", gomock.Any(), gomock.Any()).Return([]*models.DaemonGcRunner{{ID: "runner-1"}}, int64(1), nil)

			got, total, err := tt.call(&service{RepoDaemon: repo})
			require.NoError(t, err)
			require.Equal(t, int64(1), total)
			require.Len(t, got, 1)
		})
		t.Run(tt.name+"_rule_not_found", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(nil, gorm.ErrRecordNotFound)

			_, _, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeNotFound.Code)
		})
		t.Run(tt.name+"_rule_internal_error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(nil, errors.New("db error"))

			_, _, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeInternalError.Code)
		})
		t.Run(tt.name+"_list_error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(&models.DaemonGcRule{ID: "rule-1"}, nil)
			repo.EXPECT().ListGcRunners(gomock.Any(), "rule-1", gomock.Any(), gomock.Any()).Return(nil, int64(0), errors.New("db error"))

			_, _, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeInternalError.Code)
		})
	}
}

func TestListGcRecords(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		name string
		call func(*service) ([]*models.DaemonGcRecord, int64, error)
	}{
		{"artifact", func(s *service) ([]*models.DaemonGcRecord, int64, error) {
			return s.ListGcArtifactRecords(ctx, "runner-1", api.Pagination{}, api.Sortable{})
		}},
		{"repository", func(s *service) ([]*models.DaemonGcRecord, int64, error) {
			return s.ListGcRepositoryRecords(ctx, "runner-1", api.Pagination{}, api.Sortable{})
		}},
		{"tag", func(s *service) ([]*models.DaemonGcRecord, int64, error) {
			return s.ListGcTagRecords(ctx, "runner-1", api.Pagination{}, api.Sortable{})
		}},
		{"blob", func(s *service) ([]*models.DaemonGcRecord, int64, error) {
			return s.ListGcBlobRecords(ctx, "runner-1", api.Pagination{}, api.Sortable{})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_success", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().ListGcRecords(gomock.Any(), "runner-1", gomock.Any(), gomock.Any()).Return([]*models.DaemonGcRecord{{ID: "record-1"}}, int64(1), nil)

			got, total, err := tt.call(&service{RepoDaemon: repo})
			require.NoError(t, err)
			require.Equal(t, int64(1), total)
			require.Len(t, got, 1)
		})
		t.Run(tt.name+"_list_error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().ListGcRecords(gomock.Any(), "runner-1", gomock.Any(), gomock.Any()).Return(nil, int64(0), errors.New("db error"))

			_, _, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeInternalError.Code)
		})
	}
}

func TestGetGcRecord(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		name   string
		daemon enums.Daemon
		call   func(*service) (*models.DaemonGcRecord, error)
	}{
		{"artifact", enums.DaemonGcArtifact, func(s *service) (*models.DaemonGcRecord, error) {
			return s.GetGcArtifactRecord(ctx, testNamespace, "runner-1", "record-1")
		}},
		{"repository", enums.DaemonGcRepository, func(s *service) (*models.DaemonGcRecord, error) {
			return s.GetGcRepositoryRecord(ctx, testNamespace, "runner-1", "record-1")
		}},
		{"tag", enums.DaemonGcTag, func(s *service) (*models.DaemonGcRecord, error) {
			return s.GetGcTagRecord(ctx, testNamespace, "runner-1", "record-1")
		}},
		{"blob", enums.DaemonGcBlob, func(s *service) (*models.DaemonGcRecord, error) {
			return s.GetGcBlobRecord(ctx, "runner-1", "record-1")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_success", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			record := &models.DaemonGcRecord{
				ID:     "record-1",
				Runner: models.DaemonGcRunner{ID: "runner-1", Rule: models.DaemonGcRule{ID: "rule-1"}},
			}
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(&models.DaemonGcRule{ID: "rule-1"}, nil)
			repo.EXPECT().GetGcRecord(gomock.Any(), "record-1").Return(record, nil)

			got, err := tt.call(&service{RepoDaemon: repo})
			require.NoError(t, err)
			require.Equal(t, "record-1", got.ID)
		})
		t.Run(tt.name+"_rule_not_found", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(nil, gorm.ErrRecordNotFound)

			_, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeNotFound.Code)
		})
		t.Run(tt.name+"_record_not_found", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(&models.DaemonGcRule{ID: "rule-1"}, nil)
			repo.EXPECT().GetGcRecord(gomock.Any(), "record-1").Return(nil, gorm.ErrRecordNotFound)

			_, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeNotFound.Code)
		})
		t.Run(tt.name+"_mismatch", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			record := &models.DaemonGcRecord{
				ID:     "record-1",
				Runner: models.DaemonGcRunner{ID: "other-runner", Rule: models.DaemonGcRule{ID: "rule-1"}},
			}
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(&models.DaemonGcRule{ID: "rule-1"}, nil)
			repo.EXPECT().GetGcRecord(gomock.Any(), "record-1").Return(record, nil)

			_, err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeNotFound.Code)
		})
	}
}

func TestUpdateGcRuleRejectsRunning(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		name   string
		daemon enums.Daemon
		call   func(*service) error
	}{
		{"artifact", enums.DaemonGcArtifact, func(s *service) error {
			return s.UpdateGcArtifactRule(ctx, api.UpdateGcArtifactRuleRequest{NamespaceID: testNamespace})
		}},
		{"repository", enums.DaemonGcRepository, func(s *service) error {
			return s.UpdateGcRepositoryRule(ctx, api.UpdateGcRepositoryRuleRequest{NamespaceID: testNamespace})
		}},
		{"tag", enums.DaemonGcTag, func(s *service) error {
			return s.UpdateGcTagRule(ctx, api.UpdateGcTagRuleRequest{NamespaceID: testNamespace})
		}},
		{"blob", enums.DaemonGcBlob, func(s *service) error { return s.UpdateGcBlobRule(ctx, api.UpdateGcBlobRuleRequest{}) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(&models.DaemonGcRule{ID: "rule-1", IsRunning: true}, nil)

			err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeBadRequest.Code)
		})
	}
}

func TestUpdateGcRuleInternalError(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		name   string
		daemon enums.Daemon
		call   func(*service) error
	}{
		{"artifact", enums.DaemonGcArtifact, func(s *service) error {
			return s.UpdateGcArtifactRule(ctx, api.UpdateGcArtifactRuleRequest{NamespaceID: testNamespace})
		}},
		{"repository", enums.DaemonGcRepository, func(s *service) error {
			return s.UpdateGcRepositoryRule(ctx, api.UpdateGcRepositoryRuleRequest{NamespaceID: testNamespace})
		}},
		{"tag", enums.DaemonGcTag, func(s *service) error {
			return s.UpdateGcTagRule(ctx, api.UpdateGcTagRuleRequest{NamespaceID: testNamespace})
		}},
		{"blob", enums.DaemonGcBlob, func(s *service) error { return s.UpdateGcBlobRule(ctx, api.UpdateGcBlobRuleRequest{}) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(nil, errors.New("db error"))

			err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeInternalError.Code)
		})
	}
}

func TestUpdateGcBlobRuleRejectsNamespace(t *testing.T) {
	err := (&service{}).UpdateGcBlobRule(t.Context(), api.UpdateGcBlobRuleRequest{NamespaceID: testNamespace})
	requireErrCode(t, err, errcode.HTTPErrCodeUnauthorized.Code)
}

func TestGetGcBlobLatestRunnerRejectsNamespace(t *testing.T) {
	_, err := (&service{}).GetGcBlobLatestRunner(t.Context(), testNamespace)
	requireErrCode(t, err, errcode.HTTPErrCodeUnauthorized.Code)
}

func TestCreateGcRunnerErrorMapping(t *testing.T) {
	ctx := t.Context()
	tests := []struct {
		name   string
		daemon enums.Daemon
		call   func(*service) error
	}{
		{"artifact", enums.DaemonGcArtifact, func(s *service) error {
			return s.CreateGcArtifactRunner(ctx, api.CreateGcArtifactRunnerRequest{NamespaceID: testNamespace})
		}},
		{"repository", enums.DaemonGcRepository, func(s *service) error {
			return s.CreateGcRepositoryRunner(ctx, api.CreateGcRepositoryRunnerRequest{NamespaceID: testNamespace})
		}},
		{"tag", enums.DaemonGcTag, func(s *service) error {
			return s.CreateGcTagRunner(ctx, api.CreateGcTagRunnerRequest{NamespaceID: testNamespace})
		}},
		{"blob", enums.DaemonGcBlob, func(s *service) error { return s.CreateGcBlobRunner(ctx, "user-1", api.CreateGcBlobRunnerRequest{}) }},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_not_found", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(nil, gorm.ErrRecordNotFound)

			err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeNotFound.Code)
		})
		t.Run(tt.name+"_internal_error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(nil, errors.New("db error"))

			err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeInternalError.Code)
		})
		t.Run(tt.name+"_running", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repodaemon.NewMockDaemonRepository(ctrl)
			repo.EXPECT().GetGcRule(gomock.Any(), tt.daemon, gomock.Any()).Return(&models.DaemonGcRule{ID: "rule-1", IsRunning: true}, nil)

			err := tt.call(&service{RepoDaemon: repo})
			requireErrCode(t, err, errcode.HTTPErrCodeBadRequest.Code)
		})
	}
}
