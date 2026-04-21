-- +goose Up
-- +goose StatementBegin

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

CREATE INDEX idx_news_created ON news(created_at DESC);
CREATE INDEX idx_news_status ON news(status);

CREATE TABLE admin_logs (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    admin_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_admin_logs_admin_id ON admin_logs(admin_id);
CREATE INDEX idx_admin_logs_created ON admin_logs(created_at DESC);

CREATE TABLE system_settings (
    setting_key VARCHAR(100) PRIMARY KEY,
    setting_value TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'last_read_news_at'
    ) THEN
        ALTER TABLE users ADD COLUMN last_read_news_at TIMESTAMPTZ;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'news_notifications_enabled'
    ) THEN
        ALTER TABLE users ADD COLUMN news_notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE;
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE users DROP COLUMN IF EXISTS last_read_news_at;
ALTER TABLE users DROP COLUMN IF EXISTS news_notifications_enabled;

DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS admin_logs;
DROP TABLE IF EXISTS news;

-- +goose StatementEnd
