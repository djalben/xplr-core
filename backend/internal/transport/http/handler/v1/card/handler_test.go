package card_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/djalben/xplr-core/backend/internal/domain"
	handlercard "github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/card"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler/v1/card/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

var errTestCardUseCaseFail = errors.New("fail")

func reqCard(uid domain.UUID, method, path string, body *bytes.Buffer) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequestWithContext(context.Background(), method, path, body)
	} else {
		r = httptest.NewRequestWithContext(context.Background(), method, path, nil)
	}
	ctx := context.WithValue(r.Context(), "userID", uid)

	return r.WithContext(ctx)
}

func reqCardChi(uid domain.UUID, method, path, cardID string, body *bytes.Buffer) *http.Request {
	r := reqCard(uid, method, path, body)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", cardID)

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestHandler_BuyCard(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cardEnt := &domain.Card{
		ID:       uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		UserID:   uid,
		CardType: domain.CardTypeSubscriptions,
		Nickname: "n",
	}

	tests := []struct {
		name       string
		body       string
		setup      func(m *mocks.MockCardUseCase)
		wantCode   int
		wantCardID string
	}{
		{
			name: "happy path",
			body: `{"cardType":"subscriptions","nickname":"n"}`,
			setup: func(m *mocks.MockCardUseCase) {
				m.EXPECT().
					BuyCard(gomock.Any(), uid, domain.CardTypeSubscriptions, "n").
					Return(cardEnt, nil)
			},
			wantCode:   http.StatusOK,
			wantCardID: cardEnt.ID.String(),
		},
		{
			name: "use case error",
			body: `{"cardType":"subscriptions","nickname":"n"}`,
			setup: func(m *mocks.MockCardUseCase) {
				m.EXPECT().
					BuyCard(gomock.Any(), uid, domain.CardTypeSubscriptions, "n").
					Return(nil, errTestCardUseCaseFail)
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid json",
			body: `{`,
			setup: func(_ *mocks.MockCardUseCase) {
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			t.Cleanup(ctrl.Finish)

			mockUC := mocks.NewMockCardUseCase(ctrl)
			tt.setup(mockUC)

			h := handlercard.NewHandler(mockUC)
			rec := httptest.NewRecorder()
			h.BuyCard(rec, reqCard(uid, http.MethodPost, "/card/buy", bytes.NewBufferString(tt.body)))

			if rec.Code != tt.wantCode {
				t.Fatalf("code=%d want=%d body=%q", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.wantCardID == "" {
				return
			}

			var got map[string]any
			err := json.NewDecoder(rec.Body).Decode(&got)
			if err != nil {
				t.Fatal(err)
			}
			if got["id"] != tt.wantCardID {
				t.Errorf("id=%v want %s", got["id"], tt.wantCardID)
			}
		})
	}
}

func TestHandler_TopUpCard(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	cardID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockUC := mocks.NewMockCardUseCase(ctrl)
	mockUC.EXPECT().
		TopUpCard(gomock.Any(), uid, cardID, gomock.Any()).
		Return(nil)

	h := handlercard.NewHandler(mockUC)
	rec := httptest.NewRecorder()
	h.TopUpCard(rec, reqCardChi(uid, http.MethodPost, "/x/"+cardID.String()+"/topup", cardID.String(),
		bytes.NewBufferString(`{"amount":10.5}`)))

	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_SpendFromCard(t *testing.T) {
	t.Parallel()

	uid := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	cardID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockUC := mocks.NewMockCardUseCase(ctrl)
	mockUC.EXPECT().
		SpendFromCard(gomock.Any(), uid, cardID, gomock.Any()).
		Return(nil)

	h := handlercard.NewHandler(mockUC)
	rec := httptest.NewRecorder()
	h.SpendFromCard(rec, reqCardChi(uid, http.MethodPost, "/x/"+cardID.String()+"/spend", cardID.String(),
		bytes.NewBufferString(`{"amount":3}`)))

	if rec.Code != http.StatusOK {
		t.Fatalf("code=%d %s", rec.Code, rec.Body.String())
	}
}
