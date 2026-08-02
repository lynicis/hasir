package idempotency

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestIdempotencyInterceptor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("happy path", func(t *testing.T) {
		repo := NewMockIdempotencyRepository(ctrl)
		repo.EXPECT().IsIdempotencyKeyExists(gomock.Any(), "my-key").Return(nil, false)
		repo.EXPECT().ClaimIdempotencyKey(gomock.Any(), "my-key", "/TestService/IdempotentMethod", gomock.Any()).Return(nil)
		repo.EXPECT().EndOperation(gomock.Any(), "my-key", gomock.Any(), "").Return(nil)

		interceptor := NewIdempotencyInterceptor(repo)

		handler := connect.NewUnaryHandler(
			"/TestService/IdempotentMethod",
			func(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
				return connect.NewResponse[emptypb.Empty](&emptypb.Empty{}), nil
			},
			connect.WithIdempotency(connect.IdempotencyIdempotent),
			connect.WithInterceptors(interceptor.Interceptor()),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		client := connect.NewClient[emptypb.Empty, emptypb.Empty](
			http.DefaultClient,
			server.URL+"/TestService/IdempotentMethod",
		)

		req := connect.NewRequest[emptypb.Empty](&emptypb.Empty{})
		req.Header().Set("Idempotency-Key", "my-key")

		_, err := client.CallUnary(context.Background(), req)

		assert.NoError(t, err)
	})

	t.Run("missing header", func(t *testing.T) {
		repo := NewMockIdempotencyRepository(ctrl)
		interceptor := NewIdempotencyInterceptor(repo)

		handler := connect.NewUnaryHandler(
			"/TestService/IdempotentMethod",
			func(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
				return connect.NewResponse[emptypb.Empty](&emptypb.Empty{}), nil
			},
			connect.WithIdempotency(connect.IdempotencyIdempotent),
			connect.WithInterceptors(interceptor.Interceptor()),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		client := connect.NewClient[emptypb.Empty, emptypb.Empty](
			http.DefaultClient,
			server.URL+"/TestService/IdempotentMethod",
		)

		req := connect.NewRequest[emptypb.Empty](&emptypb.Empty{})
		// NO header set

		_, err := client.CallUnary(context.Background(), req)

		assert.Error(t, err)
		var connectErr *connect.Error
		assert.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
	})

	t.Run("already processed", func(t *testing.T) {
		repo := NewMockIdempotencyRepository(ctrl)
		repo.EXPECT().IsIdempotencyKeyExists(gomock.Any(), "existing-key").Return(nil, true)

		interceptor := NewIdempotencyInterceptor(repo)

		handler := connect.NewUnaryHandler(
			"/TestService/IdempotentMethod",
			func(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
				return connect.NewResponse[emptypb.Empty](&emptypb.Empty{}), nil
			},
			connect.WithIdempotency(connect.IdempotencyIdempotent),
			connect.WithInterceptors(interceptor.Interceptor()),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		client := connect.NewClient[emptypb.Empty, emptypb.Empty](
			http.DefaultClient,
			server.URL+"/TestService/IdempotentMethod",
		)

		req := connect.NewRequest[emptypb.Empty](&emptypb.Empty{})
		req.Header().Set("Idempotency-Key", "existing-key")

		_, err := client.CallUnary(context.Background(), req)

		assert.Error(t, err)
		var connectErr *connect.Error
		assert.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeAlreadyExists, connectErr.Code())
	})

	t.Run("non idempotent", func(t *testing.T) {
		repo := NewMockIdempotencyRepository(ctrl)
		interceptor := NewIdempotencyInterceptor(repo)

		handler := connect.NewUnaryHandler(
			"/TestService/NonIdempotentMethod",
			func(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
				return connect.NewResponse[emptypb.Empty](&emptypb.Empty{}), nil
			},
			// Missing WithIdempotency(connect.IdempotencyIdempotent)
			connect.WithInterceptors(interceptor.Interceptor()),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		client := connect.NewClient[emptypb.Empty, emptypb.Empty](
			http.DefaultClient,
			server.URL+"/TestService/NonIdempotentMethod",
		)

		req := connect.NewRequest[emptypb.Empty](&emptypb.Empty{})
		// No header set, but it shouldn't matter since it's not idempotent

		_, err := client.CallUnary(context.Background(), req)

		assert.NoError(t, err)
	})

	t.Run("is client", func(t *testing.T) {
		repo := NewMockIdempotencyRepository(ctrl)
		interceptor := NewIdempotencyInterceptor(repo)

		// We can't mock Spec().IsClient easily via NewUnaryHandler because that generates a server handler (IsClient=false).
		// But we can create a client and wrap the interceptor.
		// Actually, the easiest way to test IsClient is to create a client with the interceptor.
		handler := connect.NewUnaryHandler(
			"/TestService/IdempotentMethod",
			func(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
				return connect.NewResponse[emptypb.Empty](&emptypb.Empty{}), nil
			},
			connect.WithIdempotency(connect.IdempotencyIdempotent),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		client := connect.NewClient[emptypb.Empty, emptypb.Empty](
			http.DefaultClient,
			server.URL+"/TestService/IdempotentMethod",
			connect.WithInterceptors(interceptor.Interceptor()),
		)

		req := connect.NewRequest[emptypb.Empty](&emptypb.Empty{})
		// No idempotency key needed because it's a client request and should pass through.

		_, err := client.CallUnary(context.Background(), req)

		assert.NoError(t, err)
	})

	t.Run("repo is key exists error", func(t *testing.T) {
		repo := NewMockIdempotencyRepository(ctrl)
		repo.EXPECT().IsIdempotencyKeyExists(gomock.Any(), "my-key").Return(errors.New("db error"), false)

		interceptor := NewIdempotencyInterceptor(repo)

		handler := connect.NewUnaryHandler(
			"/TestService/IdempotentMethod",
			func(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
				return connect.NewResponse[emptypb.Empty](&emptypb.Empty{}), nil
			},
			connect.WithIdempotency(connect.IdempotencyIdempotent),
			connect.WithInterceptors(interceptor.Interceptor()),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		client := connect.NewClient[emptypb.Empty, emptypb.Empty](
			http.DefaultClient,
			server.URL+"/TestService/IdempotentMethod",
		)

		req := connect.NewRequest[emptypb.Empty](&emptypb.Empty{})
		req.Header().Set("Idempotency-Key", "my-key")

		_, err := client.CallUnary(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, "unknown: db error", err.Error())
	})

	t.Run("repo claim error", func(t *testing.T) {
		repo := NewMockIdempotencyRepository(ctrl)
		repo.EXPECT().IsIdempotencyKeyExists(gomock.Any(), "my-key").Return(nil, false)
		repo.EXPECT().ClaimIdempotencyKey(gomock.Any(), "my-key", "/TestService/IdempotentMethod", gomock.Any()).Return(errors.New("claim error"))

		interceptor := NewIdempotencyInterceptor(repo)

		handler := connect.NewUnaryHandler(
			"/TestService/IdempotentMethod",
			func(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
				return connect.NewResponse[emptypb.Empty](&emptypb.Empty{}), nil
			},
			connect.WithIdempotency(connect.IdempotencyIdempotent),
			connect.WithInterceptors(interceptor.Interceptor()),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		client := connect.NewClient[emptypb.Empty, emptypb.Empty](
			http.DefaultClient,
			server.URL+"/TestService/IdempotentMethod",
		)

		req := connect.NewRequest[emptypb.Empty](&emptypb.Empty{})
		req.Header().Set("Idempotency-Key", "my-key")

		_, err := client.CallUnary(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, connect.CodeInternal, connect.CodeOf(err))
	})

	t.Run("operation end error", func(t *testing.T) {
		repo := NewMockIdempotencyRepository(ctrl)
		repo.EXPECT().IsIdempotencyKeyExists(gomock.Any(), "my-key").Return(nil, false)
		repo.EXPECT().ClaimIdempotencyKey(gomock.Any(), "my-key", "/TestService/IdempotentMethod", gomock.Any()).Return(nil)
		repo.EXPECT().EndOperation(gomock.Any(), "my-key", gomock.Any(), "").Return(errors.New("end error"))

		interceptor := NewIdempotencyInterceptor(repo)

		handler := connect.NewUnaryHandler(
			"/TestService/IdempotentMethod",
			func(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
				return connect.NewResponse[emptypb.Empty](&emptypb.Empty{}), nil
			},
			connect.WithIdempotency(connect.IdempotencyIdempotent),
			connect.WithInterceptors(interceptor.Interceptor()),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		client := connect.NewClient[emptypb.Empty, emptypb.Empty](
			http.DefaultClient,
			server.URL+"/TestService/IdempotentMethod",
		)

		req := connect.NewRequest[emptypb.Empty](&emptypb.Empty{})
		req.Header().Set("Idempotency-Key", "my-key")

		_, err := client.CallUnary(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, connect.CodeInternal, connect.CodeOf(err))
	})

	t.Run("handler error", func(t *testing.T) {
		repo := NewMockIdempotencyRepository(ctrl)

		repo.EXPECT().IsIdempotencyKeyExists(gomock.Any(), "my-key").Return(nil, false)
		repo.EXPECT().ClaimIdempotencyKey(gomock.Any(), "my-key", "/TestService/IdempotentMethod", gomock.Any()).Return(nil)
		repo.EXPECT().EndOperation(gomock.Any(), "my-key", "", "handler error").Return(nil)

		interceptor := NewIdempotencyInterceptor(repo)

		handler := connect.NewUnaryHandler(
			"/TestService/IdempotentMethod",
			func(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
				return nil, errors.New("handler error")
			},
			connect.WithIdempotency(connect.IdempotencyIdempotent),
			connect.WithInterceptors(interceptor.Interceptor()),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		client := connect.NewClient[emptypb.Empty, emptypb.Empty](
			http.DefaultClient,
			server.URL+"/TestService/IdempotentMethod",
		)

		req := connect.NewRequest[emptypb.Empty](&emptypb.Empty{})
		req.Header().Set("Idempotency-Key", "my-key")

		_, err := client.CallUnary(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, "unknown: handler error", err.Error())
	})
}
