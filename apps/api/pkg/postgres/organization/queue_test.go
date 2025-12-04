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
