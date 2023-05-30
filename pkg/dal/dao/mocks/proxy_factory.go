// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ximager/ximager/pkg/dal/dao (interfaces: ProxyServiceFactory)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	dao "github.com/ximager/ximager/pkg/dal/dao"
	query "github.com/ximager/ximager/pkg/dal/query"
	reflect "reflect"
)

// MockProxyServiceFactory is a mock of ProxyServiceFactory interface
type MockProxyServiceFactory struct {
	ctrl     *gomock.Controller
	recorder *MockProxyServiceFactoryMockRecorder
}

// MockProxyServiceFactoryMockRecorder is the mock recorder for MockProxyServiceFactory
type MockProxyServiceFactoryMockRecorder struct {
	mock *MockProxyServiceFactory
}

// NewMockProxyServiceFactory creates a new mock instance
func NewMockProxyServiceFactory(ctrl *gomock.Controller) *MockProxyServiceFactory {
	mock := &MockProxyServiceFactory{ctrl: ctrl}
	mock.recorder = &MockProxyServiceFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockProxyServiceFactory) EXPECT() *MockProxyServiceFactoryMockRecorder {
	return m.recorder
}

// New mocks base method
func (m *MockProxyServiceFactory) New(arg0 ...*query.Query) dao.ProxyService {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "New", varargs...)
	ret0, _ := ret[0].(dao.ProxyService)
	return ret0
}

// New indicates an expected call of New
func (mr *MockProxyServiceFactoryMockRecorder) New(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockProxyServiceFactory)(nil).New), arg0...)
}