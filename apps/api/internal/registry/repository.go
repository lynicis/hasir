package registry

import (
	"context"

	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
)

type Repository interface {
	CreateRepository(ctx context.Context, repo *RepositoryDTO) error
	GetRepositoryByName(ctx context.Context, name string) (*RepositoryDTO, error)
	GetRepositoryById(ctx context.Context, id string) (*RepositoryDTO, error)
	GetRepositories(ctx context.Context, page, pageSize int) (*[]RepositoryDTO, error)
	GetRepositoriesByOrganizationId(ctx context.Context, organizationId string) (*[]RepositoryDTO, error)
	GetRepositoriesByUser(ctx context.Context, userId string, page, pageSize int) (*[]RepositoryDTO, error)
	GetRepositoriesByUserCount(ctx context.Context, userId string) (int, error)
	GetRepositoriesByUserAndOrganization(ctx context.Context, userId, organizationId string, page, pageSize int) (*[]RepositoryDTO, error)
	GetRepositoriesByUserAndOrganizationCount(ctx context.Context, userId, organizationId string) (int, error)
	UpdateRepository(ctx context.Context, repo *RepositoryDTO) error
	DeleteRepository(ctx context.Context, id string) error
	DeleteRepositoriesByOrganizationId(ctx context.Context, organizationId string) error
	UpdateSdkPreferences(ctx context.Context, repositoryId string, preferences []SdkPreferencesDTO) error
	GetSdkPreferences(ctx context.Context, repositoryId string) ([]SdkPreferencesDTO, error)
	GetSdkPreferencesByRepositoryIds(ctx context.Context, repositoryIds []string) (map[string][]SdkPreferencesDTO, error)
	GetCommits(ctx context.Context, repoPath string) (*registryv1.GetCommitsResponse, error)
	GetFileTree(ctx context.Context, repoPath string, subPath *string) (*registryv1.GetFileTreeResponse, error)
	GetFilePreview(ctx context.Context, repoPath, filePath string) (*registryv1.GetFilePreviewResponse, error)
}
