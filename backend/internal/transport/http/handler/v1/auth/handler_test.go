package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
		setupMocks     func(auth *mocks.MockAuthRegisterLogin, wallet *mocks.MockWalletBalanceProvider, userReader *mocks.MockUserByIDReader)
		wantStatusCode int
		wantOK         bool // true => JSON с token и user
	}

	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	regUser := &domain.User{
		ID:     uid,
		Email:  "a@example.com",
		Status: domain.UserStatusActive,
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "happy path",
			args: args{
				body: `{"email":"a@example.com","password":"secret"}`,
				setupMocks: func(authMock *mocks.MockAuthRegisterLogin, wallet *mocks.MockWalletBalanceProvider, _ *mocks.MockUserByIDReader) {
					authMock.EXPECT().
						Register(gomock.Any(), "a@example.com", "secret").
						Return(regUser, nil)
					wallet.EXPECT().
						GetBalance(gomock.Any(), uid).
						Return(domain.NewNumeric(42.5), nil)
				},
				wantStatusCode: http.StatusOK,
				wantOK:         true,
			},
		},
		{
			name: "register error -> 400",
			args: args{
				body: `{"email":"a@example.com","password":"secret"}`,
				setupMocks: func(authMock *mocks.MockAuthRegisterLogin, _ *mocks.MockWalletBalanceProvider, _ *mocks.MockUserByIDReader) {
					authMock.EXPECT().
						Register(gomock.Any(), "a@example.com", "secret").
						Return(nil, errTestRegisterEmailTaken)
				},
				wantStatusCode: http.StatusBadRequest,
				wantOK:         false,
			},
		},
		{
			name: "invalid json body -> 400",
			args: args{
				body: `{`,
				setupMocks: func(_ *mocks.MockAuthRegisterLogin, _ *mocks.MockWalletBalanceProvider, _ *mocks.MockUserByIDReader) {
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

			authMock := mocks.NewMockAuthRegisterLogin(ctrl)
			walletMock := mocks.NewMockWalletBalanceProvider(ctrl)
			userReaderMock := mocks.NewMockUserByIDReader(ctrl)
			tt.args.setupMocks(authMock, walletMock, userReaderMock)

			h := handlerauth.NewHandler(authMock, walletMock, userReaderMock, []byte(testJWTSecret))

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
			if got["token"] == nil || got["token"] == "" {
				t.Fatal("expected non-empty token")
			}
			userObj, ok := got["user"].(map[string]any)
			if !ok {
				t.Fatalf("user field: %v", got["user"])
			}
			if userObj["email"] != "a@example.com" {
				t.Errorf("user.email = %v", userObj["email"])
			}
			if userObj["balance"] != "42.5" {
				t.Errorf("user.balance = %v, want 42.5", userObj["balance"])
			}
		})
	}
}

func TestHandler_DoLogin(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	loginUser := &domain.User{
		ID:     uid,
		Email:  "b@example.com",
		Status: domain.UserStatusActive,
	}

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	authMock := mocks.NewMockAuthRegisterLogin(ctrl)
	walletMock := mocks.NewMockWalletBalanceProvider(ctrl)
	userReaderMock := mocks.NewMockUserByIDReader(ctrl)

	authMock.EXPECT().
		Login(gomock.Any(), "b@example.com", "x").
		Return(loginUser, nil)
	walletMock.EXPECT().
		GetBalance(gomock.Any(), uid).
		Return(domain.NewNumeric(0), nil)

	h := handlerauth.NewHandler(authMock, walletMock, userReaderMock, []byte(testJWTSecret))
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"b@example.com","password":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.DoLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%q", rec.Code, rec.Body.String())
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
		ID:     uid,
		Email:  "c@example.com",
		Status: domain.UserStatusActive,
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

			authMock := mocks.NewMockAuthRegisterLogin(ctrl)
			walletMock := mocks.NewMockWalletBalanceProvider(ctrl)
			userReaderMock := mocks.NewMockUserByIDReader(ctrl)

			tt.args.setupMocks(walletMock, userReaderMock, uid, u)

			h := handlerauth.NewHandler(authMock, walletMock, userReaderMock, []byte(testJWTSecret))
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
