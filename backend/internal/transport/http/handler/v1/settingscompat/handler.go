// Package settingscompat — маршруты /user/settings/* под фронт партнёра (main), прокси на наш доменный слой.
package settingscompat

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/djalben/xplr-core/backend/internal/application/auth"
	"github.com/djalben/xplr-core/backend/internal/application/kyc"
	"github.com/djalben/xplr-core/backend/internal/application/user"
	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/ports"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// Handler — совместимость с epn-killer-web (ветка main).
type Handler struct {
	userUC       *user.UseCase
	authUC       *auth.UseCase
	kycUC        *kyc.UseCase
	kycRepo      ports.KYCApplicationRepository
	sessionsRepo ports.AuthSessionsRepository
	botUsername  string
}

func NewHandler(
	userUC *user.UseCase,
	authUC *auth.UseCase,
	kycUC *kyc.UseCase,
	kycRepo ports.KYCApplicationRepository,
	sessionsRepo ports.AuthSessionsRepository,
	botUsername string,
) *Handler {
	botUsername = strings.TrimPrefix(strings.TrimSpace(botUsername), "@")

	return &Handler{
		userUC:       userUC,
		authUC:       authUC,
		kycUC:        kycUC,
		kycRepo:      kycRepo,
		sessionsRepo: sessionsRepo,
		botUsername:  botUsername,
	}
}

func (h *Handler) Register(r chi.Router) {
	r.Route("/user/settings", func(r chi.Router) {
		r.Get("/profile", h.GetProfile)
		r.Patch("/profile", h.PatchProfile)
		r.Get("/telegram-link", h.GetTelegramLink)
		r.Get("/telegram/check-status", h.TelegramCheckStatus)
		r.Post("/telegram/unlink", h.TelegramUnlink)
		r.Post("/verify-email-request", h.VerifyEmailRequest)
		r.Post("/verify-email-confirm", h.VerifyEmailConfirm)
		r.Get("/sessions", h.GetSessions)
		r.Post("/change-password", h.ChangePassword)
		r.Post("/2fa/setup", h.TwoFASetup)
		r.Post("/2fa/verify", h.TwoFAVerify)
		r.Post("/2fa/disable", h.TwoFADisable)
		r.Get("/notifications", h.GetNotifications)
		r.Patch("/notifications", h.PatchNotifications)
		r.Post("/logout-all", h.LogoutAll)
		r.Get("/kyc", h.GetKYC)
		r.Post("/kyc", h.PostKYC)
	})
}

func notificationPrefFromUser(u *domain.User) string {
	if u.NotifyEmail && u.NotifyTelegram {
		return "both"
	}

	if u.NotifyEmail {
		return "email"
	}

	return "telegram"
}

func verificationStatus(ks domain.KYCStatus) string {
	switch ks {
	case domain.KYCApproved:
		return "verified"
	case domain.KYCRejected:
		return "rejected"
	case domain.KYCPending:
		return "pending"
	default:
		return "pending"
	}
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	u, err := h.userUC.GetByID(r.Context(), uid)
	if err != nil {
		_ = handler.WriteInternalServerError(r.Context(), w, err)

		return
	}

	displayName := u.Email
	if at := strings.Index(u.Email, "@"); at > 0 {
		displayName = u.Email[:at]
	}

	role := "user"
	if u.IsAdmin {
		role = "admin"
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"id":                  u.ID.String(),
		"email":               u.Email,
		"display_name":        displayName,
		"is_verified":         u.EmailVerified,
		"verification_status": verificationStatus(u.KYCStatus),
		"two_factor_enabled":  u.TOTPEnabled,
		"telegram_linked":     u.TelegramChatID != nil && *u.TelegramChatID != 0,
		"role":                role,
		"is_admin":            u.IsAdmin,
	})
}

func (h *Handler) PatchProfile(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DisplayName string `json:"display_name"` //nolint:tagliatelle // контракт фронта партнёра (snake_case)
	}

	err := handler.ReadJSON(r, &body)
	if err != nil {
		_ = handler.WrapAndWriteError(r.Context(), w, err, http.StatusBadRequest, "Неверный запрос")

		return
	}

	_ = body.DisplayName
	// display_name в БД пока не хранится.
	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) GetTelegramLink(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	code, _, err := h.userUC.IssueTelegramLinkCode(r.Context(), uid)
	if err != nil {
		_ = handler.WrapAndWriteError(r.Context(), w, err, http.StatusBadRequest, "Не удалось создать ссылку")

		return
	}

	bot := h.botUsername
	if bot == "" {
		bot = "xplr_notify_bot"
	}

	deepLink := "https://t.me/" + bot + "?start=" + code

	handler.WriteJSON(w, http.StatusOK, map[string]string{
		"link": deepLink,
		"code": code,
	})
}

func (h *Handler) TelegramCheckStatus(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	u, err := h.userUC.GetByID(r.Context(), uid)
	if err != nil {
		_ = handler.WriteInternalServerError(r.Context(), w, err)

		return
	}

	linked := u.TelegramChatID != nil && *u.TelegramChatID != 0

	handler.WriteJSON(w, http.StatusOK, map[string]bool{"linked": linked})
}

func (h *Handler) TelegramUnlink(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	err := h.userUC.UnlinkTelegram(r.Context(), uid)
	if err != nil {
		_ = handler.WrapAndWriteError(r.Context(), w, err, http.StatusBadRequest, "Не удалось отвязать Telegram")

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) VerifyEmailRequest(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	err := h.authUC.ResendEmailVerification(r.Context(), uid)
	if err != nil {
		_ = handler.WrapAndWriteError(r.Context(), w, err, http.StatusBadRequest, "Не удалось отправить письмо")

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func (h *Handler) VerifyEmailConfirm(w http.ResponseWriter, _ *http.Request) {
	handler.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Откройте ссылку из письма для подтверждения email.",
	})
}

func (h *Handler) GetSessions(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	list, err := h.sessionsRepo.ListByUserID(r.Context(), uid, 50)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	out := make([]any, 0, len(list))
	for _, s := range list {
		if s == nil {
			continue
		}
		ip := ""
		if s.IP != nil {
			ip = *s.IP
		}
		out = append(out, map[string]any{
			"id":          s.ID.String(),
			"last_active": s.CreatedAt,
			"ip":          ip,
			"device":      s.UserAgent,
		})
	}

	handler.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	var body struct {
		OldPassword string `json:"old_password"` //nolint:tagliatelle // контракт фронта партнёра
		NewPassword string `json:"new_password"` //nolint:tagliatelle // контракт фронта партнёра
	}

	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.userUC.ChangePassword(r.Context(), uid, body.OldPassword, body.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) TwoFASetup(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	url, sec, err := h.authUC.SetupTOTP(r.Context(), uid)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{
		"secret":  sec,
		"otp_uri": url,
	})
}

func (h *Handler) TwoFAVerify(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	var body struct {
		Code string `json:"code"`
	}

	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, "Введите 6‑значный код из Google Authenticator", http.StatusBadRequest)

		return
	}

	err = h.authUC.ConfirmTOTP(r.Context(), uid, body.Code)
	if err != nil {
		http.Error(w, twoFAUserMessage(err), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) TwoFADisable(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	var body struct {
		Password string `json:"password"`
		Code     string `json:"code"`
	}

	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, "Нужны пароль и код из Google Authenticator", http.StatusBadRequest)

		return
	}

	if body.Password == "" || body.Code == "" {
		http.Error(w, "Нужны пароль и код из Google Authenticator", http.StatusBadRequest)

		return
	}

	err = h.authUC.DisableTOTP(r.Context(), uid, body.Password, body.Code)
	if err != nil {
		http.Error(w, twoFAUserMessage(err), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func twoFAUserMessage(err error) string {
	if err == nil {
		return "Ошибка"
	}

	// Скрываем внутренности wrapper (пути к файлам и т.д.).
	if errors.Is(err, domain.ErrInvalidInput) {
		s := err.Error()
		switch {
		case strings.Contains(s, "invalid totp code"):
			return "Неверный код из Google Authenticator"
		case strings.Contains(s, "code is required"):
			return "Введите 6‑значный код из Google Authenticator"
		case strings.Contains(s, "run totp setup first"):
			return "Сначала настройте Google Authenticator"
		case strings.Contains(s, "invalid password"):
			return "Неверный пароль"
		case strings.Contains(s, "totp is not enabled"):
			return "2FA уже отключена"
		case strings.Contains(s, "password and totp code are required"):
			return "Нужны пароль и код из Google Authenticator"
		}

		return "Неверные данные"
	}

	return "Ошибка. Попробуйте позже."
}

func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	u, err := h.userUC.GetByID(r.Context(), uid)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"notify_transactions":    u.NotifyTransactions,
		"notify_balance":         u.NotifyBalance,
		"notify_security":        u.NotifySecurity,
		"notify_card_operations": u.NotifyCardOperations,
		"notification_pref":      notificationPrefFromUser(u),
	})
}

func (h *Handler) PatchNotifications(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	var raw map[string]any
	err := handler.ReadJSON(r, &raw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	patch := user.PartnerNotifPatch{}

	if v, ok := raw["notification_pref"].(string); ok {
		patch.NotificationPref = &v
	}

	if v, ok := raw["notify_transactions"].(bool); ok {
		patch.NotifyTransactions = &v
	}

	if v, ok := raw["notify_balance"].(bool); ok {
		patch.NotifyBalance = &v
	}

	if v, ok := raw["notify_security"].(bool); ok {
		patch.NotifySecurity = &v
	}

	if v, ok := raw["notify_card_operations"].(bool); ok {
		patch.NotifyCardOperations = &v
	}

	err = h.userUC.PatchPartnerNotificationSettings(r.Context(), uid, patch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	err := h.authUC.RevokeAllTrustedDevices(r.Context(), uid)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	_ = h.sessionsRepo.DeleteByUserID(r.Context(), uid)

	// Удаляем trusted-device cookie на текущем устройстве.
	http.SetCookie(w, &http.Cookie{
		Name:     "xplr_trusted_device",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) GetKYC(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	u, err := h.userUC.GetByID(r.Context(), uid)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	app, err := h.kycRepo.GetLatestByUserID(r.Context(), uid)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	st := "none"
	if app != nil {
		st = strings.ToLower(string(app.Status))
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"status":          st,
		"kyc_user_status": string(u.KYCStatus),
		"payload":         app,
	})
}

func (h *Handler) PostKYC(w http.ResponseWriter, r *http.Request) {
	uid := handler.GetUserIDFromContext(r)

	var body map[string]any
	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	b, err := json.Marshal(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.kycUC.SubmitApplication(r.Context(), uid, string(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusCreated, map[string]string{"status": "submitted"})
}
