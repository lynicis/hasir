package registry

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"hasir-api/pkg/authentication"
	"hasir-api/pkg/authorization"
	"hasir-api/pkg/proto"

	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
	"buf.build/gen/go/hasir/hasir/protocolbuffers/go/shared"
)

func TestNewService(t *testing.T) {
	t.Run("default root path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := NewMockRepository(ctrl)
		mockOrgRepo := authorization.NewMockMemberRoleChecker(ctrl)

		svc := NewService(mockRepo, mockOrgRepo)
		concrete, ok := svc.(*service)
		require.True(t, ok, "NewService should return *service")
		require.Equal(t, DefaultReposPath, concrete.rootPath)
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
				require.Equal(t, repoName, repo.Name)
				require.Equal(t, filepath.Join(tmpDir, repo.Id), repo.Path)
				require.Equal(t, userID, repo.CreatedBy)
				require.Equal(t, proto.VisibilityPrivate, repo.Visibility)
				require.NotEmpty(t, repo.Id)
				return nil
			})

		err := svc.CreateRepository(ctx, &registryv1.CreateRepositoryRequest{
			Name:           repoName,
			OrganizationId: orgID,
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
				require.Equal(t, repoName, repo.Name)
				require.Equal(t, filepath.Join(tmpDir, repo.Id), repo.Path)
				require.Equal(t, userID, repo.CreatedBy)
				require.Equal(t, proto.VisibilityPublic, repo.Visibility)
				require.NotEmpty(t, repo.Id)
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
		require.NoDirExists(t, repoPath)
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
		require.NoError(t, err)
		require.NotNil(t, repo)
		require.Equal(t, repoID, repo.GetId())
		require.Equal(t, "test-repo", repo.GetName())
		require.Equal(t, shared.Visibility_VISIBILITY_PRIVATE, repo.GetVisibility())
		require.Len(t, repo.GetSdkPreferences(), 1)
		require.Equal(t, registryv1.SDK_SDK_GO_CONNECTRPC, repo.GetSdkPreferences()[0].GetSdk())
		require.True(t, repo.GetSdkPreferences()[0].GetStatus())
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
		require.Error(t, err)
		require.Nil(t, repo)
		require.ErrorContains(t, err, "repository not found")
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
		require.Error(t, err)
		require.Nil(t, repo)
		require.ErrorContains(t, err, "you are not a member of this organization")
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
		require.Error(t, err)
		require.Nil(t, repo)
		require.ErrorContains(t, err, "user not authenticated")
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
		require.NoError(t, err)
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
		require.NoError(t, err)

		require.NoDirExists(t, repoPath)
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
		require.ErrorContains(t, err, "repository not found")
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
		require.ErrorContains(t, err, "database delete failed")
		require.ErrorIs(t, err, dbErr)

		require.DirExists(t, repoPath)
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
		require.Len(t, resp.GetRepositories(), 2)
		require.Equal(t, "repo-1", resp.GetRepositories()[0].GetId())
		require.Equal(t, "first-repo", resp.GetRepositories()[0].GetName())
		require.Len(t, resp.GetRepositories()[0].GetSdkPreferences(), 1)
		require.Equal(t, "repo-2", resp.GetRepositories()[1].GetId())
		require.Equal(t, "second-repo", resp.GetRepositories()[1].GetName())
		require.Len(t, resp.GetRepositories()[1].GetSdkPreferences(), 0)
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
		}

		mockRepo.EXPECT().
			GetCommits(ctx, repoPath).
			Return(expectedCommits, nil)

		req := &registryv1.GetCommitsRequest{Id: repoID}
		resp, err := svc.GetCommits(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.GetCommits(), 1)
		require.Equal(t, "abc123", resp.GetCommits()[0].GetId())
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

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.GetNodes(), 2)
		require.Equal(t, "README.md", resp.GetNodes()[0].GetName())
		require.Equal(t, "src", resp.GetNodes()[1].GetName())
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
		require.NotNil(t, resp)
		require.Len(t, resp.GetNodes(), 1)
		require.Equal(t, "main.go", resp.GetNodes()[0].GetName())
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
		require.NotNil(t, resp)
		require.Equal(t, expectedPreview.GetContent(), resp.GetContent())
		require.Equal(t, expectedPreview.GetMimeType(), resp.GetMimeType())
		require.Equal(t, expectedPreview.GetSize(), resp.GetSize())
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
		require.NotNil(t, resp)
		require.Equal(t, expectedPreview.GetContent(), resp.GetContent())
		require.Equal(t, expectedPreview.GetMimeType(), resp.GetMimeType())
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
		require.ErrorContains(t, err, "file not found")
	})
}
