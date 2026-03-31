# Сравнение: фронт на `main` (GitHub), старый бэкенд и наш бэкенд (`sandbox`)

Репозиторий: [djalben/xplr-core](https://github.com/djalben/xplr-core/tree/main/).

## Что где лежит

| Область | Ветка `main` (типично) | Ветка `sandbox` (наш) |
|--------|-------------------------|------------------------|
| Фронт | `epn-killer-web/` — Vite + React, один `apiClient` на `/api/v1` | То же дерево; часть фич (tier, `system-settings`, `sbp-toggle`) в `sandbox` убрана как лишняя |
| Бэкенд | Корневой `backend/` — крупный монолит (wallester, куча `service/*`, `telegram/notify`, и т.д.) | `backend/` — Go clean architecture (Chi, goose, `internal/…`) |

Партнёрский фронт изначально заточен под **старый монолит**: пути вида `/user/settings/*`, числовые `id`, ответы с полями в snake_case.

Наш API — **другой контракт**: `/user/me/...`, верификация почты по **ссылке из письма**, JWT без числового id, camelCase в новых полях auth.

## Функции с фронта (`main`) vs наш бэкенд (до доработки)

| Фича (как на фронте) | Старый монолит (ожидание фронта) | Наш бэкенд |
|---------------------|-----------------------------------|------------|
| Регистрация по email | Сразу JWT в ответе | 201, **без JWT**; нужна верификация почты → вход |
| Логин | Один шаг | + **TOTP**: `mfaRequired` + `mfaToken`, затем `/auth/login/mfa` |
| Сброс пароля | `{ token, password }` | `{ token, newPassword }` (camelCase) |
| Настройки: профиль | `GET/PATCH /user/settings/profile` | Был только `GET /user/me` |
| Telegram | `GET …/telegram-link`, polling `…/check-status`, `POST …/unlink` | Были `POST /user/me/telegram/link-code` и `…/link` |
| 2FA | `POST …/2fa/setup` → `{ secret, otp_uri }` | Был `otpauth_url` без отдельного `secret` в JSON |
| Уведомления: канал | `both` / `email` / `telegram` + опции только если есть TG | Были только `notify_email` / `notify_telegram`; **не** проверялись привязка и верификация почты |
| Типы уведомлений | `notify_transactions`, `notify_balance`, `notify_security` (+ карты в копирайте) | В БД не было отдельных флагов категорий |
| KYC в настройках | `GET/POST /user/settings/kyc` | Были `POST /user/kyc/application` + админ |
| Сессии / «выйти везде» | `GET …/sessions`, `POST …/logout-all` | Нет (заглушки) |
| Смена пароля в ЛК | `POST …/change-password` | Не было |
| Повторная отправка письма верификации | `POST …/verify-email-request` | Не было отдельного эндпоинта |
| Админка | Много кастомных `/admin/...` у монолита | У нас узкий набор `/api/admin/*`; **Staff Only Zone** на фронте многого не найдёт |

## Безопасность админки

- **Фронт:** маршрут `/staff-only-zone` защищён `AdminRoute`: JWT + `isAdmin` из `/user/me` + `sessionStorage._xplr_staff === 'granted'` после PIN (тройной клик по логотипу **только если** `isAdmin`). Обычный пользователь с `/staff-only-zone` уходит на `/dashboard`.
- **Бэкенд:** группа `/api/admin/*` с `AdminOnly` — без роли админа ответ **403**, даже если URL угадать.

PIN на фронте — **не секрет сервера**; реальная защита — JWT + `is_admin` на API.

## Что мы сознательно не переносим

- Старый монолитный `backend/` целиком.
- Словари/переводы в админке, лишние экраны Staff Only под несуществующие API.
- Tier / отдельные «системные настройки» фронта, если вы их уже убрали в `sandbox`.

## Что сделано для совместимости (ветка `sandbox`)

1. **Слой BFF** `GET/POST/PATCH …/user/settings/*` на нашем Go — внутри вызываются наши use case’ы.
2. Расширение **users**: флаги категорий уведомлений + валидация каналов (email только при верифицированной почте, telegram — при привязанном чате).
3. **Auth:** доработки регистрации/ MFA / resend verification / смена пароля где не хватало.
4. Фронт: **auth** под новый контракт; вкладка уведомлений — отключение вариантов с **email**, если почта не подтверждена.

Дальнейшая работа: постепенно **срезать** совместимость и перевести фронт на прямые пути `/user/me/...`, упростить Staff Only под реальные admin-эндпоинты.
