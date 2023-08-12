// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/go-sigma/sigma/pkg/dal/dao (interfaces: BuilderService)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/go-sigma/sigma/pkg/dal/models"
	gomock "go.uber.org/mock/gomock"
)

// MockBuilderService is a mock of BuilderService interface.
type MockBuilderService struct {
	ctrl     *gomock.Controller
	recorder *MockBuilderServiceMockRecorder
}

// MockBuilderServiceMockRecorder is the mock recorder for MockBuilderService.
type MockBuilderServiceMockRecorder struct {
	mock *MockBuilderService
}

// NewMockBuilderService creates a new mock instance.
func NewMockBuilderService(ctrl *gomock.Controller) *MockBuilderService {
	mock := &MockBuilderService{ctrl: ctrl}
	mock.recorder = &MockBuilderServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBuilderService) EXPECT() *MockBuilderServiceMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockBuilderService) Create(arg0 context.Context, arg1 *models.Builder) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockBuilderServiceMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockBuilderService)(nil).Create), arg0, arg1)
}

// CreateRunner mocks base method.
func (m *MockBuilderService) CreateRunner(arg0 context.Context, arg1 *models.BuilderRunner) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateRunner", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateRunner indicates an expected call of CreateRunner.
func (mr *MockBuilderServiceMockRecorder) CreateRunner(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateRunner", reflect.TypeOf((*MockBuilderService)(nil).CreateRunner), arg0, arg1)
}

// GetByRepositoryID mocks base method.
func (m *MockBuilderService) GetByRepositoryID(arg0 context.Context, arg1 int64) (*models.Builder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByRepositoryID", arg0, arg1)
	ret0, _ := ret[0].(*models.Builder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByRepositoryID indicates an expected call of GetByRepositoryID.
func (mr *MockBuilderServiceMockRecorder) GetByRepositoryID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByRepositoryID", reflect.TypeOf((*MockBuilderService)(nil).GetByRepositoryID), arg0, arg1)
}

// GetRunner mocks base method.
func (m *MockBuilderService) GetRunner(arg0 context.Context, arg1 int64) (*models.BuilderRunner, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRunner", arg0, arg1)
	ret0, _ := ret[0].(*models.BuilderRunner)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRunner indicates an expected call of GetRunner.
func (mr *MockBuilderServiceMockRecorder) GetRunner(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRunner", reflect.TypeOf((*MockBuilderService)(nil).GetRunner), arg0, arg1)
}

// UpdateRunner mocks base method.
func (m *MockBuilderService) UpdateRunner(arg0 context.Context, arg1, arg2 int64, arg3 map[string]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateRunner", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateRunner indicates an expected call of UpdateRunner.
func (mr *MockBuilderServiceMockRecorder) UpdateRunner(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateRunner", reflect.TypeOf((*MockBuilderService)(nil).UpdateRunner), arg0, arg1, arg2, arg3)
}