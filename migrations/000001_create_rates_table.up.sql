CREATE TABLE IF NOT EXISTS rates (
    id BIGSERIAL PRIMARY KEY,
    cryptocurrency VARCHAR(20) NOT NULL,
    price_usd DECIMAL(18, 8) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rates_crypto_time ON rates(cryptocurrency, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_rates_timestamp ON rates(timestamp);

