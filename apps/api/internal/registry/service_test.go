package registry

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
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
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		tmpDir := t.TempDir()

		svc := &service{
			rootPath:   tmpDir,
			repository: mockRepo,
		}

		const repoName = "my-repo"
		ctx := context.Background()

		mockRepo.EXPECT().
			GetRepositoryByName(ctx, repoName).
			Return(nil, ErrRepositoryNotFound)

		mockRepo.EXPECT().
			CreateRepository(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, repo *RepositoryDTO) error {
				require.Equal(t, repoName, repo.Name)
				require.Equal(t, filepath.Join(tmpDir, repoName), repo.Path)
				require.NotEmpty(t, repo.Id)
				return nil
			})

		err := svc.CreateRepository(ctx, &registryv1.CreateRepositoryRequest{
			Name: repoName,
		})
		require.NoError(t, err)

		repoPath := filepath.Join(tmpDir, repoName)
		require.DirExists(t, repoPath)
		require.FileExists(t, filepath.Join(repoPath, ".git"))
	})

	t.Run("already exists in database", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)

		svc := &service{
			rootPath:   t.TempDir(),
			repository: mockRepo,
		}

		const repoName = "existing-repo"
		ctx := context.Background()

		mockRepo.EXPECT().
			GetRepositoryByName(ctx, repoName).
			Return(&RepositoryDTO{
				Id:   "existing-id",
				Name: repoName,
				Path: "/some/path",
			}, nil)

		err := svc.CreateRepository(ctx, &registryv1.CreateRepositoryRequest{
			Name: repoName,
		})
		require.EqualError(t, err, `repository "existing-repo" already exists`)
	})

	t.Run("already exists on filesystem", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		tmpDir := t.TempDir()

		svc := &service{
			rootPath:   tmpDir,
			repository: mockRepo,
		}

		const repoName = "existing-repo"
		ctx := context.Background()

		mockRepo.EXPECT().
			GetRepositoryByName(ctx, repoName).
			Return(nil, ErrRepositoryNotFound)

		mockRepo.EXPECT().
			CreateRepository(ctx, gomock.Any()).
			Return(nil)

		err := svc.CreateRepository(ctx, &registryv1.CreateRepositoryRequest{
			Name: repoName,
		})
		require.NoError(t, err)

		mockRepo.EXPECT().
			GetRepositoryByName(ctx, repoName).
			Return(&RepositoryDTO{
				Id:   "some-id",
				Name: repoName,
				Path: filepath.Join(tmpDir, repoName),
			}, nil)

		err = svc.CreateRepository(ctx, &registryv1.CreateRepositoryRequest{
			Name: repoName,
		})
		require.EqualError(t, err, `repository "existing-repo" already exists`)
	})

	t.Run("database lookup error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)

		svc := &service{
			rootPath:   t.TempDir(),
			repository: mockRepo,
		}

		const repoName = "my-repo"
		ctx := context.Background()
		dbErr := errors.New("database connection failed")

		mockRepo.EXPECT().
			GetRepositoryByName(ctx, repoName).
			Return(nil, dbErr)

		err := svc.CreateRepository(ctx, &registryv1.CreateRepositoryRequest{
			Name: repoName,
		})
		require.ErrorContains(t, err, "failed to check existing repository")
		require.ErrorIs(t, err, dbErr)
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
		ctx := context.Background()
		dbErr := errors.New("database insert failed")

		mockRepo.EXPECT().
			GetRepositoryByName(ctx, repoName).
			Return(nil, ErrRepositoryNotFound)

		mockRepo.EXPECT().
			CreateRepository(ctx, gomock.Any()).
			Return(dbErr)

		err := svc.CreateRepository(ctx, &registryv1.CreateRepositoryRequest{
			Name: repoName,
		})
		require.ErrorContains(t, err, "failed to save repository to database")
		require.ErrorIs(t, err, dbErr)

		// Verify git directory was rolled back
		repoPath := filepath.Join(tmpDir, repoName)
		require.NoDirExists(t, repoPath)
	})
}
