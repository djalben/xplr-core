package auth

import (
	"net/http"

	"github.com/djalben/xplr-core/backend/internal/pkg/utils"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type Handler struct {
	authUC    AuthRegisterLogin
	walletUC  WalletBalanceProvider
	userRepo  UserByIDReader
	jwtSecret []byte
}

func NewHandler(authUC AuthRegisterLogin, walletUC WalletBalanceProvider, userRepo UserByIDReader, jwtSecret []byte) *Handler {
	return &Handler{
		authUC:    authUC,
		walletUC:  walletUC,
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (h *Handler) Register(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.DoRegister)
		r.Post("/login", h.DoLogin)
		r.Post("/refresh-token", h.RefreshToken)
		r.Post("/reset-password-request", h.ResetPasswordRequest)
		r.Post("/reset-password", h.ResetPassword)
	})
}

func (h *Handler) DoRegister(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req request

	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	user, err := h.authUC.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	token, err := utils.GenerateJWT(h.jwtSecret, user.ID, user.Email)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	balance, _ := h.walletUC.GetBalance(r.Context(), user.ID)
	balanceStr := balance.String()

	role := "user"
	if user.IsAdmin {
		role = "admin"
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user": map[string]any{
			"id":      user.ID.String(),
			"email":   user.Email,
			"balance": balanceStr,
			"status":  string(user.Status),
			"is_admin": user.IsAdmin,
			"role":    role,
		},
	})
}

func (h *Handler) DoLogin(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req request

	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	user, err := h.authUC.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	token, err := utils.GenerateJWT(h.jwtSecret, user.ID, user.Email)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	balance, _ := h.walletUC.GetBalance(r.Context(), user.ID)
	balanceStr := balance.String()

	role := "user"
	if user.IsAdmin {
		role = "admin"
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user": map[string]any{
			"id":      user.ID.String(),
			"email":   user.Email,
			"balance": balanceStr,
			"status":  string(user.Status),
			"is_admin": user.IsAdmin,
			"role":    role,
		},
	})
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.Header.Get("Authorization")
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	userID, err := utils.ValidateJWT(h.jwtSecret, tokenStr)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)

		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)

		return
	}

	newToken, err := utils.GenerateJWT(h.jwtSecret, user.ID, user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	balance, _ := h.walletUC.GetBalance(r.Context(), user.ID)
	balanceStr := balance.String()

	role := "user"
	if user.IsAdmin {
		role = "admin"
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"token": newToken,
		"user": map[string]any{
			"id":      user.ID.String(),
			"email":   user.Email,
			"balance": balanceStr,
			"status":  string(user.Status),
			"is_admin": user.IsAdmin,
			"role":    role,
		},
	})
}

func (h *Handler) ResetPasswordRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "reset password not implemented", http.StatusNotImplemented)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "reset password not implemented", http.StatusNotImplemented)
}
