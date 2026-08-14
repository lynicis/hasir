package csrf

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

type testRequest struct {
	*connect.Request[emptypb.Empty]
	isClient bool
}

func (t *testRequest) Spec() connect.Spec {
	spec := t.Request.Spec()
	spec.IsClient = t.isClient
	return spec
}

func TestCsrfInterceptor_Interceptor(t *testing.T) {
	interceptor := NewCsrfInterceptor("http://dashboard.local", "https://api.local")

	nextFunc := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		return connect.NewResponse(new(emptypb.Empty)), nil
	}
	wrappedFunc := interceptor.Interceptor()(nextFunc)

	t.Run("skips check for client requests", func(t *testing.T) {
		realReq := connect.NewRequest(new(emptypb.Empty))
		// Intentionally add a bad origin to prove it skips the check
		realReq.Header().Set("Origin", "http://evil.com")
		req := &testRequest{
			Request:  realReq,
			isClient: true,
		}

		_, err := wrappedFunc(context.Background(), req)
		require.NoError(t, err)
	})

	t.Run("allows requests without Origin header (e.g., CLI)", func(t *testing.T) {
		realReq := connect.NewRequest(new(emptypb.Empty))
		req := &testRequest{
			Request:  realReq,
			isClient: false,
		}

		_, err := wrappedFunc(context.Background(), req)
		require.NoError(t, err)
	})

	t.Run("rejects requests from unallowed origins", func(t *testing.T) {
		realReq := connect.NewRequest(new(emptypb.Empty))
		realReq.Header().Set("Origin", "http://evil.com")
		req := &testRequest{
			Request:  realReq,
			isClient: false,
		}

		_, err := wrappedFunc(context.Background(), req)
		require.Error(t, err)
		var connectErr *connect.Error
		require.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodePermissionDenied, connectErr.Code())
		assert.Contains(t, connectErr.Message(), "cross-origin request not allowed")
	})

	t.Run("rejects requests from allowed origins without CSRF token", func(t *testing.T) {
		realReq := connect.NewRequest(new(emptypb.Empty))
		realReq.Header().Set("Origin", "http://dashboard.local")
		req := &testRequest{
			Request:  realReq,
			isClient: false,
		}

		_, err := wrappedFunc(context.Background(), req)
		require.Error(t, err)
		var connectErr *connect.Error
		require.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodePermissionDenied, connectErr.Code())
		assert.Contains(t, connectErr.Message(), "missing X-CSRF-Token header")
	})

	t.Run("allows requests from allowed origins with CSRF token", func(t *testing.T) {
		realReq := connect.NewRequest(new(emptypb.Empty))
		realReq.Header().Set("Origin", "http://dashboard.local")
		realReq.Header().Set("X-CSRF-Token", "some-token")
		req := &testRequest{
			Request:  realReq,
			isClient: false,
		}

		_, err := wrappedFunc(context.Background(), req)
		require.NoError(t, err)
	})

	t.Run("handles origins with trailing slashes gracefully", func(t *testing.T) {
		// NewCsrfInterceptor should strip trailing slashes
		strictInterceptor := NewCsrfInterceptor("http://dashboard.local/")
		wrappedStrict := strictInterceptor.Interceptor()(nextFunc)

		realReq := connect.NewRequest(new(emptypb.Empty))
		realReq.Header().Set("Origin", "http://dashboard.local") // Browser won't send trailing slash
		realReq.Header().Set("X-CSRF-Token", "some-token")
		req := &testRequest{
			Request:  realReq,
			isClient: false,
		}

		_, err := wrappedStrict(context.Background(), req)
		require.NoError(t, err)
	})
}
