// Code generated by MockGen. DO NOT EDIT.
// Source: session.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	jwt "github.com/Benzogang-Tape/Reddit/internal/models/jwt"
	gomock "github.com/golang/mock/gomock"
)

// MockSessionManager is a mock of SessionManager interface.
type MockSessionManager struct {
	ctrl     *gomock.Controller
	recorder *MockSessionManagerMockRecorder
}

// MockSessionManagerMockRecorder is the mock recorder for MockSessionManager.
type MockSessionManagerMockRecorder struct {
	mock *MockSessionManager
}

// NewMockSessionManager creates a new mock instance.
func NewMockSessionManager(ctrl *gomock.Controller) *MockSessionManager {
	mock := &MockSessionManager{ctrl: ctrl}
	mock.recorder = &MockSessionManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSessionManager) EXPECT() *MockSessionManagerMockRecorder {
	return m.recorder
}

// CheckSession mocks base method.
func (m *MockSessionManager) CheckSession(ctx context.Context, session *jwt.Session) (*jwt.TokenPayload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckSession", ctx, session)
	ret0, _ := ret[0].(*jwt.TokenPayload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckSession indicates an expected call of CheckSession.
func (mr *MockSessionManagerMockRecorder) CheckSession(ctx, session interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckSession", reflect.TypeOf((*MockSessionManager)(nil).CheckSession), ctx, session)
}

// CreateSession mocks base method.
func (m *MockSessionManager) CreateSession(ctx context.Context, session *jwt.Session, payload *jwt.TokenPayload) (*jwt.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSession", ctx, session, payload)
	ret0, _ := ret[0].(*jwt.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateSession indicates an expected call of CreateSession.
func (mr *MockSessionManagerMockRecorder) CreateSession(ctx, session, payload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSession", reflect.TypeOf((*MockSessionManager)(nil).CreateSession), ctx, session, payload)
}

// MockSessionAPI is a mock of SessionAPI interface.
type MockSessionAPI struct {
	ctrl     *gomock.Controller
	recorder *MockSessionAPIMockRecorder
}

// MockSessionAPIMockRecorder is the mock recorder for MockSessionAPI.
type MockSessionAPIMockRecorder struct {
	mock *MockSessionAPI
}

// NewMockSessionAPI creates a new mock instance.
func NewMockSessionAPI(ctrl *gomock.Controller) *MockSessionAPI {
	mock := &MockSessionAPI{ctrl: ctrl}
	mock.recorder = &MockSessionAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSessionAPI) EXPECT() *MockSessionAPIMockRecorder {
	return m.recorder
}

// New mocks base method.
func (m *MockSessionAPI) New(ctx context.Context) (*jwt.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "New", ctx)
	ret0, _ := ret[0].(*jwt.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// New indicates an expected call of New.
func (mr *MockSessionAPIMockRecorder) New(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "New", reflect.TypeOf((*MockSessionAPI)(nil).New), ctx)
}

// Verify mocks base method.
func (m *MockSessionAPI) Verify(ctx context.Context, session *jwt.Session) (*jwt.TokenPayload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Verify", ctx, session)
	ret0, _ := ret[0].(*jwt.TokenPayload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Verify indicates an expected call of Verify.
func (mr *MockSessionAPIMockRecorder) Verify(ctx, session interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Verify", reflect.TypeOf((*MockSessionAPI)(nil).Verify), ctx, session)
}
