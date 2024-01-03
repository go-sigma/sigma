// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/go-sigma/sigma/pkg/dal/dao (interfaces: UserServiceFactory)
//
// Generated by this command:
//
//	mockgen -destination=mocks/user_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao UserServiceFactory
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	dao "github.com/go-sigma/sigma/pkg/dal/dao"
	query "github.com/go-sigma/sigma/pkg/dal/query"
	gomock "go.uber.org/mock/gomock"
)

// MockUserServiceFactory is a mock of UserServiceFactory interface.
type MockUserServiceFactory struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceFactoryMockRecorder
}

// MockUserServiceFactoryMockRecorder is the mock recorder for MockUserServiceFactory.
type MockUserServiceFactoryMockRecorder struct {
	mock *MockUserServiceFactory
}

// NewMockUserServiceFactory creates a new mock instance.
func NewMockUserServiceFactory(ctrl *gomock.Controller) *MockUserServiceFactory {
	mock := &MockUserServiceFactory{ctrl: ctrl}
	mock.recorder = &MockUserServiceFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserServiceFactory) EXPECT() *MockUserServiceFactoryMockRecorder {
	return m.recorder
}

// New mocks base method.
func (m *MockUserServiceFactory) New(arg0 ...*query.Query) dao.UserService {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "New", varargs...)
	ret0, _ := ret[0].(dao.UserService)
	return ret0
}

// New indicates an expected call of New.
func (mr *MockUserServiceFactoryMockRecorder) New(arg0 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockUserServiceFactory)(nil).New), arg0...)
}
