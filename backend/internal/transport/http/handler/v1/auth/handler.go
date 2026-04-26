package auth

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/pkg/utils"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

const (
	roleUser  = "user"
	roleAdmin = "admin"

	trustedDeviceCookieName = "xplr_trusted_device"
)

type Handler struct {
	authUC    AuthFlow
	walletUC  WalletBalanceProvider
	userRepo  UserByIDReader
	limiter   RateLimiter
	jwtSecret []byte
}

func NewHandler(authUC AuthFlow, walletUC WalletBalanceProvider, userRepo UserByIDReader, limiter RateLimiter, jwtSecret []byte) *Handler {
	return &Handler{
		authUC:    authUC,
		walletUC:  walletUC,
		userRepo:  userRepo,
		limiter:   limiter,
		jwtSecret: jwtSecret,
	}
}

func (h *Handler) Register(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.DoRegister)
		r.Post("/login", h.DoLogin)
		r.Post("/login/mfa", h.DoLoginMFA)
		r.Get("/verify-email", h.VerifyEmail)
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
		if errors.Is(err, domain.ErrAlreadyExists) {
			http.Error(w, "email already registered", http.StatusConflict)

			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusCreated, map[string]any{
		"message":        "Регистрация успешна. Подтвердите email по ссылке из письма.",
		"email":          user.Email,
		"email_verified": user.EmailVerified,
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

	now := time.Now().UTC()
	ip := clientIP(r)
	ipStr := ""
	if ip != nil {
		ipStr = *ip
	}
	rlKey := "login:" + ipStr + ":" + strings.ToLower(strings.TrimSpace(req.Email))
	if h.limiter != nil {
		allowed, retryAfter, errAllow := h.limiter.Allow(r.Context(), rlKey, now)
		if errAllow != nil {
			http.Error(w, wrapper.Wrap(errAllow).Error(), http.StatusInternalServerError)

			return
		}
		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			http.Error(w, "Too many attempts", http.StatusTooManyRequests)

			return
		}
	}

	var trustedToken string
	c, errCookie := r.Cookie(trustedDeviceCookieName)
	if errCookie == nil && c != nil {
		trustedToken = strings.TrimSpace(c.Value)
	}

	out, err := h.authUC.LoginWithTrustedDevice(r.Context(), req.Email, req.Password, trustedToken, now)
	if err != nil {
		if h.limiter != nil {
			_, _ = h.limiter.Fail(r.Context(), rlKey, now)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if h.limiter != nil {
		_ = h.limiter.Success(r.Context(), rlKey, now)
	}

	if out.MFAToken != "" {
		handler.WriteJSON(w, http.StatusOK, map[string]any{
			"mfaRequired": true,
			"mfaToken":    out.MFAToken,
		})

		return
	}

	h.issueAuthToken(r.Context(), w, out.User)
}

func (h *Handler) DoLoginMFA(w http.ResponseWriter, r *http.Request) {
	type request struct {
		MFAToken       string `json:"mfaToken"`
		TOTPCode       string `json:"totpCode"`
		RememberDevice bool   `json:"rememberDevice"`
	}

	var req request

	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	now := time.Now().UTC()
	ip := clientIP(r)
	ipStr := ""
	if ip != nil {
		ipStr = *ip
	}

	userID, _, errParse := utils.ValidateMFAPendingJWT(h.jwtSecret, req.MFAToken)
	if errParse != nil {
		http.Error(w, "invalid or expired mfa token", http.StatusBadRequest)

		return
	}

	rlKey := "mfa:" + ipStr + ":" + userID.String()
	if h.limiter != nil {
		allowed, retryAfter, errAllow := h.limiter.Allow(r.Context(), rlKey, now)
		if errAllow != nil {
			http.Error(w, wrapper.Wrap(errAllow).Error(), http.StatusInternalServerError)

			return
		}
		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			http.Error(w, "Too many attempts", http.StatusTooManyRequests)

			return
		}
	}

	user, err := h.authUC.CompleteMFALogin(r.Context(), req.MFAToken, req.TOTPCode)
	if err != nil {
		if h.limiter != nil {
			_, _ = h.limiter.Fail(r.Context(), rlKey, now)
		}
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if h.limiter != nil {
		_ = h.limiter.Success(r.Context(), rlKey, now)
	}

	if req.RememberDevice {
		ua := strings.TrimSpace(r.UserAgent())
		raw, exp, errRemember := h.authUC.RememberTrustedDevice(r.Context(), user.ID, ua, ip, now)
		if errRemember != nil {
			http.Error(w, errRemember.Error(), http.StatusBadRequest)

			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     trustedDeviceCookieName,
			Value:    raw,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Expires:  exp,
		})
	}

	h.issueAuthToken(r.Context(), w, user)
}

func clientIP(r *http.Request) *string {
	// chi/middleware.RealIP уже может выставить корректный RemoteAddr, но надёжнее разобрать host:port.
	addr := strings.TrimSpace(r.RemoteAddr)
	if addr == "" {
		return nil
	}

	host, _, err := net.SplitHostPort(addr)
	if err == nil && host != "" {
		return &host
	}

	// На случай если RemoteAddr уже без порта.
	return &addr
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)

		return
	}

	err := h.authUC.VerifyEmail(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "email verified"})
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

	if !user.EmailVerified {
		http.Error(w, "email not verified", http.StatusForbidden)

		return
	}

	newToken, err := utils.GenerateJWT(h.jwtSecret, user.ID, user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	balance, _ := h.walletUC.GetBalance(r.Context(), userID)
	balanceStr := balance.String()

	h.writeAuthSuccess(w, newToken, user, balanceStr)
}

func (h *Handler) ResetPasswordRequest(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email string `json:"email"`
	}

	var req request

	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.authUC.RequestPasswordReset(r.Context(), req.Email)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "Если email зарегистрирован, мы отправили инструкции.",
	})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}

	var req request

	err := handler.ReadJSON(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if req.Token == "" {
		req.Token = r.URL.Query().Get("token")
	}

	err = h.authUC.ResetPassword(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "password updated"})
}

func (h *Handler) issueAuthToken(ctx context.Context, w http.ResponseWriter, user *domain.User) {
	token, err := utils.GenerateJWT(h.jwtSecret, user.ID, user.Email)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	balance, _ := h.walletUC.GetBalance(ctx, user.ID)
	balanceStr := balance.String()

	h.writeAuthSuccess(w, token, user, balanceStr)
}

func (h *Handler) writeAuthSuccess(w http.ResponseWriter, token string, user *domain.User, balanceStr string) {
	role := roleUser
	if user.IsAdmin {
		role = roleAdmin
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user": map[string]any{
			"id":             user.ID.String(),
			"email":          user.Email,
			"balance":        balanceStr,
			"status":         string(user.Status),
			"is_admin":       user.IsAdmin,
			"role":           role,
			"email_verified": user.EmailVerified,
			"totp_enabled":   user.TOTPEnabled,
		},
	})
}
