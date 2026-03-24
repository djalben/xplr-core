package wallet_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"github.com/djalben/xplr-core/backend/internal/domain"
	handlerwallet "github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/wallet"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/wallet/mocks"
)

func reqWallet(uid domain.UUID, method, path string, body *bytes.Buffer) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, path, body)
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	ctx := context.WithValue(r.Context(), "userID", uid)
	return r.WithContext(ctx)
}

func TestHandler_GetBalance(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockUC := mocks.NewMockWalletUseCase(ctrl)
	mockUC.EXPECT().GetBalance(gomock.Any(), uid).Return(domain.NewNumeric(99.25), nil)

	h := handlerwallet.NewHandler(mockUC)
	rec := httptest.NewRecorder()
	h.GetBalance(rec, reqWallet(uid, http.MethodGet, "/wallet/balance", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_TopUp(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockUC := mocks.NewMockWalletUseCase(ctrl)
	mockUC.EXPECT().TopUpWallet(gomock.Any(), uid, gomock.Any()).Return(nil)

	h := handlerwallet.NewHandler(mockUC)
	rec := httptest.NewRecorder()
	h.TopUp(rec, reqWallet(uid, http.MethodPost, "/wallet/topup", bytes.NewBufferString(`{"amount":5}`)))

	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_ToggleAutoTopUp(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockUC := mocks.NewMockWalletUseCase(ctrl)
	mockUC.EXPECT().ToggleAutoTopUp(gomock.Any(), uid, true).Return(nil)

	h := handlerwallet.NewHandler(mockUC)
	rec := httptest.NewRecorder()
	h.ToggleAutoTopUp(rec, reqWallet(uid, http.MethodPost, "/wallet/autotopup", bytes.NewBufferString(`{"enabled":true}`)))

	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d %s", rec.Code, rec.Body.String())
	}
}
