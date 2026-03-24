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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
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
    balance NUMERIC(20, 4) DEFAULT 0.0000 NOT NULL,
    expiry_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CHECK (card_type IN ('subscriptions', 'travel', 'premium'))
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
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
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

-- Seed дефолтных значений (один раз при создании базы)
INSERT INTO commission_config (key, value, description)
VALUES
    ('fee_standard',     6.70, 'Комиссия для грейда STANDARD (%)'),
    ('fee_silver',       5.50, 'Комиссия для грейда SILVER (%)'),
    ('fee_gold',         4.50, 'Комиссия для грейда GOLD (%)'),
    ('fee_platinum',     3.50, 'Комиссия для грейда PLATINUM (%)'),
    ('fee_black',        2.50, 'Комиссия для грейда BLACK (%)'),
    ('referral_percent', 5.00, 'Реферальная комиссия (%)'),
    ('card_issue_fee',   2.00, 'Стоимость выпуска виртуальной карты ($)')
ON CONFLICT (key) DO NOTHING;

-- Индексы (все внизу)
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_referral_code ON users(referral_code);
CREATE INDEX idx_users_referred_by ON users(referred_by);
CREATE INDEX idx_cards_user_id ON cards(user_id);
CREATE INDEX idx_cards_user_status ON cards(user_id, card_status);
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

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS commission_config;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS exchange_rates;
DROP TABLE IF EXISTS referrals;
DROP TABLE IF EXISTS user_grades;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd