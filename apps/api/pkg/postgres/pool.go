package postgres

import (
	"context"
	"fmt"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"hasir/api/pkg/config"
)

func NewPool(ctx context.Context, cfg *config.Config, traceProvider *sdktrace.TracerProvider) (*pgxpool.Pool, error) {
	if cfg == nil {
		return nil, fmt.Errorf("postgres config cannot be nil")
	}

	credential := cfg.PostgresConfig.GetPostgresDsn()
	pgConfig, err := pgxpool.ParseConfig(credential)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	if traceProvider != nil {
		pgConfig.ConnConfig.Tracer = otelpgx.NewTracer(
			otelpgx.WithTracerProvider(traceProvider),
			otelpgx.WithDisableConnectionDetailsInAttributes(),
		)
	}

	pgConnectionPool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	if traceProvider != nil {
		if err := otelpgx.RecordStats(pgConnectionPool); err != nil {
			pgConnectionPool.Close()
			return nil, fmt.Errorf("unable to record database stats: %w", err)
		}
	}

	if err := verifyPoolConnection(ctx, pgxPoolAdapter{pool: pgConnectionPool}); err != nil {
		return nil, err
	}

	return pgConnectionPool, nil
}

type poolConn interface {
	Ping(ctx context.Context) error
	Release()
}

type poolVerifier interface {
	Acquire(ctx context.Context) (poolConn, error)
	Close()
}

type pgxPoolAdapter struct {
	pool *pgxpool.Pool
}

func (a pgxPoolAdapter) Acquire(ctx context.Context) (poolConn, error) {
	return a.pool.Acquire(ctx)
}

func (a pgxPoolAdapter) Close() {
	a.pool.Close()
}

func verifyPoolConnection(ctx context.Context, pool poolVerifier) error {
	connection, err := pool.Acquire(ctx)
	if err != nil {
		pool.Close()
		return fmt.Errorf("failed to acquire connection: %w", err)
	}

	pingErr := connection.Ping(ctx)
	connection.Release()
	if pingErr != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", pingErr)
	}

	return nil
}

func MustNewPool(cfg *config.Config, traceProvider *sdktrace.TracerProvider) *pgxpool.Pool {
	pool, err := NewPool(context.Background(), cfg, traceProvider)
	if err != nil {
		zap.L().Fatal("failed to initialize database pool", zap.Error(err))
	}
	return pool
}

func NewTracer(traceProvider *sdktrace.TracerProvider, instrumentationName string) trace.Tracer {
	if traceProvider != nil {
		return traceProvider.Tracer(instrumentationName)
	}
	return noop.NewTracerProvider().Tracer(instrumentationName)
}
