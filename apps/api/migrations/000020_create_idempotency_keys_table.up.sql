CREATE TYPE idempotency_status AS ENUM (
    'pending',
    'completed',
    'failed'
);

CREATE TABLE idempotency_operations (
    key             TEXT PRIMARY KEY,
    operation       TEXT NOT NULL,
    request_hash    BYTEA NOT NULL,
    status          idempotency_status NOT NULL DEFAULT 'pending',
    response        BYTEA,
    error           BYTEA,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at      TIMESTAMP WITH TIME ZONE,
    expires_at      TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_idempotency_operations_key ON idempotency_operations(key);
CREATE INDEX idx_idempotency_operations_expires_at ON idempotency_operations(expires_at);
