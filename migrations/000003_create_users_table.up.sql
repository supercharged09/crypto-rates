CREATE TABLE IF NOT EXISTS users (
                                     id BIGSERIAL PRIMARY KEY,
                                     chat_id BIGINT NOT NULL UNIQUE,
                                     username VARCHAR(255) DEFAULT '',
    first_name VARCHAR(255) DEFAULT '',
    last_name VARCHAR(255) DEFAULT '',
    language_code VARCHAR(10) DEFAULT '',
    first_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    command_count INTEGER NOT NULL DEFAULT 0
    );