package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/google/uuid"
	"go.uber.org/zap"

	repositoryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/repository/v1"
)

const defaultReposPath = "./repos"

type Service interface {
	CreateRepository(ctx context.Context, req *repositoryv1.CreateRepositoryRequest) error
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
	req *repositoryv1.CreateRepositoryRequest,
) error {
	repoName := req.GetName()
	if repoName == "" {
		return fmt.Errorf("repository name is required")
	}

	existingRepo, err := s.repository.GetRepositoryByName(ctx, repoName)
	if err != nil && !errors.Is(err, ErrRepositoryNotFound) {
		return fmt.Errorf("failed to check existing repository: %w", err)
	}
	if existingRepo != nil {
		return fmt.Errorf("repository %q already exists", repoName)
	}

	repoPath := filepath.Join(s.rootPath, repoName)

	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	fs := osfs.New(repoPath)
	storage := filesystem.NewStorage(fs, nil)

	_, err = git.Init(storage, fs)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryAlreadyExists) {
			zap.L().Warn("repository already exists on filesystem", zap.String("path", repoPath))
			return fmt.Errorf("repository %q already exists", repoName)
		}

		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	now := time.Now().UTC()
	repoDTO := &RepositoryDTO{
		Id:        uuid.NewString(),
		Name:      repoName,
		Path:      repoPath,
		CreatedAt: now,
		UpdatedAt: now,
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
	)

	return nil
}
