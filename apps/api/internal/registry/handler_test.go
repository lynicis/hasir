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
		assert.Equal(t, mockService, h.service)
		assert.Equal(t, mockRepository, h.repository)
		assert.Empty(t, h.interceptors)
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
		assert.Len(t, h.interceptors, 2)
	})
}

func TestHandler_RegisterRoutes(t *testing.T) {
	t.Run("returns path and handler", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		h := NewHandler(mockService, mockRepository)
		path, httpHandler := h.RegisterRoutes()

		assert.Equal(t, "/"+registryv1connect.RegistryServiceName+"/", path)
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
				assert.Equal(t, "test-repo", req.GetName())
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
		assert.Equal(t, connect.CodeAlreadyExists, connectErr.Code())
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
				assert.Equal(t, "test-repo-id", req.GetId())
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
		assert.NotNil(t, resp)
		assert.Equal(t, "test-repo-id", resp.Msg.GetId())
		assert.Equal(t, "test-repo", resp.Msg.GetName())
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
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())
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
		assert.Equal(t, connect.CodePermissionDenied, connectErr.Code())
	})
}

func TestHandler_GetRepositories(t *testing.T) {
	t.Run("success with repositories", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			GetRepositories(gomock.Any(), (*string)(nil), 1, 10).
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
		assert.Len(t, resp.Msg.GetRepositories(), 2)
		assert.Equal(t, "repo-1", resp.Msg.GetRepositories()[0].GetId())
		assert.Equal(t, "first-repo", resp.Msg.GetRepositories()[0].GetName())
		assert.Equal(t, "repo-2", resp.Msg.GetRepositories()[1].GetId())
		assert.Equal(t, "second-repo", resp.Msg.GetRepositories()[1].GetName())
	})

	t.Run("success with empty repositories", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			GetRepositories(gomock.Any(), (*string)(nil), 1, 10).
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
		assert.Empty(t, resp.Msg.GetRepositories())
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			GetRepositories(gomock.Any(), (*string)(nil), 1, 10).
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
		assert.Equal(t, connect.CodeInternal, connectErr.Code())
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
				assert.Equal(t, "test-repo-id", req.GetRepositoryId())
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
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())
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
		assert.Equal(t, connect.CodeInternal, connectErr.Code())
	})
}

func TestNewGitSshHandler(t *testing.T) {
	t.Run("creates handler with service and repos path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)

		h := NewGitSshHandler(mockService, DefaultReposPath)

		require.NotNil(t, h)
		assert.Equal(t, mockService, h.service)
		assert.Equal(t, DefaultReposPath, h.reposPath)
	})
}

func TestNewGitHttpHandler(t *testing.T) {
	t.Run("creates handler with service user repo and repos path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)

		h := NewGitHttpHandler(mockService, nil, DefaultReposPath)

		require.NotNil(t, h)
		assert.Equal(t, mockService, h.service)
		assert.Equal(t, DefaultReposPath, h.reposPath)
	})
}

func TestGitHttpHandler_ServeHTTP(t *testing.T) {
	t.Run("returns 404 for empty repo path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockUserRepo := user.NewMockRepository(ctrl)

		mockUserRepo.EXPECT().
			GetUserByApiKey(gomock.Any(), "valid-key").
			Return(&user.UserDTO{Id: "user-123"}, nil)

		h := NewGitHttpHandler(mockService, mockUserRepo, DefaultReposPath)

		req := httptest.NewRequest(http.MethodGet, "/git/", nil)
		req.SetBasicAuth("user", "valid-key")
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

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Header().Get("WWW-Authenticate"), "Basic")
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
}

func TestHandler_GetCommits(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		expectedCommits := &registryv1.GetCommitsResponse{
			Commits: []*registryv1.Commit{
				{
					Id:      "abc123",
					Message: "Initial commit",
					User: &registryv1.Commit_User{
						Id:       "user@example.com",
						Username: "Test User",
					},
				},
			},
			NextPage:  int32(0),
			TotalPage: int32(1),
		}

		mockService.EXPECT().
			GetCommits(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, req *registryv1.GetCommitsRequest) (*registryv1.GetCommitsResponse, error) {
				require.Equal(t, "test-repo-id", req.GetId())
				return expectedCommits, nil
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

		resp, err := client.GetCommits(context.Background(), connect.NewRequest(&registryv1.GetCommitsRequest{
			Id: "test-repo-id",
		}))
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Msg.GetCommits(), 1)
		assert.Equal(t, "abc123", resp.Msg.GetCommits()[0].GetId())
		assert.Equal(t, "Initial commit", resp.Msg.GetCommits()[0].GetMessage())
		assert.Equal(t, "user@example.com", resp.Msg.GetCommits()[0].GetUser().GetId())
		assert.Equal(t, "Test User", resp.Msg.GetCommits()[0].GetUser().GetUsername())
	})

	t.Run("service error - repository not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			GetCommits(gomock.Any(), gomock.Any()).
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

		_, err := client.GetCommits(context.Background(), connect.NewRequest(&registryv1.GetCommitsRequest{
			Id: "non-existent-repo",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())
	})
}

func TestHandler_GetFileTree(t *testing.T) {
	t.Run("success - root directory", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		expectedFileTree := &registryv1.GetFileTreeResponse{
			Nodes: []*registryv1.FileTreeNode{
				{
					Name: "README.md",
					Path: "README.md",
					Type: registryv1.NodeType_NODE_TYPE_FILE,
				},
				{
					Name: "src",
					Path: "src",
					Type: registryv1.NodeType_NODE_TYPE_DIRECTORY,
					Children: []*registryv1.FileTreeNode{
						{
							Name: "main.go",
							Path: "src/main.go",
							Type: registryv1.NodeType_NODE_TYPE_FILE,
						},
					},
				},
			},
		}

		mockService.EXPECT().
			GetFileTree(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, req *registryv1.GetFileTreeRequest) (*registryv1.GetFileTreeResponse, error) {
				assert.Equal(t, "test-repo-id", req.GetId())
				assert.False(t, req.HasPath())
				return expectedFileTree, nil
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

		resp, err := client.GetFileTree(context.Background(), connect.NewRequest(&registryv1.GetFileTreeRequest{
			Id: "test-repo-id",
		}))
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Msg.GetNodes(), 2)
		assert.Equal(t, "README.md", resp.Msg.GetNodes()[0].GetName())
		assert.Equal(t, registryv1.NodeType_NODE_TYPE_FILE, resp.Msg.GetNodes()[0].GetType())
		assert.Equal(t, "src", resp.Msg.GetNodes()[1].GetName())
		assert.Equal(t, registryv1.NodeType_NODE_TYPE_DIRECTORY, resp.Msg.GetNodes()[1].GetType())
		assert.Len(t, resp.Msg.GetNodes()[1].GetChildren(), 1)
	})

	t.Run("success - subdirectory", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		expectedFileTree := &registryv1.GetFileTreeResponse{
			Nodes: []*registryv1.FileTreeNode{
				{
					Name: "main.go",
					Path: "src/main.go",
					Type: registryv1.NodeType_NODE_TYPE_FILE,
				},
			},
		}

		mockService.EXPECT().
			GetFileTree(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, req *registryv1.GetFileTreeRequest) (*registryv1.GetFileTreeResponse, error) {
				assert.Equal(t, "test-repo-id", req.GetId())
				assert.True(t, req.HasPath())
				assert.Equal(t, "src", req.GetPath())
				return expectedFileTree, nil
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

		subPath := "src"
		resp, err := client.GetFileTree(context.Background(), connect.NewRequest(&registryv1.GetFileTreeRequest{
			Id:   "test-repo-id",
			Path: &subPath,
		}))
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Msg.GetNodes(), 1)
		assert.Equal(t, "main.go", resp.Msg.GetNodes()[0].GetName())
	})

	t.Run("service error - repository not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			GetFileTree(gomock.Any(), gomock.Any()).
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

		_, err := client.GetFileTree(context.Background(), connect.NewRequest(&registryv1.GetFileTreeRequest{
			Id: "non-existent-repo",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())
	})
}

func TestHandler_GetFilePreview(t *testing.T) {
	t.Run("success - text file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		expectedFilePreview := &registryv1.GetFilePreviewResponse{
			Content: "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}",
		}

		mockService.EXPECT().
			GetFilePreview(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, req *registryv1.GetFilePreviewRequest) (*registryv1.GetFilePreviewResponse, error) {
				assert.Equal(t, "test-repo-id", req.GetId())
				assert.Equal(t, "main.go", req.GetPath())
				return expectedFilePreview, nil
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

		resp, err := client.GetFilePreview(context.Background(), connect.NewRequest(&registryv1.GetFilePreviewRequest{
			Id:   "test-repo-id",
			Path: "main.go",
		}))
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}", resp.Msg.GetContent())
	})

	t.Run("success - file in subdirectory", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		expectedFilePreview := &registryv1.GetFilePreviewResponse{
			Content: "# Test README\n\nThis is a test file.",
		}

		mockService.EXPECT().
			GetFilePreview(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, req *registryv1.GetFilePreviewRequest) (*registryv1.GetFilePreviewResponse, error) {
				assert.Equal(t, "test-repo-id", req.GetId())
				assert.Equal(t, "docs/README.md", req.GetPath())
				return expectedFilePreview, nil
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

		resp, err := client.GetFilePreview(context.Background(), connect.NewRequest(&registryv1.GetFilePreviewRequest{
			Id:   "test-repo-id",
			Path: "docs/README.md",
		}))
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "# Test README\n\nThis is a test file.", resp.Msg.GetContent())
	})

	t.Run("success - empty file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		expectedFilePreview := &registryv1.GetFilePreviewResponse{
			Content: "",
		}

		mockService.EXPECT().
			GetFilePreview(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, req *registryv1.GetFilePreviewRequest) (*registryv1.GetFilePreviewResponse, error) {
				assert.Equal(t, "test-repo-id", req.GetId())
				assert.Equal(t, "empty.txt", req.GetPath())
				return expectedFilePreview, nil
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

		resp, err := client.GetFilePreview(context.Background(), connect.NewRequest(&registryv1.GetFilePreviewRequest{
			Id:   "test-repo-id",
			Path: "empty.txt",
		}))
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Empty(t, resp.Msg.GetContent())
	})

	t.Run("service error - repository not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			GetFilePreview(gomock.Any(), gomock.Any()).
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

		_, err := client.GetFilePreview(context.Background(), connect.NewRequest(&registryv1.GetFilePreviewRequest{
			Id:   "non-existent-repo",
			Path: "main.go",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())
	})

	t.Run("service error - file not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			GetFilePreview(gomock.Any(), gomock.Any()).
			Return(nil, connect.NewError(connect.CodeNotFound, errors.New("file not found")))

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

		_, err := client.GetFilePreview(context.Background(), connect.NewRequest(&registryv1.GetFilePreviewRequest{
			Id:   "test-repo-id",
			Path: "nonexistent.go",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())
	})

	t.Run("service error - permission denied", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			GetFilePreview(gomock.Any(), gomock.Any()).
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

		_, err := client.GetFilePreview(context.Background(), connect.NewRequest(&registryv1.GetFilePreviewRequest{
			Id:   "test-repo-id",
			Path: "main.go",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		assert.Equal(t, connect.CodePermissionDenied, connectErr.Code())
	})

	t.Run("service error - internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			GetFilePreview(gomock.Any(), gomock.Any()).
			Return(nil, connect.NewError(connect.CodeInternal, errors.New("git error")))

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

		_, err := client.GetFilePreview(context.Background(), connect.NewRequest(&registryv1.GetFilePreviewRequest{
			Id:   "test-repo-id",
			Path: "main.go",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		assert.Equal(t, connect.CodeInternal, connectErr.Code())
	})

	t.Run("success - file with unicode and special characters", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		specialContent := "Unicode: 你好 🚀\nTabs:\t\ttabs\nNewlines:\n\nSpaces:   multiple"
		expectedFilePreview := &registryv1.GetFilePreviewResponse{
			Content: specialContent,
		}

		mockService.EXPECT().
			GetFilePreview(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, req *registryv1.GetFilePreviewRequest) (*registryv1.GetFilePreviewResponse, error) {
				assert.Equal(t, "test-repo-id", req.GetId())
				assert.Equal(t, "special.txt", req.GetPath())
				return expectedFilePreview, nil
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

		resp, err := client.GetFilePreview(context.Background(), connect.NewRequest(&registryv1.GetFilePreviewRequest{
			Id:   "test-repo-id",
			Path: "special.txt",
		}))
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, specialContent, resp.Msg.GetContent())
	})
}

func TestHandler_UpdateSdkPreferences(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			UpdateSdkPreferences(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, req *registryv1.UpdateSdkPreferencesRequest) error {
				assert.Equal(t, "test-repo-id", req.GetId())
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

		_, err := client.UpdateSdkPreferences(context.Background(), connect.NewRequest(&registryv1.UpdateSdkPreferencesRequest{
			Id: "test-repo-id",
			SdkPreferences: []*registryv1.SdkPreference{
				{
					Sdk:    registryv1.SDK_SDK_GO_PROTOBUF,
					Status: true,
				},
			},
		}))
		require.NoError(t, err)
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			UpdateSdkPreferences(gomock.Any(), gomock.Any()).
			Return(connect.NewError(connect.CodeNotFound, errors.New("repository not found")))

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

		_, err := client.UpdateSdkPreferences(context.Background(), connect.NewRequest(&registryv1.UpdateSdkPreferencesRequest{
			Id: "nonexistent-repo-id",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())
	})
}

func TestHandler_GetRecentCommit(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		expectedCommit := &registryv1.Commit{
			Id:      "abc123",
			Message: "Recent commit",
			User: &registryv1.Commit_User{
				Id:       "user@example.com",
				Username: "Test User",
			},
		}

		mockService.EXPECT().
			GetRecentCommit(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, req *registryv1.GetRecentCommitRequest) (*registryv1.Commit, error) {
				assert.Equal(t, "test-repo-id", req.GetRepositoryId())
				return expectedCommit, nil
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

		resp, err := client.GetRecentCommit(context.Background(), connect.NewRequest(&registryv1.GetRecentCommitRequest{
			RepositoryId: "test-repo-id",
		}))
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "abc123", resp.Msg.GetId())
		assert.Equal(t, "Recent commit", resp.Msg.GetMessage())
	})

	t.Run("service error - repository not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockService.EXPECT().
			GetRecentCommit(gomock.Any(), gomock.Any()).
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

		_, err := client.GetRecentCommit(context.Background(), connect.NewRequest(&registryv1.GetRecentCommitRequest{
			RepositoryId: "non-existent-repo",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())
	})
}

func TestNewSdkHttpHandler(t *testing.T) {
	t.Run("creates handler with sdk repos path", func(t *testing.T) {
		h := NewSdkHttpHandler("/sdk/repos")
		require.NotNil(t, h)
		assert.Equal(t, "/sdk/repos", h.sdkReposPath)
	})
}

func TestNewSdkSshHandler(t *testing.T) {
	t.Run("creates handler with sdk repos path", func(t *testing.T) {
		h := NewSdkSshHandler("/sdk/repos")
		require.NotNil(t, h)
		assert.Equal(t, "/sdk/repos", h.sdkReposPath)
	})
}

func TestIsValidPathComponent(t *testing.T) {
	tests := []struct {
		name      string
		component string
		want      bool
	}{
		{
			name:      "valid component",
			component: "valid-name",
			want:      true,
		},
		{
			name:      "contains ..",
			component: "../invalid",
			want:      false,
		},
		{
			name:      "absolute path",
			component: "/absolute/path",
			want:      false,
		},
		{
			name:      "contains slash",
			component: "path/with/slash",
			want:      false,
		},
		{
			name:      "contains backslash",
			component: "path\\with\\backslash",
			want:      false,
		},
		{
			name:      "cleaned path differs",
			component: "path/../other",
			want:      false,
		},
		{
			name:      "current directory",
			component: ".",
			want:      false,
		},
		{
			name:      "empty string",
			component: "",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidPathComponent(tt.component)
			assert.Equal(t, tt.want, got)
		})
	}
}
