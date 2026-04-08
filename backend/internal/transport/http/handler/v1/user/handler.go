package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/djalben/xplr-core/backend/internal/domain"
	"github.com/djalben/xplr-core/backend/internal/transport/http/handler"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type Handler struct {
	userUC   UserProfile
	walletUC UserWallet
	gradesUC UserGrades
	cardUC   UserCards
	txUC     UserTransactions
	ticketUC UserTickets
	totpUC   TOTPSettings
	kycUC    KYCApplications
}

func NewHandler(
	userUC UserProfile,
	walletUC UserWallet,
	gradesUC UserGrades,
	cardUC UserCards,
	txUC UserTransactions,
	ticketUC UserTickets,
	totpUC TOTPSettings,
	kycUC KYCApplications,
) *Handler {
	return &Handler{
		userUC:   userUC,
		walletUC: walletUC,
		gradesUC: gradesUC,
		cardUC:   cardUC,
		txUC:     txUC,
		ticketUC: ticketUC,
		totpUC:   totpUC,
		kycUC:    kycUC,
	}
}

func (h *Handler) Register(r chi.Router) {
	r.Route("/user", func(r chi.Router) {
		r.Get("/me", h.GetMe)
		r.Get("/grade", h.GetGrade)
		r.Get("/wallet", h.GetWallet)
		r.Post("/wallet/topup", h.TopUpWallet)
		r.Post("/wallet/transfer-to-card", h.TransferToCard)
		r.Patch("/wallet/auto-topup", h.SetAutoTopup)
		r.Post("/deposit", h.Deposit)
		r.Get("/transactions", h.GetTransactions)
		r.Post("/support", h.Support)
		r.Get("/referrals/info", h.GetReferralsInfo)
		r.Get("/cards", h.GetCards)
		r.Post("/cards/issue", h.IssueCards)
		r.Get("/cards/{id}/details", h.GetCardDetails)
		r.Patch("/cards/{id}/status", h.UpdateCardStatus)
		r.Patch("/cards/{id}/spending-limit", h.SetCardSpendingLimit)
		r.Post("/cards/{id}/spend", h.SpendFromCard)
		r.Post("/cards/{id}/failed-auth", h.RecordCardFailedAuth)
		r.Post("/cards/{id}/auto-replenishment", h.SetCardAutoReplenishment)
		r.Delete("/cards/{id}/auto-replenishment", h.UnsetCardAutoReplenishment)

		r.Patch("/me/notifications", h.PatchNotifications)
		r.Post("/me/telegram/link-code", h.PostTelegramLinkCode)
		r.Post("/me/telegram/link", h.PostTelegramLink)
		r.Post("/me/totp/setup", h.PostTOTPSetup)
		r.Post("/me/totp/confirm", h.PostTOTPConfirm)
		r.Post("/me/totp/disable", h.PostTOTPDisable)
		r.Post("/kyc/application", h.PostKYCApplication)
	})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	data, err := h.userUC.GetMe(r.Context(), userID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, data)
}

func (h *Handler) GetGrade(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	grade, err := h.gradesUC.GetByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"grade":       grade.Grade,
		"total_spent": grade.TotalSpent.String(),
		"fee_percent": grade.FeePercent.String(),
	})
}

func (h *Handler) GetWallet(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	balance, err := h.walletUC.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	autoEnabled, err := h.walletUC.GetAutoTopUpEnabled(r.Context(), userID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	// BFF: фронт ожидает master_balance.
	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"master_balance":        balance.String(),
		"auto_topup_enabled":    autoEnabled,
	})
}

func (h *Handler) TopUpWallet(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	type req struct {
		Amount float64 `json:"amount"`
	}

	var body req

	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.walletUC.TopUpWallet(r.Context(), userID, domain.NewNumeric(body.Amount))
	if err != nil {
		if errors.Is(err, domain.ErrSBPTopUpDisabled) {
			http.Error(w, err.Error(), http.StatusForbidden)

			return
		}

		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	balance, _ := h.walletUC.GetBalance(r.Context(), userID)
	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"master_balance": balance.String(),
	})
}

func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
	h.TopUpWallet(w, r)
}

func (h *Handler) Support(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	type req struct {
		Message string `json:"message"`
	}

	var body req

	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if body.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)

		return
	}

	_, err = h.ticketUC.Create(r.Context(), userID, "Support", body.Message, nil)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) GetReferralsInfo(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	data, err := h.userUC.GetReferralInfo(r.Context(), userID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, data)
}

func (h *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")
	limitStr := r.URL.Query().Get("limit")

	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()
	limit := 200

	if startStr != "" {
		t, err := time.Parse("2006-01-02", startStr)
		if err == nil {
			from = t
		}
	}

	if endStr != "" {
		t, err := time.Parse("2006-01-02", endStr)
		if err == nil {
			to = t
		}
	}

	if limitStr != "" {
		n, err := strconv.Atoi(limitStr)
		if err == nil && n > 0 {
			limit = n
		}
	}

	txs, err := h.txUC.GetUnifiedTransactions(r.Context(), userID, from, to, limit)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	type txResp struct {
		TransactionID   string `json:"transaction_id"`
		Amount          string `json:"amount"`
		Currency        string `json:"currency"`
		TransactionType string `json:"transaction_type"`
		SourceType      string `json:"source_type"`
		Details         string `json:"details"`
		ExecutedAt      string `json:"executed_at"`
		CardLast4       string `json:"card_last_4_digits"`
	}

	sourceMap := map[string]string{
		"TOPUP_WALLET":                  "wallet_topup",
		"TOPUP_CARD":                    "card_transfer",
		"CARD_ISSUE":                    "card_transfer",
		"AUTO_TOPUP":                    "card_transfer",
		domain.TransactionTypeCardSpend: "card_charge",
	}

	out := make([]txResp, 0, len(txs))

	for _, tx := range txs {
		srcType := sourceMap[tx.TransactionType]
		if srcType == "" {
			srcType = "card_charge"
		}

		cardLast4 := ""

		if tx.CardID != nil {
			c, errCard := h.cardUC.GetByID(r.Context(), *tx.CardID)
			if errCard == nil {
				cardLast4 = c.Last4Digits
			}
		}

		out = append(out, txResp{
			TransactionID:   tx.ID.String(),
			Amount:          tx.Amount.String(),
			Currency:        "USD",
			TransactionType: tx.TransactionType,
			SourceType:      srcType,
			Details:         tx.Details,
			ExecutedAt:      tx.ExecutedAt.Format(time.RFC3339),
			CardLast4:       cardLast4,
		})
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{"transactions": out})
}

func (h *Handler) TransferToCard(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	type req struct {
		CardID   string  `json:"card_id"`
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	}

	var body req

	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	cardID, err := domain.ParseUUID(body.CardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.cardUC.TopUpCard(r.Context(), userID, cardID, domain.NewNumeric(body.Amount))
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	balance, _ := h.walletUC.GetBalance(r.Context(), userID)
	handler.WriteJSON(w, http.StatusOK, map[string]any{"master_balance": balance.String()})
}

func (h *Handler) SetAutoTopup(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	type req struct {
		Enabled bool `json:"enabled"`
	}

	var body req

	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.walletUC.ToggleAutoTopUp(r.Context(), userID, body.Enabled)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) GetCards(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	cards, err := h.cardUC.ListByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	// BFF: маппинг в формат фронта.
	out := make([]map[string]any, 0, len(cards))

	for _, c := range cards {
		out = append(out, map[string]any{
			"id":                c.ID.String(),
			"user_id":           c.UserID.String(),
			"provider_card_id":  c.ProviderCardID,
			"bin":               c.Bin,
			"last_4_digits":     c.Last4Digits,
			"card_status":       c.CardStatus,
			"nickname":          c.Nickname,
			"daily_spend_limit": c.DailySpendLimit.String(),
			"card_type":         string(c.CardType),
			"balance":           c.Balance.String(),
			"created_at":        c.CreatedAt,
		})
	}

	handler.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) IssueCards(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	type req struct {
		Count       int    `json:"count"`
		Nickname    string `json:"nickname"`
		ServiceSlug string `json:"service_slug"`
	}

	var body req

	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	cardType := domain.CardTypeSubscriptions
	if body.ServiceSlug != "" {
		cardType = domain.CardType(body.ServiceSlug)
	}

	nickname := body.Nickname
	if nickname == "" {
		nickname = "Карта"
	}

	var results []map[string]any

	for range max(1, body.Count) {
		card, err := h.cardUC.BuyCard(r.Context(), userID, cardType, nickname)
		if err != nil {
			handler.WriteJSON(w, http.StatusOK, map[string]any{
				"successful_count": 0,
				"failed_count":     1,
				"results": []map[string]any{{
					"success": false, "message": err.Error(),
				}},
			})

			return
		}

		results = append(results, map[string]any{
			"success":     true,
			"status":      "issued",
			"card_last_4": card.Last4Digits,
			"nickname":    card.Nickname,
			"message":     "Карта выпущена",
			"card": map[string]any{
				"id": card.ID.String(), "last_4_digits": card.Last4Digits,
			},
		})
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"successful_count": len(results),
		"failed_count":     0,
		"results":          results,
	})
}

func (h *Handler) GetCardDetails(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	idStr := chi.URLParam(r, "id")
	cardID, err := domain.ParseUUID(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	card, err := h.cardUC.GetByID(r.Context(), cardID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusNotFound)

		return
	}

	if card.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"card_id":     card.ID.String(),
		"full_number": "424242******" + card.Last4Digits,
		"cvv":         "***",
		"expiry":      "MM/YY",
		"card_type":   string(card.CardType),
		"bin":         card.Bin,
		"last_4":      card.Last4Digits,
	})
}

func (h *Handler) UpdateCardStatus(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	idStr := chi.URLParam(r, "id")
	cardID, err := domain.ParseUUID(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	type req struct {
		Status string `json:"status"`
	}

	var body req

	err = handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.cardUC.UpdateStatus(r.Context(), userID, cardID, body.Status)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) SetCardSpendingLimit(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	idStr := chi.URLParam(r, "id")
	cardID, err := domain.ParseUUID(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	type req struct {
		SpendingLimit float64 `json:"spending_limit"`
	}

	var body req

	err = handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.cardUC.SetSpendingLimit(r.Context(), userID, cardID, domain.NewNumeric(body.SpendingLimit))
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// SpendFromCard — POST /user/cards/{id}/spend (списание с карты: лимиты день/месяц по типу).
func (h *Handler) SpendFromCard(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	idStr := chi.URLParam(r, "id")
	cardID, err := domain.ParseUUID(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	type req struct {
		Amount float64 `json:"amount"`
	}

	var body req

	err = handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.cardUC.SpendFromCard(r.Context(), userID, cardID, domain.NewNumeric(body.Amount))
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) SetCardAutoReplenishment(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	type req struct {
		Enabled   bool    `json:"enabled"`
		Threshold float64 `json:"threshold"`
		Amount    float64 `json:"amount"`
	}

	var body req

	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.walletUC.ToggleAutoTopUp(r.Context(), userID, body.Enabled)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) UnsetCardAutoReplenishment(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	err := h.walletUC.ToggleAutoTopUp(r.Context(), userID, false)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// RecordCardFailedAuth — POST /user/cards/{id}/failed-auth (неудачная авторизация у провайдера; антифрод).
func (h *Handler) RecordCardFailedAuth(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	idStr := chi.URLParam(r, "id")
	cardID, err := domain.ParseUUID(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.cardUC.RecordFailedAuthorization(r.Context(), userID, cardID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "recorded"})
}

func (h *Handler) PatchNotifications(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	var body struct {
		NotifyEmail    bool `json:"notify_email"`
		NotifyTelegram bool `json:"notify_telegram"`
	}

	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.userUC.SetNotificationPreferences(r.Context(), userID, body.NotifyEmail, body.NotifyTelegram)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) PostTelegramLinkCode(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	code, exp, err := h.userUC.IssueTelegramLinkCode(r.Context(), userID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusInternalServerError)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]any{
		"link_code":  code,
		"expires_at": exp.Format(time.RFC3339),
	})
}

func (h *Handler) PostTelegramLink(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	var body struct {
		TelegramChatID int64  `json:"telegram_chat_id"`
		LinkCode       string `json:"link_code"`
	}

	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.userUC.LinkTelegram(r.Context(), userID, body.TelegramChatID, body.LinkCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *Handler) PostTOTPSetup(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	url, sec, err := h.totpUC.SetupTOTP(r.Context(), userID)
	if err != nil {
		http.Error(w, wrapper.Wrap(err).Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"otpauth_url": url, "secret": sec})
}

func (h *Handler) PostTOTPConfirm(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	var body struct {
		Code string `json:"code"`
	}

	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.totpUC.ConfirmTOTP(r.Context(), userID, body.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "totp enabled"})
}

func (h *Handler) PostTOTPDisable(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	var body struct {
		Password string `json:"password"`
		Code     string `json:"code"`
	}

	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	err = h.totpUC.DisableTOTP(r.Context(), userID, body.Password, body.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "totp disabled"})
}

func (h *Handler) PostKYCApplication(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)

		return
	}

	var body struct {
		Payload map[string]any `json:"payload"`
	}

	err := handler.ReadJSON(r, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	payloadJSON := "{}"
	if len(body.Payload) > 0 {
		b, jerr := json.Marshal(body.Payload)
		if jerr != nil {
			http.Error(w, jerr.Error(), http.StatusBadRequest)

			return
		}

		payloadJSON = string(b)
	}

	err = h.kycUC.SubmitApplication(r.Context(), userID, payloadJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	handler.WriteJSON(w, http.StatusCreated, map[string]string{"status": "submitted"})
}
