package registry

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"hasir-api/pkg/auth"

	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
	"buf.build/gen/go/hasir/hasir/protocolbuffers/go/shared"
)

func TestNewService(t *testing.T) {
	t.Run("default root path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)

		svc := NewService(mockRepo)
		concrete, ok := svc.(*service)
		require.True(t, ok, "NewService should return *service")
		require.Equal(t, defaultReposPath, concrete.rootPath)
	})
}

func TestService_CreateRepository(t *testing.T) {
	t.Run("success with default visibility (private)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		tmpDir := t.TempDir()

		svc := &service{
			rootPath:   tmpDir,
			repository: mockRepo,
		}

		const repoName = "my-repo"
		const userID = "test-user-id"
		ctx := context.WithValue(context.Background(), auth.UserIDKey, userID)

		mockRepo.EXPECT().
			CreateRepository(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, repo *RepositoryDTO) error {
				require.Equal(t, repoName, repo.Name)
				require.Equal(t, filepath.Join(tmpDir, repo.Id), repo.Path)
				require.Equal(t, userID, repo.CreatedBy)
				require.Equal(t, VisibilityPrivate, repo.Visibility)
				require.NotEmpty(t, repo.Id)
				return nil
			})

		err := svc.CreateRepository(ctx, &registryv1.CreateRepositoryRequest{
			Name: repoName,
		})
		require.NoError(t, err)

		require.NoError(t, err)

		dirs, err := os.ReadDir(tmpDir)
		require.NoError(t, err)
		require.Len(t, dirs, 1)

		repoPath := filepath.Join(tmpDir, dirs[0].Name())
		require.DirExists(t, repoPath)
		require.FileExists(t, filepath.Join(repoPath, "HEAD"))
	})

	t.Run("success with explicit public visibility", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		tmpDir := t.TempDir()

		svc := &service{
			rootPath:   tmpDir,
			repository: mockRepo,
		}

		const repoName = "public-repo"
		const userID = "test-user-id"
		ctx := context.WithValue(context.Background(), auth.UserIDKey, userID)

		mockRepo.EXPECT().
			CreateRepository(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, repo *RepositoryDTO) error {
				require.Equal(t, repoName, repo.Name)
				require.Equal(t, filepath.Join(tmpDir, repo.Id), repo.Path)
				require.Equal(t, userID, repo.CreatedBy)
				require.Equal(t, VisibilityPublic, repo.Visibility)
				require.NotEmpty(t, repo.Id)
				return nil
			})

		err := svc.CreateRepository(ctx, &registryv1.CreateRepositoryRequest{
			Name:       repoName,
			Visibility: shared.Visibility_VISIBILITY_PUBLIC,
		})
		require.NoError(t, err)
	})

	t.Run("database save error rolls back git directory", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		tmpDir := t.TempDir()

		svc := &service{
			rootPath:   tmpDir,
			repository: mockRepo,
		}

		const repoName = "my-repo"
		const userID = "test-user-id"
		ctx := context.WithValue(context.Background(), auth.UserIDKey, userID)
		dbErr := errors.New("database insert failed")

		mockRepo.EXPECT().
			CreateRepository(ctx, gomock.Any()).
			Return(dbErr)

		err := svc.CreateRepository(ctx, &registryv1.CreateRepositoryRequest{
			Name: repoName,
		})
		require.ErrorContains(t, err, "failed to save repository to database")
		require.ErrorIs(t, err, dbErr)

		repoPath := filepath.Join(tmpDir, repoName)
		require.NoDirExists(t, repoPath)
	})
}

func TestService_DeleteRepository(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		tmpDir := t.TempDir()

		svc := &service{
			rootPath:   tmpDir,
			repository: mockRepo,
		}

		repoId := "test-repo-id"
		repoPath := filepath.Join(tmpDir, repoId)
		repoName := "test-repo"
		ctx := context.Background()

		require.NoError(t, os.MkdirAll(repoPath, 0o755))

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoId).
			Return(&RepositoryDTO{
				Id:   repoId,
				Name: repoName,
				Path: repoPath,
			}, nil)

		mockRepo.EXPECT().
			DeleteRepository(ctx, repoId).
			Return(nil)

		err := svc.DeleteRepository(ctx, &registryv1.DeleteRepositoryRequest{
			RepositoryId: repoId,
		})
		require.NoError(t, err)

		require.NoDirExists(t, repoPath)
	})

	t.Run("repository not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)

		svc := &service{
			rootPath:   t.TempDir(),
			repository: mockRepo,
		}

		repoId := "nonexistent-repo-id"
		ctx := context.Background()

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoId).
			Return(nil, ErrRepositoryNotFound)

		err := svc.DeleteRepository(ctx, &registryv1.DeleteRepositoryRequest{
			RepositoryId: repoId,
		})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to get repository")
	})

	t.Run("database delete error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		tmpDir := t.TempDir()

		svc := &service{
			rootPath:   tmpDir,
			repository: mockRepo,
		}

		repoId := "test-repo-id"
		repoPath := filepath.Join(tmpDir, repoId)
		repoName := "test-repo"
		ctx := context.Background()

		require.NoError(t, os.MkdirAll(repoPath, 0o755))

		dbErr := errors.New("database delete failed")

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoId).
			Return(&RepositoryDTO{
				Id:   repoId,
				Name: repoName,
				Path: repoPath,
			}, nil)

		mockRepo.EXPECT().
			DeleteRepository(ctx, repoId).
			Return(dbErr)

		err := svc.DeleteRepository(ctx, &registryv1.DeleteRepositoryRequest{
			RepositoryId: repoId,
		})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to delete repository from database")
		require.ErrorIs(t, err, dbErr)

		require.DirExists(t, repoPath)
	})

	t.Run("filesystem removal failure returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		tmpDir := t.TempDir()

		svc := &service{
			rootPath:   tmpDir,
			repository: mockRepo,
		}

		repoId := "test-repo-id"
		repoPath := filepath.Join(tmpDir, repoId)
		repoName := "test-repo"
		ctx := context.Background()

		require.NoError(t, os.MkdirAll(repoPath, 0o755))

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoId).
			Return(&RepositoryDTO{
				Id:   repoId,
				Name: repoName,
				Path: repoPath,
			}, nil)

		mockRepo.EXPECT().
			DeleteRepository(ctx, repoId).
			Return(nil)

		testFile := filepath.Join(repoPath, "test-file")
		require.NoError(t, os.WriteFile(testFile, []byte("test"), 0o644))
		require.NoError(t, os.Chmod(repoPath, 0o444))

		err := svc.DeleteRepository(ctx, &registryv1.DeleteRepositoryRequest{
			RepositoryId: repoId,
		})

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to remove repository directory")

		// Cleanup: restore permissions so temp dir can be cleaned up
		if _, err := os.Stat(repoPath); err == nil {
			require.NoError(t, os.Chmod(repoPath, 0o755))
		}
	})
}
