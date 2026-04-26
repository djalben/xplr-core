package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authapp "github.com/djalben/xplr-core/backend/internal/application/auth"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/pkg/utils"
	handlerauth "github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/auth"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/auth/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

const testJWTSecret = "test-jwt-secret-for-handler-tests-ok"

var (
	errTestRegisterEmailTaken = errors.New("email already registered")
	errTestUserNotFound       = errors.New("not found")
)

func TestHandler_DoRegister(t *testing.T) {
	t.Parallel()

	type args struct {
		body           string
		setupMocks     func(auth *mocks.MockAuthFlow, wallet *mocks.MockWalletBalanceProvider, userReader *mocks.MockUserByIDReader)
		wantStatusCode int
		wantOK         bool // true => JSON с сообщением о письме (без JWT)
	}

	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	regUser := &domain.User{
		ID:            uid,
		Email:         "a@example.com",
		Status:        domain.UserStatusActive,
		EmailVerified: false,
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "happy path",
			args: args{
				body: `{"email":"a@example.com","password":"secret"}`,
				setupMocks: func(authMock *mocks.MockAuthFlow, _ *mocks.MockWalletBalanceProvider, _ *mocks.MockUserByIDReader) {
					authMock.EXPECT().
						Register(gomock.Any(), "a@example.com", "secret").
						Return(regUser, nil)
				},
				wantStatusCode: http.StatusCreated,
				wantOK:         true,
			},
		},
		{
			name: "register generic error -> 400",
			args: args{
				body: `{"email":"a@example.com","password":"secret"}`,
				setupMocks: func(authMock *mocks.MockAuthFlow, _ *mocks.MockWalletBalanceProvider, _ *mocks.MockUserByIDReader) {
					authMock.EXPECT().
						Register(gomock.Any(), "a@example.com", "secret").
						Return(nil, errTestRegisterEmailTaken)
				},
				wantStatusCode: http.StatusBadRequest,
				wantOK:         false,
			},
		},
		{
			name: "email taken -> 409",
			args: args{
				body: `{"email":"a@example.com","password":"secret"}`,
				setupMocks: func(authMock *mocks.MockAuthFlow, _ *mocks.MockWalletBalanceProvider, _ *mocks.MockUserByIDReader) {
					authMock.EXPECT().
						Register(gomock.Any(), "a@example.com", "secret").
						Return(nil, domain.NewAlreadyExists("email already registered"))
				},
				wantStatusCode: http.StatusConflict,
				wantOK:         false,
			},
		},
		{
			name: "invalid json body -> 400",
			args: args{
				body: `{`,
				setupMocks: func(_ *mocks.MockAuthFlow, _ *mocks.MockWalletBalanceProvider, _ *mocks.MockUserByIDReader) {
				},
				wantStatusCode: http.StatusBadRequest,
				wantOK:         false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			authMock := mocks.NewMockAuthFlow(ctrl)
			walletMock := mocks.NewMockWalletBalanceProvider(ctrl)
			userReaderMock := mocks.NewMockUserByIDReader(ctrl)
			limiterMock := mocks.NewMockRateLimiter(ctrl)
			tt.args.setupMocks(authMock, walletMock, userReaderMock)

			h := handlerauth.NewHandler(authMock, walletMock, userReaderMock, limiterMock, []byte(testJWTSecret))

			req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/register", bytes.NewBufferString(tt.args.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.DoRegister(rec, req)

			if rec.Code != tt.args.wantStatusCode {
				t.Fatalf("status = %d, want %d, body=%q", rec.Code, tt.args.wantStatusCode, rec.Body.String())
			}

			if !tt.args.wantOK {
				return
			}

			var got map[string]any
			err := json.NewDecoder(rec.Body).Decode(&got)
			if err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got["email"] != "a@example.com" {
				t.Errorf("email = %v", got["email"])
			}
			if got["email_verified"] != false {
				t.Errorf("email_verified = %v, want false", got["email_verified"])
			}
		})
	}
}

func TestHandler_DoLogin(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	loginUser := &domain.User{
		ID:            uid,
		Email:         "b@example.com",
		Status:        domain.UserStatusActive,
		EmailVerified: true,
	}

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	authMock := mocks.NewMockAuthFlow(ctrl)
	walletMock := mocks.NewMockWalletBalanceProvider(ctrl)
	userReaderMock := mocks.NewMockUserByIDReader(ctrl)
	limiterMock := mocks.NewMockRateLimiter(ctrl)

	limiterMock.EXPECT().Allow(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, time.Duration(0), nil)
	authMock.EXPECT().
		LoginWithTrustedDevice(gomock.Any(), "b@example.com", "x", "", gomock.Any()).
		Return(&authapp.LoginResult{User: loginUser}, nil)
	walletMock.EXPECT().
		GetBalance(gomock.Any(), uid).
		Return(domain.NewNumeric(0), nil)

	limiterMock.EXPECT().Success(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	h := handlerauth.NewHandler(authMock, walletMock, userReaderMock, limiterMock, []byte(testJWTSecret))
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"b@example.com","password":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.DoLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%q", rec.Code, rec.Body.String())
	}
}

func TestHandler_DoLogin_WithTrustedDeviceCookie(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	loginUser := &domain.User{
		ID:            uid,
		Email:         "b@example.com",
		Status:        domain.UserStatusActive,
		EmailVerified: true,
	}

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	authMock := mocks.NewMockAuthFlow(ctrl)
	walletMock := mocks.NewMockWalletBalanceProvider(ctrl)
	userReaderMock := mocks.NewMockUserByIDReader(ctrl)
	limiterMock := mocks.NewMockRateLimiter(ctrl)

	limiterMock.EXPECT().Allow(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, time.Duration(0), nil)
	authMock.EXPECT().
		LoginWithTrustedDevice(gomock.Any(), "b@example.com", "x", "raw-device-token", gomock.Any()).
		Return(&authapp.LoginResult{User: loginUser}, nil)
	walletMock.EXPECT().
		GetBalance(gomock.Any(), uid).
		Return(domain.NewNumeric(0), nil)

	limiterMock.EXPECT().Success(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	h := handlerauth.NewHandler(authMock, walletMock, userReaderMock, limiterMock, []byte(testJWTSecret))
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"b@example.com","password":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "xplr_trusted_device", Value: "raw-device-token"})
	rec := httptest.NewRecorder()

	h.DoLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%q", rec.Code, rec.Body.String())
	}
}

func TestHandler_DoLoginMFA_RememberDevice_SetsCookie(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	u := &domain.User{
		ID:            uid,
		Email:         "mfa@example.com",
		Status:        domain.UserStatusActive,
		EmailVerified: true,
	}

	mfaToken, err := utils.GenerateMFAPendingJWT([]byte(testJWTSecret), uid, u.Email)
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	authMock := mocks.NewMockAuthFlow(ctrl)
	walletMock := mocks.NewMockWalletBalanceProvider(ctrl)
	userReaderMock := mocks.NewMockUserByIDReader(ctrl)
	limiterMock := mocks.NewMockRateLimiter(ctrl)

	authMock.EXPECT().
		CompleteMFALogin(gomock.Any(), mfaToken, "123456").
		Return(u, nil)

	exp := time.Now().UTC().Add(30 * 24 * time.Hour)
	authMock.EXPECT().
		RememberTrustedDevice(gomock.Any(), uid, gomock.Any(), gomock.Any(), gomock.Any()).
		Return("raw-trusted", exp, nil)

	limiterMock.EXPECT().Allow(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, time.Duration(0), nil)
	limiterMock.EXPECT().Success(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	walletMock.EXPECT().
		GetBalance(gomock.Any(), uid).
		Return(domain.NewNumeric(0), nil)

	h := handlerauth.NewHandler(authMock, walletMock, userReaderMock, limiterMock, []byte(testJWTSecret))
	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/auth/login/mfa",
		bytes.NewBufferString(`{"mfaToken":"`+mfaToken+`","totpCode":"123456","rememberDevice":true}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()

	h.DoLoginMFA(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%q", rec.Code, rec.Body.String())
	}

	setCookie := rec.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Fatalf("expected Set-Cookie header")
	}
	if !strings.Contains(setCookie, "xplr_trusted_device=raw-trusted") {
		t.Fatalf("Set-Cookie=%q, want trusted device cookie", setCookie)
	}
}

func TestHandler_RefreshToken(t *testing.T) {
	t.Parallel()

	type args struct {
		header         string
		setupMocks     func(wallet *mocks.MockWalletBalanceProvider, userReader *mocks.MockUserByIDReader, uid uuid.UUID, u *domain.User)
		wantStatusCode int
		decodeJSON     bool
	}

	uid := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	u := &domain.User{
		ID:            uid,
		Email:         "c@example.com",
		Status:        domain.UserStatusActive,
		EmailVerified: true,
	}

	validTok, err := utils.GenerateJWT([]byte(testJWTSecret), uid, u.Email)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "missing header -> 401",
			args: args{
				header:         "",
				setupMocks:     func(_ *mocks.MockWalletBalanceProvider, _ *mocks.MockUserByIDReader, _ uuid.UUID, _ *domain.User) {},
				wantStatusCode: http.StatusUnauthorized,
				decodeJSON:     false,
			},
		},
		{
			name: "invalid token -> 401",
			args: args{
				header:         "Bearer not-a-jwt",
				setupMocks:     func(_ *mocks.MockWalletBalanceProvider, _ *mocks.MockUserByIDReader, _ uuid.UUID, _ *domain.User) {},
				wantStatusCode: http.StatusUnauthorized,
				decodeJSON:     false,
			},
		},
		{
			name: "user not found -> 401",
			args: args{
				header: "Bearer " + validTok,
				setupMocks: func(_ *mocks.MockWalletBalanceProvider, ur *mocks.MockUserByIDReader, id uuid.UUID, _ *domain.User) {
					ur.EXPECT().GetByID(gomock.Any(), id).Return(nil, errTestUserNotFound)
				},
				wantStatusCode: http.StatusUnauthorized,
				decodeJSON:     false,
			},
		},
		{
			name: "happy path",
			args: args{
				header: "Bearer " + validTok,
				setupMocks: func(wallet *mocks.MockWalletBalanceProvider, ur *mocks.MockUserByIDReader, id uuid.UUID, user *domain.User) {
					ur.EXPECT().GetByID(gomock.Any(), id).Return(user, nil)
					wallet.EXPECT().GetBalance(gomock.Any(), id).Return(domain.NewNumeric(1), nil)
				},
				wantStatusCode: http.StatusOK,
				decodeJSON:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			authMock := mocks.NewMockAuthFlow(ctrl)
			walletMock := mocks.NewMockWalletBalanceProvider(ctrl)
			userReaderMock := mocks.NewMockUserByIDReader(ctrl)
			limiterMock := mocks.NewMockRateLimiter(ctrl)

			tt.args.setupMocks(walletMock, userReaderMock, uid, u)

			h := handlerauth.NewHandler(authMock, walletMock, userReaderMock, limiterMock, []byte(testJWTSecret))
			req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/refresh-token", nil)
			if tt.args.header != "" {
				req.Header.Set("Authorization", tt.args.header)
			}
			rec := httptest.NewRecorder()

			h.RefreshToken(rec, req)

			if rec.Code != tt.args.wantStatusCode {
				t.Fatalf("status = %d, want %d, body=%q", rec.Code, tt.args.wantStatusCode, rec.Body.String())
			}

			if !tt.args.decodeJSON {
				return
			}

			var got map[string]any
			err := json.NewDecoder(rec.Body).Decode(&got)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got["token"] == nil || got["token"] == "" {
				t.Fatal("expected token in JSON body")
			}
		})
	}
}
