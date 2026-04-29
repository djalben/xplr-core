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
	Status    string  `json:"status"` // "ok", "maintenance", "error"
}

// cleanBalanceString removes currency symbols and whitespace, replaces comma decimal separators.
var balanceCleanRe = regexp.MustCompile(`[^\d.,\-]`)

const (
	aezaMaxRetries = 3
	aezaBaseDelay  = 1 * time.Second
	aezaPrimaryURL = "https://my.aeza.net/api/account"
)

var (
	aezaLastAlert     time.Time
	aezaAlertMu       sync.Mutex
	aezaAlertCooldown = 6 * time.Hour
)

// doAezaRequest performs a single HTTP request to the Aeza API.
// Returns body, status code, and error.
func doAezaRequest(apiKey string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", aezaPrimaryURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-API-Key", strings.TrimSpace(apiKey))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("aeza API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return body, resp.StatusCode, nil
}

// extractTraceID tries to pull a traceId from an Aeza error JSON response.
func extractTraceID(body []byte) string {
	var errResp struct {
		TraceID string `json:"traceId"`
	}
	if json.Unmarshal(body, &errResp) == nil && errResp.TraceID != "" {
		return errResp.TraceID
	}
	// fallback: look for traceId in raw body
	var raw map[string]any
	if json.Unmarshal(body, &raw) == nil {
		if tid, ok := raw["traceId"]; ok {
			return fmt.Sprintf("%v", tid)
		}
		if tid, ok := raw["trace_id"]; ok {
			return fmt.Sprintf("%v", tid)
		}
	}
	return ""
}

// GetAezaBalance fetches the current account balance from the Aeza API.
// Implements retry logic (up to 3 attempts) with exponential backoff.
// On persistent 5xx errors, returns a MAINTENANCE status instead of an error.
func GetAezaBalance() (*AezaBalance, error) {
	apiKey := os.Getenv("AEZA_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("AEZA_API_KEY not configured")
	}

	var lastBody []byte
	var lastStatus int
	var lastErr error

	for attempt := 1; attempt <= aezaMaxRetries; attempt++ {
		body, status, err := doAezaRequest(apiKey)
		lastBody = body
		lastStatus = status
		lastErr = err

		if err != nil {
			log.Printf("[AEZA] ⚠️ Attempt %d/%d network error: %v", attempt, aezaMaxRetries, err)
			if attempt < aezaMaxRetries {
				time.Sleep(aezaBaseDelay * time.Duration(attempt))
			}
			continue
		}

		// Success — parse the response
		if status == http.StatusOK {
			log.Printf("[AEZA] Raw API response (attempt %d, status %d): %s", attempt, status, string(body))
			return parseAezaBalanceBody(body)
		}

		// 5xx — log traceId and retry
		if status >= 500 {
			traceID := extractTraceID(body)
			traceLog := ""
			if traceID != "" {
				traceLog = fmt.Sprintf(" traceId=%s", traceID)
			}
			log.Printf("[AEZA] ⚠️ Attempt %d/%d server error %d%s: %s", attempt, aezaMaxRetries, status, traceLog, string(body))
			if attempt < aezaMaxRetries {
				time.Sleep(aezaBaseDelay * time.Duration(attempt))
			}
			continue
		}

		// 4xx — no point retrying
		return nil, fmt.Errorf("aeza API returned %d: %s", status, string(body))
	}

	// All retries exhausted — return graceful MAINTENANCE status for 5xx
	if lastStatus >= 500 {
		traceID := extractTraceID(lastBody)
		log.Printf("[AEZA] 🔧 All %d retries failed (last status %d, traceId=%s). Returning MAINTENANCE status.", aezaMaxRetries, lastStatus, traceID)
		return &AezaBalance{
			Balance:   -1,
			Currency:  "EUR",
			UpdatedAt: time.Now().Format(time.RFC3339),
			Status:    "maintenance",
		}, nil
	}

	// Network error after all retries
	return nil, fmt.Errorf("aeza API unreachable after %d attempts: %w", aezaMaxRetries, lastErr)
}

// parseAezaBalanceBody parses a successful Aeza API response body into AezaBalance.
func parseAezaBalanceBody(body []byte) (*AezaBalance, error) {
	var result struct {
		Data struct {
			Balance json.RawMessage `json:"balance"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse aeza response: %w (body: %s)", err, string(body))
	}

	var balanceVal float64
	if err := json.Unmarshal(result.Data.Balance, &balanceVal); err != nil {
		var balanceStr string
		if err2 := json.Unmarshal(result.Data.Balance, &balanceStr); err2 != nil {
			return nil, fmt.Errorf("failed to parse balance value: float err=%v, string err=%v, raw=%s", err, err2, string(result.Data.Balance))
		}
		cleaned := balanceCleanRe.ReplaceAllString(balanceStr, "")
		cleaned = strings.TrimSpace(cleaned)
		cleaned = strings.Replace(cleaned, ",", ".", 1)
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
		Status:    "ok",
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

	if balance.Status == "maintenance" {
		log.Printf("[AEZA] 🔧 API in maintenance — skipping threshold check")
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

// ══════════════════════════════════════════════════════════════
// Aeza Server Info — pulls server status, cost, expiry, specs
// via GET https://my.aeza.net/api/services/{id}
// ══════════════════════════════════════════════════════════════

// AezaServerInfo holds data about a specific Aeza VPS.
type AezaServerInfo struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Status      string  `json:"status"` // "active", "stopped", etc.
	IP          string  `json:"ip"`
	CostEUR     float64 `json:"cost_eur"`
	ExpiresAt   string  `json:"expires_at"` // ISO date
	CPU         int     `json:"cpu"`
	RAMMB       int     `json:"ram_mb"`
	DiskGB      int     `json:"disk_gb"`
	BandwidthGB int     `json:"bandwidth_gb"` // monthly traffic limit from Aeza plan
	DiskType    string  `json:"disk_type"`    // "SSD", "NVMe", etc.
	OS          string  `json:"os"`
	Location    string  `json:"location"`
	UpdatedAt   string  `json:"updated_at"`
	APIStatus   string  `json:"api_status"` // "ok", "error", "maintenance"
}

// GetAezaServerInfo fetches info about a specific server from the Aeza API.
func GetAezaServerInfo() (*AezaServerInfo, error) {
	apiKey := os.Getenv("AEZA_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("AEZA_API_KEY not configured")
	}

	serverID := os.Getenv("AEZA_SERVER_ID")
	if serverID == "" {
		serverID = "1767112" // default XPLR VPN server
	}

	apiURL := fmt.Sprintf("https://my.aeza.net/api/services/%s", serverID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-API-Key", strings.TrimSpace(apiKey))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aeza server info request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("[AEZA-SERVER] API response (status %d): %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode >= 500 {
			return &AezaServerInfo{
				ID:        mustAtoi(serverID),
				APIStatus: "maintenance",
				UpdatedAt: time.Now().Format(time.RFC3339),
			}, nil
		}
		return nil, fmt.Errorf("aeza server info API returned %d: %s", resp.StatusCode, string(body))
	}

	// Parse Aeza response — structure: { "data": { ... } }
	var raw struct {
		Data struct {
			ID           int            `json:"id"`
			Name         string         `json:"name"`
			Status       string         `json:"status"`
			IP           string         `json:"ip"`
			Cost         any            `json:"cost"` // may be string or float
			EndDate      string         `json:"endDate"`
			Products     map[string]any `json:"products"`     // dynamic map to catch ALL field names
			TrafficLimit any            `json:"trafficLimit"` // some plans put limit at top level
			Traffic      any            `json:"traffic"`      // alternative top-level field
			DiskType     string         `json:"diskType"`
			OsName       string         `json:"osName"`
			LocationKey  string         `json:"locationKey"`
			LocationName string         `json:"locationName"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse aeza server response: %w (body: %s)", err, string(body))
	}

	d := raw.Data

	// Nil-safety: if products map is nil, initialize empty map
	if d.Products == nil {
		d.Products = make(map[string]any)
		log.Printf("[AEZA-SERVER] ⚠️ Products field is nil/missing in API response!")
	}

	// Log ALL products fields for diagnostics (critical for debugging SWEs-2 plan changes)
	log.Printf("[AEZA-SERVER] Raw products map (%d keys): %+v", len(d.Products), d.Products)
	if d.TrafficLimit != nil {
		log.Printf("[AEZA-SERVER] Top-level trafficLimit: %v (type: %T)", d.TrafficLimit, d.TrafficLimit)
	}
	if d.Traffic != nil {
		log.Printf("[AEZA-SERVER] Top-level traffic: %v (type: %T)", d.Traffic, d.Traffic)
	}

	// Parse bandwidth from products map — try ALL known field names
	bwGB := 0
	bandwidthKeys := []string{"bandwidth", "traffic", "trafficLimit", "traffic_limit", "bw", "monthlyTraffic", "monthly_traffic", "limit", "Bandwidth", "Traffic", "TrafficLimit"}
	for _, key := range bandwidthKeys {
		if val, ok := d.Products[key]; ok && val != nil {
			parsed := int(parseAnyFloat(val))
			if parsed > 0 {
				bwGB = parsed
				log.Printf("[AEZA-SERVER] ✅ Found bandwidth in products.%s = %d GB", key, bwGB)
				break
			}
		}
	}
	// Also try case-insensitive match on all products keys
	if bwGB == 0 {
		for key, val := range d.Products {
			keyLower := strings.ToLower(key)
			if strings.Contains(keyLower, "traffic") || strings.Contains(keyLower, "bandwidth") || strings.Contains(keyLower, "bw") {
				parsed := int(parseAnyFloat(val))
				if parsed > 0 {
					bwGB = parsed
					log.Printf("[AEZA-SERVER] ✅ Found bandwidth via fuzzy match: products.%s = %d GB", key, bwGB)
					break
				}
			}
		}
	}
	// Try top-level trafficLimit / traffic fields
	if bwGB == 0 && d.TrafficLimit != nil {
		bwGB = int(parseAnyFloat(d.TrafficLimit))
		if bwGB > 0 {
			log.Printf("[AEZA-SERVER] ✅ Found bandwidth in top-level trafficLimit = %d GB", bwGB)
		}
	}
	if bwGB == 0 && d.Traffic != nil {
		bwGB = int(parseAnyFloat(d.Traffic))
		if bwGB > 0 {
			log.Printf("[AEZA-SERVER] ✅ Found bandwidth in top-level traffic = %d GB", bwGB)
		}
	}

	// Normalize units: if bwGB > 1000, it's likely in MB; if > 1000000, likely bytes
	if bwGB > 1000000 {
		bwGB = bwGB / (1024 * 1024 * 1024)
		log.Printf("[AEZA-SERVER] Converted bandwidth from bytes to GB: %d GB", bwGB)
	} else if bwGB > 1000 {
		bwGB = bwGB / 1024
		log.Printf("[AEZA-SERVER] Converted bandwidth from MB to GB: %d GB", bwGB)
	}

	// Last resort: scan ALL products values — pick the largest numeric value >= 30 (likely bandwidth)
	if bwGB == 0 {
		var largestVal int
		var largestKey string
		for key, val := range d.Products {
			parsed := int(parseAnyFloat(val))
			if parsed >= 30 && parsed > largestVal {
				largestVal = parsed
				largestKey = key
			}
		}
		if largestVal > 0 {
			bwGB = largestVal
			// Also normalize the guessed value
			if bwGB > 1000000 {
				bwGB = bwGB / (1024 * 1024 * 1024)
			} else if bwGB > 1000 {
				bwGB = bwGB / 1024
			}
			log.Printf("[AEZA-SERVER] ⚠️ Guessed bandwidth from largest products value: products.%s = %d GB", largestKey, bwGB)
		}
	}

	if bwGB == 0 {
		log.Printf("[AEZA-SERVER] ❌ CRITICAL: Could not find bandwidth in any field! Products keys: %v", func() []string {
			keys := make([]string, 0, len(d.Products))
			for k := range d.Products {
				keys = append(keys, fmt.Sprintf("%s=%v", k, d.Products[k]))
			}
			return keys
		}())
	}

	// Parse CPU/RAM/Disk from products map
	cpuVal := int(parseAnyFloat(d.Products["cpu"]))
	ramVal := int(parseAnyFloat(d.Products["ram"]))
	diskVal := int(parseAnyFloat(d.Products["disk"]))

	info := &AezaServerInfo{
		ID:          d.ID,
		Name:        d.Name,
		Status:      d.Status,
		IP:          d.IP,
		CostEUR:     parseAnyFloat(d.Cost),
		ExpiresAt:   d.EndDate,
		CPU:         cpuVal,
		RAMMB:       ramVal,
		DiskGB:      diskVal,
		BandwidthGB: bwGB,
		DiskType:    d.DiskType,
		OS:          d.OsName,
		Location:    d.LocationName,
		UpdatedAt:   time.Now().Format(time.RFC3339),
		APIStatus:   "ok",
	}

	// Fallback defaults if API returned zero values
	if info.CostEUR == 0 {
		info.CostEUR = 4.94
	}
	if info.CPU == 0 {
		info.CPU = 1
	}
	if info.RAMMB == 0 {
		info.RAMMB = 2048
	}
	if info.DiskGB == 0 {
		info.DiskGB = 30
	}
	if info.DiskType == "" {
		info.DiskType = "SSD"
	}
	if info.ExpiresAt == "" {
		info.ExpiresAt = "2026-05-18"
	}
	if info.Status == "" {
		info.Status = "active"
	}

	log.Printf("[AEZA-SERVER] ✅ Server %d: status=%s, cost=€%.2f, expires=%s, %dCPU/%dMB/%dGB %s, bandwidth=%dGB",
		info.ID, info.Status, info.CostEUR, info.ExpiresAt, info.CPU, info.RAMMB, info.DiskGB, info.DiskType, info.BandwidthGB)
	return info, nil
}

// parseAnyFloat converts interface{} (string or number) to float64.
func parseAnyFloat(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case string:
		cleaned := balanceCleanRe.ReplaceAllString(val, "")
		cleaned = strings.Replace(cleaned, ",", ".", 1)
		f, _ := strconv.ParseFloat(strings.TrimSpace(cleaned), 64)
		return f
	case json.Number:
		f, _ := val.Float64()
		return f
	}
	return 0
}

func mustAtoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
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
