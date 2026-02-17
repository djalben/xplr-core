package repository

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/djalben/xplr-core/backend/models"
	"github.com/djalben/xplr-core/backend/notification"
	"github.com/shopspring/decimal"
)

// WallesterRepository - репозиторий для работы с API Wallester
// Реализует принципы Clean Architecture: изоляция внешних зависимостей
type WallesterRepository struct {
	apiKey    string
	apiURL    string
	httpClient *http.Client
}

// NewWallesterRepository создает новый экземпляр репозитория Wallester
func NewWallesterRepository() *WallesterRepository {
	apiKey := os.Getenv("WALLESTER_API_KEY")
	apiURL := os.Getenv("WALLESTER_API_URL")
	
	if apiKey == "" {
		log.Println("⚠️  WALLESTER_API_KEY not set in environment")
	}
	if apiURL == "" {
		log.Println("⚠️  WALLESTER_API_URL not set in environment")
		apiURL = "https://api.wallester.com/v1" // Дефолтный URL
	}

	return &WallesterRepository{
		apiKey: apiKey,
		apiURL: apiURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WallesterCardRequest - запрос на создание карты в Wallester
type WallesterCardRequest struct {
	Currency    string `json:"currency"`
	CardType    string `json:"card_type"`    // VISA, MasterCard
	CardName    string `json:"card_name,omitempty"`
	Amount      string `json:"amount,omitempty"` // Начальный баланс карты
}

// WallesterCardResponse - ответ от Wallester при создании карты
type WallesterCardResponse struct {
	Success bool   `json:"success"`
	CardID  string `json:"card_id,omitempty"`
	PAN     string `json:"pan,omitempty"`      // Маскированный номер карты
	BIN     string `json:"bin,omitempty"`
	Last4   string `json:"last_4,omitempty"`
	Status  string `json:"status,omitempty"`
	Balance string `json:"balance,omitempty"`
	Error   string `json:"error,omitempty"`
}

// WallesterCardDetailsResponse - ответ с деталями карты (PAN, CVV, expiry)
type WallesterCardDetailsResponse struct {
	Success bool   `json:"success"`
	PAN     string `json:"pan,omitempty"`
	CVV     string `json:"cvv,omitempty"`
	Expiry  string `json:"expiry,omitempty"` // MM/YY
	Error   string `json:"error,omitempty"`
}

// WallesterBalanceResponse - ответ с балансом карты
type WallesterBalanceResponse struct {
	Success bool   `json:"success"`
	Balance string `json:"balance,omitempty"`
	Error   string `json:"error,omitempty"`
}

// GetServiceIDBySlug получает service_id из таблицы services по slug
func GetServiceIDBySlug(slug string) (*int, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	var serviceID int
	query := `SELECT id FROM services WHERE slug = $1 LIMIT 1`
	err := GlobalDB.QueryRow(query, slug).Scan(&serviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("service with slug '%s' not found", slug)
		}
		return nil, fmt.Errorf("failed to get service_id: %w", err)
	}

	return &serviceID, nil
}

// IssueCard - Выпуск виртуальной карты через API Wallester
// Принимает userID, serviceSlug и параметры карты
// Проверяет баланс пользователя перед выпуском
func (wr *WallesterRepository) IssueCard(
	userID int,
	serviceSlug string,
	cardType string,
	nickname string,
	dailyLimit decimal.Decimal,
) (*models.Card, error) {
	if GlobalDB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	// 1. Проверка баланса пользователя (balance_rub)
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Минимальная сумма для выпуска карты + начальное пополнение (например, 100 руб)
	minRequired := decimal.NewFromInt(100)
	if user.BalanceRub.LessThan(minRequired) {
		return nil, fmt.Errorf("insufficient funds: required %s, available %s", 
			minRequired.String(), user.BalanceRub.String())
	}

	// 2. Получение service_id по slug
	serviceID, err := GetServiceIDBySlug(serviceSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get service_id: %w", err)
	}

	// 3. Подготовка запроса к Wallester API
	reqBody := WallesterCardRequest{
		Currency: "RUB",
		CardType: cardType,
		CardName: nickname,
		Amount:   "0", // Начальный баланс 0, пополнение через наш баланс
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 4. POST-запрос к Wallester
	url := fmt.Sprintf("%s/cards", wr.apiURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", wr.apiKey))

	resp, err := wr.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wallester API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("wallester API error (status %d): %s", resp.StatusCode, string(body))
	}

	var wallesterResp WallesterCardResponse
	if err := json.Unmarshal(body, &wallesterResp); err != nil {
		return nil, fmt.Errorf("failed to parse wallester response: %w", err)
	}

	if !wallesterResp.Success || wallesterResp.CardID == "" {
		return nil, fmt.Errorf("wallester card creation failed: %s", wallesterResp.Error)
	}

	// 5. Сохранение карты в таблицу cards Supabase
	card := models.Card{
		UserID:          userID,
		ServiceID:       serviceID,
		ProviderCardID:  wallesterResp.CardID,
		ExternalID:      wallesterResp.CardID, // external_id = card_id от Wallester
		BIN:             wallesterResp.BIN,
		Last4Digits:     wallesterResp.Last4,
		CardStatus:      wallesterResp.Status,
		Status:          wallesterResp.Status,
		Nickname:        nickname,
		ServiceSlug:     serviceSlug,
		DailySpendLimit: dailyLimit,
		CardType:        cardType,
		CardBalance:     decimal.Zero,
		CreatedAt:       time.Now(),
	}

	// Вставка в БД
	query := `
		INSERT INTO cards (
			user_id, service_id, provider_card_id, external_id, bin, last_4_digits,
			card_status, status, nickname, service_slug, daily_spend_limit,
			card_type, card_balance, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
	`

	var cardID int
	err = GlobalDB.QueryRow(
		query,
		card.UserID,
		card.ServiceID,
		card.ProviderCardID,
		card.ExternalID,
		card.BIN,
		card.Last4Digits,
		card.CardStatus,
		card.Status,
		card.Nickname,
		card.ServiceSlug,
		card.DailySpendLimit,
		card.CardType,
		card.CardBalance,
		card.CreatedAt,
	).Scan(&cardID)

	if err != nil {
		return nil, fmt.Errorf("failed to save card to database: %w", err)
	}

	card.ID = cardID
	log.Printf("✅ Card issued via Wallester: ID=%d, ExternalID=%s, Last4=%s", 
		cardID, card.ExternalID, card.Last4Digits)

	return &card, nil
}

// GetCardDetails - Получение реквизитов карты (PAN, CVV, expiry) из Wallester
func (wr *WallesterRepository) GetCardDetails(externalID string) (*WallesterCardDetailsResponse, error) {
	if wr.apiKey == "" {
		return nil, fmt.Errorf("WALLESTER_API_KEY not configured")
	}

	url := fmt.Sprintf("%s/cards/%s/details", wr.apiURL, externalID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", wr.apiKey))

	resp, err := wr.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wallester API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wallester API error (status %d): %s", resp.StatusCode, string(body))
	}

	var details WallesterCardDetailsResponse
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !details.Success {
		return nil, fmt.Errorf("wallester error: %s", details.Error)
	}

	return &details, nil
}

// SyncBalance - Синхронизация баланса карты из Wallester в нашу БД
func (wr *WallesterRepository) SyncBalance(cardID int, externalID string) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	if wr.apiKey == "" {
		return fmt.Errorf("WALLESTER_API_KEY not configured")
	}

	// Запрос баланса из Wallester
	url := fmt.Sprintf("%s/cards/%s/balance", wr.apiURL, externalID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", wr.apiKey))

	resp, err := wr.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("wallester API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wallester API error (status %d): %s", resp.StatusCode, string(body))
	}

	var balanceResp WallesterBalanceResponse
	if err := json.Unmarshal(body, &balanceResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !balanceResp.Success {
		return fmt.Errorf("wallester error: %s", balanceResp.Error)
	}

	// Парсинг баланса
	balance, err := decimal.NewFromString(balanceResp.Balance)
	if err != nil {
		return fmt.Errorf("invalid balance format: %w", err)
	}

	// Обновление card_balance в БД
	_, err = GlobalDB.Exec(
		"UPDATE cards SET card_balance = $1 WHERE id = $2",
		balance, cardID,
	)

	if err != nil {
		return fmt.Errorf("failed to update card balance: %w", err)
	}

	log.Printf("✅ Synced balance for card %d (external_id=%s): %s", 
		cardID, externalID, balance.String())

	return nil
}

// SyncAllCardsBalances - Синхронизация балансов всех активных карт
func (wr *WallesterRepository) SyncAllCardsBalances() error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Получаем все активные карты с external_id
	query := `
		SELECT id, external_id, provider_card_id 
		FROM cards 
		WHERE card_status = 'ACTIVE' 
		  AND (external_id IS NOT NULL AND external_id != '' OR provider_card_id IS NOT NULL AND provider_card_id != '')
	`

	rows, err := GlobalDB.Query(query)
	if err != nil {
		return fmt.Errorf("failed to fetch cards: %w", err)
	}
	defer rows.Close()

	var synced, failed int
	for rows.Next() {
		var cardID int
		var externalID, providerCardID sql.NullString

		if err := rows.Scan(&cardID, &externalID, &providerCardID); err != nil {
			log.Printf("Error scanning card: %v", err)
			continue
		}

		// Используем external_id, если есть, иначе provider_card_id
		extID := externalID.String
		if extID == "" {
			extID = providerCardID.String
		}

		if extID == "" {
			continue
		}

		if err := wr.SyncBalance(cardID, extID); err != nil {
			log.Printf("Failed to sync balance for card %d: %v", cardID, err)
			failed++
		} else {
			synced++
		}
	}

	log.Printf("Balance sync completed: %d synced, %d failed", synced, failed)
	return nil
}

// WallesterWebhookPayload - структура для обработки webhook от Wallester
type WallesterWebhookPayload struct {
	EventType     string                 `json:"event_type"`      // transaction, balance_update, 3ds_authentication, etc.
	CardID        string                 `json:"card_id"`         // external_id карты
	TransactionID string                 `json:"transaction_id"`  // ID транзакции от Wallester (для idempotency)
	Amount        string                 `json:"amount"`
	Currency      string                 `json:"currency"`
	Status        string                 `json:"status"`
	Timestamp     string                 `json:"timestamp"`
	AuthCode      string                 `json:"auth_code,omitempty"`      // Код подтверждения для 3DS
	MerchantName  string                 `json:"merchant_name,omitempty"`  // Название магазина
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CheckIPWhitelist проверяет, что IP-адрес находится в whitelist Wallester
func CheckIPWhitelist(clientIP string) bool {
	allowedIPs := os.Getenv("WALLESTER_WEBHOOK_IPS")
	if allowedIPs == "" {
		log.Println("⚠️  WALLESTER_WEBHOOK_IPS not set, allowing all IPs (not recommended for production)")
		return true // В development разрешаем все IP, если не настроено
	}

	// Разделяем список IP через запятую
	ips := strings.Split(allowedIPs, ",")
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		// Поддержка CIDR нотации (например, 192.168.1.0/24)
		if strings.Contains(ip, "/") {
			_, ipNet, err := net.ParseCIDR(ip)
			if err != nil {
				log.Printf("Invalid CIDR in whitelist: %s", ip)
				continue
			}
			clientAddr := net.ParseIP(clientIP)
			if clientAddr != nil && ipNet.Contains(clientAddr) {
				return true
			}
		} else {
			// Точное совпадение IP
			if clientIP == ip {
				return true
			}
		}
	}

	return false
}

// VerifyWebhookSignature проверяет подпись webhook от Wallester
// Использует HMAC-SHA256 для проверки заголовка X-Wallester-Signature
func VerifyWebhookSignature(body []byte, signature string) bool {
	secretKey := os.Getenv("WALLESTER_WEBHOOK_SECRET")
	if secretKey == "" {
		log.Println("⚠️  WALLESTER_WEBHOOK_SECRET not set, skipping signature verification")
		return true // В development пропускаем проверку, если не настроено
	}

	if signature == "" {
		return false
	}

	// Вычисляем HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Сравниваем подписи (защита от timing attacks через hmac.Equal)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// CheckTransactionIdempotency проверяет, не обрабатывалась ли уже транзакция с данным provider_tx_id
func CheckTransactionIdempotency(providerTxID string) (bool, error) {
	if GlobalDB == nil {
		return false, fmt.Errorf("database connection not initialized")
	}

	if providerTxID == "" {
		return false, nil // Если нет ID, считаем новой транзакцией
	}

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM transactions WHERE provider_tx_id = $1)`
	err := GlobalDB.QueryRow(query, providerTxID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check idempotency: %w", err)
	}

	return exists, nil
}

// ProcessWebhook - Обработка webhook от Wallester
// Обновляет balance_rub пользователя на основе транзакций
// Включает проверку idempotency через provider_tx_id
func (wr *WallesterRepository) ProcessWebhook(payload WallesterWebhookPayload) error {
	if GlobalDB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// IDEMPOTENCY: Проверяем, не обрабатывалась ли уже эта транзакция
	if payload.TransactionID != "" {
		exists, err := CheckTransactionIdempotency(payload.TransactionID)
		if err != nil {
			return fmt.Errorf("failed to check idempotency: %w", err)
		}
		if exists {
			log.Printf("⚠️  Webhook already processed: transaction_id=%s, skipping", payload.TransactionID)
			return nil // Возвращаем успех, но не обрабатываем повторно
		}
	}

	// Находим карту по external_id (получаем также last_4_digits для уведомлений)
	var cardID, userID int
	var last4Digits string
	query := `SELECT id, user_id, last_4_digits FROM cards WHERE external_id = $1 OR provider_card_id = $1 LIMIT 1`
	err := GlobalDB.QueryRow(query, payload.CardID).Scan(&cardID, &userID, &last4Digits)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("card not found: external_id=%s", payload.CardID)
		}
		return fmt.Errorf("failed to find card: %w", err)
	}

	// Парсим сумму транзакции (может быть пустой для некоторых событий, например 3DS)
	var amount decimal.Decimal
	if payload.Amount != "" {
		amount, err = decimal.NewFromString(payload.Amount)
		if err != nil {
			return fmt.Errorf("invalid amount format: %w", err)
		}
	}

	// Обработка разных типов событий
	switch payload.EventType {
	case "3ds_authentication":
		// Обработка события 3DS аутентификации
		// Извлекаем auth_code и merchant_name из payload
		authCode := payload.AuthCode
		merchantName := payload.MerchantName
		
		if authCode == "" {
			log.Printf("⚠️  3DS webhook received but auth_code is empty")
			return nil // Не критичная ошибка, просто логируем
		}

		// Получаем пользователя для отправки уведомления
		user, err := GetUserByID(userID)
		if err != nil {
			log.Printf("⚠️  Failed to get user for 3DS notification: %v", err)
			return nil // Не прерываем обработку, только логируем
		}

		// Отправляем уведомление о 3DS коде
		if user.TelegramChatID.Valid {
			message := fmt.Sprintf(
				"🔑 Код подтверждения: %s | Магазин: %s\n\n⚠️ Внимание: не сообщайте код третьим лицам!",
				authCode,
				merchantName,
			)
			// Отправка уведомления (не блокируем обработку при ошибке)
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("⚠️  Panic in 3DS notification: %v", r)
					}
				}()
				notification.SendTelegramMessage(user.TelegramChatID.Int64, message)
			}()
		} else {
			log.Printf("⚠️  User %d has no Telegram chat ID, skipping 3DS notification", userID)
		}

		log.Printf("✅ 3DS notification sent: auth_code=%s, merchant=%s, user=%d", 
			authCode, merchantName, userID)
		return nil

	case "transaction", "capture", "authorization", "payment_success":
		// Списание средств с баланса пользователя
		if payload.Status == "approved" || payload.Status == "completed" || payload.EventType == "payment_success" {
			// Проверяем баланс перед списанием
			user, err := GetUserByID(userID)
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}

			if user.BalanceRub.LessThan(amount) {
				return fmt.Errorf("insufficient funds for transaction: required %s, available %s",
					amount.String(), user.BalanceRub.String())
			}

			// Начинаем транзакцию БД для атомарности
			tx, err := GlobalDB.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}
			defer tx.Rollback()

			// Списание с balance_rub
			_, err = tx.Exec(
				"UPDATE users SET balance_rub = balance_rub - $1 WHERE id = $2",
				amount, userID,
			)
			if err != nil {
				return fmt.Errorf("failed to deduct balance: %w", err)
			}

			// Получаем новый баланс после списания
			var newBalance decimal.Decimal
			err = tx.QueryRow("SELECT balance_rub FROM users WHERE id = $1", userID).Scan(&newBalance)
			if err != nil {
				return fmt.Errorf("failed to get new balance: %w", err)
			}

			// Записываем транзакцию с provider_tx_id для idempotency
			merchantName := payload.MerchantName
			if merchantName == "" {
				merchantName = "Unknown"
			}
			_, err = tx.Exec(
				`INSERT INTO transactions (user_id, card_id, amount, fee, transaction_type, status, details, provider_tx_id, executed_at)
				 VALUES ($1, $2, $3, $4, 'CAPTURE', 'APPROVED', $5, $6, $7)`,
				userID,
				cardID,
				amount,
				decimal.Zero, // Комиссия для webhook транзакций
				fmt.Sprintf("Wallester webhook: %s from card %s, merchant: %s", payload.EventType, payload.CardID, merchantName),
				payload.TransactionID,
				time.Now(),
			)
			if err != nil {
				return fmt.Errorf("failed to record transaction: %w", err)
			}

			// Коммитим транзакцию
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit transaction: %w", err)
			}

			log.Printf("✅ Webhook: Deducted %s from user %d (card %s, tx_id=%s)", 
				amount.String(), userID, payload.CardID, payload.TransactionID)

			// Отправляем уведомление об успешном списании (не блокируем обработку при ошибке)
			if user.TelegramChatID.Valid {
				currency := payload.Currency
				if currency == "" {
					currency = "RUB"
				}
				message := fmt.Sprintf(
					"💸 Списание: %s %s | Карта: *%s | Магазин: %s\n\nВаш новый баланс: %s₽",
					amount.String(),
					currency,
					last4Digits,
					merchantName,
					newBalance.String(),
				)
				// Отправка уведомления (не блокируем обработку при ошибке)
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("⚠️  Panic in transaction notification: %v", r)
						}
					}()
					notification.SendTelegramMessage(user.TelegramChatID.Int64, message)
				}()
			} else {
				log.Printf("⚠️  User %d has no Telegram chat ID, skipping transaction notification", userID)
			}
		}

	case "refund", "reversal":
		// Возврат средств на баланс пользователя
		tx, err := GlobalDB.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		_, err = tx.Exec(
			"UPDATE users SET balance_rub = balance_rub + $1 WHERE id = $2",
			amount, userID,
		)
		if err != nil {
			return fmt.Errorf("failed to refund balance: %w", err)
		}

		// Записываем транзакцию возврата
		_, err = tx.Exec(
			`INSERT INTO transactions (user_id, card_id, amount, fee, transaction_type, status, details, provider_tx_id, executed_at)
			 VALUES ($1, $2, $3, $4, 'REFUND', 'APPROVED', $5, $6, $7)`,
			userID,
			cardID,
			amount,
			decimal.Zero,
			fmt.Sprintf("Wallester webhook: %s for card %s", payload.EventType, payload.CardID),
			payload.TransactionID,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to record refund transaction: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit refund transaction: %w", err)
		}

		log.Printf("✅ Webhook: Refunded %s to user %d (card %s, tx_id=%s)", 
			amount.String(), userID, payload.CardID, payload.TransactionID)

	case "balance_update":
		// Прямое обновление баланса карты (синхронизация)
		return wr.SyncBalance(cardID, payload.CardID)

	default:
		log.Printf("⚠️  Unknown webhook event type: %s", payload.EventType)
	}

	return nil
}
