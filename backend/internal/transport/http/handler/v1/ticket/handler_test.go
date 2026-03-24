package ticket_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"github.com/djalben/xplr-core/backend/internal/domain"
	handlerticket "github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/ticket"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/ticket/mocks"
	"github.com/go-chi/chi/v5"
)

func reqTicket(uid domain.UUID, method, path string, body *bytes.Buffer) *http.Request {
	r := httptest.NewRequest(method, path, body)
	ctx := context.WithValue(r.Context(), "userID", uid)
	return r.WithContext(ctx)
}

func reqTicketChi(method, path, idParam string, body *bytes.Buffer) *http.Request {
	r := httptest.NewRequest(method, path, body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", idParam)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestHandler_Create(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ticketID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	tk := &domain.Ticket{
		ID:        ticketID,
		UserID:    uid,
		Subject:   "s",
		CreatedAt: time.Now().UTC(),
	}

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockUC := mocks.NewMockTicketUseCase(ctrl)
	mockUC.EXPECT().
		Create(gomock.Any(), uid, "sub", "msg", (*int64)(nil)).
		Return(tk, nil)

	h := handlerticket.NewHandler(mockUC)
	rec := httptest.NewRecorder()
	h.Create(rec, reqTicket(uid, http.MethodPost, "/ticket/create", bytes.NewBufferString(
		`{"subject":"sub","message":"msg"}`)))

	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_Take(t *testing.T) {
	t.Parallel()

	adminID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	ticketID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockUC := mocks.NewMockTicketUseCase(ctrl)
	mockUC.EXPECT().Take(gomock.Any(), ticketID, adminID).Return(nil)

	h := handlerticket.NewHandler(mockUC)
	rec := httptest.NewRecorder()

	ctx := context.WithValue(context.Background(), "userID", adminID)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", ticketID.String())
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	r := httptest.NewRequest(http.MethodPut, "/ticket/"+ticketID.String()+"/take", nil).WithContext(ctx)

	h.Take(rec, r)

	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_Close(t *testing.T) {
	t.Parallel()

	ticketID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockUC := mocks.NewMockTicketUseCase(ctrl)
	mockUC.EXPECT().Close(gomock.Any(), ticketID, "done").Return(errors.New("x"))

	h := handlerticket.NewHandler(mockUC)
	rec := httptest.NewRecorder()
	r := reqTicketChi(http.MethodPut, "/t/x/close", ticketID.String(), bytes.NewBufferString(`{"reply":"done"}`))
	h.Close(rec, r)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d want 400", rec.Code)
	}
}
