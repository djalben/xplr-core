# Wallester Repository - Clean Architecture Integration

## Описание

Репозиторий для работы с API Wallester, реализованный по принципам Clean Architecture. Изолирует внешние зависимости (HTTP-клиент, API Wallester) от бизнес-логики.

## Структура

### Файлы
- `backend/repository/wallester.go` - Репозиторий Wallester
- `backend/handlers/wallester.go` - HTTP-хендлеры для работы с Wallester

## Переменные окружения

Добавьте в `.env`:
```env
WALLESTER_API_KEY=your_wallester_api_key
WALLESTER_API_URL=https://api.wallester.com/v1
```

## Методы репозитория

### IssueCard
Выпуск виртуальной карты через Wallester API.

**Параметры:**
- `userID` - ID пользователя
- `serviceSlug` - slug сервиса ('arbitrage', 'travel', 'subscriptions')
- `cardType` - тип карты ('VISA', 'MasterCard')
- `nickname` - имя карты
- `dailyLimit` - дневной лимит

**Логика:**
1. Проверяет `balance_rub` пользователя (минимум 100 руб)
2. Получает `service_id` из таблицы `services` по slug
3. Отправляет POST-запрос к Wallester API
4. Сохраняет карту в таблицу `cards` с привязкой к `service_id`

### GetCardDetails
Получение реквизитов карты (PAN, CVV, expiry) из Wallester.

**Параметры:**
- `externalID` - внешний ID карты (от Wallester)

**Возвращает:**
- `WallesterCardDetailsResponse` с PAN, CVV, expiry

### SyncBalance
Синхронизация баланса конкретной карты из Wallester в БД.

**Параметры:**
- `cardID` - ID карты в нашей БД
- `externalID` - внешний ID карты

**Обновляет:**
- Поле `card_balance` в таблице `cards`

### SyncAllCardsBalances
Синхронизация балансов всех активных карт (вызывается периодически).

### ProcessWebhook
Обработка webhook от Wallester.

**Поддерживаемые события:**
- `transaction`, `capture`, `authorization` - списание с `balance_rub`
- `refund`, `reversal` - возврат на `balance_rub`
- `balance_update` - синхронизация баланса

## API Endpoints

### Webhook (публичный)
```
POST /api/v1/webhooks/wallester
```

### Получение реквизитов карты (защищенный)
```
GET /api/v1/user/cards/{id}/details
```

### Синхронизация баланса (защищенный)
```
POST /api/v1/user/cards/{id}/sync-balance
```

## Автоматическая синхронизация

В `main.go` запущен фоновый процесс, который каждые 5 минут синхронизирует балансы всех активных карт из Wallester.

## Интеграция с Supabase

Репозиторий использует:
- Таблицу `services` для получения `service_id` по `slug`
- Таблицу `cards` для сохранения карт с полями:
  - `service_id` - связь с таблицей services
  - `external_id` - ID карты в Wallester
  - `provider_card_id` - дублирует external_id для совместимости
- Таблицу `users` для обновления `balance_rub` при транзакциях
