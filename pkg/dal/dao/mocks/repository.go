// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ximager/ximager/pkg/dal/dao (interfaces: RepositoryService)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	models "github.com/ximager/ximager/pkg/dal/models"
	types "github.com/ximager/ximager/pkg/types"
	reflect "reflect"
)

// MockRepositoryService is a mock of RepositoryService interface
type MockRepositoryService struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryServiceMockRecorder
}

// MockRepositoryServiceMockRecorder is the mock recorder for MockRepositoryService
type MockRepositoryServiceMockRecorder struct {
	mock *MockRepositoryService
}

// NewMockRepositoryService creates a new mock instance
func NewMockRepositoryService(ctrl *gomock.Controller) *MockRepositoryService {
	mock := &MockRepositoryService{ctrl: ctrl}
	mock.recorder = &MockRepositoryServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRepositoryService) EXPECT() *MockRepositoryServiceMockRecorder {
	return m.recorder
}

// CountRepository mocks base method
func (m *MockRepositoryService) CountRepository(arg0 context.Context, arg1 types.ListRepositoryRequest) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CountRepository", arg0, arg1)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CountRepository indicates an expected call of CountRepository
func (mr *MockRepositoryServiceMockRecorder) CountRepository(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountRepository", reflect.TypeOf((*MockRepositoryService)(nil).CountRepository), arg0, arg1)
}

// Create mocks base method
func (m *MockRepositoryService) Create(arg0 context.Context, arg1 *models.Repository) (*models.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(*models.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create
func (mr *MockRepositoryServiceMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockRepositoryService)(nil).Create), arg0, arg1)
}

// DeleteByID mocks base method
func (m *MockRepositoryService) DeleteByID(arg0 context.Context, arg1 uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteByID", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteByID indicates an expected call of DeleteByID
func (mr *MockRepositoryServiceMockRecorder) DeleteByID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteByID", reflect.TypeOf((*MockRepositoryService)(nil).DeleteByID), arg0, arg1)
}

// Get mocks base method
func (m *MockRepositoryService) Get(arg0 context.Context, arg1 uint64) (*models.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1)
	ret0, _ := ret[0].(*models.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockRepositoryServiceMockRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockRepositoryService)(nil).Get), arg0, arg1)
}

// GetByName mocks base method
func (m *MockRepositoryService) GetByName(arg0 context.Context, arg1 string) (*models.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByName", arg0, arg1)
	ret0, _ := ret[0].(*models.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByName indicates an expected call of GetByName
func (mr *MockRepositoryServiceMockRecorder) GetByName(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByName", reflect.TypeOf((*MockRepositoryService)(nil).GetByName), arg0, arg1)
}

// ListByDtPagination mocks base method
func (m *MockRepositoryService) ListByDtPagination(arg0 context.Context, arg1 int, arg2 ...uint64) ([]*models.Repository, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListByDtPagination", varargs...)
	ret0, _ := ret[0].([]*models.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByDtPagination indicates an expected call of ListByDtPagination
func (mr *MockRepositoryServiceMockRecorder) ListByDtPagination(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByDtPagination", reflect.TypeOf((*MockRepositoryService)(nil).ListByDtPagination), varargs...)
}

// ListRepository mocks base method
func (m *MockRepositoryService) ListRepository(arg0 context.Context, arg1 types.ListRepositoryRequest) ([]*models.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListRepository", arg0, arg1)
	ret0, _ := ret[0].([]*models.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListRepository indicates an expected call of ListRepository
func (mr *MockRepositoryServiceMockRecorder) ListRepository(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListRepository", reflect.TypeOf((*MockRepositoryService)(nil).ListRepository), arg0, arg1)
}

// Save mocks base method
func (m *MockRepositoryService) Save(arg0 context.Context, arg1 *models.Repository) (*models.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", arg0, arg1)
	ret0, _ := ret[0].(*models.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Save indicates an expected call of Save
func (mr *MockRepositoryServiceMockRecorder) Save(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockRepositoryService)(nil).Save), arg0, arg1)
}
