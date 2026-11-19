package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"hasir/api/pkg/config"
)

func TestNewTracer(t *testing.T) {
	t.Run("returns noop tracer when traceProvider is nil", func(t *testing.T) {
		tracer := NewTracer(nil, "TestRepo")
		require.NotNil(t, tracer)

		_, span := tracer.Start(context.Background(), "op")
		assert.False(t, span.SpanContext().IsValid())
		span.End()
	})

	t.Run("returns sdk tracer when traceProvider is provided", func(t *testing.T) {
		tp := sdktrace.NewTracerProvider()
		t.Cleanup(func() {
			_ = tp.Shutdown(context.Background())
		})

		tracer := NewTracer(tp, "TestRepo")
		require.NotNil(t, tracer)

		_, span := tracer.Start(context.Background(), "op")
		assert.True(t, span.SpanContext().IsValid())
		span.End()
	})
}

func TestNewPool_Errors(t *testing.T) {
	t.Run("nil config returns error", func(t *testing.T) {
		pool, err := NewPool(context.Background(), nil, nil)
		assert.Nil(t, pool)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("invalid connection string returns parse error", func(t *testing.T) {
		cfg := &config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: "postgres://[::1:invalid",
			},
		}
		pool, err := NewPool(context.Background(), cfg, nil)
		assert.Nil(t, pool)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse database config")
	})

	t.Run("unreachable host returns acquire error", func(t *testing.T) {
		cfg := &config.Config{
			PostgresConfig: config.PostgresConfig{
				ConnectionString: "postgres://user:pass@127.0.0.1:1/db?sslmode=disable",
			},
		}
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		tp := sdktrace.NewTracerProvider()
		t.Cleanup(func() {
			_ = tp.Shutdown(context.Background())
		})

		pool, err := NewPool(ctx, cfg, tp)
		assert.Nil(t, pool)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to acquire connection")
	})
}
