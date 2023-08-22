// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/go-sigma/sigma/pkg/storage (interfaces: StorageDriver)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	io "io"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockStorageDriver is a mock of StorageDriver interface.
type MockStorageDriver struct {
	ctrl     *gomock.Controller
	recorder *MockStorageDriverMockRecorder
}

// MockStorageDriverMockRecorder is the mock recorder for MockStorageDriver.
type MockStorageDriverMockRecorder struct {
	mock *MockStorageDriver
}

// NewMockStorageDriver creates a new mock instance.
func NewMockStorageDriver(ctrl *gomock.Controller) *MockStorageDriver {
	mock := &MockStorageDriver{ctrl: ctrl}
	mock.recorder = &MockStorageDriverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorageDriver) EXPECT() *MockStorageDriverMockRecorder {
	return m.recorder
}

// AbortUpload mocks base method.
func (m *MockStorageDriver) AbortUpload(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AbortUpload", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// AbortUpload indicates an expected call of AbortUpload.
func (mr *MockStorageDriverMockRecorder) AbortUpload(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AbortUpload", reflect.TypeOf((*MockStorageDriver)(nil).AbortUpload), arg0, arg1, arg2)
}

// CommitUpload mocks base method.
func (m *MockStorageDriver) CommitUpload(arg0 context.Context, arg1, arg2 string, arg3 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CommitUpload", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// CommitUpload indicates an expected call of CommitUpload.
func (mr *MockStorageDriverMockRecorder) CommitUpload(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CommitUpload", reflect.TypeOf((*MockStorageDriver)(nil).CommitUpload), arg0, arg1, arg2, arg3)
}

// CreateUploadID mocks base method.
func (m *MockStorageDriver) CreateUploadID(arg0 context.Context, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUploadID", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUploadID indicates an expected call of CreateUploadID.
func (mr *MockStorageDriverMockRecorder) CreateUploadID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUploadID", reflect.TypeOf((*MockStorageDriver)(nil).CreateUploadID), arg0, arg1)
}

// Delete mocks base method.
func (m *MockStorageDriver) Delete(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockStorageDriverMockRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockStorageDriver)(nil).Delete), arg0, arg1)
}

// Move mocks base method.
func (m *MockStorageDriver) Move(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Move", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Move indicates an expected call of Move.
func (mr *MockStorageDriverMockRecorder) Move(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Move", reflect.TypeOf((*MockStorageDriver)(nil).Move), arg0, arg1, arg2)
}

// Reader mocks base method.
func (m *MockStorageDriver) Reader(arg0 context.Context, arg1 string, arg2 int64) (io.ReadCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reader", arg0, arg1, arg2)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Reader indicates an expected call of Reader.
func (mr *MockStorageDriverMockRecorder) Reader(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reader", reflect.TypeOf((*MockStorageDriver)(nil).Reader), arg0, arg1, arg2)
}

// Upload mocks base method.
func (m *MockStorageDriver) Upload(arg0 context.Context, arg1 string, arg2 io.Reader) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upload", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upload indicates an expected call of Upload.
func (mr *MockStorageDriverMockRecorder) Upload(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upload", reflect.TypeOf((*MockStorageDriver)(nil).Upload), arg0, arg1, arg2)
}

// UploadPart mocks base method.
func (m *MockStorageDriver) UploadPart(arg0 context.Context, arg1, arg2 string, arg3 int64, arg4 io.Reader) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadPart", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UploadPart indicates an expected call of UploadPart.
func (mr *MockStorageDriverMockRecorder) UploadPart(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadPart", reflect.TypeOf((*MockStorageDriver)(nil).UploadPart), arg0, arg1, arg2, arg3, arg4)
}
