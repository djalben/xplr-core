---
name: use-envparse
description: Парсит конфиг через gitlab.com/libs-artifex/envparse. Use when loading configuration, parsing env vars, or modifying config in XPLR.
---

# Использование envparse

## Правило

**Всегда** парси конфиг через `gitlab.com/libs-artifex/envparse`.

## Импорт

```go
import "gitlab.com/libs-artifex/envparse"
```

## Путь конфига

- Конфиг парсится в `internal/config/config.go`
- Структура конфига соответствует тегам envparse

## Запрещено

- `config.MustLoad` или самописный парсинг
- Ручной разбор `os.Getenv` для конфигурации
