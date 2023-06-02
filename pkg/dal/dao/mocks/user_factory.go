// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ximager/ximager/pkg/dal/dao (interfaces: UserServiceFactory)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	dao "github.com/ximager/ximager/pkg/dal/dao"
	query "github.com/ximager/ximager/pkg/dal/query"
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
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "New", varargs...)
	ret0, _ := ret[0].(dao.UserService)
	return ret0
}

// New indicates an expected call of New.
func (mr *MockUserServiceFactoryMockRecorder) New(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockUserServiceFactory)(nil).New), arg0...)
}