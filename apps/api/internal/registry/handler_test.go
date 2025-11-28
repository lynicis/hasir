package registry

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/registry/v1/registryv1connect"
	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
)

func TestNewHandler(t *testing.T) {
	t.Run("creates handler with service and repository", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		h := NewHandler(mockService, mockRepository)

		require.NotNil(t, h)
		require.Equal(t, mockService, h.service)
		require.Equal(t, mockRepository, h.repository)
		require.Empty(t, h.interceptors)
	})

	t.Run("creates handler with interceptors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		interceptor1 := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
			return next
		})
		interceptor2 := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
			return next
		})

		h := NewHandler(mockService, mockRepository, interceptor1, interceptor2)

		require.NotNil(t, h)
		require.Len(t, h.interceptors, 2)
	})
}

func TestHandler_RegisterRoutes(t *testing.T) {
	t.Run("returns path and handler", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		h := NewHandler(mockService, mockRepository)
		path, httpHandler := h.RegisterRoutes()

		require.Equal(t, "/"+registryv1connect.RegistryServiceName+"/", path)
		require.NotNil(t, httpHandler)
	})
}

func TestHandler_CreateRepository(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			CreateRepository(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, req *registryv1.CreateRepositoryRequest) error {
				require.Equal(t, "test-repo", req.GetName())
				return nil
			})

		h := NewHandler(mockService, mockRepository)
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := registryv1connect.NewRegistryServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.CreateRepository(context.Background(), connect.NewRequest(&registryv1.CreateRepositoryRequest{
			Name: "test-repo",
		}))
		require.NoError(t, err)
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			CreateRepository(gomock.Any(), gomock.Any()).
			Return(connect.NewError(connect.CodeAlreadyExists, errors.New("repository already exists")))

		h := NewHandler(mockService, mockRepository)
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := registryv1connect.NewRegistryServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.CreateRepository(context.Background(), connect.NewRequest(&registryv1.CreateRepositoryRequest{
			Name: "existing-repo",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeAlreadyExists, connectErr.Code())
	})
}

func TestHandler_GetRepositories(t *testing.T) {
	t.Run("success with repositories", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		repos := &[]RepositoryDTO{
			{Id: "repo-1", Name: "first-repo"},
			{Id: "repo-2", Name: "second-repo"},
		}

		mockRepository.EXPECT().
			GetRepositoriesCount(gomock.Any()).
			Return(2, nil)
		mockRepository.EXPECT().
			GetRepositories(gomock.Any(), 1, 10).
			Return(repos, nil)

		h := NewHandler(mockService, mockRepository)
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := registryv1connect.NewRegistryServiceClient(
			http.DefaultClient,
			server.URL,
		)

		resp, err := client.GetRepositories(context.Background(), connect.NewRequest(&registryv1.GetRepositoriesRequest{}))
		require.NoError(t, err)
		require.Len(t, resp.Msg.GetRepositories(), 2)
		require.Equal(t, "repo-1", resp.Msg.GetRepositories()[0].GetId())
		require.Equal(t, "first-repo", resp.Msg.GetRepositories()[0].GetName())
		require.Equal(t, "repo-2", resp.Msg.GetRepositories()[1].GetId())
		require.Equal(t, "second-repo", resp.Msg.GetRepositories()[1].GetName())
	})

	t.Run("success with empty repositories", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		repos := &[]RepositoryDTO{}

		mockRepository.EXPECT().
			GetRepositoriesCount(gomock.Any()).
			Return(2, nil)
		mockRepository.EXPECT().
			GetRepositories(gomock.Any(), 1, 10).
			Return(repos, nil)

		h := NewHandler(mockService, mockRepository)
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := registryv1connect.NewRegistryServiceClient(
			http.DefaultClient,
			server.URL,
		)

		resp, err := client.GetRepositories(context.Background(), connect.NewRequest(&registryv1.GetRepositoriesRequest{}))
		require.NoError(t, err)
		require.Empty(t, resp.Msg.GetRepositories())
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockRepository.EXPECT().
			GetRepositoriesCount(gomock.Any()).
			Return(0, connect.NewError(connect.CodeInternal, errors.New("database error")))

		h := NewHandler(mockService, mockRepository)
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := registryv1connect.NewRegistryServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.GetRepositories(context.Background(), connect.NewRequest(&registryv1.GetRepositoriesRequest{}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeInternal, connectErr.Code())
	})
}
