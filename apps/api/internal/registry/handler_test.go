package registry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/registry/v1/registryv1connect"
	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
	"buf.build/gen/go/hasir/hasir/protocolbuffers/go/shared"

	"hasir-api/internal/user"
)

var ErrRepositoryNotFound = connect.NewError(connect.CodeNotFound, errors.New("repository not found"))

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

func TestHandler_GetRepository(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			GetRepository(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, req *registryv1.GetRepositoryRequest) (*registryv1.Repository, error) {
				require.Equal(t, "test-repo-id", req.GetId())
				return &registryv1.Repository{
					Id:   "test-repo-id",
					Name: "test-repo",
				}, nil
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

		resp, err := client.GetRepository(context.Background(), connect.NewRequest(&registryv1.GetRepositoryRequest{
			Id: "test-repo-id",
		}))
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "test-repo-id", resp.Msg.GetId())
		require.Equal(t, "test-repo", resp.Msg.GetName())
	})

	t.Run("service error - repository not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			GetRepository(gomock.Any(), gomock.Any()).
			Return(nil, ErrRepositoryNotFound)

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

		_, err := client.GetRepository(context.Background(), connect.NewRequest(&registryv1.GetRepositoryRequest{
			Id: "nonexistent-repo-id",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeNotFound, connectErr.Code())
	})

	t.Run("service error - permission denied", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			GetRepository(gomock.Any(), gomock.Any()).
			Return(nil, connect.NewError(connect.CodePermissionDenied, errors.New("you are not a member of this organization")))

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

		_, err := client.GetRepository(context.Background(), connect.NewRequest(&registryv1.GetRepositoryRequest{
			Id: "test-repo-id",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodePermissionDenied, connectErr.Code())
	})
}

func TestHandler_GetRepositories(t *testing.T) {
	t.Run("success with repositories", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			GetRepositories(gomock.Any(), 1, 10).
			Return(&registryv1.GetRepositoriesResponse{
				Repositories: []*registryv1.Repository{
					{Id: "repo-1", Name: "first-repo"},
					{Id: "repo-2", Name: "second-repo"},
				},
				NextPage:  0,
				TotalPage: 1,
			}, nil)

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

		mockService.EXPECT().
			GetRepositories(gomock.Any(), 1, 10).
			Return(&registryv1.GetRepositoriesResponse{
				Repositories: []*registryv1.Repository{},
				NextPage:     0,
				TotalPage:    1,
			}, nil)

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

		mockService.EXPECT().
			GetRepositories(gomock.Any(), 1, 10).
			Return(nil, connect.NewError(connect.CodeInternal, errors.New("database error")))

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

func TestHandler_UpdateRepository(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			UpdateRepository(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

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

		resp, err := client.UpdateRepository(context.Background(), connect.NewRequest(&registryv1.UpdateRepositoryRequest{
			Id:         "test-repo-id",
			Name:       "test-repo",
			Visibility: shared.Visibility_VISIBILITY_PRIVATE,
		}))

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			UpdateRepository(gomock.Any(), gomock.Any()).
			Return(connect.NewError(connect.CodeInternal, errors.New("database error"))).
			Times(1)

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

		resp, err := client.UpdateRepository(context.Background(), connect.NewRequest(&registryv1.UpdateRepositoryRequest{
			Id:         "test-repo-id",
			Name:       "test-repo",
			Visibility: shared.Visibility_VISIBILITY_PRIVATE,
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestHandler_DeleteRepository(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			DeleteRepository(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, req *registryv1.DeleteRepositoryRequest) error {
				require.Equal(t, "test-repo-id", req.GetRepositoryId())
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

		_, err := client.DeleteRepository(context.Background(), connect.NewRequest(&registryv1.DeleteRepositoryRequest{
			RepositoryId: "test-repo-id",
		}))
		require.NoError(t, err)
	})

	t.Run("service error - repository not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			DeleteRepository(gomock.Any(), gomock.Any()).
			Return(ErrRepositoryNotFound)

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

		_, err := client.DeleteRepository(context.Background(), connect.NewRequest(&registryv1.DeleteRepositoryRequest{
			RepositoryId: "nonexistent-repo-id",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeNotFound, connectErr.Code())
	})

	t.Run("service error - internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			DeleteRepository(gomock.Any(), gomock.Any()).
			Return(connect.NewError(connect.CodeInternal, errors.New("database error")))

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

		_, err := client.DeleteRepository(context.Background(), connect.NewRequest(&registryv1.DeleteRepositoryRequest{
			RepositoryId: "test-repo-id",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeInternal, connectErr.Code())
	})
}

func TestNewGitSshHandler(t *testing.T) {
	t.Run("creates handler with service and repos path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)

		h := NewGitSshHandler(mockService, DefaultReposPath)

		require.NotNil(t, h)
		require.Equal(t, mockService, h.service)
		require.Equal(t, DefaultReposPath, h.reposPath)
	})
}

func TestNewGitHttpHandler(t *testing.T) {
	t.Run("creates handler with service user repo and repos path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)

		h := NewGitHttpHandler(mockService, nil, DefaultReposPath)

		require.NotNil(t, h)
		require.Equal(t, mockService, h.service)
		require.Equal(t, DefaultReposPath, h.reposPath)
	})
}

func TestGitHttpHandler_ServeHTTP(t *testing.T) {
	t.Run("returns 404 for empty repo path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)

		h := NewGitHttpHandler(mockService, nil, DefaultReposPath)

		req := httptest.NewRequest(http.MethodGet, "/git/", nil)
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 401 when no auth provided", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)

		h := NewGitHttpHandler(mockService, nil, DefaultReposPath)

		req := httptest.NewRequest(http.MethodGet, "/git/repo-uuid/info/refs?service=git-upload-pack", nil)
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		require.Contains(t, w.Header().Get("WWW-Authenticate"), "Basic")
	})

	t.Run("returns 401 for invalid API key", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockUserRepo := user.NewMockRepository(ctrl)

		mockUserRepo.EXPECT().
			GetUserByApiKey(gomock.Any(), "invalid-key").
			Return(nil, connect.NewError(connect.CodeNotFound, errors.New("api key not found")))

		h := NewGitHttpHandler(mockService, mockUserRepo, DefaultReposPath)

		req := httptest.NewRequest(http.MethodGet, "/git/repo-uuid/info/refs?service=git-upload-pack", nil)
		req.SetBasicAuth("user", "invalid-key")
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 403 when access denied", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockUserRepo := user.NewMockRepository(ctrl)

		mockUserRepo.EXPECT().
			GetUserByApiKey(gomock.Any(), "valid-key").
			Return(&user.UserDTO{Id: "user-123"}, nil)

		mockService.EXPECT().
			ValidateSshAccess(gomock.Any(), "user-123", "./repos/repo-uuid", SshOperationRead).
			Return(false, nil)

		h := NewGitHttpHandler(mockService, mockUserRepo, DefaultReposPath)

		req := httptest.NewRequest(http.MethodGet, "/git/repo-uuid/info/refs?service=git-upload-pack", nil)
		req.SetBasicAuth("user", "valid-key")
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("returns 404 for unknown subpath after auth", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockUserRepo := user.NewMockRepository(ctrl)

		mockUserRepo.EXPECT().
			GetUserByApiKey(gomock.Any(), "valid-key").
			Return(&user.UserDTO{Id: "user-123"}, nil)

		mockService.EXPECT().
			ValidateSshAccess(gomock.Any(), "user-123", "./repos/repo-uuid", SshOperationRead).
			Return(true, nil)

		h := NewGitHttpHandler(mockService, mockUserRepo, DefaultReposPath)

		req := httptest.NewRequest(http.MethodGet, "/git/repo-uuid/unknown", nil)
		req.SetBasicAuth("user", "valid-key")
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("successfully handles info/refs", func(t *testing.T) {
		if _, err := exec.LookPath("git"); err != nil {
			t.Skip("git not installed")
		}

		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockUserRepo := user.NewMockRepository(ctrl)

		tempDir, err := os.MkdirTemp("", "git-test")
		require.NoError(t, err)
		defer func() {
			_ = os.RemoveAll(tempDir)
		}()

		repoName := "test-repo"
		repoPath := filepath.Join(tempDir, repoName)
		err = os.MkdirAll(repoPath, 0755)
		require.NoError(t, err)

		cmd := exec.Command("git", "init", "--bare", repoPath)
		err = cmd.Run()
		require.NoError(t, err)

		mockUserRepo.EXPECT().
			GetUserByApiKey(gomock.Any(), "valid-key").
			Return(&user.UserDTO{Id: "user-123"}, nil)

		mockService.EXPECT().
			ValidateSshAccess(gomock.Any(), "user-123", repoPath, SshOperationRead).
			Return(true, nil)

		h := NewGitHttpHandler(mockService, mockUserRepo, tempDir)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/git/%s/info/refs?service=git-upload-pack", repoName), nil)
		req.SetBasicAuth("user", "valid-key")
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Header().Get("Content-Type"), "application/x-git-upload-pack-advertisement")
		require.Contains(t, w.Body.String(), "# service=git-upload-pack")
	})

	t.Run("successfully handles info/refs with .git suffix", func(t *testing.T) {
		if _, err := exec.LookPath("git"); err != nil {
			t.Skip("git not installed")
		}

		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockUserRepo := user.NewMockRepository(ctrl)

		tempDir, err := os.MkdirTemp("", "git-test")
		require.NoError(t, err)
		defer func() {
			_ = os.RemoveAll(tempDir)
		}()

		repoName := "test-repo"
		repoPath := filepath.Join(tempDir, repoName)
		err = os.MkdirAll(repoPath, 0755)
		require.NoError(t, err)

		cmd := exec.Command("git", "init", "--bare", repoPath)
		err = cmd.Run()
		require.NoError(t, err)

		mockUserRepo.EXPECT().
			GetUserByApiKey(gomock.Any(), "valid-key").
			Return(&user.UserDTO{Id: "user-123"}, nil)

		mockService.EXPECT().
			ValidateSshAccess(gomock.Any(), "user-123", repoPath, SshOperationRead).
			Return(true, nil)

		h := NewGitHttpHandler(mockService, mockUserRepo, tempDir)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/git/%s.git/info/refs?service=git-upload-pack", repoName), nil)
		req.SetBasicAuth("user", "valid-key")
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Header().Get("Content-Type"), "application/x-git-upload-pack-advertisement")
	})

	t.Run("cleans repository path when using relative paths", func(t *testing.T) {
		if _, err := exec.LookPath("git"); err != nil {
			t.Skip("git not installed")
		}

		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockUserRepo := user.NewMockRepository(ctrl)

		tempDir, err := os.MkdirTemp("", "git-test")
		require.NoError(t, err)
		defer func() {
			_ = os.RemoveAll(tempDir)
		}()

		repoName := "test-repo"
		absUrl := filepath.Join(tempDir, repoName)
		err = os.MkdirAll(absUrl, 0755)
		require.NoError(t, err)

		cmd := exec.Command("git", "init", "--bare", absUrl)
		err = cmd.Run()
		require.NoError(t, err)

		mockUserRepo.EXPECT().
			GetUserByApiKey(gomock.Any(), "valid-key").
			Return(&user.UserDTO{Id: "user-123"}, nil)

		messyBase := filepath.Dir(tempDir) + "/./" + filepath.Base(tempDir)
		cleanedPath := filepath.Clean(absUrl)

		mockService.EXPECT().
			ValidateSshAccess(gomock.Any(), "user-123", cleanedPath, SshOperationRead).
			Return(true, nil)

		h := NewGitHttpHandler(mockService, mockUserRepo, messyBase)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/git/%s/info/refs?service=git-upload-pack", repoName), nil)
		req.SetBasicAuth("user", "valid-key")
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})
}
