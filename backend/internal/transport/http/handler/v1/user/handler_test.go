package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"github.com/djalben/xplr-core/backend/internal/domain"
	handleruser "github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/user"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/user/mocks"
)

func reqUser(uid domain.UUID, method, path string, body *bytes.Buffer) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, path, body)
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	if uid != uuid.Nil {
		ctx := context.WithValue(r.Context(), "userID", uid)
		r = r.WithContext(ctx)
	}
	return r
}

func newHandlerAllMocks(ctrl *gomock.Controller) (
	*handleruser.Handler,
	*mocks.MockUserProfile,
	*mocks.MockUserWallet,
	*mocks.MockUserGrades,
	*mocks.MockUserCards,
	*mocks.MockUserTransactions,
	*mocks.MockUserTickets,
) {
	up := mocks.NewMockUserProfile(ctrl)
	w := mocks.NewMockUserWallet(ctrl)
	g := mocks.NewMockUserGrades(ctrl)
	c := mocks.NewMockUserCards(ctrl)
	tx := mocks.NewMockUserTransactions(ctrl)
	tk := mocks.NewMockUserTickets(ctrl)
	return handleruser.NewHandler(up, w, g, c, tx, tk), up, w, g, c, tx, tk
}

func TestHandler_GetMe_unauthorized(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	h, _, _, _, _, _, _ := newHandlerAllMocks(ctrl)
	rec := httptest.NewRecorder()
	h.GetMe(rec, reqUser(uuid.Nil, http.MethodGet, "/user/me", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestHandler_GetMe_ok(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	h, up, _, _, _, _, _ := newHandlerAllMocks(ctrl)
	up.EXPECT().GetMe(gomock.Any(), uid).Return(map[string]any{
		"id": uid.String(), "email": "a@b.c", "balance": "0",
	}, nil)

	rec := httptest.NewRecorder()
	h.GetMe(rec, reqUser(uid, http.MethodGet, "/user/me", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_GetGrade_ok(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	ug := &domain.UserGrade{
		UserID:     uid,
		Grade:      "GOLD",
		TotalSpent: domain.NewNumeric(10),
		FeePercent: domain.NewNumeric(1.5),
	}

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	h, _, _, g, _, _, _ := newHandlerAllMocks(ctrl)
	g.EXPECT().GetByUserID(gomock.Any(), uid).Return(ug, nil)

	rec := httptest.NewRecorder()
	h.GetGrade(rec, reqUser(uid, http.MethodGet, "/user/grade", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_GetWallet_ok(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	h, _, w, _, _, _, _ := newHandlerAllMocks(ctrl)
	w.EXPECT().GetBalance(gomock.Any(), uid).Return(domain.NewNumeric(7), nil)

	rec := httptest.NewRecorder()
	h.GetWallet(rec, reqUser(uid, http.MethodGet, "/user/wallet", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d %s", rec.Code, rec.Body.String())
	}
	var got map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if got["master_balance"] != "7" {
		t.Errorf("master_balance=%v", got["master_balance"])
	}
}

func TestHandler_Support_emptyMessage(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	h, _, _, _, _, _, _ := newHandlerAllMocks(ctrl)
	rec := httptest.NewRecorder()
	h.Support(rec, reqUser(uid, http.MethodPost, "/user/support", bytes.NewBufferString(`{"message":""}`)))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestHandler_Support_ok(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	ticketID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	h, _, _, _, _, _, tk := newHandlerAllMocks(ctrl)
	tk.EXPECT().
		Create(gomock.Any(), uid, "Support", "hello", (*int64)(nil)).
		Return(&domain.Ticket{ID: ticketID, UserID: uid, CreatedAt: time.Now().UTC()}, nil)

	rec := httptest.NewRecorder()
	h.Support(rec, reqUser(uid, http.MethodPost, "/user/support", bytes.NewBufferString(`{"message":"hello"}`)))

	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_GetReferralsInfo_error(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("66666666-6666-6666-6666-666666666666")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	h, up, _, _, _, _, _ := newHandlerAllMocks(ctrl)
	up.EXPECT().GetReferralInfo(gomock.Any(), uid).Return(nil, errors.New("db"))

	rec := httptest.NewRecorder()
	h.GetReferralsInfo(rec, reqUser(uid, http.MethodGet, "/user/referrals/info", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d", rec.Code)
	}
}
