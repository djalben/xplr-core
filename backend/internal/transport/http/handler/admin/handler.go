package admin

import (
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/application/card"
	"github.com/djalben/xplr-core/backend/internal/application/commission"
	"github.com/djalben/xplr-core/backend/internal/application/grades"
	"github.com/djalben/xplr-core/backend/internal/application/kyc"
	"github.com/djalben/xplr-core/backend/internal/application/ticket"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type Handler struct {
	cardUseCase       *card.UseCase
	commissionUseCase *commission.UseCase
	ticketUseCase     *ticket.UseCase
	gradesUseCase     *grades.UseCase
	kycUseCase        *kyc.UseCase
}

func NewHandler(cardUC *card.UseCase, commissionUC *commission.UseCase, ticketUC *ticket.UseCase, gradesUC *grades.UseCase, kycUC *kyc.UseCase) *Handler {
	return &Handler{
		cardUseCase:       cardUC,
		commissionUseCase: commissionUC,
		ticketUseCase:     ticketUC,
		gradesUseCase:     gradesUC,
		kycUseCase:        kycUC,
	}
}

func (h *Handler) Register(r chi.Router) {
	r.Post("/tariffs", h.ChangeTariffs)
	r.Post("/referrals", h.ChangeReferralBonuses)
	r.Put("/cards/{id}/block", h.BlockCard)
	r.Put("/cards/{id}/unblock", h.UnblockCard)
	r.Put("/tickets/{id}/take", h.TakeTicket)
	r.Put("/tickets/{id}/close", h.CloseTicket)
	r.Put("/users/{id}/grade", h.ChangeUserGrade)
	r.Get("/kyc-applications", h.ListKYCPending)
	r.Put("/kyc-applications/{id}/decision", h.DecideKYC)
}

// ChangeTariffs — POST /admin/tariffs.
func (h *Handler) ChangeTariffs(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Key   string  `json:"key"`
		Value float64 `json:"value"`
	}

	var req request

	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	cfg, err := h.commissionUseCase.GetByKey(r.Context(), req.Key)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	cfg.Value = domain.NewNumeric(req.Value)

	err = h.commissionUseCase.Update(r.Context(), cfg)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// ChangeReferralBonuses — POST /admin/referrals.
func (h *Handler) ChangeReferralBonuses(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Percent float64 `json:"percent"`
	}

	var req request

	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	cfg, err := h.commissionUseCase.GetByKey(r.Context(), "referral_percent")
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	cfg.Value = domain.NewNumeric(req.Percent)

	err = h.commissionUseCase.Update(r.Context(), cfg)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// BlockCard — PUT /admin/cards/{id}/block.
func (h *Handler) BlockCard(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := domain.ParseUUID(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.cardUseCase.BlockCard(r.Context(), id)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// UnblockCard — PUT /admin/cards/{id}/unblock.
func (h *Handler) UnblockCard(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := domain.ParseUUID(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.cardUseCase.UnblockCard(r.Context(), id)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// TakeTicket — PUT /admin/tickets/{id}/take.
func (h *Handler) TakeTicket(w http.ResponseWriter, r *http.Request) {
	adminID := handler.GetUserIDFromContext(r)

	idStr := chi.URLParam(r, "id")

	id, err := domain.ParseUUID(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.ticketUseCase.Take(r.Context(), id, adminID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// CloseTicket — PUT /admin/tickets/{id}/close.
func (h *Handler) CloseTicket(w http.ResponseWriter, r *http.Request) {
	id, ok := adminChiUUID(w, r)
	if !ok {
		return
	}

	var req struct {
		Reply string `json:"reply"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}

	err := h.ticketUseCase.Close(r.Context(), id, req.Reply)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// ChangeUserGrade — PUT /admin/users/{id}/grade.
func (h *Handler) ChangeUserGrade(w http.ResponseWriter, r *http.Request) {
	id, ok := adminChiUUID(w, r)
	if !ok {
		return
	}

	var req struct {
		Grade string `json:"grade"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}

	err := h.gradesUseCase.ChangeGrade(r.Context(), id, req.Grade)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func adminChiUUID(w http.ResponseWriter, r *http.Request) (domain.UUID, bool) {
	idStr := chi.URLParam(r, "id")

	id, err := domain.ParseUUID(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return domain.UUID{}, false
	}

	return id, true
}

func adminReadJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	err := handler.ReadJSON(r, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return false
	}

	return true
}

// ListKYCPending — GET /admin/kyc-applications.
func (h *Handler) ListKYCPending(w http.ResponseWriter, r *http.Request) {
	list, err := h.kycUseCase.ListPending(r.Context(), 200)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	out := make([]map[string]any, 0, len(list))

	for _, a := range list {
		out = append(out, map[string]any{
			"id":         a.ID.String(),
			"user_id":    a.UserID.String(),
			"status":     string(a.Status),
			"payload":    a.PayloadJSON,
			"created_at": a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{"applications": out})
}

// DecideKYC — PUT /admin/kyc-applications/{id}/decision.
func (h *Handler) DecideKYC(w http.ResponseWriter, r *http.Request) {
	adminID := handler.GetUserIDFromContext(r)

	idStr := chi.URLParam(r, "id")
	appID, err := domain.ParseUUID(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	var req struct {
		Approve bool   `json:"approve"`
		Comment string `json:"comment"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}

	err = h.kycUseCase.DecideApplication(r.Context(), appID, adminID, req.Approve, req.Comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
