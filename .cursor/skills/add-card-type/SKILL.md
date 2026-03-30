---
name: add-card-type
description: Добавляет новый тип карты в проект XPLR. Use when adding a new card type (Subscription, Travel, Premium or custom), when user asks for new card type, or when modifying domain/card types.
---

# Добавление нового типа карты

## Workflow

1. **domain/card/newtype.go** — определи новый тип (const или type)
2. **domain/card/type.go** — обнови enum/валидацию типов карт
3. **application/card/service.go** — добавь логику для нового типа
4. **Миграция** — создай миграцию goose (см. @add-migration)

## Пути (могут быть под internal/)

| Слой | Путь |
|------|------|
| domain | domain/card/ или internal/domain/card/ |
| application | application/card/ или internal/application/card/ |

## Ограничения (из правил)

- Только 3 типа: Subscription, Travel, Prime — не добавляй новые без явного запроса
- Соблюдай Dependency Rule: domain не импортирует application/infrastructure
