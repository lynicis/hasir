package registry

import (
	"context"
	"errors"
	"html"
	"os"
	"path/filepath"
	"time"

	"connectrpc.com/connect"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"hasir/api/pkg/authentication"
	"hasir/api/pkg/authorization"
	"hasir/api/pkg/config"
	"hasir/api/pkg/pagination"
	"hasir/api/pkg/proto"
	"hasir/api/pkg/sdkgenerator"
	registryv1 "hasir/proto/gen/go/registry/v1"
)

const DefaultReposPath = "./repos"

type Service interface {
	SdkGenerator
	SdkTriggerProcessor
	CreateRepository(ctx context.Context, req *registryv1.CreateRepositoryRequest) error
	GetRepository(ctx context.Context, req *registryv1.GetRepositoryRequest) (*registryv1.Repository, error)
	GetRepositories(ctx context.Context, organizationId *string, page, pageSize int) (*registryv1.GetRepositoriesResponse, error)
	UpdateRepository(ctx context.Context, req *registryv1.UpdateRepositoryRequest) error
	DeleteRepository(ctx context.Context, req *registryv1.DeleteRepositoryRequest) error
	DeleteRepositoriesByOrganization(ctx context.Context, organizationId string) error
	UpdateSdkPreferences(ctx context.Context, req *registryv1.UpdateSdkPreferencesRequest) error
	GetCommits(ctx context.Context, req *registryv1.GetCommitsRequest) (*registryv1.GetCommitsResponse, error)
	GetRecentCommit(ctx context.Context, req *registryv1.GetRecentCommitRequest) (*registryv1.Commit, error)
	GetFileTree(ctx context.Context, req *registryv1.GetFileTreeRequest) (*registryv1.GetFileTreeResponse, error)
	GetFilePreview(ctx context.Context, req *registryv1.GetFilePreviewRequest) (*registryv1.GetFilePreviewResponse, error)
	ValidateSshAccess(ctx context.Context, userId, repoPath string, operation SshOperation) (bool, error)
	HasProtoFiles(ctx context.Context, repoPath, commitHash string) (bool, error)
	TriggerSdkGeneration(ctx context.Context, repositoryId, commitHash string) error
}

type service struct {
	rootPath     string
	repository   Repository
	orgRepo      authorization.MemberRoleChecker
	sdkQueue     SdkGenerationQueue
	cfg          *config.Config
	sdkPath      string
	runner       sdkgenerator.CommandRunner
	sdkRegistry  *sdkgenerator.Registry
	bufRegistry  *sdkgenerator.Registry
	docGenerator *sdkgenerator.DocumentationGenerator
}

func NewService(repository Repository, orgRepo authorization.MemberRoleChecker, sdkQueue SdkGenerationQueue, cfg *config.Config) Service {
	sdkPath := "./sdk"
	if cfg != nil && cfg.SdkGeneration.OutputPath != "" {
		sdkPath = cfg.SdkGeneration.OutputPath
	}

	runner := sdkgenerator.NewDefaultCommandRunner()
	return &service{
		rootPath:     DefaultReposPath,
		repository:   repository,
		orgRepo:      orgRepo,
		sdkQueue:     sdkQueue,
		cfg:          cfg,
		sdkPath:      sdkPath,
		runner:       runner,
		sdkRegistry:  sdkgenerator.NewRegistry(runner),
		bufRegistry:  sdkgenerator.NewBufRegistry(runner),
		docGenerator: sdkgenerator.NewDocumentationGenerator(runner),
	}
}

func (s *service) commandRunner() sdkgenerator.CommandRunner {
	if s.runner != nil {
		return s.runner
	}
	return sdkgenerator.NewDefaultCommandRunner()
}

func (s *service) CreateRepository(
	ctx context.Context,
	req *registryv1.CreateRepositoryRequest,
) error {
	repoName := html.EscapeString(req.GetName())
	organizationId := req.GetOrganizationId()

	visibility, ok := proto.VisibilityMap[req.GetVisibility()]
	if !ok {
		visibility = proto.VisibilityPrivate
	}

	createdBy, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return err
	}

	if err := authorization.IsUserOwner(ctx, s.orgRepo, organizationId, createdBy); err != nil {
		return err
	}

	repoId := uuid.NewString()
	repoPath := filepath.Join(s.rootPath, repoId)

	if err := os.MkdirAll(repoPath, 0o750); err != nil {
		return connect.NewError(connect.CodeInternal, errors.New("failed to create repository directory"))
	}

	gitRepo, err := git.PlainInit(repoPath, true)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryAlreadyExists) {
			zap.L().Warn("repository already exists on filesystem", zap.String("path", repoPath))
			return connect.NewError(connect.CodeAlreadyExists, errors.New("repository path already exists"))
		}

		return connect.NewError(connect.CodeInternal, errors.New("failed to initialize git repository"))
	}
	_ = gitRepo.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.ReferenceName("refs/heads/main")))

	now := time.Now().UTC()
	repoDTO := &RepositoryDTO{
		Id:             repoId,
		Name:           repoName,
		CreatedBy:      createdBy,
		OrganizationId: organizationId,
		Path:           repoPath,
		Visibility:     visibility,
		ManagedByBuf:   req.GetManagedByBuf(),
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

		return connect.NewError(connect.CodeInternal, errors.New("failed to save repository to database"))
	}

	zap.L().Info("git repository created and synced with database",
		zap.String("id", repoDTO.Id),
		zap.String("name", repoName),
		zap.String("path", repoPath),
		zap.String("organizationId", organizationId),
	)

	return nil
}

func (s *service) GetRepository(
	ctx context.Context,
	req *registryv1.GetRepositoryRequest,
) (*registryv1.Repository, error) {
	repoId := req.GetId()

	repo, err := s.repository.GetRepositoryById(ctx, repoId)
	if err != nil {
		return nil, err
	}

	userId, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := authorization.IsUserMember(ctx, s.orgRepo, repo.OrganizationId, userId); err != nil {
		return nil, err
	}

	sdkPreferences, err := s.repository.GetSdkPreferences(ctx, repoId)
	if err != nil {
		return nil, err
	}

	return &registryv1.Repository{
		Id:             repo.Id,
		Name:           repo.Name,
		OrganizationId: repo.OrganizationId,
		Visibility:     proto.ReverseVisibilityMap[repo.Visibility],
		SdkPreferences: toProtoSdkPreferences(sdkPreferences),
		ManagedByBuf:   repo.ManagedByBuf,
	}, nil
}

func (s *service) GetRepositories(
	ctx context.Context,
	organizationId *string,
	page, pageSize int,
) (*registryv1.GetRepositoriesResponse, error) {
	userId, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	var totalCount int
	var repositories *[]RepositoryDTO

	if organizationId != nil && *organizationId != "" {
		if err := authorization.IsUserMember(ctx, s.orgRepo, *organizationId, userId); err != nil {
			return nil, err
		}

		totalCount, err = s.repository.GetRepositoriesByUserAndOrganizationCount(ctx, userId, *organizationId)
		if err != nil {
			return nil, err
		}

		repositories, err = s.repository.GetRepositoriesByUserAndOrganization(ctx, userId, *organizationId, page, pageSize)
		if err != nil {
			return nil, err
		}
	} else {
		totalCount, err = s.repository.GetRepositoriesByUserCount(ctx, userId)
		if err != nil {
			return nil, err
		}

		repositories, err = s.repository.GetRepositoriesByUser(ctx, userId, page, pageSize)
		if err != nil {
			return nil, err
		}
	}

	repoIds := make([]string, 0, len(*repositories))
	for _, repo := range *repositories {
		repoIds = append(repoIds, repo.Id)
	}

	var sdkPrefsMap map[string][]SdkPreferencesDTO
	if len(repoIds) > 0 {
		sdkPrefsMap, err = s.repository.GetSdkPreferencesByRepositoryIds(ctx, repoIds)
		if err != nil {
			return nil, err
		}
	}

	var resp []*registryv1.Repository
	for _, repository := range *repositories {
		resp = append(resp, &registryv1.Repository{
			Id:             repository.Id,
			Name:           repository.Name,
			Visibility:     proto.ReverseVisibilityMap[repository.Visibility],
			SdkPreferences: toProtoSdkPreferences(sdkPrefsMap[repository.Id]),
			ManagedByBuf:   repository.ManagedByBuf,
		})
	}

	nextPage, totalPages, err := pagination.ComputePageMeta(totalCount, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &registryv1.GetRepositoriesResponse{
		Repositories: resp,
		NextPage:     nextPage,
		TotalPage:    totalPages,
	}, nil
}

func toProtoSdkPreferences(prefs []SdkPreferencesDTO) []*registryv1.SdkPreference {
	var result []*registryv1.SdkPreference
	for _, pref := range prefs {
		result = append(result, &registryv1.SdkPreference{
			Sdk:    SdkDbToProtoEnum[pref.Sdk],
			Status: pref.Status,
		})
	}
	return result
}

func (s *service) UpdateRepository(
	ctx context.Context,
	req *registryv1.UpdateRepositoryRequest,
) error {
	repoId := req.GetId()
	repo, err := s.repository.GetRepositoryById(ctx, repoId)
	if err != nil {
		return err
	}

	userId, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return err
	}

	if err := authorization.IsUserOwner(ctx, s.orgRepo, repo.OrganizationId, userId); err != nil {
		return err
	}

	repo.Name = html.EscapeString(req.GetName())
	repo.Visibility = proto.VisibilityMap[req.GetVisibility()]
	repo.ManagedByBuf = req.GetManagedByBuf()

	if err := s.repository.UpdateRepository(ctx, repo); err != nil {
		return err
	}

	zap.L().Info("repository updated",
		zap.String("id", repoId),
		zap.String("name", repo.Name),
		zap.String("visibility", string(repo.Visibility)),
	)

	return nil
}

func (s *service) DeleteRepository(
	ctx context.Context,
	req *registryv1.DeleteRepositoryRequest,
) error {
	repoId := req.GetRepositoryId()

	repo, err := s.repository.GetRepositoryById(ctx, repoId)
	if err != nil {
		return err
	}

	userId, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return err
	}

	if err := authorization.IsUserOwner(ctx, s.orgRepo, repo.OrganizationId, userId); err != nil {
		return err
	}

	if err := s.repository.DeleteRepository(ctx, repoId); err != nil {
		return err
	}

	if err := os.RemoveAll(repo.Path); err != nil {
		zap.L().Error("failed to remove repository directory after database deletion",
			zap.String("id", repoId),
			zap.String("path", repo.Path),
			zap.Error(err),
		)

		return connect.NewError(connect.CodeInternal, errors.New("failed to remove repository directory"))
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
		return err
	}

	if err := s.repository.DeleteRepositoriesByOrganizationId(ctx, organizationId); err != nil {
		return err
	}

	errGroup := &errgroup.Group{}

	for _, repo := range *repos {
		errGroup.Go(func() error {
			if err := os.RemoveAll(repo.Path); err != nil {
				zap.L().Error("failed to remove repository directory after database deletion",
					zap.String("id", repo.Id),
					zap.String("path", repo.Path),
					zap.Error(err),
				)

				return connect.NewError(connect.CodeInternal, errors.New("failed to remove repository directory"))
			}

			zap.L().Info("repository deleted as part of organization deletion",
				zap.String("id", repo.Id),
				zap.String("name", repo.Name),
				zap.String("path", repo.Path),
				zap.String("organizationId", organizationId),
			)

			return nil
		})
	}

	return errGroup.Wait()
}
