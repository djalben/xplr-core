package admin

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
	"golang.org/x/crypto/bcrypt"
)

var staffPinRe = regexp.MustCompile(`^\d{4}$`)

func (h *Handler) RegisterStaffPIN(r chi.Router) {
	r.Patch("/staff-pin", h.PatchStaffPIN)
}

func (h *Handler) PatchStaffPIN(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PIN string `json:"pin"`
	}

	if !adminReadJSON(w, r, &req) {
		return
	}

	pin := strings.TrimSpace(req.PIN)
	if !staffPinRe.MatchString(pin) {
		http.Error(w, "pin must be exactly 4 digits", http.StatusBadRequest)

		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	row := &domain.SystemSetting{
		Key:         "admin_pin",
		Value:       string(hash),
		Description: "PIN для входа в админку (4 цифры). Можно хранить как plain или bcrypt-хэш.",
	}

	err = h.systemRepo.Upsert(r.Context(), row)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSONWithContext(r.Context(), w, http.StatusOK, map[string]string{"status": "success"})
}
