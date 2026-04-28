package staffpin

import (
	"context"
	"net/http"
	"strings"

	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	systemRepo ports.SystemSettingsRepository
}

func NewHandler(systemRepo ports.SystemSettingsRepository) *Handler {
	return &Handler{systemRepo: systemRepo}
}

func (h *Handler) Register(r chi.Router) {
	r.Post("/verify-staff-pin", h.VerifyStaffPIN)
}

func (h *Handler) VerifyStaffPIN(w http.ResponseWriter, r *http.Request) {
	type request struct {
		PIN string `json:"pin"`
	}

	var req request

	err := handler.ReadJSON(r, &req)
	if err != nil {
		_ = handler.WrapAndWriteError(r.Context(), w, err, http.StatusBadRequest, "Неверный запрос")

		return
	}

	pin := strings.TrimSpace(req.PIN)
	if pin == "" {
		handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, map[string]any{"access": "denied"})

		return
	}

	stored, err := h.getAdminPIN(r.Context())
	if err != nil {
		_ = handler.WriteInternalServerError(r.Context(), w, err)

		return
	}

	access := "denied"
	if verifyPIN(stored, pin) {
		access = "granted"
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, map[string]any{"access": access})
}

func (h *Handler) getAdminPIN(ctx context.Context) (string, error) {
	list, err := h.systemRepo.ListAll(ctx)
	if err != nil {
		return "", wrapper.Wrap(err)
	}

	for _, row := range list {
		if row == nil {
			continue
		}
		if row.Key == "admin_pin" {
			return strings.TrimSpace(row.Value), nil
		}
	}

	return "", nil
}

func verifyPIN(stored, plain string) bool {
	stored = strings.TrimSpace(stored)
	plain = strings.TrimSpace(plain)
	if stored == "" || plain == "" {
		return false
	}

	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain))

		return err == nil
	}

	// Legacy/plain storage (seeded as 0000). On next update, we store bcrypt.
	return stored == plain
}
