---
name: use-wrapper
description: Оборачивает ошибки и логирование через gitlab.com/libs-artifex/wrapper. Use when handling errors, logging, or when user asks about error handling in XPLR.
---

# Использование wrapper

## Правило

**Всегда** используй `gitlab.com/libs-artifex/wrapper` вместо:
- `errors.New`
- `fmt.Errorf`
- `log.Printf` / `slog`

## Импорт

```go
import "gitlab.com/libs-artifex/wrapper"
```

## Типичные функции

- `wrapper.WithLogAndWrap` — обёртка с логированием (в usecases/handlers)
- `wrapper.Wrap` — простая обёртка
- Используй функции библиотеки согласно документации

## Запрещено

- Самописные обёртки ошибок
- Прямое использование стандартного `log`/`slog` для ошибок
