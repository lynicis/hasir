DROP INDEX IF EXISTS idx_idempotency_operations_key;
DROP INDEX IF EXISTS idx_idempotency_operations_expires_at;
DROP TABLE IF EXISTS idempotency_operations;
DROP TYPE IF EXISTS idempotency_status;
