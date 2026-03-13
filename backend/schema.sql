-- Активация расширения для генерации UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    balance NUMERIC(20, 4) DEFAULT 0.0000 NOT NULL,
    balance_rub NUMERIC(20, 4) DEFAULT 0.0000 NOT NULL,
    kyc_status VARCHAR(50) DEFAULT 'pending',
    active_mode VARCHAR(50) DEFAULT 'personal',
    status VARCHAR(50) DEFAULT 'ACTIVE',
    telegram_chat_id BIGINT DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Таблица карт
CREATE TABLE IF NOT EXISTS cards (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    -- ID карты от Wallester/Банка-эмитента
    provider_card_id VARCHAR(100) NOT NULL,
    bin VARCHAR(6) NOT NULL DEFAULT '424242',
    last_4_digits VARCHAR(4) NOT NULL,
    card_status VARCHAR(50) DEFAULT 'ACTIVE',
    nickname VARCHAR(100),
    -- Лимит для контроля спенда (анти-фрод)
    daily_spend_limit NUMERIC(20, 4) DEFAULT 1000.0000,
    -- Счетчик неудачных авторизаций (анти-фрод)
    failed_auth_count INTEGER DEFAULT 0,
    -- Тип карты (VISA, MasterCard)
    card_type VARCHAR(20) DEFAULT 'VISA',
    -- Автопополнение карты
    auto_replenish_enabled BOOLEAN DEFAULT FALSE,
    auto_replenish_threshold NUMERIC(20, 4) DEFAULT 0.0000,
    auto_replenish_amount NUMERIC(20, 4) DEFAULT 0.0000,
    card_balance NUMERIC(20, 4) DEFAULT 0.0000,
    service_slug VARCHAR(50) DEFAULT 'arbitrage',
    team_id INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Таблица транзакций
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    card_id INTEGER REFERENCES cards(id),
    amount NUMERIC(20, 4) NOT NULL,
    fee NUMERIC(20, 4) DEFAULT 0.0000,
    transaction_type VARCHAR(50) NOT NULL, -- 'FUND', 'AUTH', 'CAPTURE', 'DECLINE', 'REFUND'
    status VARCHAR(50) NOT NULL,
    details TEXT,
    provider_tx_id VARCHAR(255), -- ID транзакции от провайдера (Wallester) для idempotency
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индекс для быстрой проверки idempotency по provider_tx_id
CREATE INDEX IF NOT EXISTS idx_transactions_provider_tx_id ON transactions(provider_tx_id) WHERE provider_tx_id IS NOT NULL;

-- 4. API Ключи (Для трекеров)
CREATE TABLE IF NOT EXISTS api_keys (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    api_key UUID UNIQUE DEFAULT uuid_generate_v4(),
    permissions VARCHAR(50) DEFAULT 'READ_ONLY', -- READ_ONLY для Keitaro/Binom
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 5. Таблица команд
CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    owner_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 6. Таблица участников команд
CREATE TABLE IF NOT EXISTS team_members (
    id SERIAL PRIMARY KEY,
    team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'member', -- 'owner', 'admin', 'member'
    invited_by INTEGER REFERENCES users(id),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(team_id, user_id)
);

-- 7. Таблица Grade пользователей (система уровней)
CREATE TABLE IF NOT EXISTS user_grades (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    grade VARCHAR(50) DEFAULT 'STANDARD', -- 'STANDARD', 'SILVER', 'GOLD', 'PLATINUM', 'BLACK'
    total_spent NUMERIC(20, 4) DEFAULT 0.0000, -- Общая сумма трат
    fee_percent NUMERIC(5, 2) DEFAULT 6.70, -- Комиссия в процентах (6.7% для STANDARD)
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 8. Таблица реферальной программы
CREATE TABLE IF NOT EXISTS referrals (
    id SERIAL PRIMARY KEY,
    referrer_id INTEGER REFERENCES users(id) ON DELETE CASCADE, -- Кто пригласил
    referred_id INTEGER REFERENCES users(id) ON DELETE CASCADE, -- Кого пригласили
    referral_code VARCHAR(50) UNIQUE NOT NULL, -- Уникальный код реферала
    status VARCHAR(50) DEFAULT 'PENDING', -- 'PENDING', 'ACTIVE', 'COMPLETED'
    commission_earned NUMERIC(20, 4) DEFAULT 0.0000, -- Заработанная комиссия
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Обновить таблицу cards для поддержки команд (если еще не добавлено)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cards' AND column_name = 'team_id') THEN
        ALTER TABLE cards ADD COLUMN team_id INTEGER REFERENCES teams(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Индексы для скорости
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_cards_user_id ON cards(user_id);
CREATE INDEX idx_cards_card_id ON cards(id);
CREATE INDEX idx_cards_team_id ON cards(team_id);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_card_id ON transactions(card_id);
CREATE INDEX idx_teams_owner_id ON teams(owner_id);
CREATE INDEX idx_team_members_team_id ON team_members(team_id);
CREATE INDEX idx_team_members_user_id ON team_members(user_id);
CREATE INDEX idx_user_grades_user_id ON user_grades(user_id);
CREATE INDEX idx_referrals_referrer_id ON referrals(referrer_id);
CREATE INDEX idx_referrals_referred_id ON referrals(referred_id);
CREATE INDEX idx_referrals_code ON referrals(referral_code);

-- 9. Таблица внутренних балансов (Кошелёк) — один master_balance на пользователя
CREATE TABLE IF NOT EXISTS internal_balances (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    master_balance NUMERIC(20, 4) DEFAULT 0.0000 NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_internal_balances_user_id ON internal_balances(user_id);

-- Миграции XPLR: новые поля users и cards (идемпотентно)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'balance_rub') THEN
        ALTER TABLE users ADD COLUMN balance_rub NUMERIC(20, 4) DEFAULT 0.0000 NOT NULL;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'kyc_status') THEN
        ALTER TABLE users ADD COLUMN kyc_status VARCHAR(50) DEFAULT 'pending';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'active_mode') THEN
        ALTER TABLE users ADD COLUMN active_mode VARCHAR(50) DEFAULT 'personal';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cards' AND column_name = 'service_slug') THEN
        ALTER TABLE cards ADD COLUMN service_slug VARCHAR(50) DEFAULT 'arbitrage';
    END IF;
    -- Миграция для provider_tx_id в transactions (для idempotency webhook)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'transactions' AND column_name = 'provider_tx_id') THEN
        ALTER TABLE transactions ADD COLUMN provider_tx_id VARCHAR(255);
        CREATE INDEX IF NOT EXISTS idx_transactions_provider_tx_id ON transactions(provider_tx_id) WHERE provider_tx_id IS NOT NULL;
    END IF;
    -- Лимит списания карты из Кошелька (максимум, который карта может потратить)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cards' AND column_name = 'spending_limit') THEN
        ALTER TABLE cards ADD COLUMN spending_limit NUMERIC(20, 4) DEFAULT 0.0000;
    END IF;
    -- Дата истечения срока карты (для cron-возврата остатка в Кошелёк)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cards' AND column_name = 'expiry_date') THEN
        ALTER TABLE cards ADD COLUMN expiry_date TIMESTAMP WITH TIME ZONE;
    END IF;
    -- Сколько карта реально потратила из Кошелька (для отслеживания остатка)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cards' AND column_name = 'spent_from_vault') THEN
        ALTER TABLE cards ADD COLUMN spent_from_vault NUMERIC(20, 4) DEFAULT 0.0000;
    END IF;
    -- Telegram ID для интеграции с ботом
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'telegram_id') THEN
        ALTER TABLE users ADD COLUMN telegram_id BIGINT;
        CREATE UNIQUE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id) WHERE telegram_id IS NOT NULL;
    END IF;
    -- Дефолтный лимит карты по типу (устанавливается при выпуске)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cards' AND column_name = 'default_max_limit') THEN
        ALTER TABLE cards ADD COLUMN default_max_limit NUMERIC(20, 4) DEFAULT 0.0000;
    END IF;
    -- Автопополнение карт из Кошелька
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'internal_balances' AND column_name = 'auto_topup_enabled') THEN
        ALTER TABLE internal_balances ADD COLUMN auto_topup_enabled BOOLEAN DEFAULT FALSE;
    END IF;
    -- Email verification
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'is_verified') THEN
        ALTER TABLE users ADD COLUMN is_verified BOOLEAN DEFAULT FALSE;
    END IF;
    -- Rebrand: spent_from_vault → spent_from_wallet
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cards' AND column_name = 'spent_from_vault')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cards' AND column_name = 'spent_from_wallet') THEN
        ALTER TABLE cards RENAME COLUMN spent_from_vault TO spent_from_wallet;
    END IF;
    -- Auth method preference (email / telegram) — подготовка для Telegram-бот кодов
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'auth_method_preference') THEN
        ALTER TABLE users ADD COLUMN auth_method_preference VARCHAR(20) DEFAULT 'email';
    END IF;
END $$;

-- 10. Таблица токенов подтверждения email
CREATE TABLE IF NOT EXISTS verification_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_verification_tokens_token ON verification_tokens(token);
CREATE INDEX IF NOT EXISTS idx_verification_tokens_user_id ON verification_tokens(user_id);

-- 11. Таблица токенов сброса пароля
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token ON password_reset_tokens(token);

-- 12. Миграция: расширение transactions для единой истории денежных потоков
DO $$
BEGIN
    -- source_type: 'wallet_topup', 'card_transfer', 'card_charge', 'referral_bonus', 'refund', 'commission'
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'transactions' AND column_name = 'source_type') THEN
        ALTER TABLE transactions ADD COLUMN source_type VARCHAR(50) DEFAULT 'card_charge';
    END IF;
    -- source_id: ID карты или кошелька, с которым связана операция
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'transactions' AND column_name = 'source_id') THEN
        ALTER TABLE transactions ADD COLUMN source_id INTEGER;
    END IF;
    -- currency для мультивалютности
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'transactions' AND column_name = 'currency') THEN
        ALTER TABLE transactions ADD COLUMN currency VARCHAR(10) DEFAULT 'USD';
    END IF;
END $$;

-- 13. Миграция: реферальная система — индивидуальный трекинг
DO $$
BEGIN
    -- Уникальный реферальный код пользователя (генерируется при регистрации)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'referral_code') THEN
        ALTER TABLE users ADD COLUMN referral_code VARCHAR(20) UNIQUE;
    END IF;
    -- Кто пригласил (ID пользователя-реферера)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'referred_by') THEN
        ALTER TABLE users ADD COLUMN referred_by INTEGER REFERENCES users(id);
    END IF;
END $$;
CREATE INDEX IF NOT EXISTS idx_users_referral_code ON users(referral_code) WHERE referral_code IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_referred_by ON users(referred_by) WHERE referred_by IS NOT NULL;

-- 14. Таблица реферальных комиссий
CREATE TABLE IF NOT EXISTS referral_commissions (
    id SERIAL PRIMARY KEY,
    referrer_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    referred_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    source_transaction_id INTEGER REFERENCES transactions(id),
    commission_amount NUMERIC(20, 4) NOT NULL,
    commission_percent NUMERIC(5, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_referral_commissions_referrer ON referral_commissions(referrer_id);
CREATE INDEX IF NOT EXISTS idx_referral_commissions_referred ON referral_commissions(referred_id);

-- 15. Миграция: роли пользователей и verification_token
DO $$
BEGIN
    -- role: 'user', 'admin', 'support'
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'role') THEN
        ALTER TABLE users ADD COLUMN role VARCHAR(20) DEFAULT 'user';
    END IF;
    -- verification_token: прямая ссылка на токен верификации (для быстрого доступа)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'verification_token') THEN
        ALTER TABLE users ADD COLUMN verification_token VARCHAR(255);
    END IF;
    -- is_admin: обратная совместимость (если ещё не существует)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'is_admin') THEN
        ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE;
    END IF;
END $$;
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- 16. Таблица тикетов поддержки
CREATE TABLE IF NOT EXISTS support_tickets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    subject VARCHAR(500) NOT NULL,
    status VARCHAR(50) DEFAULT 'open', -- 'open', 'in_progress', 'resolved', 'closed'
    tg_chat_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_support_tickets_user_id ON support_tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_status ON support_tickets(status);

-- 17. Лог действий администраторов
CREATE TABLE IF NOT EXISTS admin_logs (
    id SERIAL PRIMARY KEY,
    admin_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_admin_logs_admin_id ON admin_logs(admin_id);

-- 18. Конфигурация комиссий (управляется через админку)
CREATE TABLE IF NOT EXISTS commission_config (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,       -- e.g. 'fee_standard', 'fee_silver', 'referral_percent'
    value NUMERIC(20, 4) NOT NULL,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
-- Seed defaults (idempotent)
INSERT INTO commission_config (key, value, description)
VALUES
    ('fee_standard', 6.70, 'Комиссия для грейда STANDARD (%)'),
    ('fee_silver',   5.50, 'Комиссия для грейда SILVER (%)'),
    ('fee_gold',     4.50, 'Комиссия для грейда GOLD (%)'),
    ('fee_platinum', 3.50, 'Комиссия для грейда PLATINUM (%)'),
    ('fee_black',    2.50, 'Комиссия для грейда BLACK (%)'),
    ('referral_percent', 5.00, 'Процент реферальной комиссии'),
    ('card_issue_fee', 2.00, 'Стоимость выпуска карты ($)')
ON CONFLICT (key) DO NOTHING;