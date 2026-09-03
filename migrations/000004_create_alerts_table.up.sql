CREATE TABLE IF NOT EXISTS alerts (
                                      id BIGSERIAL PRIMARY KEY,
                                      chat_id BIGINT NOT NULL,
                                      cryptocurrency VARCHAR(20) NOT NULL,
    direction VARCHAR(5) NOT NULL CHECK (direction IN ('above', 'below')),
    price_threshold DECIMAL(18, 8) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    triggered_at TIMESTAMPTZ
    );

CREATE INDEX IF NOT EXISTS idx_alerts_active ON alerts(is_active);
CREATE INDEX IF NOT EXISTS idx_alerts_chat_id ON alerts(chat_id);
