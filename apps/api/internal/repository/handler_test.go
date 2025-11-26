package repository

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/repository/v1/repositoryv1connect"
	repositoryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/repository/v1"
)

func TestNewHandler(t *testing.T) {
	t.Run("creates handler with service and interceptors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)

		h := NewHandler(mockService)

		require.NotNil(t, h)
		require.Equal(t, mockService, h.service)
		require.Empty(t, h.interceptors)
	})

	t.Run("creates handler with interceptors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)

		interceptor1 := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
			return next
		})
		interceptor2 := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
			return next
		})

		h := NewHandler(mockService, interceptor1, interceptor2)

		require.NotNil(t, h)
		require.Len(t, h.interceptors, 2)
	})
}

func TestHandler_RegisterRoutes(t *testing.T) {
	t.Run("returns path and handler", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)

		h := NewHandler(mockService)
		path, httpHandler := h.RegisterRoutes()

		require.Equal(t, "/"+repositoryv1connect.RepositoryServiceName+"/", path)
		require.NotNil(t, httpHandler)
	})
}

func TestHandler_CreateRepository(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)

		mockService.EXPECT().
			CreateRepository(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, req *repositoryv1.CreateRepositoryRequest) error {
				require.Equal(t, "test-repo", req.GetName())
				return nil
			})

		h := NewHandler(mockService)
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := repositoryv1connect.NewRepositoryServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.CreateRepository(context.Background(), connect.NewRequest(&repositoryv1.CreateRepositoryRequest{
			Name: "test-repo",
		}))
		require.NoError(t, err)
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)

		mockService.EXPECT().
			CreateRepository(gomock.Any(), gomock.Any()).
			Return(connect.NewError(connect.CodeAlreadyExists, errors.New("repository already exists")))

		h := NewHandler(mockService)
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := repositoryv1connect.NewRepositoryServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.CreateRepository(context.Background(), connect.NewRequest(&repositoryv1.CreateRepositoryRequest{
			Name: "existing-repo",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeAlreadyExists, connectErr.Code())
	})
}

