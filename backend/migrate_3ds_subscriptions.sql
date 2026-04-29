-- ═══════════════════════════════════════════════════════════
-- Migration: 3DS Hub + Subscription Control
-- Run: psql $DATABASE_URL -f backend/migrate_3ds_subscriptions.sql
-- ═══════════════════════════════════════════════════════════

-- 1. Auto-pay toggle on cards (default TRUE = all recurring allowed)
ALTER TABLE cards ADD COLUMN IF NOT EXISTS is_auto_pay_enabled BOOLEAN DEFAULT TRUE;

-- 2. Merchant blocks (per-card merchant blocking)
CREATE TABLE IF NOT EXISTS merchant_blocks (
    id SERIAL PRIMARY KEY,
    card_id INTEGER REFERENCES cards(id) ON DELETE CASCADE,
    merchant_name VARCHAR(500) NOT NULL,
    is_blocked BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_merchant_blocks_card_merchant ON merchant_blocks(card_id, merchant_name);
ALTER TABLE merchant_blocks DISABLE ROW LEVEL SECURITY;

-- 3. Card subscriptions (auto-tracked merchants on charge)
CREATE TABLE IF NOT EXISTS card_subscriptions (
    id SERIAL PRIMARY KEY,
    card_id INTEGER REFERENCES cards(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    merchant_name VARCHAR(500) NOT NULL,
    last_amount NUMERIC(20, 4),
    last_currency VARCHAR(10) DEFAULT 'USD',
    charge_count INTEGER DEFAULT 1,
    is_allowed BOOLEAN DEFAULT TRUE,
    first_seen_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_card_subs_card_merchant ON card_subscriptions(card_id, merchant_name);
CREATE INDEX IF NOT EXISTS idx_card_subs_user ON card_subscriptions(user_id);
ALTER TABLE card_subscriptions DISABLE ROW LEVEL SECURITY;

-- 4. SMS/3DS codes (hub storage + delivery tracking)
CREATE TABLE IF NOT EXISTS sms_codes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    card_id INTEGER,
    code VARCHAR(10) NOT NULL,
    merchant_name VARCHAR(500),
    raw_message TEXT,
    delivered_ws BOOLEAN DEFAULT FALSE,
    delivered_tg BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sms_codes_user_created ON sms_codes(user_id, created_at DESC);
ALTER TABLE sms_codes DISABLE ROW LEVEL SECURITY;

-- Verify
DO $$
BEGIN
    RAISE NOTICE '✅ Migration complete. Tables: merchant_blocks, card_subscriptions, sms_codes. Column: cards.is_auto_pay_enabled';
END $$;
