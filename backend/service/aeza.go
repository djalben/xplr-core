package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// ══════════════════════════════════════════════════════════════
// Aeza Hosting Balance Monitor
// Checks account balance via Aeza API and alerts admins
// when it drops below the configured threshold.
// ══════════════════════════════════════════════════════════════

// AezaBalance holds the current balance info.
type AezaBalance struct {
	BalanceRUB float64 `json:"balance_rub"`
	Currency   string  `json:"currency"`
	UpdatedAt  string  `json:"updated_at"`
}

var (
	aezaLastAlert   time.Time
	aezaAlertMu     sync.Mutex
	aezaAlertCooldown = 6 * time.Hour
)

// GetAezaBalance fetches the current account balance from the Aeza API.
// Returns nil if AEZA_API_KEY is not configured.
func GetAezaBalance() (*AezaBalance, error) {
	apiKey := os.Getenv("AEZA_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("AEZA_API_KEY not configured")
	}

	req, err := http.NewRequest("GET", "https://my.aeza.net/api/balance", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aeza API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("aeza API returned %d: %s", resp.StatusCode, string(body))
	}

	// Aeza API response format: {"data": {"balance": "1234.56"}}
	var result struct {
		Data struct {
			Balance string `json:"balance"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse aeza response: %w (body: %s)", err, string(body))
	}

	balanceVal, err := strconv.ParseFloat(result.Data.Balance, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance value %q: %w", result.Data.Balance, err)
	}

	return &AezaBalance{
		BalanceRUB: balanceVal,
		Currency:   "RUB",
		UpdatedAt:  time.Now().Format(time.RFC3339),
	}, nil
}

// CheckAezaBalanceAndNotify checks the Aeza balance and sends
// admin notifications if it drops below the threshold.
// Called periodically from a background goroutine.
func CheckAezaBalanceAndNotify() {
	thresholdStr := os.Getenv("AEZA_BALANCE_THRESHOLD")
	threshold := 500.0 // default 500 RUB
	if thresholdStr != "" {
		if v, err := strconv.ParseFloat(thresholdStr, 64); err == nil {
			threshold = v
		}
	}

	balance, err := GetAezaBalance()
	if err != nil {
		log.Printf("[AEZA] ⚠️ Balance check failed: %v", err)
		return
	}

	log.Printf("[AEZA] 💰 Current balance: %.2f RUB (threshold: %.0f RUB)", balance.BalanceRUB, threshold)

	if balance.BalanceRUB < threshold {
		aezaAlertMu.Lock()
		canAlert := time.Since(aezaLastAlert) > aezaAlertCooldown
		if canAlert {
			aezaLastAlert = time.Now()
		}
		aezaAlertMu.Unlock()

		if canAlert {
			subject := "⚠️ Низкий баланс Aeza"
			msg := fmt.Sprintf(
				"<b>⚠️ Низкий баланс инфраструктуры</b>\n\n"+
					"Баланс Aeza: <b>%.2f ₽</b>\n"+
					"Порог: <b>%.0f ₽</b>\n\n"+
					"Необходимо пополнить баланс хостинга для продолжения работы VPN-серверов.\n\n"+
					"<a href=\"https://my.aeza.net/billing\">Перейти в Aeza</a>",
				balance.BalanceRUB, threshold)

			NotifyAdmins(subject, msg)
			log.Printf("[AEZA] 🚨 Low balance alert sent: %.2f RUB < %.0f RUB", balance.BalanceRUB, threshold)
		} else {
			log.Printf("[AEZA] ⚠️ Low balance (%.2f RUB) but alert cooldown active", balance.BalanceRUB)
		}
	}
}

// StartAezaBalanceMonitor starts a background goroutine that checks
// the Aeza balance periodically (every 4 hours).
func StartAezaBalanceMonitor() {
	apiKey := os.Getenv("AEZA_API_KEY")
	if apiKey == "" {
		log.Println("[AEZA] ⚠️ AEZA_API_KEY not set — balance monitor disabled")
		return
	}

	go func() {
		// Initial check after 30 seconds
		time.Sleep(30 * time.Second)
		CheckAezaBalanceAndNotify()

		ticker := time.NewTicker(4 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			CheckAezaBalanceAndNotify()
		}
	}()

	log.Println("[AEZA] ✅ Balance monitor started (interval=4h, threshold from AEZA_BALANCE_THRESHOLD or 500 RUB)")
}
