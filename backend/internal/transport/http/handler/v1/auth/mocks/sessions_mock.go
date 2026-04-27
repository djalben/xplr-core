package mocks

import (
	context "context"
	reflect "reflect"

	domain "github.com/djalben/xplr-core/backend/internal/domain"
	gomock "github.com/golang/mock/gomock"
)

type MockAuthSessions struct {
	ctrl     *gomock.Controller
	recorder *MockAuthSessionsMockRecorder
}

type MockAuthSessionsMockRecorder struct {
	mock *MockAuthSessions
}

func NewMockAuthSessions(ctrl *gomock.Controller) *MockAuthSessions {
	mock := &MockAuthSessions{ctrl: ctrl}
	mock.recorder = &MockAuthSessionsMockRecorder{mock}

	return mock
}

func (m *MockAuthSessions) EXPECT() *MockAuthSessionsMockRecorder {
	return m.recorder
}

func (m *MockAuthSessions) Add(ctx context.Context, s *domain.AuthSession) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", ctx, s)
	ret0, _ := ret[0].(error)

	return ret0
}

func (mr *MockAuthSessionsMockRecorder) Add(ctx, s any) *gomock.Call {
	mr.mock.ctrl.T.Helper()

	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockAuthSessions)(nil).Add), ctx, s)
}

func (m *MockAuthSessions) DeleteOlderThan(ctx context.Context, userID domain.UUID, keepLast int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteOlderThan", ctx, userID, keepLast)
	ret0, _ := ret[0].(error)

	return ret0
}

func (mr *MockAuthSessionsMockRecorder) DeleteOlderThan(ctx, userID, keepLast any) *gomock.Call {
	mr.mock.ctrl.T.Helper()

	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteOlderThan", reflect.TypeOf((*MockAuthSessions)(nil).DeleteOlderThan), ctx, userID, keepLast)
}
