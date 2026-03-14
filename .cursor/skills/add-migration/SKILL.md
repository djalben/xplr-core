---
name: add-migration
description: Создаёт миграции goose для Postgres. Use when adding database migrations, schema changes, or when user asks to create a migration.
---

# Создание миграций

## Единственный способ

```bash
make migrate-create <имя_миграции>
```

Примеры:
- `make migrate-create add_wallet_table`
- `make migrate-create add_card_type_column`

## Важно

- **Никогда** не создавай миграции вручную (копированием, созданием файла)
- Миграции находятся в `./migrations/postgres/`
- Goose генерирует timestamp в имени файла
