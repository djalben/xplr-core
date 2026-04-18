package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
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
	Balance   float64 `json:"balance"`
	Currency  string  `json:"currency"`
	UpdatedAt string  `json:"updated_at"`
}

// cleanBalanceString removes currency symbols and whitespace, replaces comma decimal separators.
var balanceCleanRe = regexp.MustCompile(`[^\d.,\-]`)

var (
	aezaLastAlert     time.Time
	aezaAlertMu       sync.Mutex
	aezaAlertCooldown = 6 * time.Hour
)

// GetAezaBalance fetches the current account balance from the Aeza API.
// Returns nil if AEZA_API_KEY is not configured.
func GetAezaBalance() (*AezaBalance, error) {
	apiKey := os.Getenv("AEZA_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("AEZA_API_KEY not configured")
	}

	req, err := http.NewRequest("GET", "https://my.aeza.net/api/account", nil)
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

	log.Printf("[AEZA] Raw API response (status %d): %s", resp.StatusCode, string(body))

	// Aeza API response: {"data": {"balance": 1234.56, ...}} or {"data": {"balance": "12.34 \u20ac", ...}}
	// Try float first, then fall back to string parsing
	var result struct {
		Data struct {
			Balance json.RawMessage `json:"balance"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse aeza response: %w (body: %s)", err, string(body))
	}

	var balanceVal float64
	// Try parsing as float64 directly
	if err := json.Unmarshal(result.Data.Balance, &balanceVal); err != nil {
		// Fall back: parse as string, strip currency symbols
		var balanceStr string
		if err2 := json.Unmarshal(result.Data.Balance, &balanceStr); err2 != nil {
			return nil, fmt.Errorf("failed to parse balance value: float err=%v, string err=%v, raw=%s", err, err2, string(result.Data.Balance))
		}
		cleaned := balanceCleanRe.ReplaceAllString(balanceStr, "")
		cleaned = strings.TrimSpace(cleaned)
		cleaned = strings.Replace(cleaned, ",", ".", 1) // handle comma decimal separator
		balanceVal, err = strconv.ParseFloat(cleaned, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse cleaned balance %q (from %q): %w", cleaned, balanceStr, err)
		}
		log.Printf("[AEZA] Parsed balance from string: %q -> %.2f", balanceStr, balanceVal)
	}

	log.Printf("[AEZA] Parsed balance: %.4f", balanceVal)

	return &AezaBalance{
		Balance:   balanceVal,
		Currency:  "EUR",
		UpdatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// CheckAezaBalanceAndNotify checks the Aeza balance and sends
// admin notifications if it drops below the threshold.
// Called periodically from a background goroutine.
func CheckAezaBalanceAndNotify() {
	thresholdStr := os.Getenv("AEZA_BALANCE_THRESHOLD")
	threshold := 2.0 // default 2 EUR
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

	log.Printf("[AEZA] 💰 Current balance: %.2f EUR (threshold: %.2f EUR)", balance.Balance, threshold)

	if balance.Balance < threshold {
		aezaAlertMu.Lock()
		canAlert := time.Since(aezaLastAlert) > aezaAlertCooldown
		if canAlert {
			aezaLastAlert = time.Now()
		}
		aezaAlertMu.Unlock()

		if canAlert {
			subject := "\u26a0\ufe0f \u0411\u0430\u043b\u0430\u043d\u0441 Aeza \u043d\u0430 \u0438\u0441\u0445\u043e\u0434\u0435!"
			msg := fmt.Sprintf(
				"<b>\u26a0\ufe0f \u0412\u043d\u0438\u043c\u0430\u043d\u0438\u0435! \u0411\u0430\u043b\u0430\u043d\u0441 Aeza (XPLR Infrastructure) \u043d\u0430 \u0438\u0441\u0445\u043e\u0434\u0435: %.2f \u20ac</b>\n\n"+
					"\u041f\u043e\u0436\u0430\u043b\u0443\u0439\u0441\u0442\u0430, \u043f\u043e\u043f\u043e\u043b\u043d\u0438\u0442\u0435 \u0441\u0447\u0435\u0442, \u0447\u0442\u043e\u0431\u044b \u0438\u0437\u0431\u0435\u0436\u0430\u0442\u044c \u043e\u0442\u043a\u043b\u044e\u0447\u0435\u043d\u0438\u044f \u0441\u0435\u0440\u0432\u0435\u0440\u043e\u0432.\n"+
					"\u0421\u0440\u043e\u0447\u043d\u043e \u043f\u043e\u043f\u043e\u043b\u043d\u0438\u0442\u0435 \u0441\u0447\u0435\u0442 \u0434\u043b\u044f \u0440\u0430\u0431\u043e\u0442\u044b \u0411\u0435\u0437\u043e\u043f\u0430\u0441\u043d\u043e\u0433\u043e \u0434\u043e\u0441\u0442\u0443\u043f\u0430.\n\n"+
					"\u041f\u043e\u0440\u043e\u0433: <b>%.2f \u20ac</b>\n"+
					"<a href=\"https://my.aeza.net/billing\">\u041f\u0435\u0440\u0435\u0439\u0442\u0438 \u0432 Aeza \u2192</a>",
				balance.Balance, threshold)

			NotifyAdmins(subject, msg)
			log.Printf("[AEZA] 🚨 Low balance alert sent: %.2f EUR < %.2f EUR", balance.Balance, threshold)
		} else {
			log.Printf("[AEZA] ⚠️ Low balance (%.2f EUR) but alert cooldown active", balance.Balance)
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

	log.Println("[AEZA] ✅ Balance monitor started (interval=4h, threshold from AEZA_BALANCE_THRESHOLD or 2 EUR)")
}
