package mocks

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockRateLimiter is a mock of RateLimiter interface.
type MockRateLimiter struct {
	ctrl     *gomock.Controller
	recorder *MockRateLimiterMockRecorder
}

// MockRateLimiterMockRecorder is the mock recorder for MockRateLimiter.
type MockRateLimiterMockRecorder struct {
	mock *MockRateLimiter
}

func NewMockRateLimiter(ctrl *gomock.Controller) *MockRateLimiter {
	mock := &MockRateLimiter{ctrl: ctrl}
	mock.recorder = &MockRateLimiterMockRecorder{mock}

	return mock
}

func (m *MockRateLimiter) EXPECT() *MockRateLimiterMockRecorder {
	return m.recorder
}

func (m *MockRateLimiter) Allow(ctx context.Context, key string, now time.Time) (bool, time.Duration, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Allow", ctx, key, now)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(time.Duration)
	ret2, _ := ret[2].(error)

	return ret0, ret1, ret2
}

func (mr *MockRateLimiterMockRecorder) Allow(ctx, key, now any) *gomock.Call {
	mr.mock.ctrl.T.Helper()

	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Allow", reflect.TypeOf((*MockRateLimiter)(nil).Allow), ctx, key, now)
}

func (m *MockRateLimiter) Fail(ctx context.Context, key string, now time.Time) (time.Duration, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Fail", ctx, key, now)
	ret0, _ := ret[0].(time.Duration)
	ret1, _ := ret[1].(error)

	return ret0, ret1
}

func (mr *MockRateLimiterMockRecorder) Fail(ctx, key, now any) *gomock.Call {
	mr.mock.ctrl.T.Helper()

	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fail", reflect.TypeOf((*MockRateLimiter)(nil).Fail), ctx, key, now)
}

func (m *MockRateLimiter) Success(ctx context.Context, key string, now time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Success", ctx, key, now)
	ret0, _ := ret[0].(error)

	return ret0
}

func (mr *MockRateLimiterMockRecorder) Success(ctx, key, now any) *gomock.Call {
	mr.mock.ctrl.T.Helper()

	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Success", reflect.TypeOf((*MockRateLimiter)(nil).Success), ctx, key, now)
}
