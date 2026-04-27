-- +goose Up
-- +goose StatementBegin

-- UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 1. Таблица пользователей
CREATE TABLE users (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    kyc_status VARCHAR(50) DEFAULT 'PENDING',
    status VARCHAR(50) DEFAULT 'ACTIVE',
    telegram_chat_id BIGINT DEFAULT NULL,
    referral_code VARCHAR(20) UNIQUE,
    referred_by UUID REFERENCES users(id),
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verify_token_hash VARCHAR(64),
    email_verify_expires_at TIMESTAMPTZ,
    password_reset_token_hash VARCHAR(64),
    password_reset_expires_at TIMESTAMPTZ,
    totp_secret VARCHAR(64),
    totp_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    notify_email BOOLEAN NOT NULL DEFAULT TRUE,
    notify_telegram BOOLEAN NOT NULL DEFAULT TRUE,
    notify_transactions BOOLEAN NOT NULL DEFAULT TRUE,
    notify_balance BOOLEAN NOT NULL DEFAULT TRUE,
    notify_security BOOLEAN NOT NULL DEFAULT TRUE,
    notify_card_operations BOOLEAN NOT NULL DEFAULT TRUE,
    telegram_link_code VARCHAR(32),
    telegram_link_expires_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    last_login_ip INET,
    last_login_user_agent TEXT NOT NULL DEFAULT '',
    last_read_news_at TIMESTAMPTZ,
    news_notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    CONSTRAINT users_notify_channel_check CHECK (notify_email OR notify_telegram)
);

-- 1.1 Доверенные устройства (пропуск TOTP на 14–30 дней через HttpOnly cookie)
CREATE TABLE trusted_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL,
    user_agent TEXT NOT NULL DEFAULT '',
    ip INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT trusted_devices_token_hash_unq UNIQUE(token_hash)
);

CREATE INDEX idx_trusted_devices_user_id ON trusted_devices(user_id);
CREATE INDEX idx_trusted_devices_expires_at ON trusted_devices(expires_at);

-- 1.2 Rate limit для auth (анти-брутфорс). Храним блокировки/попытки в БД, чтобы работало в serverless.
CREATE TABLE auth_rate_limits (
    key TEXT PRIMARY KEY,
    attempts INT NOT NULL DEFAULT 0,
    window_ends_at TIMESTAMPTZ NOT NULL,
    blocked_until TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_rate_limits_blocked_until ON auth_rate_limits(blocked_until);

-- 1.3 Сессии/активности входа (для экрана «Последняя активность» как в main).
CREATE TABLE auth_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip INET,
    user_agent TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_auth_sessions_user_created ON auth_sessions(user_id, created_at DESC);

-- 2. Заявки KYC
CREATE TABLE kyc_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    payload_json TEXT NOT NULL DEFAULT '{}',
    admin_comment TEXT,
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT kyc_applications_status_check CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED'))
);

-- 3. Таблица карт
CREATE TABLE cards (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    provider_card_id VARCHAR(100) NOT NULL,
    bin VARCHAR(6) NOT NULL DEFAULT '424242',
    last_4_digits VARCHAR(4) NOT NULL,
    card_status VARCHAR(50) DEFAULT 'ACTIVE',
    nickname VARCHAR(100),
    daily_spend_limit NUMERIC(20, 4) DEFAULT 1000.0000,
    failed_auth_count BIGINT DEFAULT 0,
    card_type VARCHAR(20) DEFAULT 'subscriptions',
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    balance NUMERIC(20, 4) DEFAULT 0.0000 NOT NULL,
    expiry_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CHECK (card_type IN ('subscriptions', 'travel', 'premium')),
    CHECK (currency IN ('USD', 'EUR'))
);

-- 4. Таблица транзакций
CREATE TABLE transactions (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    card_id UUID REFERENCES cards(id),
    amount NUMERIC(20, 4) NOT NULL,
    fee NUMERIC(20, 4) DEFAULT 0.0000,
    transaction_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    details TEXT,
    provider_tx_id VARCHAR(255),
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 5. API Ключи
CREATE TABLE api_keys (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    api_key UUID UNIQUE DEFAULT gen_random_uuid(),
    permissions VARCHAR(50) DEFAULT 'READ_ONLY',
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. Wallets
CREATE TABLE wallets (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    balance NUMERIC(20, 4) DEFAULT 0.0000 NOT NULL,
    auto_topup_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 7. Грейды пользователей
CREATE TABLE user_grades (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    grade VARCHAR(50) DEFAULT 'STANDARD',
    total_spent NUMERIC(20, 4) DEFAULT 0.0000,
    fee_percent NUMERIC(5, 2) DEFAULT 6.70,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_grades_grade_check CHECK (grade IN ('STANDARD', 'GOLD'))
);

-- 8. Рефералка
CREATE TABLE referrals (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    referrer_id UUID REFERENCES users(id) ON DELETE CASCADE,
    referred_id UUID REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'PENDING',
    commission_earned NUMERIC(20, 4) DEFAULT 0.0000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 9. Курсы валют
CREATE TABLE exchange_rates (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    currency_from VARCHAR(10) NOT NULL,
    currency_to VARCHAR(10) NOT NULL,
    base_rate NUMERIC(20, 8) NOT NULL,
    markup_percent NUMERIC(6, 2) NOT NULL DEFAULT 0.00,
    final_rate NUMERIC(20, 8) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT exchange_rates_currency_pair_unq UNIQUE(currency_from, currency_to)
);

-- 10. Тикеты поддержки
CREATE TABLE tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    admin_id UUID REFERENCES users(id),
    tg_chat_id BIGINT,
    subject VARCHAR(255),
    status VARCHAR(20) DEFAULT 'NEW',
    user_message TEXT NOT NULL,
    admin_reply TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    
    CHECK (status IN ('NEW', 'IN_PROGRESS', 'DONE'))
);

-- 11. Конфигурация комиссий (для админки — обязательно!)
CREATE TABLE commission_config (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,           -- fee_standard, referral_percent, card_issue_fee и т.д.
    value NUMERIC(20, 4) NOT NULL,
    description TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 12. Магазин (категории/товары/заказы)
CREATE TABLE store_categories (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    slug VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    icon VARCHAR(50) DEFAULT '',
    image_url TEXT DEFAULT '',
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE store_products (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    category_id UUID NOT NULL REFERENCES store_categories(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,                  -- mobimatter, vless, razer, demo
    external_id VARCHAR(100) NOT NULL,              -- ID в системе поставщика
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    country VARCHAR(100) DEFAULT '',
    country_code VARCHAR(2) DEFAULT '',
    product_type VARCHAR(20) NOT NULL,              -- esim | digital | vpn
    price_usd NUMERIC(20, 4) NOT NULL DEFAULT 0.0000,
    cost_price NUMERIC(20, 4) NOT NULL DEFAULT 0.0000,
    markup_percent NUMERIC(6, 2) NOT NULL DEFAULT 0.00,
    data_gb VARCHAR(20) DEFAULT '',
    validity_days INT NOT NULL DEFAULT 0,
    image_url TEXT DEFAULT '',
    in_stock BOOLEAN NOT NULL DEFAULT TRUE,
    meta TEXT NOT NULL DEFAULT '{}',
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT store_products_type_check CHECK (product_type IN ('esim', 'digital', 'vpn'))
);

CREATE TABLE store_orders (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID REFERENCES store_products(id) ON DELETE SET NULL,
    product_name VARCHAR(255) NOT NULL,
    price_usd NUMERIC(20, 4) NOT NULL DEFAULT 0.0000,
    status VARCHAR(20) NOT NULL DEFAULT 'COMPLETED',
    activation_key TEXT DEFAULT '',
    qr_data TEXT DEFAULT '',
    provider_ref VARCHAR(100) DEFAULT '',
    meta TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT store_orders_status_check CHECK (status IN ('PENDING', 'READY', 'FAILED', 'COMPLETED', 'DELETED'))
);

-- 13. Новости / системные настройки / админ-логи
CREATE TABLE news (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    image_url TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT news_status_check CHECK (status IN ('draft', 'published', 'archived'))
);

CREATE TABLE admin_logs (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    admin_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE system_settings (
    setting_key VARCHAR(100) PRIMARY KEY,
    setting_value TEXT NOT NULL DEFAULT '',
    setting_bool BOOLEAN,
    description TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индексы (все внизу)
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_referral_code ON users(referral_code);
CREATE INDEX idx_users_referred_by ON users(referred_by);
CREATE UNIQUE INDEX idx_users_telegram_chat_id_unique ON users(telegram_chat_id) WHERE telegram_chat_id IS NOT NULL;
CREATE INDEX idx_kyc_applications_user_id ON kyc_applications(user_id);
CREATE INDEX idx_kyc_applications_status ON kyc_applications(status);
CREATE INDEX idx_cards_user_id ON cards(user_id);
CREATE INDEX idx_cards_user_status ON cards(user_id, card_status);
CREATE INDEX idx_cards_currency ON cards(currency);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_card_id ON transactions(card_id);
CREATE INDEX idx_transactions_executed_at ON transactions(executed_at DESC);
CREATE INDEX idx_transactions_provider_tx_id ON transactions(provider_tx_id) WHERE provider_tx_id IS NOT NULL;
CREATE INDEX idx_user_grades_user_id ON user_grades(user_id);
CREATE INDEX idx_referrals_referrer_id ON referrals(referrer_id);
CREATE INDEX idx_referrals_referred_id ON referrals(referred_id);
CREATE INDEX idx_exchange_rates_pair ON exchange_rates(currency_from, currency_to);
CREATE INDEX idx_tickets_user_id ON tickets(user_id);
CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_cards_card_type ON cards(card_type);
CREATE INDEX idx_commission_config_key ON commission_config(key);
CREATE INDEX idx_store_categories_slug ON store_categories(slug);
CREATE INDEX idx_store_products_category ON store_products(category_id);
CREATE INDEX idx_store_products_type ON store_products(product_type);
CREATE INDEX idx_store_products_provider ON store_products(provider);
CREATE INDEX idx_store_orders_user_created ON store_orders(user_id, created_at DESC);
CREATE INDEX idx_store_orders_provider_ref ON store_orders(provider_ref) WHERE provider_ref IS NOT NULL AND provider_ref <> '';
CREATE INDEX idx_news_created ON news(created_at DESC);
CREATE INDEX idx_news_status ON news(status);
CREATE INDEX idx_admin_logs_admin_id ON admin_logs(admin_id);
CREATE INDEX idx_admin_logs_created ON admin_logs(created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS admin_logs;
DROP TABLE IF EXISTS news;
DROP TABLE IF EXISTS commission_config;
DROP TABLE IF EXISTS store_orders;
DROP TABLE IF EXISTS store_products;
DROP TABLE IF EXISTS store_categories;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS exchange_rates;
DROP TABLE IF EXISTS referrals;
DROP TABLE IF EXISTS user_grades;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS kyc_applications;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
