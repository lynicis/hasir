package idempotency

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestIdempotencyInterceptor_MissingHeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
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
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect.Error, got %v", err)
	}

	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Errorf("expected CodeInvalidArgument, got %v", connectErr.Code())
	}
}

func TestIdempotencyInterceptor_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
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
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestIdempotencyInterceptor_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
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
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect.Error, got %v", err)
	}

	if connectErr.Code() != connect.CodeAlreadyExists {
		t.Errorf("expected CodeAlreadyExists, got %v", connectErr.Code())
	}
}

func TestIdempotencyInterceptor_NonIdempotent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
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
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestIdempotencyInterceptor_IsClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
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
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestIdempotencyInterceptor_RepoIsKeyExistsErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
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
	if err == nil || err.Error() != "unknown: db error" {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestIdempotencyInterceptor_RepoClaimErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
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
	if err == nil || connect.CodeOf(err) != connect.CodeInternal {
		t.Fatalf("expected CodeInternal, got %v", err)
	}
}

func TestIdempotencyInterceptor_RepoEndErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
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
	if err == nil || connect.CodeOf(err) != connect.CodeInternal {
		t.Fatalf("expected CodeInternal, got %v", err)
	}
}

func TestIdempotencyInterceptor_HandlerErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
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
	if err == nil || err.Error() != "unknown: handler error" {
		t.Fatalf("expected handler error, got %v", err)
	}
}
