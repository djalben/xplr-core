package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
	"github.com/shopspring/decimal"
)

// ExternalTopUpPayload — входящий webhook от зарубежной организации
// для подтверждения пополнения (депозита) пользователя.
type ExternalTopUpPayload struct {
	UserID       int    `json:"user_id"`
	ExternalTxID string `json:"external_tx_id"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	Status       string `json:"status"` // "confirmed", "pending", "failed"
	ProviderName string `json:"provider_name,omitempty"`
}

// ExternalTopUpWebhookHandler — POST /api/v1/webhooks/external-topup
// Принимает входящий вебхук от зарубежной организации, проверяет подпись,
// и зачисляет средства в Кошелёк пользователя.
func ExternalTopUpWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if repository.GlobalDB == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	// Decode payload
	var payload ExternalTopUpPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if payload.UserID <= 0 || payload.ExternalTxID == "" || payload.Amount == "" {
		http.Error(w, "Missing required fields: user_id, external_tx_id, amount", http.StatusBadRequest)
		return
	}

	// Only process confirmed transactions
	if payload.Status != "confirmed" {
		log.Printf("[EXT-WEBHOOK] Ignoring non-confirmed tx: %s status=%s", payload.ExternalTxID, payload.Status)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "not confirmed"})
		return
	}

	// Idempotency check
	exists, err := repository.CheckTransactionIdempotency(payload.ExternalTxID)
	if err != nil {
		log.Printf("[EXT-WEBHOOK] Idempotency check error: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if exists {
		log.Printf("[EXT-WEBHOOK] Duplicate tx: %s, skipping", payload.ExternalTxID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "duplicate"})
		return
	}

	// Parse amount
	amount, err := decimal.NewFromString(payload.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	// Atomic: credit Wallet + record transaction
	tx, err := repository.GlobalDB.Begin()
	if err != nil {
		http.Error(w, "Transaction begin failed", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Upsert into internal_balances
	_, err = tx.Exec(
		`INSERT INTO internal_balances (user_id, master_balance, updated_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (user_id) DO UPDATE SET master_balance = internal_balances.master_balance + $2, updated_at = NOW()`,
		payload.UserID, amount,
	)
	if err != nil {
		log.Printf("[EXT-WEBHOOK] Failed to credit wallet: %v", err)
		http.Error(w, "Failed to credit wallet", http.StatusInternalServerError)
		return
	}

	// Record transaction with provider_tx_id for idempotency
	providerName := payload.ProviderName
	if providerName == "" {
		providerName = "external"
	}
	_, err = tx.Exec(
		`INSERT INTO transactions (user_id, amount, fee, transaction_type, status, details, provider_tx_id, executed_at)
		 VALUES ($1, $2, 0, 'DEPOSIT', 'APPROVED', $3, $4, $5)`,
		payload.UserID, amount,
		fmt.Sprintf("External top-up via %s: +%s %s", providerName, amount.String(), payload.Currency),
		payload.ExternalTxID,
		time.Now(),
	)
	if err != nil {
		log.Printf("[EXT-WEBHOOK] Failed to record tx: %v", err)
		http.Error(w, "Failed to record transaction", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Commit failed", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [EXT-WEBHOOK] Credited %s %s to wallet (user %d, tx=%s)",
		amount.String(), payload.Currency, payload.UserID, payload.ExternalTxID)
	log.Printf("[EVENT] User %d performed external_topup (amount=%s %s, tx=%s). Triggering notifications...",
		payload.UserID, amount.String(), payload.Currency, payload.ExternalTxID)

	// Notify user about successful top-up
	go func() {
		service.NotifyUser(payload.UserID, "Пополнение баланса",
			fmt.Sprintf("💰 <b>Кошелёк пополнен</b>\n\n"+
				"Сумма: <b>%s %s</b>\n"+
				"Источник: <b>%s</b>\n\n"+
				"<a href=\"https://xplr.pro/wallet\">Открыть кошелёк</a>",
				amount.String(), payload.Currency, providerName))
		log.Printf("[NOTIFY] Message sent to UserID: %d via NotifyUser (external topup)", payload.UserID)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "credited",
		"tx_id":   payload.ExternalTxID,
		"amount":  amount.String(),
		"user_id": fmt.Sprintf("%d", payload.UserID),
	})
}
