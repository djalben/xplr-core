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
END $$;