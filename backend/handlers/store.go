package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/djalben/xplr-core/backend/middleware"
	"github.com/djalben/xplr-core/backend/providers"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
	"github.com/shopspring/decimal"
)

// ══════════════════════════════════════════════════════════════
// Store types
// ══════════════════════════════════════════════════════════════

type StoreCategory struct {
	ID          int    `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	ImageURL    string `json:"image_url"`
	SortOrder   int    `json:"sort_order"`
}

type StoreProduct struct {
	ID            int             `json:"id"`
	CategoryID    int             `json:"category_id"`
	CategorySlug  string          `json:"category_slug"`
	Provider      string          `json:"provider"`
	ExternalID    string          `json:"external_id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Country       string          `json:"country"`
	CountryCode   string          `json:"country_code"`
	PriceUSD      decimal.Decimal `json:"price_usd"`
	CostPrice     decimal.Decimal `json:"cost_price"`
	MarkupPercent decimal.Decimal `json:"markup_percent"`
	OldPrice      decimal.Decimal `json:"old_price"`
	DataGB        string          `json:"data_gb"`
	ValidityDays  int             `json:"validity_days"`
	ImageURL      string          `json:"image_url"`
	ProductType   string          `json:"product_type"`
	InStock       bool            `json:"in_stock"`
	Meta          json.RawMessage `json:"meta"`
	SortOrder     int             `json:"sort_order"`
}

// calculatePrice computes the retail price from cost and markup,
// then rounds UP to the nearest .90 (e.g. $12.15 → $12.90).
func calculatePrice(cost, markupPct decimal.Decimal) decimal.Decimal {
	if cost.IsZero() {
		return decimal.Zero
	}
	// retail = cost * (1 + markup/100)
	multiplier := decimal.NewFromInt(1).Add(markupPct.Div(decimal.NewFromInt(100)))
	raw := cost.Mul(multiplier)

	// Round up to nearest .90
	floor := raw.Floor()   // integer part
	frac := raw.Sub(floor) // fractional part
	threshold := decimal.NewFromFloat(0.90)
	if frac.LessThanOrEqual(threshold) {
		return floor.Add(threshold) // e.g. 12.15 → 12.90
	}
	// frac > 0.90 → next integer + .90
	return floor.Add(decimal.NewFromInt(1)).Add(threshold) // e.g. 12.95 → 13.90
}

// applyMarkup recalculates PriceUSD and OldPrice from CostPrice + MarkupPercent.
func applyMarkup(p *StoreProduct) {
	if p.CostPrice.IsPositive() && p.MarkupPercent.IsPositive() {
		p.PriceUSD = calculatePrice(p.CostPrice, p.MarkupPercent)
	}
	// Fake "old price" = current price * 1.20 (rounded to .90)
	if p.PriceUSD.IsPositive() {
		p.OldPrice = calculatePrice(p.PriceUSD, decimal.NewFromInt(20))
	}
}

type StoreOrder struct {
	ID            int             `json:"id"`
	UserID        int             `json:"user_id"`
	ProductID     int             `json:"product_id"`
	ProductName   string          `json:"product_name"`
	PriceUSD      decimal.Decimal `json:"price_usd"`
	Status        string          `json:"status"`
	ActivationKey string          `json:"activation_key"`
	QRData        string          `json:"qr_data"`
	ProviderRef   string          `json:"provider_ref"`
	CreatedAt     time.Time       `json:"created_at"`
}

// ══════════════════════════════════════════════════════════════
// GET /api/v1/store/catalog — returns categories + products
// ══════════════════════════════════════════════════════════════

func StoreCatalogHandler(w http.ResponseWriter, r *http.Request) {
	if GlobalDB == nil {
		http.Error(w, "DB not ready", http.StatusInternalServerError)
		return
	}

	// Fetch categories
	catRows, err := GlobalDB.Query(`SELECT id, slug, name, description, icon, COALESCE(image_url, ''), sort_order FROM store_categories ORDER BY sort_order, id`)
	if err != nil {
		log.Printf("[STORE] ❌ Failed to fetch categories: %v", err)
		http.Error(w, "Failed to fetch catalog", http.StatusInternalServerError)
		return
	}
	defer catRows.Close()

	var categories []StoreCategory
	for catRows.Next() {
		var c StoreCategory
		if err := catRows.Scan(&c.ID, &c.Slug, &c.Name, &c.Description, &c.Icon, &c.ImageURL, &c.SortOrder); err != nil {
			continue
		}
		categories = append(categories, c)
	}
	if categories == nil {
		categories = []StoreCategory{}
	}

	// Optional filters
	categorySlug := r.URL.Query().Get("category")
	country := r.URL.Query().Get("country")
	search := r.URL.Query().Get("search")

	// Build product query with optional filters
	query := `SELECT p.id, p.category_id, c.slug, p.provider, p.external_id, p.name, p.description,
		COALESCE(p.country, ''), COALESCE(p.country_code, ''), p.price_usd,
		COALESCE(p.cost_price, 0), COALESCE(p.markup_percent, 20),
		COALESCE(p.data_gb, ''),
		COALESCE(p.validity_days, 0), COALESCE(p.image_url, ''), p.product_type, p.in_stock,
		COALESCE(p.meta, '{}'), p.sort_order
		FROM store_products p
		JOIN store_categories c ON c.id = p.category_id
		WHERE p.in_stock = TRUE`
	args := []interface{}{}
	argIdx := 1

	if categorySlug != "" {
		query += fmt.Sprintf(" AND c.slug = $%d", argIdx)
		args = append(args, categorySlug)
		argIdx++
	}
	if country != "" {
		query += fmt.Sprintf(" AND (LOWER(p.country) LIKE LOWER($%d) OR LOWER(p.country_code) = LOWER($%d))", argIdx, argIdx+1)
		args = append(args, "%"+country+"%", country)
		argIdx += 2
	}
	if search != "" {
		query += fmt.Sprintf(" AND (LOWER(p.name) LIKE LOWER($%d) OR LOWER(p.country) LIKE LOWER($%d))", argIdx, argIdx+1)
		args = append(args, "%"+search+"%", "%"+search+"%")
		argIdx += 2
	}
	query += " ORDER BY p.sort_order, p.id"

	prodRows, err := GlobalDB.Query(query, args...)
	if err != nil {
		log.Printf("[STORE] ❌ Failed to fetch products: %v", err)
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}
	defer prodRows.Close()

	var products []StoreProduct
	for prodRows.Next() {
		var p StoreProduct
		if err := prodRows.Scan(&p.ID, &p.CategoryID, &p.CategorySlug, &p.Provider, &p.ExternalID,
			&p.Name, &p.Description, &p.Country, &p.CountryCode, &p.PriceUSD,
			&p.CostPrice, &p.MarkupPercent,
			&p.DataGB,
			&p.ValidityDays, &p.ImageURL, &p.ProductType, &p.InStock, &p.Meta, &p.SortOrder); err != nil {
			log.Printf("[STORE] product scan error: %v", err)
			continue
		}
		applyMarkup(&p)
		products = append(products, p)
	}
	if products == nil {
		products = []StoreProduct{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories": categories,
		"products":   products,
	})
}

// ══════════════════════════════════════════════════════════════
// POST /api/v1/store/purchase — buy a digital product
// ══════════════════════════════════════════════════════════════

func StorePurchaseHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ProductID int `json:"product_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProductID <= 0 {
		http.Error(w, "Invalid product_id", http.StatusBadRequest)
		return
	}

	log.Printf("[STORE-PURCHASE] User %d → product %d", userID, req.ProductID)

	// 1. Fetch product
	var product StoreProduct
	var metaBytes []byte
	err := GlobalDB.QueryRow(`
		SELECT p.id, p.category_id, c.slug, p.provider, p.external_id, p.name, p.description,
			COALESCE(p.country, ''), COALESCE(p.country_code, ''), p.price_usd,
			COALESCE(p.cost_price, 0), COALESCE(p.markup_percent, 20),
			COALESCE(p.data_gb, ''),
			COALESCE(p.validity_days, 0), COALESCE(p.image_url, ''), p.product_type, p.in_stock,
			COALESCE(p.meta, '{}'), p.sort_order
		FROM store_products p
		JOIN store_categories c ON c.id = p.category_id
		WHERE p.id = $1
	`, req.ProductID).Scan(&product.ID, &product.CategoryID, &product.CategorySlug, &product.Provider,
		&product.ExternalID, &product.Name, &product.Description, &product.Country, &product.CountryCode,
		&product.PriceUSD, &product.CostPrice, &product.MarkupPercent,
		&product.DataGB, &product.ValidityDays, &product.ImageURL, &product.ProductType,
		&product.InStock, &metaBytes, &product.SortOrder)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			log.Printf("[STORE-PURCHASE] ❌ DB error fetching product %d: %v", req.ProductID, err)
			http.Error(w, "Failed to fetch product", http.StatusInternalServerError)
		}
		return
	}
	product.Meta = metaBytes
	applyMarkup(&product)

	// 2. Check stock
	if !product.InStock {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Товар временно недоступен у поставщика",
			"code":  "OUT_OF_STOCK",
		})
		return
	}

	// 3. Call provider API (stub — returns simulated activation data)
	activationKey, qrData, providerRef, providerErr := callProvider(product)
	if providerErr != nil {
		log.Printf("[STORE-PURCHASE] ❌ Provider error for product %d: %v", product.ID, providerErr)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Ошибка поставщика: " + providerErr.Error(),
			"code":  "PROVIDER_ERROR",
		})
		return
	}

	// 4. Payment via Card (direct wallet deduction FORBIDDEN)
	details := fmt.Sprintf("Покупка товара ID_%d (%s) — $%s", product.ID, product.Name, product.PriceUSD.StringFixed(2))
	cardID, cardLast4, payErr := repository.PurchaseViaCard(userID, product.PriceUSD, details)
	if payErr != nil {
		errMsg := payErr.Error()
		if errMsg == "NO_ACTIVE_CARD" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Для покупки товаров необходимо иметь активную карту. Пожалуйста, приобретите карту в разделе «Карты» и пополните её с кошелька XPLR.",
				"code":  "NO_ACTIVE_CARD",
			})
			return
		}
		if len(errMsg) > 18 && errMsg[:18] == "INSUFFICIENT_FUNDS" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Недостаточно средств в системе XPLR для проведения операции",
				"code":  "INSUFFICIENT_FUNDS",
			})
			return
		}
		log.Printf("[STORE-PURCHASE] ❌ Payment failed for user %d: %v", userID, payErr)
		http.Error(w, "Payment failed: "+errMsg, http.StatusInternalServerError)
		return
	}

	log.Printf("[STORE-PURCHASE] Покупка товара ID_%d через Карту ID_%d (*%s)", product.ID, cardID, cardLast4)

	// 5. Record order
	var orderID int
	err = GlobalDB.QueryRow(`
		INSERT INTO store_orders (user_id, product_id, product_name, price_usd, status, activation_key, qr_data, provider_ref)
		VALUES ($1, $2, $3, $4, 'completed', $5, $6, $7) RETURNING id`,
		userID, product.ID, product.Name, product.PriceUSD, activationKey, qrData, providerRef,
	).Scan(&orderID)
	if err != nil {
		log.Printf("[STORE-PURCHASE] ❌ Failed to record order: %v", err)
	}

	log.Printf("[STORE-PURCHASE] ✅ User %d purchased '%s' for $%s via Card %d (order #%d)",
		userID, product.Name, product.PriceUSD.StringFixed(2), cardID, orderID)

	// 6. Notify user
	go notifyStorePurchase(userID, product, activationKey, qrData)

	// 7. Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order_id":       orderID,
		"product_name":   product.Name,
		"price_usd":      product.PriceUSD.StringFixed(2),
		"activation_key": activationKey,
		"qr_data":        qrData,
		"status":         "completed",
	})
}

// ══════════════════════════════════════════════════════════════
// GET /api/v1/store/orders — user's purchase history
// ══════════════════════════════════════════════════════════════

func StoreOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	rows, err := GlobalDB.Query(`
		SELECT id, user_id, product_id, product_name, price_usd, status, COALESCE(activation_key, ''), COALESCE(qr_data, ''), COALESCE(provider_ref, ''), created_at
		FROM store_orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []StoreOrder
	for rows.Next() {
		var o StoreOrder
		if err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.ProductName, &o.PriceUSD,
			&o.Status, &o.ActivationKey, &o.QRData, &o.ProviderRef, &o.CreatedAt); err != nil {
			continue
		}
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []StoreOrder{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"orders": orders})
}

// ══════════════════════════════════════════════════════════════
// Provider abstraction — stub for MobiMatter / Razer Gold
// ══════════════════════════════════════════════════════════════

func callProvider(product StoreProduct) (activationKey string, qrData string, providerRef string, err error) {
	switch product.Provider {
	case "mobimatter":
		return callMobiMatter(product)
	case "razer":
		return callRazerGold(product)
	case "demo":
		return callDemoProvider(product)
	default:
		return callDemoProvider(product)
	}
}

// Demo provider — returns test activation data
func callDemoProvider(product StoreProduct) (string, string, string, error) {
	ref := fmt.Sprintf("DEMO-%d-%d", product.ID, time.Now().UnixMilli())

	if product.ProductType == "esim" {
		// Return simulated eSIM QR data (LPA URI format)
		qrData := fmt.Sprintf("LPA:1$smdp.example.com$%s", ref)
		return "", qrData, ref, nil
	}

	// Digital product — return activation key
	key := fmt.Sprintf("XPLR-%s-%d", product.ExternalID, time.Now().UnixMilli())
	return key, "", ref, nil
}

// MobiMatter — uses the real eSIM provider wrapper
func callMobiMatter(product StoreProduct) (string, string, string, error) {
	p := providers.GetESIMProvider()

	// Check availability before ordering
	available, err := p.CheckAvailability(product.ExternalID)
	if err != nil {
		log.Printf("[MOBIMATTER] ❌ Availability check error for %s: %v", product.ExternalID, err)
	}
	if !available {
		return "", "", "", fmt.Errorf("план %s временно недоступен у поставщика", product.Name)
	}

	result, err := p.OrderESIM(product.ExternalID)
	if err != nil {
		return "", "", "", fmt.Errorf("eSIM order failed: %w", err)
	}

	log.Printf("[MOBIMATTER] ✅ Order placed: ref=%s, ICCID=%s", result.ProviderRef, result.ICCID)
	return "", result.QRData, result.ProviderRef, nil
}

// Razer Gold stub — will be replaced with real API integration
func callRazerGold(product StoreProduct) (string, string, string, error) {
	// TODO: Implement real Razer Gold API call
	log.Printf("[RAZER] 🔧 Stub call for product %s (external_id=%s)", product.Name, product.ExternalID)
	return callDemoProvider(product)
}

// ══════════════════════════════════════════════════════════════
// Notification after purchase
// ══════════════════════════════════════════════════════════════

func notifyStorePurchase(userID int, product StoreProduct, activationKey, qrData string) {
	var resultInfo string
	if qrData != "" {
		resultInfo = "QR-код для активации eSIM отправлен ниже."
	} else if activationKey != "" {
		resultInfo = fmt.Sprintf("Ваш ключ активации: <code>%s</code>", activationKey)
	}

	// Build flag emoji for eSIM products
	flagEmoji := ""
	if product.ProductType == "esim" && len(product.CountryCode) == 2 {
		cc := []rune(product.CountryCode)
		flagEmoji = string(rune(0x1F1E6+int(cc[0])-'A')) + string(rune(0x1F1E6+int(cc[1])-'A')) + " "
	}

	tgMsg := fmt.Sprintf("🛒 <b>Покупка успешна!</b>\n\n"+
		"%sТовар: <b>%s</b>\n"+
		"Цена: <b>$%s</b>\n\n"+
		"%s\n\n"+
		"<a href=\"https://xplr.pro/store\">Открыть магазин</a>",
		flagEmoji, product.Name, product.PriceUSD.StringFixed(2), resultInfo)

	emailBody := fmt.Sprintf(`
		<p style="color:#cbd5e1;font-size:16px;line-height:1.5;margin:0 0 16px;font-weight:700;">🛒 Покупка успешна!</p>
		<table style="width:100%%;border-collapse:collapse;margin:0 0 24px;">
			<tr><td style="color:#94a3b8;padding:8px 0;border-bottom:1px solid rgba(255,255,255,0.06);">Товар</td><td style="color:#fff;padding:8px 0;border-bottom:1px solid rgba(255,255,255,0.06);text-align:right;font-weight:600;">%s</td></tr>
			<tr><td style="color:#94a3b8;padding:8px 0;border-bottom:1px solid rgba(255,255,255,0.06);">Цена</td><td style="color:#fff;padding:8px 0;border-bottom:1px solid rgba(255,255,255,0.06);text-align:right;font-weight:600;">$%s</td></tr>
		</table>`,
		product.Name, product.PriceUSD.StringFixed(2))

	if activationKey != "" {
		emailBody += fmt.Sprintf(`
		<div style="background:rgba(59,130,246,0.1);border:1px solid rgba(59,130,246,0.3);border-radius:12px;padding:16px;margin:0 0 24px;text-align:center;">
			<p style="color:#94a3b8;font-size:12px;margin:0 0 8px;">Ключ активации</p>
			<p style="color:#fff;font-size:18px;font-weight:700;font-family:monospace;letter-spacing:2px;margin:0;">%s</p>
		</div>`, activationKey)
	}
	if qrData != "" {
		emailBody += `
		<div style="background:rgba(59,130,246,0.1);border:1px solid rgba(59,130,246,0.3);border-radius:12px;padding:16px;margin:0 0 24px;text-align:center;">
			<p style="color:#94a3b8;font-size:12px;margin:0 0 8px;">QR-код для активации eSIM</p>
			<p style="color:#fff;font-size:14px;margin:0;">QR-код доступен в приложении XPLR</p>
		</div>`
	}

	emailBody += `
		<div style="text-align:center;">
			<a href="https://xplr.pro/store" style="display:inline-block;padding:14px 40px;background:linear-gradient(135deg,#3b82f6,#8b5cf6);color:#fff;text-decoration:none;border-radius:12px;font-size:14px;font-weight:600;">Открыть магазин</a>
		</div>`

	// If product has an image, use NotifyUserNews (sends photo in TG + image in email)
	if product.ImageURL != "" {
		service.NotifyUserNews(userID, "Покупка в XPLR Store", tgMsg, emailBody, product.ImageURL)
	} else {
		service.NotifyUser(userID, "Покупка в XPLR Store", tgMsg)
		go func() {
			user, err := repository.GetUserByID(userID)
			if err != nil || user.Email == "" {
				return
			}
			if err := service.SendGenericEmail(user.Email, "Чек покупки — XPLR Store", emailBody); err != nil {
				log.Printf("[STORE-NOTIFY] ❌ Email to user %d failed: %v", userID, err)
			}
		}()
	}
}

// ══════════════════════════════════════════════════════════════
// eSIM API endpoints (provider wrapper)
// ══════════════════════════════════════════════════════════════

// GET /api/v1/store/esim/destinations
func ESIMDestinationsHandler(w http.ResponseWriter, r *http.Request) {
	p := providers.GetESIMProvider()
	dests, err := p.GetDestinations()
	if err != nil {
		log.Printf("[ESIM] ❌ GetDestinations error: %v", err)
		http.Error(w, "Failed to fetch destinations", http.StatusInternalServerError)
		return
	}

	// Optional search filter
	search := r.URL.Query().Get("search")
	if search != "" {
		searchLower := toLower(search)
		var filtered []providers.ESIMDestination
		for _, d := range dests {
			if containsLower(d.CountryName, searchLower) || containsLower(d.CountryCode, searchLower) {
				filtered = append(filtered, d)
			}
		}
		dests = filtered
	}
	if dests == nil {
		dests = []providers.ESIMDestination{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"destinations": dests})
}

// GET /api/v1/store/esim/plans?country=XX
func ESIMPlansHandler(w http.ResponseWriter, r *http.Request) {
	cc := r.URL.Query().Get("country")
	if cc == "" {
		http.Error(w, "country parameter required", http.StatusBadRequest)
		return
	}

	prov := providers.GetESIMProvider()
	plans, err := prov.GetPlans(cc)
	if err != nil {
		log.Printf("[ESIM] ❌ GetPlans error for %s: %v", cc, err)
		http.Error(w, "Failed to fetch plans", http.StatusInternalServerError)
		return
	}
	if plans == nil {
		plans = []providers.ESIMPlan{}
	}

	// Read eSIM default markup from DB (or use 150%)
	var esimMarkup float64 = 150
	if GlobalDB != nil {
		row := GlobalDB.QueryRow(`SELECT COALESCE(AVG(markup_percent), 150) FROM store_products WHERE product_type = 'esim' AND markup_percent > 0`)
		row.Scan(&esimMarkup)
	}
	markupDec := decimal.NewFromFloat(esimMarkup)

	// Apply markup to each plan
	for i := range plans {
		costDec := decimal.NewFromFloat(plans[i].PriceUSD)
		plans[i].CostPrice = plans[i].PriceUSD
		retailDec := calculatePrice(costDec, markupDec)
		plans[i].PriceUSD, _ = retailDec.Float64()
		oldDec := calculatePrice(retailDec, decimal.NewFromInt(20))
		plans[i].OldPrice, _ = oldDec.Float64()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"plans": plans})
}

// POST /api/v1/store/esim/order — full eSIM purchase flow
func ESIMOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		PlanID      string  `json:"plan_id"`
		PlanName    string  `json:"plan_name"`
		Country     string  `json:"country"`
		CountryCode string  `json:"country_code"`
		DataGB      string  `json:"data_gb"`
		Days        int     `json:"validity_days"`
		PriceUSD    float64 `json:"price_usd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlanID == "" || req.PriceUSD <= 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	price := decimal.NewFromFloat(req.PriceUSD)
	log.Printf("[ESIM-ORDER] User %d → plan %s (%s) $%s", userID, req.PlanID, req.PlanName, price.StringFixed(2))

	// 1. Check availability at provider
	p := providers.GetESIMProvider()
	available, err := p.CheckAvailability(req.PlanID)
	if err != nil {
		log.Printf("[ESIM-ORDER] ⚠️ Availability check error: %v (proceeding anyway)", err)
	}
	if !available {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Этот план временно недоступен у поставщика. Попробуйте другой.",
			"code":  "OUT_OF_STOCK",
		})
		return
	}

	// 2. Order from provider
	result, orderErr := p.OrderESIM(req.PlanID)
	if orderErr != nil {
		log.Printf("[ESIM-ORDER] ❌ Provider order error: %v", orderErr)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Ошибка поставщика: " + orderErr.Error(),
			"code":  "PROVIDER_ERROR",
		})
		return
	}

	// 3. Payment via Card (direct wallet deduction FORBIDDEN)
	productName := req.PlanName
	if productName == "" {
		productName = "eSIM " + req.CountryCode
	}
	details := fmt.Sprintf("Покупка eSIM ID_%s (%s) — $%s", req.PlanID, productName, price.StringFixed(2))
	cardID, cardLast4, payErr := repository.PurchaseViaCard(userID, price, details)
	if payErr != nil {
		errMsg := payErr.Error()
		if errMsg == "NO_ACTIVE_CARD" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Для покупки товаров необходимо иметь активную карту. Пожалуйста, приобретите карту в разделе «Карты» и пополните её с кошелька XPLR.",
				"code":  "NO_ACTIVE_CARD",
			})
			return
		}
		if len(errMsg) > 18 && errMsg[:18] == "INSUFFICIENT_FUNDS" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Недостаточно средств в системе XPLR для проведения операции",
				"code":  "INSUFFICIENT_FUNDS",
			})
			return
		}
		log.Printf("[ESIM-ORDER] ❌ Payment failed for user %d: %v", userID, payErr)
		http.Error(w, "Payment failed: "+errMsg, http.StatusInternalServerError)
		return
	}

	log.Printf("[ESIM-ORDER] Покупка eSIM ID_%s через Карту ID_%d (*%s)", req.PlanID, cardID, cardLast4)

	// 4. Record order in store_orders
	var orderID int
	err = GlobalDB.QueryRow(`
		INSERT INTO store_orders (user_id, product_id, product_name, price_usd, status, activation_key, qr_data, provider_ref)
		VALUES ($1, 0, $2, $3, 'completed', $4, $5, $6) RETURNING id`,
		userID, productName, price, result.ICCID, result.QRData, result.ProviderRef,
	).Scan(&orderID)
	if err != nil {
		log.Printf("[ESIM-ORDER] ❌ Failed to record order: %v", err)
	}

	log.Printf("[ESIM-ORDER] ✅ User %d ordered '%s' for $%s via Card %d (order #%d, ref=%s)",
		userID, productName, price.StringFixed(2), cardID, orderID, result.ProviderRef)

	// 5. Notify
	go func() {
		product := StoreProduct{
			Name:     productName,
			PriceUSD: price,
		}
		notifyStorePurchase(userID, product, "", result.QRData)
	}()

	// 6. Return full result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order_id":     orderID,
		"product_name": productName,
		"price_usd":    price.StringFixed(2),
		"qr_data":      result.QRData,
		"lpa":          result.LPA,
		"smdp":         result.SMDP,
		"matching_id":  result.MatchingID,
		"iccid":        result.ICCID,
		"provider_ref": result.ProviderRef,
		"status":       "completed",
	})
}

// ══════════════════════════════════════════════════════════════
// Utility helpers
// ══════════════════════════════════════════════════════════════

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func containsLower(s, sub string) bool {
	s = toLower(s)
	return len(s) >= len(sub) && findSubstring(s, sub)
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ══════════════════════════════════════════════════════════════
// Admin: Store Price Management
// ══════════════════════════════════════════════════════════════

// AdminStoreProduct — admin view includes cost_price and markup
type AdminStoreProduct struct {
	ID            int             `json:"id"`
	Name          string          `json:"name"`
	ProductType   string          `json:"product_type"`
	Provider      string          `json:"provider"`
	CostPrice     decimal.Decimal `json:"cost_price"`
	MarkupPercent decimal.Decimal `json:"markup_percent"`
	RetailPrice   decimal.Decimal `json:"retail_price"`
	OldPrice      decimal.Decimal `json:"old_price"`
	InStock       bool            `json:"in_stock"`
	ExternalID    string          `json:"external_id"`
	ImageURL      string          `json:"image_url"`
}

// GET /api/v1/admin/store/products — list all products with pricing
func AdminStoreProductsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := GlobalDB.Query(`
		SELECT id, name, product_type, provider,
			COALESCE(cost_price, 0), COALESCE(markup_percent, 20),
			price_usd, in_stock, COALESCE(external_id, ''), COALESCE(image_url, '')
		FROM store_products ORDER BY product_type, sort_order, id`)
	if err != nil {
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []AdminStoreProduct
	for rows.Next() {
		var p AdminStoreProduct
		if err := rows.Scan(&p.ID, &p.Name, &p.ProductType, &p.Provider,
			&p.CostPrice, &p.MarkupPercent, &p.RetailPrice, &p.InStock, &p.ExternalID, &p.ImageURL); err != nil {
			continue
		}
		// Recalculate retail price from cost + markup
		if p.CostPrice.IsPositive() && p.MarkupPercent.IsPositive() {
			p.RetailPrice = calculatePrice(p.CostPrice, p.MarkupPercent)
		}
		if p.RetailPrice.IsPositive() {
			p.OldPrice = calculatePrice(p.RetailPrice, decimal.NewFromInt(20))
		}
		products = append(products, p)
	}
	if products == nil {
		products = []AdminStoreProduct{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// PATCH /api/v1/admin/store/products/{id} — update cost_price and/or markup_percent
func AdminUpdateStoreProductHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/v1/admin/store/products/"):]
	productID, _ := strconv.Atoi(idStr)
	if productID <= 0 {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req struct {
		CostPrice     *float64 `json:"cost_price"`
		MarkupPercent *float64 `json:"markup_percent"`
		ImageURL      *string  `json:"image_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.CostPrice != nil {
		if _, err := GlobalDB.Exec(`UPDATE store_products SET cost_price = $1 WHERE id = $2`, *req.CostPrice, productID); err != nil {
			http.Error(w, "Failed to update cost_price", http.StatusInternalServerError)
			return
		}
	}
	if req.MarkupPercent != nil {
		if _, err := GlobalDB.Exec(`UPDATE store_products SET markup_percent = $1 WHERE id = $2`, *req.MarkupPercent, productID); err != nil {
			http.Error(w, "Failed to update markup_percent", http.StatusInternalServerError)
			return
		}
	}
	if req.ImageURL != nil {
		if _, err := GlobalDB.Exec(`UPDATE store_products SET image_url = $1 WHERE id = $2`, *req.ImageURL, productID); err != nil {
			http.Error(w, "Failed to update image_url", http.StatusInternalServerError)
			return
		}
	}

	// Also update the stored price_usd column for consistency
	GlobalDB.Exec(`
		UPDATE store_products SET price_usd = cost_price * (1 + markup_percent / 100) WHERE id = $1`, productID)

	log.Printf("[ADMIN-STORE] Updated product %d: cost=%v markup=%v", productID, req.CostPrice, req.MarkupPercent)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// POST /api/v1/admin/store/bulk-markup — increase markup_percent by delta for all or filtered products
func AdminBulkMarkupHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Delta       float64 `json:"delta"`
		ProductType string  `json:"product_type"` // optional: "esim", "digital", or "" for all
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Delta == 0 {
		http.Error(w, "Invalid request (delta required)", http.StatusBadRequest)
		return
	}

	query := `UPDATE store_products SET markup_percent = markup_percent + $1`
	args := []interface{}{req.Delta}
	if req.ProductType != "" {
		query += ` WHERE product_type = $2`
		args = append(args, req.ProductType)
	}

	res, err := GlobalDB.Exec(query, args...)
	if err != nil {
		http.Error(w, "Failed to update markups", http.StatusInternalServerError)
		return
	}
	affected, _ := res.RowsAffected()

	// Also update stored price_usd for consistency
	GlobalDB.Exec(`UPDATE store_products SET price_usd = cost_price * (1 + markup_percent / 100) WHERE cost_price > 0`)

	log.Printf("[ADMIN-STORE] Bulk markup +%.1f%% applied to %d products (type=%s)", req.Delta, affected, req.ProductType)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ok",
		"affected": affected,
		"delta":    req.Delta,
	})
}
