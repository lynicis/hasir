package idempotency

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"hasir/api/pkg/config"
)

var (
	ErrFailedAcquireConnection      = connect.NewError(connect.CodeInternal, errors.New("failed to acquire connection"))
	ErrIdempotencyOperationNotFound = connect.NewError(connect.CodeNotFound, errors.New("no idempotency operation found"))
)

type OperationStatus string

var (
	OperationPending   OperationStatus = "pending"
	OperationCompleted OperationStatus = "completed"
	OperationFailed    OperationStatus = "failed"
)

type IdempotencyRepository struct {
	connectionPool *pgxpool.Pool
	tracer         trace.Tracer
}

func NewIdempotencyRepository(
	cfg *config.Config,
	traceProvider *sdktrace.TracerProvider,
) *IdempotencyRepository {
	credential := cfg.PostgresConfig.GetPostgresDsn()
	pgConfig, err := pgxpool.ParseConfig(credential)
	if err != nil {
		zap.L().Fatal("failed to parse database config", zap.Error(err))
	}

	if traceProvider != nil {
		pgConfig.ConnConfig.Tracer = otelpgx.NewTracer(
			otelpgx.WithTracerProvider(traceProvider),
			otelpgx.WithDisableConnectionDetailsInAttributes(),
		)
	}

	var pgConnectionPool *pgxpool.Pool
	pgConnectionPool, err = pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		zap.L().Fatal("failed to connect database", zap.Error(err))
	}

	if traceProvider != nil {
		if err := otelpgx.RecordStats(pgConnectionPool); err != nil {
			zap.L().Fatal("unable to record database stats", zap.Error(err))
		}
	}

	var connection *pgxpool.Conn
	connection, err = pgConnectionPool.Acquire(context.Background())
	if err != nil {
		zap.L().Fatal("failed to acquire connection", zap.Error(err))
	}
	defer connection.Release()

	err = connection.Ping(context.Background())
	if err != nil {
		zap.L().Fatal("failed to ping database", zap.Error(err))
	}

	var tracer trace.Tracer
	if traceProvider != nil {
		tracer = traceProvider.Tracer("IdempotencyPostgreSQLRepository")
	} else {
		tracer = noop.NewTracerProvider().Tracer("IdempotencyPostgreSQLRepository")
	}

	return &IdempotencyRepository{
		connectionPool: pgConnectionPool,
		tracer:         tracer,
	}
}

func (r *IdempotencyRepository) IsIdempotencyKeyExists(ctx context.Context, key string) (error, bool) {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "IsIdempotencyKeyExists", trace.WithAttributes(attribute.KeyValue{
		Key:   "idempotencyKey",
		Value: attribute.StringValue(key),
	}))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection, false
	}
	defer connection.Release()

	sql := "SELECT EXISTS (SELECT 1 FROM idempotency_operations WHERE key = $1) AS exists"

	var exists bool
	err = connection.QueryRow(ctx, sql, key).Scan(&exists)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(
			connect.CodeInternal,
			errors.New("failed to check idempotency key"),
		), false
	}

	return nil, exists
}

func (r *IdempotencyRepository) ClaimIdempotencyKey(ctx context.Context, key, operation, requestHash string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "ClaimIdempotencyKey", trace.WithAttributes(attribute.KeyValue{
		Key:   "idempotencyKey",
		Value: attribute.StringValue(key),
	}))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	sql := `
		INSERT INTO idempotency_operations (key, operation, request_hash, status, created_at, expires_at)
		VALUES (@Key, @Operation, @RequestHash, @Status, @CreatedAt, @ExpiresAt)`

	now := time.Now().UTC()
	sqlArgs := pgx.NamedArgs{
		"Key":         key,
		"Operation":   operation,
		"RequestHash": []byte(requestHash),
		"Status":      OperationPending,
		"CreatedAt":   now,
		"ExpiresAt":   time.Now().UTC().Add(10 * time.Minute),
	}

	if _, err = connection.Exec(ctx, sql, sqlArgs); err != nil {
		span.RecordError(err)
		return connect.NewError(
			connect.CodeInternal,
			errors.New("failed to execute insert idempotency operation query"),
		)
	}

	return nil
}

func (r *IdempotencyRepository) EndOperation(ctx context.Context, key, response, errStr string) error {
	var span trace.Span
	ctx, span = r.tracer.Start(ctx, "EndOperation", trace.WithAttributes(attribute.KeyValue{
		Key:   "idempotencyKey",
		Value: attribute.StringValue(key),
	}))
	defer span.End()

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return ErrFailedAcquireConnection
	}
	defer connection.Release()

	status := OperationCompleted
	if errStr != "" {
		status = OperationFailed
	}

	var responseBytes, errorBytes []byte
	if response != "" {
		responseBytes = []byte(response)
	}
	if errStr != "" {
		errorBytes = []byte(errStr)
	}

	sql := `
		UPDATE idempotency_operations
		SET status = @Status,
			response = @Response,
			error = @Error,
			updated_at = @UpdatedAt
		WHERE key = @Key`
	sqlArgs := pgx.NamedArgs{
		"Key":       key,
		"Status":    status,
		"Response":  responseBytes,
		"Error":     errorBytes,
		"UpdatedAt": time.Now().UTC(),
	}

	result, err := connection.Exec(ctx, sql, sqlArgs)
	if err != nil {
		span.RecordError(err)
		return connect.NewError(
			connect.CodeInternal,
			errors.New("failed to execute update idempotency operation query"),
		)
	}

	if result.RowsAffected() == 0 {
		return ErrIdempotencyOperationNotFound
	}

	return nil
}
