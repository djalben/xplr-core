package core

import (
	"fmt"
	"log"

	"github.com/djalben/xplr-core/backend/config"
	"github.com/djalben/xplr-core/backend/models"
	"github.com/djalben/xplr-core/backend/repository"
	"github.com/djalben/xplr-core/backend/service"
	"github.com/shopspring/decimal"
)

// AuthorizeCardRequest - запрос авторизации от провайдера
type AuthorizeCardRequest struct {
	CardID       int             `json:"card_id"`       // ID карты в нашей системе
	Amount       decimal.Decimal `json:"amount"`        // Сумма транзакции
	MerchantName string          `json:"merchant_name"` // Название мерчанта
}

// authorizeCard - Центральная функция, которая обрабатывает все проверки и записывает транзакцию.
// Это ядро Zero Decline Logic.
func AuthorizeCard(req AuthorizeCardRequest) models.AuthResponse {
	// 1. Получаем карту по ID
	card, err := repository.GetCardByID(req.CardID)
	if err != nil {
		return models.AuthResponse{
			Success: false,
			Status:  "DECLINED",
			Message: "Card not found.",
			Fee:     decimal.NewFromFloat(config.DeclineFee),
		}
	}

	// 2. Получаем пользователя
	user, err := repository.GetUserByID(card.UserID)
	if err != nil {
		return models.AuthResponse{
			Success: false,
			Status:  "DECLINED",
			Message: "User not found.",
			Fee:     decimal.NewFromFloat(config.DeclineFee),
		}
	}

	// 2.5. АНТИ-ФРОД: Автоматическая блокировка после MaxFailedAttempts
	if card.FailedAuthCount >= config.MaxFailedAttempts {
		log.Printf("ANTI-FRAUD: Card %d has %d failed attempts, blocking card", card.ID, card.FailedAuthCount)

		// Блокируем карту в БД
		err := repository.BlockCard(card.ID)
		if err != nil {
			log.Printf("ERROR: Failed to block card %d: %v", card.ID, err)
		}

		// Уведомление о блокировке (TG + Email)
		go service.NotifyUser(user.ID, "Карта заблокирована",
			fmt.Sprintf("🔒 <b>Карта *%s заблокирована</b>\n\n"+
				"Причина: множественные неудачные попытки авторизации.\n\n"+
				"<a href=\"https://xplr.pro/cards\">Открыть карты</a>",
				card.Last4Digits))

		return models.AuthResponse{
			Success: false,
			Status:  "DECLINED",
			Message: "Card blocked due to multiple failed attempts.",
			Fee:     decimal.NewFromFloat(config.DeclineFee),
		}
	}

	// 3. БИЗНЕС-ЛОГИКА: Проверка баланса и лимитов (Zero Decline Logic)

	// Проверка 3.1: Статус карты (блокировка)
	if card.CardStatus != "ACTIVE" {
		log.Printf("DECLINED: Card %d is not active (Status: %s)", req.CardID, card.CardStatus)

		// Уведомление о DECLINE (TG + Email)
		go service.NotifyUser(user.ID, "Транзакция отклонена",
			fmt.Sprintf("❌ <b>Транзакция по карте *%s отклонена</b>\n\n"+
				"Причина: карта не активна (статус: %s).\n\n"+
				"<a href=\"https://xplr.pro/cards\">Открыть карты</a>",
				card.Last4Digits, card.CardStatus))

		return models.AuthResponse{
			Success: false,
			Status:  "DECLINED",
			Message: "Card is blocked or inactive.",
			Fee:     decimal.NewFromFloat(config.DeclineFee),
		}
	}

	// Проверка 3.2: Баланс (XPLR: BalanceRub — основной баланс в рублях)
	if user.BalanceRub.LessThan(req.Amount) {
		log.Printf("DECLINED: User %d balance_rub (%s) is insufficient for transaction %s", user.ID, user.BalanceRub.String(), req.Amount.String())

		go service.NotifyUser(user.ID, "Транзакция отклонена",
			fmt.Sprintf("❌ <b>Транзакция по карте *%s отклонена</b>\n\n"+
				"Причина: недостаточно средств.\n"+
				"Баланс: <b>%s</b>, сумма: <b>%s</b>\n\n"+
				"<a href=\"https://xplr.pro/wallet\">Пополнить кошелёк</a>",
				card.Last4Digits, user.BalanceRub.String(), req.Amount.String()))

		repository.IncrementFailedAuthCount(card.ID)

		return models.AuthResponse{
			Success: false,
			Status:  "DECLINED",
			Message: "Insufficient user balance.",
			Fee:     decimal.NewFromFloat(config.DeclineFee),
		}
	}

	// Проверка 3.3: Дневной лимит (Rule-Based Blocking)
	if req.Amount.GreaterThan(card.DailySpendLimit) && card.DailySpendLimit.GreaterThan(decimal.Zero) {
		log.Printf("DECLINED: Card %d daily limit (%s) exceeded by transaction %s", req.CardID, card.DailySpendLimit.String(), req.Amount.String())

		// Уведомление о DECLINE (TG + Email)
		go service.NotifyUser(user.ID, "Транзакция отклонена",
			fmt.Sprintf("❌ <b>Транзакция по карте *%s отклонена</b>\n\n"+
				"Причина: превышен дневной лимит (лимит: <b>$%s</b>).\n\n"+
				"<a href=\"https://xplr.pro/cards\">Открыть карты</a>",
				card.Last4Digits, card.DailySpendLimit.String()))

		return models.AuthResponse{
			Success: false,
			Status:  "DECLINED",
			Message: "Daily spend limit exceeded.",
			Fee:     decimal.NewFromFloat(config.DeclineFee),
		}
	}

	// 4. УСПЕХ (APPROVED)

	// 4.1. Получить Grade пользователя и вычислить комиссию ПЕРЕД обработкой платежа
	userGrade, err := repository.GetUserGrade(user.ID)
	if err != nil {
		log.Printf("Warning: Failed to get user grade for user %d: %v", user.ID, err)
		// Используем стандартную комиссию 6.7% если Grade не найден (как у e.pn)
		userGrade = &models.UserGrade{
			FeePercent: decimal.NewFromFloat(6.70),
		}
	}

	// Вычислить комиссию на основе Grade (fee_percent в процентах, например 6.70 = 6.7%)
	fee := req.Amount.Mul(userGrade.FeePercent).Div(decimal.NewFromInt(100))

	// 4.2. Списание средств и запись транзакции в рамках атомарной операции (с комиссией)
	err = repository.ProcessCardPayment(user.ID, card.ID, req.Amount, fee, req.MerchantName, card.Last4Digits)
	if err != nil {
		log.Printf("CRITICAL DB ERROR: Failed to process payment for user %d: %v", user.ID, err)
		// Если произошла ошибка БД, отклоняем списание, но БЕЗ комиссии.
		return models.AuthResponse{
			Success: false,
			Status:  "DECLINED",
			Message: "Internal system error during payment processing.",
			Fee:     decimal.NewFromFloat(config.DeclineFee),
		}
	}

	// 4.3. Уведомление об УСПЕШНОЙ транзакции (TG + Email)
	go service.NotifyUser(user.ID, "Списание с карты",
		fmt.Sprintf("💸 <b>Списание с карты *%s</b>\n\n"+
			"Сумма: <b>%s</b>\n"+
			"Магазин: %s\n"+
			"Комиссия: %s\n\n"+
			"<a href=\"https://xplr.pro/cards\">Открыть карты</a>",
			card.Last4Digits, req.Amount.String(), req.MerchantName, fee.String()))

	return models.AuthResponse{
		Success: true,
		Status:  "APPROVED",
		Message: "Transaction approved.",
		Fee:     fee, // Комиссия на основе Grade пользователя
	}
}

// TestAuthorizeCard - Заглушка для тестирования (POST /v1/authorize)
func TestAuthorizeCard(req AuthorizeCardRequest) models.AuthResponse {
	log.Printf("Test authorize request received: CardID=%d, Amount=%.2f, Merchant=%s", req.CardID, req.Amount, req.MerchantName)
	return AuthorizeCard(req)
}
