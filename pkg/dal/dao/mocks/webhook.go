// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/go-sigma/sigma/pkg/dal/dao (interfaces: WebhookService)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/go-sigma/sigma/pkg/dal/models"
	types "github.com/go-sigma/sigma/pkg/types"
	gomock "go.uber.org/mock/gomock"
)

// MockWebhookService is a mock of WebhookService interface.
type MockWebhookService struct {
	ctrl     *gomock.Controller
	recorder *MockWebhookServiceMockRecorder
}

// MockWebhookServiceMockRecorder is the mock recorder for MockWebhookService.
type MockWebhookServiceMockRecorder struct {
	mock *MockWebhookService
}

// NewMockWebhookService creates a new mock instance.
func NewMockWebhookService(ctrl *gomock.Controller) *MockWebhookService {
	mock := &MockWebhookService{ctrl: ctrl}
	mock.recorder = &MockWebhookServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWebhookService) EXPECT() *MockWebhookServiceMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockWebhookService) Create(arg0 context.Context, arg1 *models.Webhook) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockWebhookServiceMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockWebhookService)(nil).Create), arg0, arg1)
}

// CreateLog mocks base method.
func (m *MockWebhookService) CreateLog(arg0 context.Context, arg1 *models.WebhookLog) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateLog", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateLog indicates an expected call of CreateLog.
func (mr *MockWebhookServiceMockRecorder) CreateLog(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateLog", reflect.TypeOf((*MockWebhookService)(nil).CreateLog), arg0, arg1)
}

// DeleteByID mocks base method.
func (m *MockWebhookService) DeleteByID(arg0 context.Context, arg1 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteByID", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteByID indicates an expected call of DeleteByID.
func (mr *MockWebhookServiceMockRecorder) DeleteByID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteByID", reflect.TypeOf((*MockWebhookService)(nil).DeleteByID), arg0, arg1)
}

// Get mocks base method.
func (m *MockWebhookService) Get(arg0 context.Context, arg1 int64) (*models.Webhook, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1)
	ret0, _ := ret[0].(*models.Webhook)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockWebhookServiceMockRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockWebhookService)(nil).Get), arg0, arg1)
}

// GetByFilter mocks base method.
func (m *MockWebhookService) GetByFilter(arg0 context.Context, arg1 map[string]interface{}) ([]*models.Webhook, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByFilter", arg0, arg1)
	ret0, _ := ret[0].([]*models.Webhook)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByFilter indicates an expected call of GetByFilter.
func (mr *MockWebhookServiceMockRecorder) GetByFilter(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByFilter", reflect.TypeOf((*MockWebhookService)(nil).GetByFilter), arg0, arg1)
}

// GetLog mocks base method.
func (m *MockWebhookService) GetLog(arg0 context.Context, arg1 int64) (*models.WebhookLog, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLog", arg0, arg1)
	ret0, _ := ret[0].(*models.WebhookLog)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLog indicates an expected call of GetLog.
func (mr *MockWebhookServiceMockRecorder) GetLog(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLog", reflect.TypeOf((*MockWebhookService)(nil).GetLog), arg0, arg1)
}

// List mocks base method.
func (m *MockWebhookService) List(arg0 context.Context, arg1 *int64, arg2 types.Pagination, arg3 types.Sortable) ([]*models.Webhook, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*models.Webhook)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// List indicates an expected call of List.
func (mr *MockWebhookServiceMockRecorder) List(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockWebhookService)(nil).List), arg0, arg1, arg2, arg3)
}

// ListLogs mocks base method.
func (m *MockWebhookService) ListLogs(arg0 context.Context, arg1 int64, arg2 types.Pagination, arg3 types.Sortable) ([]*models.WebhookLog, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListLogs", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*models.WebhookLog)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ListLogs indicates an expected call of ListLogs.
func (mr *MockWebhookServiceMockRecorder) ListLogs(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListLogs", reflect.TypeOf((*MockWebhookService)(nil).ListLogs), arg0, arg1, arg2, arg3)
}

// UpdateByID mocks base method.
func (m *MockWebhookService) UpdateByID(arg0 context.Context, arg1 int64, arg2 map[string]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateByID", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateByID indicates an expected call of UpdateByID.
func (mr *MockWebhookServiceMockRecorder) UpdateByID(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateByID", reflect.TypeOf((*MockWebhookService)(nil).UpdateByID), arg0, arg1, arg2)
}
