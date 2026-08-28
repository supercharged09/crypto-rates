CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    cryptocurrency VARCHAR(10) NOT NULL DEFAULT 'all',
    interval_min INTEGER NOT NULL DEFAULT 10,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(chat_id, cryptocurrency)
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_active ON subscriptions(is_active);
