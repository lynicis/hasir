package registry

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"hasir-api/pkg/authentication"
	"hasir-api/pkg/authorization"
	"hasir-api/pkg/config"
	"hasir-api/pkg/proto"
	"hasir-api/pkg/sdkgenerator"

	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
	"buf.build/gen/go/hasir/hasir/protocolbuffers/go/shared"
)

func initGitRepoWithEmptyCommit(t *testing.T, repoPath string) string {
	t.Helper()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "commit", "--allow-empty", "-m", "initial"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repoPath
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "command %v failed: %s", args, string(out))
	}

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	require.NoError(t, err)
	return strings.TrimSpace(string(out))
}

func initGitRepoWithProtoFile(t *testing.T, repoPath string) string {
	t.Helper()

	protoFile := filepath.Join(repoPath, "test.proto")
	require.NoError(t, os.WriteFile(protoFile, []byte("syntax = \"proto3\";"), 0o644))

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "add", "."},
		{"git", "commit", "-m", "add proto file"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repoPath
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "command %v failed: %s", args, string(out))
	}

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	require.NoError(t, err)
	return strings.TrimSpace(string(out))
}

func TestNewService(t *testing.T) {
	t.Run("default root path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := NewService(mockRepo, mockOrgRepo, nil, nil)
		concrete, ok := svc.(*service)
		require.True(t, ok, "NewService should return *service")
		assert.Equal(t, DefaultReposPath, concrete.rootPath)
	})
}

func testAuthInterceptor(userID string) context.Context {
	return context.WithValue(context.Background(), authentication.UserIDKey, userID)
}

func TestService_CreateRepository(t *testing.T) {
	t.Run("success with default visibility (private)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)
		tmpDir := t.TempDir()

		svc := &service{
			rootPath:   tmpDir,
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const repoName = "my-repo"
		const orgID = "org-123"
		const userID = "test-user-id"
		ctx := testAuthInterceptor(userID)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleOwner, nil)

		mockRepo.EXPECT().
			CreateRepository(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, repo *RepositoryDTO) error {
				assert.Equal(t, repoName, repo.Name)
				assert.Equal(t, filepath.Join(tmpDir, repo.Id), repo.Path)
				assert.Equal(t, userID, repo.CreatedBy)
				assert.Equal(t, proto.VisibilityPrivate, repo.Visibility)
				assert.NotEmpty(t, repo.Id)
				return nil
			})

		err := svc.CreateRepository(ctx, &registryv1.CreateRepositoryRequest{
			Name:           repoName,
			OrganizationId: orgID,
		})
		require.NoError(t, err)

		dirs, err := os.ReadDir(tmpDir)
		require.NoError(t, err)
		assert.Len(t, dirs, 1)

		repoPath := filepath.Join(tmpDir, dirs[0].Name())
		assert.DirExists(t, repoPath)
		assert.FileExists(t, filepath.Join(repoPath, "HEAD"))
	})

	t.Run("success with explicit public visibility", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)
		tmpDir := t.TempDir()

		svc := &service{
			rootPath:   tmpDir,
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const repoName = "public-repo"
		const orgID = "org-123"
		const userID = "test-user-id"
		ctx := testAuthInterceptor(userID)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleOwner, nil)

		mockRepo.EXPECT().
			CreateRepository(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, repo *RepositoryDTO) error {
				assert.Equal(t, repoName, repo.Name)
				assert.Equal(t, filepath.Join(tmpDir, repo.Id), repo.Path)
				assert.Equal(t, userID, repo.CreatedBy)
				assert.Equal(t, proto.VisibilityPublic, repo.Visibility)
				assert.NotEmpty(t, repo.Id)
				return nil
			})

		err := svc.CreateRepository(ctx, &registryv1.CreateRepositoryRequest{
			Name:           repoName,
			OrganizationId: orgID,
			Visibility:     shared.Visibility_VISIBILITY_PUBLIC,
		})
		require.NoError(t, err)
	})

	t.Run("database save error rolls back git directory", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)
		tmpDir := t.TempDir()

		svc := &service{
			rootPath:   tmpDir,
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const repoName = "my-repo"
		const orgID = "org-123"
		const userID = "test-user-id"
		ctx := testAuthInterceptor(userID)
		dbErr := errors.New("failed to save repository to database")

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleOwner, nil)

		mockRepo.EXPECT().
			CreateRepository(ctx, gomock.Any()).
			Return(dbErr)

		err := svc.CreateRepository(ctx, &registryv1.CreateRepositoryRequest{
			Name:           repoName,
			OrganizationId: orgID,
		})
		require.ErrorContains(t, err, "failed to save repository to database")

		repoPath := filepath.Join(tmpDir, repoName)
		assert.NoDirExists(t, repoPath)
	})
}

func TestService_GetRepository(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   t.TempDir(),
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const repoID = "repo-123"
		const orgID = "org-123"
		const userID = "user-123"
		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Visibility:     proto.VisibilityPrivate,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleReader, nil)

		sdkPrefs := []SdkPreferencesDTO{
			{
				Id:           "pref-1",
				RepositoryId: repoID,
				Sdk:          SdkGoConnectRpc,
				Status:       true,
			},
		}

		mockRepo.EXPECT().
			GetSdkPreferences(ctx, repoID).
			Return(sdkPrefs, nil)

		repo, err := svc.GetRepository(ctx, &registryv1.GetRepositoryRequest{
			Id: repoID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, repo)
		assert.Equal(t, repoID, repo.GetId())
		assert.Equal(t, "test-repo", repo.GetName())
		assert.Equal(t, orgID, repo.GetOrganizationId())
		assert.Equal(t, shared.Visibility_VISIBILITY_PRIVATE, repo.GetVisibility())
		assert.Len(t, repo.GetSdkPreferences(), 1)
		assert.Equal(t, registryv1.SDK_SDK_GO_CONNECTRPC, repo.GetSdkPreferences()[0].GetSdk())
		assert.True(t, repo.GetSdkPreferences()[0].GetStatus())
	})

	t.Run("repository not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   t.TempDir(),
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const repoID = "nonexistent-repo"
		const userID = "user-123"
		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(nil, ErrRepositoryNotFound)

		repo, err := svc.GetRepository(ctx, &registryv1.GetRepositoryRequest{
			Id: repoID,
		})
		assert.Error(t, err)
		assert.Nil(t, repo)
		assert.ErrorContains(t, err, "repository not found")
	})

	t.Run("user not member of organization", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   t.TempDir(),
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const repoID = "repo-123"
		const orgID = "org-123"
		const userID = "user-123"
		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Visibility:     proto.VisibilityPrivate,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return("", authorization.ErrMemberNotFound)

		repo, err := svc.GetRepository(ctx, &registryv1.GetRepositoryRequest{
			Id: repoID,
		})
		assert.Error(t, err)
		assert.Nil(t, repo)
		assert.ErrorContains(t, err, "you are not a member of this organization")
	})

	t.Run("missing user ID in context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   t.TempDir(),
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const repoID = "repo-123"
		const orgID = "org-123"
		ctx := context.Background()

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Visibility:     proto.VisibilityPrivate,
			}, nil)

		repo, err := svc.GetRepository(ctx, &registryv1.GetRepositoryRequest{
			Id: repoID,
		})
		assert.Error(t, err)
		assert.Nil(t, repo)
		assert.ErrorContains(t, err, "user not authenticated")
	})
}

func TestService_UpdateRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   t.TempDir(),
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const repoID = "repo-123"
		const repoName = "test-repo"
		const orgID = "org-123"
		const userID = "user-123"
		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           repoName,
				OrganizationId: orgID,
				Visibility:     proto.VisibilityPrivate,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleOwner, nil)

		mockRepo.EXPECT().
			UpdateRepository(ctx, gomock.Any()).
			Return(nil)

		err := svc.UpdateRepository(ctx, &registryv1.UpdateRepositoryRequest{
			Id:         repoID,
			Name:       repoName,
			Visibility: shared.Visibility_VISIBILITY_PRIVATE,
		})
		assert.NoError(t, err)
	})
}

func TestService_DeleteRepository(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)
		tmpDir := t.TempDir()

		svc := &service{
			rootPath:   tmpDir,
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		repoId := "test-repo-id"
		orgID := "org-123"
		userID := "user-123"
		repoPath := filepath.Join(tmpDir, repoId)
		repoName := "test-repo"
		ctx := testAuthInterceptor(userID)

		require.NoError(t, os.MkdirAll(repoPath, 0o755))

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoId).
			Return(&RepositoryDTO{
				Id:             repoId,
				Name:           repoName,
				Path:           repoPath,
				OrganizationId: orgID,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleOwner, nil)

		mockRepo.EXPECT().
			DeleteRepository(ctx, repoId).
			Return(nil)

		err := svc.DeleteRepository(ctx, &registryv1.DeleteRepositoryRequest{
			RepositoryId: repoId,
		})

		assert.NoError(t, err)
		assert.NoDirExists(t, repoPath)
	})

	t.Run("repository not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   t.TempDir(),
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		repoId := "nonexistent-repo-id"
		userID := "user-123"
		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoId).
			Return(nil, ErrRepositoryNotFound)

		err := svc.DeleteRepository(ctx, &registryv1.DeleteRepositoryRequest{
			RepositoryId: repoId,
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "repository not found")
	})

	t.Run("database delete error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)
		tmpDir := t.TempDir()

		svc := &service{
			rootPath:   tmpDir,
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		repoId := "test-repo-id"
		orgID := "org-123"
		userID := "user-123"
		repoPath := filepath.Join(tmpDir, repoId)
		repoName := "test-repo"
		ctx := testAuthInterceptor(userID)

		require.NoError(t, os.MkdirAll(repoPath, 0o755))

		dbErr := errors.New("database delete failed")

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoId).
			Return(&RepositoryDTO{
				Id:             repoId,
				Name:           repoName,
				Path:           repoPath,
				OrganizationId: orgID,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleOwner, nil)

		mockRepo.EXPECT().
			DeleteRepository(ctx, repoId).
			Return(dbErr)

		err := svc.DeleteRepository(ctx, &registryv1.DeleteRepositoryRequest{
			RepositoryId: repoId,
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "database delete failed")
		assert.ErrorIs(t, err, dbErr)

		assert.DirExists(t, repoPath)
	})

	t.Run("filesystem removal failure returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)
		tmpDir := t.TempDir()

		svc := &service{
			rootPath:   tmpDir,
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		repoId := "test-repo-id"
		orgID := "org-123"
		userID := "user-123"
		repoPath := filepath.Join(tmpDir, repoId)
		repoName := "test-repo"
		ctx := testAuthInterceptor(userID)

		require.NoError(t, os.MkdirAll(repoPath, 0o755))

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoId).
			Return(&RepositoryDTO{
				Id:             repoId,
				Name:           repoName,
				Path:           repoPath,
				OrganizationId: orgID,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleOwner, nil)

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

		if _, err := os.Stat(repoPath); err == nil {
			require.NoError(t, os.Chmod(repoPath, 0o755))
		}
	})

	t.Run("GetRepositories returns repositories for authenticated user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   t.TempDir(),
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const userID = "user-123"
		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoriesByUserCount(ctx, userID).
			Return(2, nil)

		repos := &[]RepositoryDTO{
			{Id: "repo-1", Name: "first-repo", Visibility: proto.VisibilityPrivate},
			{Id: "repo-2", Name: "second-repo", Visibility: proto.VisibilityPublic},
		}

		mockRepo.EXPECT().
			GetRepositoriesByUser(ctx, userID, 1, 10).
			Return(repos, nil)

		sdkPrefsMap := map[string][]SdkPreferencesDTO{
			"repo-1": {
				{
					Id:           "pref-1",
					RepositoryId: "repo-1",
					Sdk:          SdkGoConnectRpc,
					Status:       true,
				},
			},
		}

		mockRepo.EXPECT().
			GetSdkPreferencesByRepositoryIds(ctx, []string{"repo-1", "repo-2"}).
			Return(sdkPrefsMap, nil)

		resp, err := svc.GetRepositories(ctx, nil, 1, 10)
		require.NoError(t, err)
		assert.Len(t, resp.GetRepositories(), 2)
		assert.Equal(t, "repo-1", resp.GetRepositories()[0].GetId())
		assert.Equal(t, "first-repo", resp.GetRepositories()[0].GetName())
		assert.Len(t, resp.GetRepositories()[0].GetSdkPreferences(), 1)
		assert.Equal(t, "repo-2", resp.GetRepositories()[1].GetId())
		assert.Equal(t, "second-repo", resp.GetRepositories()[1].GetName())
		assert.Len(t, resp.GetRepositories()[1].GetSdkPreferences(), 0)
	})
}

func TestService_GetCommits(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const userID = "user-123"
		const orgID = "org-123"
		const repoID = "repo-123"
		repoPath := filepath.Join("./repos", repoID)

		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Path:           repoPath,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleReader, nil)

		expectedCommits := []*registryv1.Commit{
			{
				Id:      "abc123",
				Message: "Initial commit",
				User: &registryv1.Commit_User{
					Id:       "user@example.com",
					Username: "Test User",
				},
			},
		}

		mockRepo.EXPECT().
			GetCommits(ctx, repoPath, 1, 10).
			Return(expectedCommits, 1, nil)

		req := &registryv1.GetCommitsRequest{Id: repoID}
		resp, err := svc.GetCommits(ctx, req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.GetCommits(), 1)
		assert.Equal(t, "abc123", resp.GetCommits()[0].GetId())
		assert.Equal(t, int32(1), resp.GetTotalPage())
		assert.Equal(t, int32(0), resp.GetNextPage())
	})

	t.Run("repository not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const userID = "user-123"
		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, "non-existent").
			Return(nil, errors.New("repository not found"))

		req := &registryv1.GetCommitsRequest{Id: "non-existent"}
		_, err := svc.GetCommits(ctx, req)

		require.Error(t, err)
	})

	t.Run("user not member of organization", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const userID = "user-123"
		const orgID = "org-123"
		const repoID = "repo-123"

		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Path:           "./repos/repo-123",
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return("", errors.New("user is not a member"))

		req := &registryv1.GetCommitsRequest{Id: repoID}
		_, err := svc.GetCommits(ctx, req)

		require.Error(t, err)
	})

	t.Run("success with pagination", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const userID = "user-123"
		const orgID = "org-123"
		const repoID = "repo-123"
		repoPath := filepath.Join("./repos", repoID)

		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Path:           repoPath,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleReader, nil)

		expectedCommits := []*registryv1.Commit{
			{
				Id:      "commit-1",
				Message: "First commit",
				User: &registryv1.Commit_User{
					Id:       "user1@example.com",
					Username: "User 1",
				},
			},
			{
				Id:      "commit-2",
				Message: "Second commit",
				User: &registryv1.Commit_User{
					Id:       "user2@example.com",
					Username: "User 2",
				},
			},
		}

		mockRepo.EXPECT().
			GetCommits(ctx, repoPath, 2, 5).
			Return(expectedCommits, 12, nil)

		req := &registryv1.GetCommitsRequest{
			Id: repoID,
			Pagination: &shared.Pagination{
				Page:      2,
				PageLimit: 5,
			},
		}
		resp, err := svc.GetCommits(ctx, req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.GetCommits(), 2)
		assert.Equal(t, int32(3), resp.GetTotalPage()) // 12 commits / 5 per page = 3 pages
		assert.Equal(t, int32(3), resp.GetNextPage())  // page 2 < total pages, so next is 3
	})
}

func TestService_GetFileTree(t *testing.T) {
	t.Run("success - root directory", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const userID = "user-123"
		const orgID = "org-123"
		const repoID = "repo-123"
		repoPath := filepath.Join("./repos", repoID)

		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Path:           repoPath,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleReader, nil)

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
				},
			},
		}

		mockRepo.EXPECT().
			GetFileTree(ctx, repoPath, (*string)(nil)).
			Return(expectedFileTree, nil)

		req := &registryv1.GetFileTreeRequest{Id: repoID}
		resp, err := svc.GetFileTree(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.GetNodes(), 2)
		assert.Equal(t, "README.md", resp.GetNodes()[0].GetName())
		assert.Equal(t, "src", resp.GetNodes()[1].GetName())
	})

	t.Run("success - subdirectory", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const userID = "user-123"
		const orgID = "org-123"
		const repoID = "repo-123"
		repoPath := filepath.Join("./repos", repoID)
		subPath := "src"

		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Path:           repoPath,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleReader, nil)

		expectedFileTree := &registryv1.GetFileTreeResponse{
			Nodes: []*registryv1.FileTreeNode{
				{
					Name: "main.go",
					Path: "src/main.go",
					Type: registryv1.NodeType_NODE_TYPE_FILE,
				},
			},
		}

		mockRepo.EXPECT().
			GetFileTree(ctx, repoPath, &subPath).
			Return(expectedFileTree, nil)

		req := &registryv1.GetFileTreeRequest{
			Id:   repoID,
			Path: &subPath,
		}
		resp, err := svc.GetFileTree(ctx, req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.GetNodes(), 1)
		assert.Equal(t, "main.go", resp.GetNodes()[0].GetName())
	})

	t.Run("repository not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const userID = "user-123"
		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, "non-existent").
			Return(nil, errors.New("repository not found"))

		req := &registryv1.GetFileTreeRequest{Id: "non-existent"}
		_, err := svc.GetFileTree(ctx, req)

		require.Error(t, err)
	})

	t.Run("user not member of organization", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const userID = "user-123"
		const orgID = "org-123"
		const repoID = "repo-123"

		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Path:           "./repos/repo-123",
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return("", errors.New("user is not a member"))

		req := &registryv1.GetFileTreeRequest{Id: repoID}
		_, err := svc.GetFileTree(ctx, req)

		require.Error(t, err)
	})
}

func TestService_GetFilePreview(t *testing.T) {
	t.Run("success - text file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const userID = "user-123"
		const orgID = "org-123"
		const repoID = "repo-123"
		const filePath = "README.md"
		repoPath := filepath.Join("./repos", repoID)

		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Path:           repoPath,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleReader, nil)

		expectedPreview := &registryv1.GetFilePreviewResponse{
			Content:  "# Test Repository\n\nThis is a test.",
			MimeType: "text/markdown",
			Size:     35,
		}

		mockRepo.EXPECT().
			GetFilePreview(ctx, repoPath, filePath).
			Return(expectedPreview, nil)

		req := &registryv1.GetFilePreviewRequest{
			Id:   repoID,
			Path: filePath,
		}
		resp, err := svc.GetFilePreview(ctx, req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedPreview.GetContent(), resp.GetContent())
		assert.Equal(t, expectedPreview.GetMimeType(), resp.GetMimeType())
		assert.Equal(t, expectedPreview.GetSize(), resp.GetSize())
	})

	t.Run("success - go file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const userID = "user-123"
		const orgID = "org-123"
		const repoID = "repo-123"
		const filePath = "main.go"
		repoPath := filepath.Join("./repos", repoID)

		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Path:           repoPath,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleReader, nil)

		expectedPreview := &registryv1.GetFilePreviewResponse{
			Content:  "package main\n\nfunc main() {}",
			MimeType: "text/x-go",
			Size:     30,
		}

		mockRepo.EXPECT().
			GetFilePreview(ctx, repoPath, filePath).
			Return(expectedPreview, nil)

		req := &registryv1.GetFilePreviewRequest{
			Id:   repoID,
			Path: filePath,
		}
		resp, err := svc.GetFilePreview(ctx, req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedPreview.GetContent(), resp.GetContent())
		assert.Equal(t, expectedPreview.GetMimeType(), resp.GetMimeType())
	})

	t.Run("repository not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const userID = "user-123"
		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, "non-existent").
			Return(nil, errors.New("repository not found"))

		req := &registryv1.GetFilePreviewRequest{
			Id:   "non-existent",
			Path: "README.md",
		}
		_, err := svc.GetFilePreview(ctx, req)

		require.Error(t, err)
	})

	t.Run("user not member of organization", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const userID = "user-123"
		const orgID = "org-123"
		const repoID = "repo-123"

		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Path:           "./repos/repo-123",
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return("", errors.New("user is not a member"))

		req := &registryv1.GetFilePreviewRequest{
			Id:   repoID,
			Path: "README.md",
		}
		_, err := svc.GetFilePreview(ctx, req)

		require.Error(t, err)
	})

	t.Run("file not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const userID = "user-123"
		const orgID = "org-123"
		const repoID = "repo-123"
		const filePath = "nonexistent.txt"
		repoPath := filepath.Join("./repos", repoID)

		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Path:           repoPath,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleReader, nil)

		mockRepo.EXPECT().
			GetFilePreview(ctx, repoPath, filePath).
			Return(nil, errors.New("file not found in repository"))

		req := &registryv1.GetFilePreviewRequest{
			Id:   repoID,
			Path: filePath,
		}
		_, err := svc.GetFilePreview(ctx, req)

		require.Error(t, err)
		assert.ErrorContains(t, err, "file not found")
	})
}

func TestService_TriggerSdkGeneration(t *testing.T) {
	t.Run("success - enqueues jobs for enabled SDK preferences", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockQueue := NewMockSdkGenerationQueue(ctrl)

		svc := &service{
			repository: mockRepo,
			sdkQueue:   mockQueue,
		}

		ctx := context.Background()
		repoID := "repo-123"
		commitHash := "abc123def456"

		sdkPreferences := []SdkPreferencesDTO{
			{
				Id:           uuid.NewString(),
				RepositoryId: repoID,
				Sdk:          SdkGoProtobuf,
				Status:       true,
			},
			{
				Id:           uuid.NewString(),
				RepositoryId: repoID,
				Sdk:          SdkGoConnectRpc,
				Status:       true,
			},
			{
				Id:           uuid.NewString(),
				RepositoryId: repoID,
				Sdk:          SdkGoGrpc,
				Status:       false,
			},
		}

		mockRepo.EXPECT().
			GetSdkPreferences(ctx, repoID).
			Return(sdkPreferences, nil)

		mockQueue.EXPECT().
			EnqueueSdkGenerationJobs(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, jobs []*SdkGenerationJobDTO) error {
				assert.Len(t, jobs, 2)
				assert.Equal(t, repoID, jobs[0].RepositoryId)
				assert.Equal(t, commitHash, jobs[0].CommitHash)
				assert.Equal(t, SdkGenerationJobStatusPending, jobs[0].Status)
				assert.Equal(t, 0, jobs[0].Attempts)
				assert.Equal(t, 5, jobs[0].MaxAttempts)
				return nil
			})

		err := svc.TriggerSdkGeneration(ctx, repoID, commitHash)
		assert.NoError(t, err)
	})

	t.Run("success - no jobs when no SDK preferences enabled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockQueue := NewMockSdkGenerationQueue(ctrl)

		svc := &service{
			repository: mockRepo,
			sdkQueue:   mockQueue,
		}

		ctx := context.Background()
		repoID := "repo-123"
		commitHash := "abc123def456"

		sdkPreferences := []SdkPreferencesDTO{
			{
				Id:           uuid.NewString(),
				RepositoryId: repoID,
				Sdk:          SdkGoProtobuf,
				Status:       false,
			},
		}

		mockRepo.EXPECT().
			GetSdkPreferences(ctx, repoID).
			Return(sdkPreferences, nil)

		err := svc.TriggerSdkGeneration(ctx, repoID, commitHash)
		require.NoError(t, err)
	})

	t.Run("error - failed to get SDK preferences", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockQueue := NewMockSdkGenerationQueue(ctrl)

		svc := &service{
			repository: mockRepo,
			sdkQueue:   mockQueue,
		}

		ctx := context.Background()
		repoID := "repo-123"
		commitHash := "abc123def456"

		mockRepo.EXPECT().
			GetSdkPreferences(ctx, repoID).
			Return(nil, errors.New("database error"))

		err := svc.TriggerSdkGeneration(ctx, repoID, commitHash)
		require.Error(t, err)
	})

	t.Run("success - queue not configured", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)

		svc := &service{
			repository: mockRepo,
			sdkQueue:   nil,
		}

		ctx := context.Background()
		repoID := "repo-123"
		commitHash := "abc123def456"

		err := svc.TriggerSdkGeneration(ctx, repoID, commitHash)
		require.NoError(t, err)
	})
}

func TestService_GenerateSDK(t *testing.T) {
	t.Run("error - repository not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			sdkPath:    "./sdk",
		}

		ctx := context.Background()
		repoID := "repo-123"
		commitHash := "abc123"

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(nil, errors.New("repository not found"))

		err := svc.GenerateSDK(ctx, repoID, commitHash, SdkGoProtobuf)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch repository")
	})

	t.Run("error - no proto files found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		tmpDir := t.TempDir()
		sdkDir := t.TempDir()

		repoID := uuid.NewString()
		repoPath := filepath.Join(tmpDir, repoID)
		require.NoError(t, os.MkdirAll(repoPath, 0o750))

		// Initialize git repo with empty commit
		commitHash := initGitRepoWithEmptyCommit(t, repoPath)

		svc := &service{
			rootPath:    tmpDir,
			repository:  mockRepo,
			sdkPath:     sdkDir,
			sdkRegistry: sdkgenerator.NewRegistry(nil),
		}

		ctx := context.Background()
		orgID := "org-123"

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				OrganizationId: orgID,
				Path:           repoPath,
			}, nil)

		err := svc.GenerateSDK(ctx, repoID, commitHash, SdkGoProtobuf)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no proto files found")
	})

	t.Run("error - unsupported SDK type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		tmpDir := t.TempDir()
		sdkDir := t.TempDir()

		repoID := uuid.NewString()
		repoPath := filepath.Join(tmpDir, repoID)
		require.NoError(t, os.MkdirAll(repoPath, 0o750))

		commitHash := initGitRepoWithProtoFile(t, repoPath)

		cfg := &config.Config{
			SdkGeneration: config.SdkGenerationConfig{
				OutputPath: sdkDir,
			},
		}

		svc := NewService(mockRepo, nil, nil, cfg).(*service)
		svc.rootPath = tmpDir

		ctx := context.Background()
		orgID := "org-123"

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				OrganizationId: orgID,
				Path:           repoPath,
			}, nil)

		err := svc.GenerateSDK(ctx, repoID, commitHash, SDK("INVALID_SDK"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported SDK type")
	})
}

func TestService_FindProtoFiles(t *testing.T) {
	t.Run("finds proto files in nested directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "proto", "v1"), 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "proto", "user.proto"), []byte("syntax = \"proto3\";"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "proto", "v1", "api.proto"), []byte("syntax = \"proto3\";"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0o644))

		protoFiles, err := sdkgenerator.FindProtoFiles(tmpDir)
		require.NoError(t, err)
		assert.Len(t, protoFiles, 2)

		assert.Contains(t, protoFiles, filepath.Join("proto", "user.proto"))
		assert.Contains(t, protoFiles, filepath.Join("proto", "v1", "api.proto"))
	})

	t.Run("returns empty when no proto files", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0o644))

		protoFiles, err := sdkgenerator.FindProtoFiles(tmpDir)
		require.NoError(t, err)
		assert.Len(t, protoFiles, 0)
	})

	t.Run("error on invalid directory", func(t *testing.T) {
		_, err := sdkgenerator.FindProtoFiles("/nonexistent/directory")
		require.Error(t, err)
	})
}

func TestService_GetSdkDirName(t *testing.T) {
	tests := []struct {
		sdk      SDK
		expected string
	}{
		{SdkGoProtobuf, "go-protobuf"},
		{SdkGoConnectRpc, "go-connectrpc"},
		{SdkGoGrpc, "go-grpc"},
		{SdkJsBufbuildEs, "js-bufbuild-es"},
		{SdkJsProtobuf, "js-protobuf"},
		{SdkJsConnectrpc, "js-connectrpc"},
		{SDK("UNKNOWN"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.sdk), func(t *testing.T) {
			result := sdkgenerator.SDK(tt.sdk).DirName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestService_GenerateDocumentation(t *testing.T) {
	t.Run("error - no proto files found", func(t *testing.T) {
		tmpDir := t.TempDir()
		sdkDir := t.TempDir()

		svc := &service{
			sdkPath: sdkDir,
		}

		ctx := context.Background()
		repoID := "repo-123"
		commitHash := "abc123"
		orgID := "org-123"

		err := svc.GenerateDocumentation(ctx, repoID, commitHash, tmpDir, orgID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no proto files found")
	})

	t.Run("error - invalid directory", func(t *testing.T) {
		sdkDir := t.TempDir()

		svc := &service{
			sdkPath: sdkDir,
		}

		ctx := context.Background()
		repoID := "repo-123"
		commitHash := "abc123"
		orgID := "org-123"

		err := svc.GenerateDocumentation(ctx, repoID, commitHash, "/nonexistent/path", orgID)
		require.Error(t, err)
	})
}

func TestNewService_WithConfig(t *testing.T) {
	t.Run("uses default SDK path when config is nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)

		svc := NewService(mockRepo, nil, nil, nil)
		concrete, ok := svc.(*service)
		require.True(t, ok)
		assert.Equal(t, "./sdk", concrete.sdkPath)
	})

	t.Run("uses custom SDK path from config", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)

		cfg := &config.Config{
			SdkGeneration: config.SdkGenerationConfig{
				OutputPath: "/custom/sdk/path",
			},
		}

		svc := NewService(mockRepo, nil, nil, cfg)
		concrete, ok := svc.(*service)
		require.True(t, ok)
		assert.Equal(t, "/custom/sdk/path", concrete.sdkPath)
	})

	t.Run("initializes SDK registry", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)

		cfg := &config.Config{
			SdkGeneration: config.SdkGenerationConfig{
				OutputPath: "./sdk",
			},
		}

		svc := NewService(mockRepo, nil, nil, cfg)
		concrete, ok := svc.(*service)
		require.True(t, ok)

		require.NotNil(t, concrete.sdkRegistry)

		generators := concrete.sdkRegistry.List()
		assert.Len(t, generators, 6)

		_, err := concrete.sdkRegistry.Get(sdkgenerator.SdkGoProtobuf)
		assert.NoError(t, err)
		_, err = concrete.sdkRegistry.Get(sdkgenerator.SdkGoConnectRpc)
		assert.NoError(t, err)
		_, err = concrete.sdkRegistry.Get(sdkgenerator.SdkGoGrpc)
		assert.NoError(t, err)
		_, err = concrete.sdkRegistry.Get(sdkgenerator.SdkJsBufbuildEs)
		assert.NoError(t, err)
		_, err = concrete.sdkRegistry.Get(sdkgenerator.SdkJsProtobuf)
		assert.NoError(t, err)
		_, err = concrete.sdkRegistry.Get(sdkgenerator.SdkJsConnectrpc)
		assert.NoError(t, err)
	})

	t.Run("SDK DirName returns correct values", func(t *testing.T) {
		assert.Equal(t, "go-protobuf", sdkgenerator.SdkGoProtobuf.DirName())
		assert.Equal(t, "go-connectrpc", sdkgenerator.SdkGoConnectRpc.DirName())
		assert.Equal(t, "go-grpc", sdkgenerator.SdkGoGrpc.DirName())
		assert.Equal(t, "js-bufbuild-es", sdkgenerator.SdkJsBufbuildEs.DirName())
		assert.Equal(t, "js-protobuf", sdkgenerator.SdkJsProtobuf.DirName())
		assert.Equal(t, "js-connectrpc", sdkgenerator.SdkJsConnectrpc.DirName())
	})
}

func TestService_UpdateSdkPreferences(t *testing.T) {
	t.Run("success - updates preferences and enqueues trigger job", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)
		mockQueue := NewMockSdkGenerationQueue(ctrl)

		svc := &service{
			rootPath:   t.TempDir(),
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
			sdkQueue:   mockQueue,
		}

		const repoID = "repo-123"
		const orgID = "org-123"
		const userID = "user-123"
		const repoPath = "./repos/repo-123"
		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Path:           repoPath,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleOwner, nil)

		mockRepo.EXPECT().
			UpdateSdkPreferences(ctx, repoID, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, prefs []SdkPreferencesDTO) error {
				assert.Len(t, prefs, 2)
				return nil
			})

		mockQueue.EXPECT().
			EnqueueSdkTriggerJob(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, job *SdkTriggerJobDTO) error {
				assert.Equal(t, repoID, job.RepositoryId)
				assert.Equal(t, repoPath, job.RepoPath)
				assert.Equal(t, SdkGenerationJobStatusPending, job.Status)
				return nil
			})

		err := svc.UpdateSdkPreferences(ctx, &registryv1.UpdateSdkPreferencesRequest{
			Id: repoID,
			SdkPreferences: []*registryv1.SdkPreference{
				{Sdk: registryv1.SDK_SDK_GO_PROTOBUF, Status: true},
				{Sdk: registryv1.SDK_SDK_GO_CONNECTRPC, Status: true},
			},
		})
		require.NoError(t, err)
	})

	t.Run("success - no trigger job when queue is nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   t.TempDir(),
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
			sdkQueue:   nil,
		}

		const repoID = "repo-123"
		const orgID = "org-123"
		const userID = "user-123"
		const repoPath = "./repos/repo-123"
		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
				Path:           repoPath,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleOwner, nil)

		mockRepo.EXPECT().
			UpdateSdkPreferences(ctx, repoID, gomock.Any()).
			Return(nil)

		err := svc.UpdateSdkPreferences(ctx, &registryv1.UpdateSdkPreferencesRequest{
			Id: repoID,
			SdkPreferences: []*registryv1.SdkPreference{
				{Sdk: registryv1.SDK_SDK_GO_PROTOBUF, Status: true},
			},
		})
		require.NoError(t, err)
	})

	t.Run("error - repository not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   t.TempDir(),
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const repoID = "repo-123"
		const userID = "user-123"
		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(nil, ErrRepositoryNotFound)

		err := svc.UpdateSdkPreferences(ctx, &registryv1.UpdateSdkPreferencesRequest{
			Id: repoID,
			SdkPreferences: []*registryv1.SdkPreference{
				{Sdk: registryv1.SDK_SDK_GO_PROTOBUF, Status: true},
			},
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "repository not found")
	})

	t.Run("error - user not owner", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := &service{
			rootPath:   t.TempDir(),
			repository: mockRepo,
			orgRepo:    mockOrgRepo,
		}

		const repoID = "repo-123"
		const orgID = "org-123"
		const userID = "user-123"
		ctx := testAuthInterceptor(userID)

		mockRepo.EXPECT().
			GetRepositoryById(ctx, repoID).
			Return(&RepositoryDTO{
				Id:             repoID,
				Name:           "test-repo",
				OrganizationId: orgID,
			}, nil)

		mockOrgRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(authorization.MemberRoleReader, nil)

		err := svc.UpdateSdkPreferences(ctx, &registryv1.UpdateSdkPreferencesRequest{
			Id: repoID,
			SdkPreferences: []*registryv1.SdkPreference{
				{Sdk: registryv1.SDK_SDK_GO_PROTOBUF, Status: true},
			},
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "only organization owners can perform this operation")
	})
}

func TestService_ProcessSdkTrigger(t *testing.T) {
	t.Run("success - creates SDK generation jobs for all commits", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockQueue := NewMockSdkGenerationQueue(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			sdkQueue:   mockQueue,
		}

		ctx := context.Background()
		const repoID = "repo-123"
		const repoPath = "./repos/repo-123"

		mockRepo.EXPECT().
			GetSdkPreferences(ctx, repoID).
			Return([]SdkPreferencesDTO{
				{Id: "pref-1", RepositoryId: repoID, Sdk: SdkGoProtobuf, Status: true},
				{Id: "pref-2", RepositoryId: repoID, Sdk: SdkGoConnectRpc, Status: true},
			}, nil)

		mockRepo.EXPECT().
			GetCommits(ctx, repoPath, 1, 10000).
			Return([]*registryv1.Commit{
				{Id: "commit-1", Message: "First commit"},
				{Id: "commit-2", Message: "Second commit"},
			}, 2, nil)

		mockQueue.EXPECT().
			EnqueueSdkGenerationJobs(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, jobs []*SdkGenerationJobDTO) error {
				require.Len(t, jobs, 4)
				return nil
			})

		err := svc.ProcessSdkTrigger(ctx, repoID, repoPath)
		require.NoError(t, err)
	})

	t.Run("success - no jobs when no commits exist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockQueue := NewMockSdkGenerationQueue(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			sdkQueue:   mockQueue,
		}

		ctx := context.Background()
		const repoID = "repo-123"
		const repoPath = "./repos/repo-123"

		mockRepo.EXPECT().
			GetSdkPreferences(ctx, repoID).
			Return([]SdkPreferencesDTO{
				{Id: "pref-1", RepositoryId: repoID, Sdk: SdkGoProtobuf, Status: true},
			}, nil)

		mockRepo.EXPECT().
			GetCommits(ctx, repoPath, 1, 10000).
			Return([]*registryv1.Commit{}, 0, nil)

		err := svc.ProcessSdkTrigger(ctx, repoID, repoPath)
		require.NoError(t, err)
	})

	t.Run("success - no jobs when all preferences disabled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockQueue := NewMockSdkGenerationQueue(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			sdkQueue:   mockQueue,
		}

		ctx := context.Background()
		const repoID = "repo-123"
		const repoPath = "./repos/repo-123"

		mockRepo.EXPECT().
			GetSdkPreferences(ctx, repoID).
			Return([]SdkPreferencesDTO{
				{Id: "pref-1", RepositoryId: repoID, Sdk: SdkGoProtobuf, Status: false},
			}, nil)

		err := svc.ProcessSdkTrigger(ctx, repoID, repoPath)
		require.NoError(t, err)
	})

	t.Run("error - failed to get SDK preferences", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockQueue := NewMockSdkGenerationQueue(ctrl)

		svc := &service{
			rootPath:   "./repos",
			repository: mockRepo,
			sdkQueue:   mockQueue,
		}

		ctx := context.Background()
		const repoID = "repo-123"
		const repoPath = "./repos/repo-123"

		mockRepo.EXPECT().
			GetSdkPreferences(ctx, repoID).
			Return(nil, errors.New("database error"))

		err := svc.ProcessSdkTrigger(ctx, repoID, repoPath)
		require.Error(t, err)
	})
}

func TestSdkGenerationJobDTO(t *testing.T) {
	t.Run("creates valid DTO", func(t *testing.T) {
		now := time.Now().UTC()
		dto := &SdkGenerationJobDTO{
			Id:           uuid.NewString(),
			RepositoryId: uuid.NewString(),
			CommitHash:   "abc123def456",
			Sdk:          SdkGoProtobuf,
			Status:       SdkGenerationJobStatusPending,
			Attempts:     0,
			MaxAttempts:  5,
			CreatedAt:    now,
		}

		assert.NotEmpty(t, dto.Id)
		assert.NotEmpty(t, dto.RepositoryId)
		assert.Equal(t, "abc123def456", dto.CommitHash)
		assert.Equal(t, SdkGoProtobuf, dto.Sdk)
		assert.Equal(t, SdkGenerationJobStatusPending, dto.Status)
		assert.Equal(t, 0, dto.Attempts)
		assert.Equal(t, 5, dto.MaxAttempts)
		assert.Equal(t, now, dto.CreatedAt)
	})
}
