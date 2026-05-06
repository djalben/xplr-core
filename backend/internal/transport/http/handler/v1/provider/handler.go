package provider

import (
	"net/http"

	subscriptionUC "github.com/djalben/xplr-core/backend/internal/application/subscription"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	subUC SubscriptionAuthorizationUseCase
}

func NewHandler(subUC SubscriptionAuthorizationUseCase) *Handler {
	return &Handler{subUC: subUC}
}

func (h *Handler) Register(r chi.Router) {
	r.Route("/provider", func(r chi.Router) {
		r.Post("/authorization", h.Authorization)
	})
}

func (h *Handler) Authorization(w http.ResponseWriter, r *http.Request) {
	var body subscriptionUC.AuthorizationEvent

	err := handler.ReadJSON(r, &body)
	if err != nil {
		_ = handler.WrapAndWriteError(r.Context(), w, err, http.StatusBadRequest, "Неверный запрос")

		return
	}

	res, err := h.subUC.HandleAuthorization(r.Context(), body)
	if err != nil {
		_ = handler.WriteInternalServerError(r.Context(), w, err)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, res)
}

