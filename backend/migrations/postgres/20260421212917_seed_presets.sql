-- +goose Up
-- +goose StatementBegin

-- Seed дефолтных значений (отдельно от init структуры).
INSERT INTO commission_config (key, value, description)
VALUES
    ('fee_standard',      6.70,   'Комиссия для грейда STANDARD (%)'),
    ('fee_gold',          4.50,   'Комиссия для грейда GOLD (%)'),
    ('referral_percent',  5.00,   'Реферальная комиссия (%)'),
    ('card_issue_fee',    2.00,   'Стоимость выпуска виртуальной карты ($)')
ON CONFLICT (key) DO NOTHING;

-- Seed: курсы валют (как в main). Внутренний курс = база + наценка.
INSERT INTO exchange_rates (currency_from, currency_to, base_rate, markup_percent, final_rate, updated_at)
VALUES
    ('RUB', 'USD', 75.04, 3.00, 77.29, NOW()),
    ('RUB', 'EUR', 87.98, 3.00, 90.62, NOW())
ON CONFLICT (currency_from, currency_to) DO UPDATE SET
    base_rate = EXCLUDED.base_rate,
    markup_percent = EXCLUDED.markup_percent,
    final_rate = EXCLUDED.final_rate,
    updated_at = EXCLUDED.updated_at;

-- Seed: PIN для входа в админку (по умолчанию 0000).
-- Хранится в system_settings; бэкенд поддерживает как plain (0000), так и bcrypt-хэш.
INSERT INTO system_settings (setting_key, setting_value, setting_bool, description, updated_at)
VALUES
    ('admin_pin', '0000', NULL, 'PIN для входа в админку (4 цифры). Можно хранить как plain или bcrypt-хэш.', NOW()),
    ('sbp_topup_enabled', '', TRUE, 'Пополнение через СБП включено/отключено (boolean).', NOW())
ON CONFLICT (setting_key) DO NOTHING;

-- Seed: VPN infra defaults (for Admin Dashboard block; 1:1 with main UI).
INSERT INTO system_settings (setting_key, setting_value, setting_bool, description, updated_at)
VALUES
    ('vpn_server_limit_gb', '30', NULL, 'Лимит трафика VPN-сервера (ГБ).', NOW()),
    ('vpn_server_cost_eur', '4.94', NULL, 'Стоимость сервера в EUR (для расчёта маржи).', NOW()),
    ('vpn_server_id', '0', NULL, 'ID сервера у хостера (Aeza). 0 = не настроено.', NOW()),
    ('vpn_server_name', 'VPN-сервер', NULL, 'Имя VPN-сервера (админка).', NOW()),
    ('vpn_server_status', 'active', NULL, 'Статус сервера (админка).', NOW()),
    ('vpn_server_ip', '', NULL, 'IP VPN-сервера.', NOW()),
    ('vpn_server_expires_at', '', NULL, 'Дата окончания оплаты (RFC3339).', NOW()),
    ('vpn_server_cpu', '1', NULL, 'vCPU', NOW()),
    ('vpn_server_ram_mb', '2048', NULL, 'RAM в МБ', NOW()),
    ('vpn_server_disk_gb', '30', NULL, 'Диск в ГБ', NOW()),
    ('vpn_server_disk_type', 'SSD', NULL, 'Тип диска', NOW()),
    ('vpn_server_os', '', NULL, 'OS', NOW()),
    ('vpn_server_location', '', NULL, 'Локация', NOW()),
    ('aeza_balance', '0', NULL, 'Баланс Aeza (если интеграция подключена).', NOW()),
    ('aeza_currency', 'EUR', NULL, 'Валюта баланса Aeza.', NOW())
ON CONFLICT (setting_key) DO NOTHING;

-- Seed: категории магазина.
INSERT INTO store_categories (slug, name, description, icon, sort_order)
VALUES
    ('esim', 'eSIM и Сим-карты', 'Мобильный интернет в любой точке мира', 'esim', 10),
    ('digital', 'Цифровые товары', 'Подарочные карты и подписки', 'digital', 20),
    ('vpn', 'Безопасный доступ', 'Зашифрованный канал, защита данных и стабильный интернет', 'vpn', 30)
ON CONFLICT (slug) DO NOTHING;

-- Seed: VPN планы (в стиле main/vless_provider.go).
INSERT INTO store_products (category_id, provider, external_id, name, description, country, country_code, product_type, price_usd, cost_price, markup_percent, data_gb, validity_days, image_url, in_stock, meta, sort_order)
SELECT
    c.id,
    'vless',
    v.external_id,
    v.name,
    v.description,
    v.country,
    v.country_code,
    'vpn',
    v.retail_price,
    v.cost_price,
    0.00,
    v.data_gb,
    v.validity_days,
    '',
    TRUE,
    v.meta,
    v.sort_order
FROM store_categories c
JOIN (
    VALUES
        ('vless-stockholm-7d',   'Безопасный доступ — 7 дней',   'VLESS+Reality VPN ключ (Швеция). Лимит 15 ГБ, 7 дней.',   'Швеция', 'SE', 5.00::numeric, 0.88::numeric,   '15', 7,   '{"duration_days":7,"traffic_bytes":16106127360}'::text,   10),
        ('vless-stockholm-30d',  'Безопасный доступ — 30 дней',  'VLESS+Reality VPN ключ (Швеция). Лимит 60 ГБ, 30 дней.',  'Швеция', 'SE', 10.00::numeric, 5.30::numeric,  '60', 30,  '{"duration_days":30,"traffic_bytes":64424509440}'::text,  20),
        ('vless-stockholm-180d', 'Безопасный доступ — 180 дней', 'VLESS+Reality VPN ключ (Швеция). Лимит 300 ГБ, 180 дней.', 'Швеция','SE', 35.00::numeric, 26.50::numeric, '300', 180, '{"duration_days":180,"traffic_bytes":322122547200}'::text, 30),
        ('vless-stockholm-365d', 'Безопасный доступ — 365 дней', 'VLESS+Reality VPN ключ (Швеция). Лимит 600 ГБ, 365 дней.', 'Швеция','SE', 55.00::numeric, 48.00::numeric, '600', 365, '{"duration_days":365,"traffic_bytes":644245094400}'::text, 40)
) AS v(external_id, name, description, country, country_code, retail_price, cost_price, data_gb, validity_days, meta, sort_order)
ON c.slug = 'vpn'
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM store_products
WHERE provider = 'vless'
  AND external_id IN (
    'vless-stockholm-7d',
    'vless-stockholm-30d',
    'vless-stockholm-180d',
    'vless-stockholm-365d'
  );

DELETE FROM store_categories
WHERE slug IN ('esim', 'digital', 'vpn');

DELETE FROM commission_config
WHERE key IN ('fee_standard', 'fee_gold', 'referral_percent', 'card_issue_fee');

-- +goose StatementEnd
