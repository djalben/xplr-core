---
name: implement-antifraud
description: Реализует антифрод: подсчёт неудачных попыток, блокировку карты после 3 попыток. Use when implementing antifraud logic, card blocking, failed auth attempts, or fraud prevention.
---

# Реализация антифрода

## Workflow

1. **transaction_repo** — считай попытки (инкремент failed_auth_count при неудачной авторизации)
2. **Блокировка** — после 3 неудачных попыток блокируй карту
3. **Ошибки** — используй `domain/transaction/errors.go` (sentinel errors для блокировки)
4. **Логирование** — только через wrapper (см. @use-wrapper)

## Логика

- При неудачной auth: `failed_auth_count++`
- Если `failed_auth_count >= 3` → блокировка карты
- Уведомить пользователя (Telegram и т.д.)

## Ограничения

- НЕ используй log.Printf / fmt.Errorf — только wrapper
- Ошибки — через `errors.Is` / `errors.As` и sentinel errors
