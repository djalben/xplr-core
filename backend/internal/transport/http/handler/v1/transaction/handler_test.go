package transaction_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	handlertr "github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/transaction"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/transaction/mocks"
	"github.com/djalben/xplr-core/backend/internal/transport/http/httpctx"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

func reqTx(uid domain.UUID, method, path string) *http.Request {
	r := httptest.NewRequestWithContext(context.Background(), method, path, nil)
	ctx := httpctx.WithUserID(r.Context(), uid)

	return r.WithContext(ctx)
}

func reqTxChi(method, path, cardID string) *http.Request {
	r := httptest.NewRequestWithContext(context.Background(), method, path, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", cardID)

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestHandler_GetWalletTransactions(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tx := &domain.Transaction{
		ID:              uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		UserID:          uid,
		Amount:          domain.NewNumeric(1),
		TransactionType: "TOPUP_WALLET",
		ExecutedAt:      time.Now().UTC(),
	}

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockUC := mocks.NewMockTransactionUseCase(ctrl)
	mockUC.EXPECT().
		GetWalletTransactions(gomock.Any(), uid, gomock.Any(), gomock.Any()).
		Return([]*domain.Transaction{tx}, nil)

	h := handlertr.NewHandler(mockUC)
	rec := httptest.NewRecorder()
	h.GetWalletTransactions(rec, reqTx(uid, http.MethodGet, "/transaction/wallet"))

	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_GetCardTransactions_badUUID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockUC := mocks.NewMockTransactionUseCase(ctrl)
	h := handlertr.NewHandler(mockUC)
	rec := httptest.NewRecorder()
	h.GetCardTransactions(rec, reqTxChi(http.MethodGet, "/x/not-uuid", "not-uuid"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestHandler_GetCardTransactions_ok(t *testing.T) {
	t.Parallel()

	cardID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockUC := mocks.NewMockTransactionUseCase(ctrl)
	mockUC.EXPECT().
		GetCardTransactions(gomock.Any(), cardID, gomock.Any(), gomock.Any()).
		Return([]*domain.Transaction{}, nil)

	h := handlertr.NewHandler(mockUC)
	rec := httptest.NewRecorder()
	h.GetCardTransactions(rec, reqTxChi(http.MethodGet, "/c/"+cardID.String(), cardID.String()))

	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d %s", rec.Code, rec.Body.String())
	}
}
