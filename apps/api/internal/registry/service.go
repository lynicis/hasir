package registry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"hasir-api/pkg/auth"

	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
)

const defaultReposPath = "./repos"

type Service interface {
	CreateRepository(ctx context.Context, req *registryv1.CreateRepositoryRequest) error
	GetRepositories(ctx context.Context, page, pageSize int) (*registryv1.GetRepositoriesResponse, error)
	DeleteRepository(ctx context.Context, req *registryv1.DeleteRepositoryRequest) error
	DeleteRepositoriesByOrganization(ctx context.Context, organizationId string) error
}

type service struct {
	rootPath   string
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{
		rootPath:   defaultReposPath,
		repository: repository,
	}
}

func (s *service) CreateRepository(
	ctx context.Context,
	req *registryv1.CreateRepositoryRequest,
) error {
	repoName := req.GetName()
	organizationId := req.GetOrganizationId()

	createdBy, err := auth.MustGetUserID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get repository owner from context: %w", err)
	}

	repoId := uuid.NewString()
	repoPath := filepath.Join(s.rootPath, repoId)

	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	_, err = git.PlainInit(repoPath, true)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryAlreadyExists) {
			zap.L().Warn("repository already exists on filesystem", zap.String("path", repoPath))
			return fmt.Errorf("repository path already exists")
		}

		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	now := time.Now().UTC()
	repoDTO := &RepositoryDTO{
		Id:             repoId,
		Name:           repoName,
		CreatedBy:      createdBy,
		OrganizationId: organizationId,
		Path:           repoPath,
		CreatedAt:      now,
		UpdatedAt:      &now,
	}

	if err := s.repository.CreateRepository(ctx, repoDTO); err != nil {
		if removeErr := os.RemoveAll(repoPath); removeErr != nil {
			zap.L().Error("failed to rollback git repository after db error",
				zap.String("path", repoPath),
				zap.Error(removeErr),
			)
		}

		return fmt.Errorf("failed to save repository to database: %w", err)
	}

	zap.L().Info("git repository created and synced with database",
		zap.String("id", repoDTO.Id),
		zap.String("name", repoName),
		zap.String("path", repoPath),
		zap.String("organizationId", organizationId),
	)

	return nil
}

func (s *service) GetRepositories(
	ctx context.Context,
	page, pageSize int,
) (*registryv1.GetRepositoriesResponse, error) {
	totalCount, err := s.repository.GetRepositoriesCount(ctx)
	if err != nil {
		return nil, err
	}

	repositories, err := s.repository.GetRepositories(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	var resp []*registryv1.Repository
	for _, repository := range *repositories {
		resp = append(resp, &registryv1.Repository{
			Id:   repository.Id,
			Name: repository.Name,
		})
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	nextPage := int32(page + 1)
	if page >= totalPages {
		nextPage = 0
	}

	return &registryv1.GetRepositoriesResponse{
		Repositories: resp,
		NextPage:     nextPage,
		TotalPage:    int32(totalPages),
	}, nil
}

func (s *service) DeleteRepository(
	ctx context.Context,
	req *registryv1.DeleteRepositoryRequest,
) error {
	repoId := req.GetRepositoryId()

	repo, err := s.repository.GetRepositoryById(ctx, repoId)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	if err := s.repository.DeleteRepository(ctx, repoId); err != nil {
		return fmt.Errorf("failed to delete repository from database: %w", err)
	}

	if err := os.RemoveAll(repo.Path); err != nil {
		zap.L().Error("failed to remove repository directory after database deletion",
			zap.String("id", repoId),
			zap.String("path", repo.Path),
			zap.Error(err),
		)

		return fmt.Errorf("failed to remove repository directory: %w", err)
	}

	zap.L().Info("repository deleted",
		zap.String("id", repoId),
		zap.String("name", repo.Name),
		zap.String("path", repo.Path),
	)

	return nil
}

func (s *service) DeleteRepositoriesByOrganization(
	ctx context.Context,
	organizationId string,
) error {
	repos, err := s.repository.GetRepositoriesByOrganizationId(ctx, organizationId)
	if err != nil {
		return fmt.Errorf("failed to get repositories for organization: %w", err)
	}

	for _, repo := range *repos {
		if err := s.repository.DeleteRepository(ctx, repo.Id); err != nil {
			return fmt.Errorf("failed to delete repository %s from database: %w", repo.Id, err)
		}

		if err := os.RemoveAll(repo.Path); err != nil {
			zap.L().Error("failed to remove repository directory after database deletion",
				zap.String("id", repo.Id),
				zap.String("path", repo.Path),
				zap.Error(err),
			)

			return fmt.Errorf("failed to remove repository directory for %s: %w", repo.Id, err)
		}

		zap.L().Info("repository deleted as part of organization deletion",
			zap.String("id", repo.Id),
			zap.String("name", repo.Name),
			zap.String("path", repo.Path),
			zap.String("organizationId", organizationId),
		)
	}

	return nil
}
