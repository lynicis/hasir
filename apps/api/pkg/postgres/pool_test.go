package postgres

import (
	"context"
	"errors"
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

type fakePoolConn struct {
	pingErr error
	onRel   func()
}

func (c *fakePoolConn) Ping(_ context.Context) error { return c.pingErr }
func (c *fakePoolConn) Release()                     { c.onRel() }

type fakePoolVerifier struct {
	acquireErr error
	pingErr    error
	released   bool
	closed     bool
	events     []string
}

func (p *fakePoolVerifier) Acquire(_ context.Context) (poolConn, error) {
	if p.acquireErr != nil {
		return nil, p.acquireErr
	}
	return &fakePoolConn{
		pingErr: p.pingErr,
		onRel: func() {
			p.released = true
			p.events = append(p.events, "release")
		},
	}, nil
}

func (p *fakePoolVerifier) Close() {
	p.closed = true
	p.events = append(p.events, "close")
}

func TestVerifyPoolConnection(t *testing.T) {
	t.Run("releases connection before closing pool when ping fails", func(t *testing.T) {
		pool := &fakePoolVerifier{pingErr: errors.New("ping timeout")}

		err := verifyPoolConnection(context.Background(), pool)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to ping database")
		assert.True(t, pool.released)
		assert.True(t, pool.closed)
		assert.Equal(t, []string{"release", "close"}, pool.events)
	})

	t.Run("releases connection and keeps pool open when ping succeeds", func(t *testing.T) {
		pool := &fakePoolVerifier{}

		err := verifyPoolConnection(context.Background(), pool)
		require.NoError(t, err)
		assert.True(t, pool.released)
		assert.False(t, pool.closed)
		assert.Equal(t, []string{"release"}, pool.events)
	})
}
