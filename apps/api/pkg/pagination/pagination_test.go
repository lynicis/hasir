package pagination

import (
	"math"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hasir/proto/gen/go/shared"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name         string
		req          *shared.Pagination
		wantPage     int
		wantPageSize int
	}{
		{
			name:         "nil request returns defaults",
			req:          nil,
			wantPage:     DefaultPage,
			wantPageSize: DefaultPageSize,
		},
		{
			name:         "zero and negative values return defaults",
			req:          &shared.Pagination{Page: -2, PageLimit: 0},
			wantPage:     DefaultPage,
			wantPageSize: DefaultPageSize,
		},
		{
			name:         "page limit exceeding max is clamped",
			req:          &shared.Pagination{Page: 3, PageLimit: 500},
			wantPage:     3,
			wantPageSize: MaxPageSize,
		},
		{
			name:         "valid values pass through",
			req:          &shared.Pagination{Page: 2, PageLimit: 25},
			wantPage:     2,
			wantPageSize: 25,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			page, pageSize := Normalize(tc.req)
			assert.Equal(t, tc.wantPage, page)
			assert.Equal(t, tc.wantPageSize, pageSize)
		})
	}
}

func TestClamp(t *testing.T) {
	page, size := Clamp(0, 200)
	assert.Equal(t, DefaultPage, page)
	assert.Equal(t, MaxPageSize, size)

	page, size = Clamp(4, -5)
	assert.Equal(t, 4, page)
	assert.Equal(t, DefaultPageSize, size)
}

func TestComputePageMeta(t *testing.T) {
	t.Run("zero items yields one total page and zero next page", func(t *testing.T) {
		next, total, err := ComputePageMeta(0, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int32(0), next)
		assert.Equal(t, int32(1), total)
	})

	t.Run("middle page computes next page", func(t *testing.T) {
		next, total, err := ComputePageMeta(25, 2, 10)
		require.NoError(t, err)
		assert.Equal(t, int32(3), next)
		assert.Equal(t, int32(3), total)
	})

	t.Run("last page has zero next page", func(t *testing.T) {
		next, total, err := ComputePageMeta(25, 3, 10)
		require.NoError(t, err)
		assert.Equal(t, int32(0), next)
		assert.Equal(t, int32(3), total)
	})

	t.Run("invalid pageSize falls back to default", func(t *testing.T) {
		next, total, err := ComputePageMeta(25, 1, 0)
		require.NoError(t, err)
		assert.Equal(t, int32(2), next)
		assert.Equal(t, int32(3), total)
	})

	t.Run("overflows return internal error", func(t *testing.T) {
		_, _, err := ComputePageMeta(math.MaxInt, 1, 1)
		require.Error(t, err)
		var connectErr *connect.Error
		require.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeInternal, connectErr.Code())
	})
}
