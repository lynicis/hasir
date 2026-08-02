package idempotency

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"hasir/api/pkg/config"
)

const (
	pgDb       = "test"
	pgUsername = "test"
	pgPassword = "test"
)

func setupPgContainer(t *testing.T) *postgres.PostgresContainer {
	ctx := context.Background()

	// Use the same migrations dir as the real app
	migrationsDir := filepath.Join("..", "..", "..", "migrations")

	postgresContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase(pgDb),
		postgres.WithUsername(pgUsername),
		postgres.WithPassword(pgPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
		postgres.WithInitScripts(filepath.Join(migrationsDir, "000020_create_idempotency_keys_table.up.sql")),
	)
	require.NoError(t, err)

	return postgresContainer
}

func TestIdempotencyRepository(t *testing.T) {
	container := setupPgContainer(t)
	defer func() {
		err := container.Terminate(context.Background())
		require.NoError(t, err)
	}()

	connString, err := container.ConnectionString(context.Background())
	require.NoError(t, err)

	traceProvider := sdktrace.NewTracerProvider()

	repo := NewIdempotencyRepository(&config.Config{
		PostgresConfig: config.PostgresConfig{
			ConnectionString: connString,
		},
	}, traceProvider)

	ctx := context.Background()

	t.Run("IsIdempotencyKeyExists on new key", func(t *testing.T) {
		err, exists := repo.IsIdempotencyKeyExists(ctx, "key1")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("ClaimIdempotencyKey", func(t *testing.T) {
		err := repo.ClaimIdempotencyKey(ctx, "key1", "/TestService/Method", "requestHash123")
		assert.NoError(t, err)

		err, exists := repo.IsIdempotencyKeyExists(ctx, "key1")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("EndOperation", func(t *testing.T) {
		err := repo.EndOperation(ctx, "key1", "responseHash456", "")
		assert.NoError(t, err)
		// We could verify the DB state here, but since the interface doesn't expose a Get,
		// verifying IsIdempotencyKeyExists is still true is enough for the interface contract.
		err, exists := repo.IsIdempotencyKeyExists(ctx, "key1")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("EndOperation with error", func(t *testing.T) {
		err := repo.ClaimIdempotencyKey(ctx, "key2", "/TestService/Method", "reqHash")
		assert.NoError(t, err)

		err = repo.EndOperation(ctx, "key2", "", "some internal error")
		assert.NoError(t, err)
	})

	t.Run("EndOperation non-existent", func(t *testing.T) {
		err := repo.EndOperation(ctx, "non-existent", "", "")
		assert.ErrorIs(t, err, ErrIdempotencyOperationNotFound)
	})

	t.Run("IsIdempotencyKeyExists context cancelled", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately
		err, _ := repo.IsIdempotencyKeyExists(cancelledCtx, "key1")
		assert.Error(t, err)
	})

	t.Run("ClaimIdempotencyKey context cancelled", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		err := repo.ClaimIdempotencyKey(cancelledCtx, "key1", "/Test/Op", "hash")
		assert.Error(t, err)
	})

	t.Run("EndOperation context cancelled", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		err := repo.EndOperation(cancelledCtx, "key1", "response", "")
		assert.Error(t, err)
	})

	t.Run("ClaimIdempotencyKey duplicate error", func(t *testing.T) {
		// key1 was already claimed in a previous test step
		err := repo.ClaimIdempotencyKey(ctx, "key1", "/Test/Op", "hash")
		assert.Error(t, err)
	})
}
