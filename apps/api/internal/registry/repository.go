package registry

import "context"

type Repository interface {
	CreateRepository(ctx context.Context, repo *RepositoryDTO) error
	GetRepositoryByName(ctx context.Context, name string) (*RepositoryDTO, error)
	GetRepositoryById(ctx context.Context, id string) (*RepositoryDTO, error)
	GetRepositories(ctx context.Context, page, pageSize int) (*[]RepositoryDTO, error)
	GetRepositoriesByOrganizationId(ctx context.Context, organizationId string) (*[]RepositoryDTO, error)
	GetRepositoriesCount(ctx context.Context) (int, error)
	GetRepositoriesByUser(ctx context.Context, userId string, page, pageSize int) (*[]RepositoryDTO, error)
	GetRepositoriesByUserCount(ctx context.Context, userId string) (int, error)
	UpdateRepository(ctx context.Context, repo *RepositoryDTO) error
	DeleteRepository(ctx context.Context, id string) error
	UpdateSdkPreferences(ctx context.Context, repositoryId string, preferences []SdkPreferencesDTO) error
	GetSdkPreferences(ctx context.Context, repositoryId string) ([]SdkPreferencesDTO, error)
}
