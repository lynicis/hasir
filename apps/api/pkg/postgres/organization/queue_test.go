package organization

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestNewEmailJobQueue(t *testing.T) {
	t.Run("creates queue with valid parameters", func(t *testing.T) {
		tracer := noop.NewTracerProvider().Tracer("test")
		queue := NewEmailJobQueue(nil, tracer)

		assert.NotNil(t, queue)
		assert.NotNil(t, queue.tracer)
		assert.NotNil(t, queue.stopChan)
	})
}

// Note: Additional tests for queue functionality (Start, Stop, processEmailJobs, etc.)
// should be implemented as integration tests using testcontainers with a real PostgreSQL
// database, since the queue now implements database operations directly and cannot be
// easily mocked.

// TODO: Add integration tests for:
// - Start and Stop functionality
// - processEmailJobs with various scenarios
// - EnqueueEmailJobs
// - GetPendingEmailJobs
// - UpdateEmailJobStatus
