package shop

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

// ══════════════════════════════════════════════════════════════
// Fulfillment Engine — auto-delivers products after successful payment.
//
// Flow:
//   1. Client pays → StorePurchaseHandler calls FulfillOrder()
//   2. FulfillOrder creates a "pending" order in DB
//   3. Calls supplier API via ProductProvider.CreateOrder()
//   4. On success: saves activation_key / QR, sets status = "ready"
//   5. Sends premium email + TG notification to user
//   6. On failure: sets status = "failed", notifies admins
// ══════════════════════════════════════════════════════════════

// FulfillmentRequest contains all data needed to fulfill an order.
type FulfillmentRequest struct {
	UserID          int
	UserEmail       string
	ProductID       int             // internal store_products.id (0 for eSIM)
	ProductName     string
	ExternalID      string          // supplier's product ID
	ProviderName    string          // "mobimatter", "razer", "demo"
	PriceUSD        decimal.Decimal // retail price charged to user
	CostPrice       decimal.Decimal // cost from supplier
	ProductType     string          // "esim", "digital"
	CardLast4       string          // card used for payment
}

// FulfillmentResult is returned after the fulfillment attempt.
type FulfillmentResult struct {
	OrderID       int    `json:"order_id"`
	Status        string `json:"status"` // "ready", "failed"
	ActivationKey string `json:"activation_key"`
	QRData        string `json:"qr_data"`
	ProviderRef   string `json:"provider_ref"`
	Error         string `json:"error,omitempty"`
}

// UserNotifier sends a notification (email + TG) to a user about their purchase.
// Injected to avoid circular imports with service package.
type UserNotifier func(userID int, subject string, tgMsg string, emailBody string)

// PremiumEmailSender sends the premium purchase receipt email.
// Signature: (toEmail, orderID, productName, priceUSD, cardLast4, isESIM, activationData)
type PremiumEmailSender func(toEmail string, orderID int, productName string, priceUSD string, cardLast4 string, isESIM bool, activationData map[string]string) error

// FulfillmentEngine orchestrates the auto-delivery pipeline.
type FulfillmentEngine struct {
	db            *sql.DB
	registry      *Registry
	notifyUser    UserNotifier
	notifyAdmins  AdminNotifier
	sendReceipt   PremiumEmailSender
}

// NewFulfillmentEngine creates a new engine.
func NewFulfillmentEngine(
	db *sql.DB,
	registry *Registry,
	notifyUser UserNotifier,
	notifyAdmins AdminNotifier,
	sendReceipt PremiumEmailSender,
) *FulfillmentEngine {
	return &FulfillmentEngine{
		db:           db,
		registry:     registry,
		notifyUser:   notifyUser,
		notifyAdmins: notifyAdmins,
		sendReceipt:  sendReceipt,
	}
}

// FulfillOrder is the main entry point — called immediately after successful payment.
// It calls the supplier, saves the result, and triggers notifications.
func (fe *FulfillmentEngine) FulfillOrder(req FulfillmentRequest) (*FulfillmentResult, error) {
	log.Printf("[FULFILLMENT] ▶ Start: user=%d product=%q provider=%q external=%s",
		req.UserID, req.ProductName, req.ProviderName, req.ExternalID)

	// 1. Insert a "pending" order row
	orderID, err := fe.createPendingOrder(req)
	if err != nil {
		log.Printf("[FULFILLMENT] ❌ Failed to create pending order: %v", err)
		return nil, fmt.Errorf("failed to create order record: %w", err)
	}

	// 2. Resolve provider
	provider, err := fe.registry.MustGet(req.ProviderName)
	if err != nil {
		fe.markOrderFailed(orderID, "provider not found: "+req.ProviderName)
		return &FulfillmentResult{OrderID: orderID, Status: "failed", Error: err.Error()}, err
	}

	// 3. Call supplier API
	result, err := provider.CreateOrder(req.ExternalID)
	if err != nil {
		log.Printf("[FULFILLMENT] ❌ Supplier error for order #%d: %v", orderID, err)
		fe.markOrderFailed(orderID, err.Error())
		fe.notifyAdminOrderFailed(orderID, req, err)
		return &FulfillmentResult{OrderID: orderID, Status: "failed", Error: err.Error()}, err
	}

	// 4. Save activation data → status = "ready"
	fe.markOrderReady(orderID, result)

	log.Printf("[FULFILLMENT] ✅ Order #%d fulfilled: ref=%s key=%s qr=%v",
		orderID, result.ProviderRef, maskKey(result.ActivationKey), result.QRData != "")

	// 5. Send premium email + notifications (async)
	go fe.sendNotifications(req, orderID, result)

	return &FulfillmentResult{
		OrderID:       orderID,
		Status:        "ready",
		ActivationKey: result.ActivationKey,
		QRData:        result.QRData,
		ProviderRef:   result.ProviderRef,
	}, nil
}

// ── DB Operations ──

func (fe *FulfillmentEngine) createPendingOrder(req FulfillmentRequest) (int, error) {
	var orderID int
	err := fe.db.QueryRow(`
		INSERT INTO store_orders (user_id, product_id, product_name, price_usd, status, activation_key, qr_data, provider_ref)
		VALUES ($1, $2, $3, $4, 'pending', '', '', '')
		RETURNING id`,
		req.UserID, req.ProductID, req.ProductName, req.PriceUSD,
	).Scan(&orderID)
	return orderID, err
}

func (fe *FulfillmentEngine) markOrderReady(orderID int, result *OrderResult) {
	_, err := fe.db.Exec(`
		UPDATE store_orders
		SET status = 'ready', activation_key = $1, qr_data = $2, provider_ref = $3
		WHERE id = $4`,
		result.ActivationKey, result.QRData, result.ProviderRef, orderID,
	)
	if err != nil {
		log.Printf("[FULFILLMENT] ❌ Failed to mark order #%d as ready: %v", orderID, err)
	}
}

func (fe *FulfillmentEngine) markOrderFailed(orderID int, reason string) {
	_, err := fe.db.Exec(`
		UPDATE store_orders SET status = 'failed', provider_ref = $1 WHERE id = $2`,
		truncate(reason, 255), orderID,
	)
	if err != nil {
		log.Printf("[FULFILLMENT] ❌ Failed to mark order #%d as failed: %v", orderID, err)
	}
}

// ── Notifications ──

func (fe *FulfillmentEngine) sendNotifications(req FulfillmentRequest, orderID int, result *OrderResult) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[FULFILLMENT] PANIC in notifications for order #%d: %v", orderID, r)
		}
	}()

	isESIM := req.ProductType == "esim"

	// Build flag emoji for eSIM
	flagEmoji := ""
	if isESIM && len(req.ExternalID) >= 2 {
		// CountryCode might be embedded; use product name as fallback
	}

	// Telegram + generic notification
	var resultInfo string
	if result.QRData != "" {
		resultInfo = "QR-код для активации eSIM доступен в вашем кабинете."
	} else if result.ActivationKey != "" {
		resultInfo = fmt.Sprintf("Ваш ключ активации: <code>%s</code>", result.ActivationKey)
	}

	tgMsg := fmt.Sprintf("🛒 <b>Заказ #%d — Готов к выдаче!</b>\n\n"+
		"%sТовар: <b>%s</b>\n"+
		"Цена: <b>$%s</b>\n\n"+
		"%s\n\n"+
		`<a href="https://xplr.pro/purchases">Мои покупки</a>`,
		orderID, flagEmoji, req.ProductName, req.PriceUSD.StringFixed(2), resultInfo)

	if fe.notifyUser != nil {
		fe.notifyUser(req.UserID, "Заказ готов — XPLR Store", tgMsg, "")
	}

	// Premium email with receipt
	if fe.sendReceipt != nil && req.UserEmail != "" {
		activationData := map[string]string{
			"activation_key": result.ActivationKey,
			"qr_data":        result.QRData,
			"provider_ref":   result.ProviderRef,
		}
		if err := fe.sendReceipt(
			req.UserEmail, orderID, req.ProductName,
			req.PriceUSD.StringFixed(2), req.CardLast4,
			isESIM, activationData,
		); err != nil {
			log.Printf("[FULFILLMENT] ⚠️ Premium email failed for order #%d: %v", orderID, err)
		} else {
			log.Printf("[FULFILLMENT] 📧 Premium email sent for order #%d to %s", orderID, req.UserEmail)
		}
	}
}

func (fe *FulfillmentEngine) notifyAdminOrderFailed(orderID int, req FulfillmentRequest, orderErr error) {
	if fe.notifyAdmins == nil {
		return
	}

	subject := fmt.Sprintf("❌ Ошибка выполнения заказа #%d", orderID)
	htmlMsg := fmt.Sprintf(
		"<b>❌ Ошибка автовыдачи</b>\n\n"+
			"Заказ: <b>#%d</b>\n"+
			"Юзер: <b>#%d</b> (%s)\n"+
			"Товар: <b>%s</b>\n"+
			"Поставщик: <b>%s</b>\n"+
			"Цена: <b>$%s</b>\n"+
			"Ошибка: <code>%s</code>\n\n"+
			"Необходимо обработать вручную или вернуть средства.",
		orderID, req.UserID, req.UserEmail, req.ProductName,
		req.ProviderName, req.PriceUSD.StringFixed(2),
		truncate(orderErr.Error(), 200))

	fe.notifyAdmins(subject, htmlMsg)
}

// ── Helpers ──

func maskKey(key string) string {
	if len(key) <= 6 {
		return key
	}
	return key[:3] + "***" + key[len(key)-3:]
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// CheckOrderStatus checks a specific order status via the supplier API.
func (fe *FulfillmentEngine) CheckOrderStatus(providerName, providerRef string) (*OrderStatus, error) {
	provider, err := fe.registry.MustGet(providerName)
	if err != nil {
		return nil, err
	}
	return provider.CheckStatus(providerRef)
}

// RetryPendingOrders scans for stuck "pending" orders older than 5 minutes
// and attempts to re-check their status at the supplier.
func (fe *FulfillmentEngine) RetryPendingOrders() {
	rows, err := fe.db.Query(`
		SELECT id, provider_ref, product_name
		FROM store_orders
		WHERE status = 'pending' AND created_at < NOW() - INTERVAL '5 minutes'
		ORDER BY created_at ASC LIMIT 20`)
	if err != nil {
		log.Printf("[FULFILLMENT] ❌ Failed to query pending orders: %v", err)
		return
	}
	defer rows.Close()

	checked := 0
	for rows.Next() {
		var id int
		var ref, name string
		if err := rows.Scan(&id, &ref, &name); err != nil {
			continue
		}
		if ref == "" {
			continue
		}

		log.Printf("[FULFILLMENT] 🔄 Re-checking pending order #%d (ref=%s)", id, ref)

		// Try all registered providers (we don't store provider name in orders table yet)
		for _, p := range fe.registry.All() {
			if p.Name() == "demo" {
				continue
			}
			status, err := p.CheckStatus(ref)
			if err != nil || status == nil {
				continue
			}
			if status.Status == "completed" {
				fe.db.Exec(`
					UPDATE store_orders SET status = 'ready', activation_key = $1, qr_data = $2 WHERE id = $3`,
					status.ActivationKey, status.QRData, id)
				log.Printf("[FULFILLMENT] ✅ Pending order #%d resolved → ready", id)
				break
			}
			if status.Status == "failed" {
				fe.markOrderFailed(id, status.ErrorMessage)
				break
			}
		}
		checked++
	}

	if checked > 0 {
		log.Printf("[FULFILLMENT] 🔄 Re-checked %d pending orders", checked)
	}
}

// StartRetryLoop starts a background goroutine that periodically retries pending orders.
func (fe *FulfillmentEngine) StartRetryLoop() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			fe.RetryPendingOrders()
		}
	}()
	log.Println("[FULFILLMENT] ✅ Retry loop started (interval=5m)")
}
