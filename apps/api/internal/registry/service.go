package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
	"connectrpc.com/connect"
	"github.com/go-git/go-git/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"hasir-api/pkg/authentication"
	"hasir-api/pkg/authorization"
	"hasir-api/pkg/config"
	"hasir-api/pkg/proto"
	"hasir-api/pkg/sdkgenerator"
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
	HasProtoFiles(ctx context.Context, repoPath string) (bool, error)
	TriggerSdkGeneration(ctx context.Context, repositoryId, commitHash string) error
}

type service struct {
	rootPath    string
	repository  Repository
	orgRepo     authorization.MemberRoleChecker
	sdkQueue    SdkGenerationQueue
	cfg         *config.Config
	sdkPath     string
	sdkRegistry *sdkgenerator.Registry
}

func NewService(repository Repository, orgRepo authorization.MemberRoleChecker, sdkQueue SdkGenerationQueue, cfg *config.Config) Service {
	sdkPath := "./sdk"
	if cfg != nil && cfg.SdkGeneration.OutputPath != "" {
		sdkPath = cfg.SdkGeneration.OutputPath
	}

	return &service{
		rootPath:    DefaultReposPath,
		repository:  repository,
		orgRepo:     orgRepo,
		sdkQueue:    sdkQueue,
		cfg:         cfg,
		sdkPath:     sdkPath,
		sdkRegistry: sdkgenerator.NewRegistry(nil),
	}
}

func (s *service) CreateRepository(
	ctx context.Context,
	req *registryv1.CreateRepositoryRequest,
) error {
	repoName := req.GetName()
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

	_, err = git.PlainInit(repoPath, true)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryAlreadyExists) {
			zap.L().Warn("repository already exists on filesystem", zap.String("path", repoPath))
			return connect.NewError(connect.CodeAlreadyExists, errors.New("repository path already exists"))
		}

		return connect.NewError(connect.CodeInternal, errors.New("failed to initialize git repository"))
	}

	now := time.Now().UTC()
	repoDTO := &RepositoryDTO{
		Id:             repoId,
		Name:           repoName,
		CreatedBy:      createdBy,
		OrganizationId: organizationId,
		Path:           repoPath,
		Visibility:     visibility,
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

	var protoSdkPreferences []*registryv1.SdkPreference
	for _, pref := range sdkPreferences {
		protoSdkPreferences = append(protoSdkPreferences, &registryv1.SdkPreference{
			Sdk:    SdkDbToProtoEnum[pref.Sdk],
			Status: pref.Status,
		})
	}

	return &registryv1.Repository{
		Id:             repo.Id,
		Name:           repo.Name,
		OrganizationId: repo.OrganizationId,
		Visibility:     proto.ReverseVisibilityMap[repo.Visibility],
		SdkPreferences: protoSdkPreferences,
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

	var repoIds []string
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
		var protoSdkPreferences []*registryv1.SdkPreference
		if sdkPrefs, exists := sdkPrefsMap[repository.Id]; exists {
			for _, pref := range sdkPrefs {
				protoSdkPreferences = append(protoSdkPreferences, &registryv1.SdkPreference{
					Sdk:    SdkDbToProtoEnum[pref.Sdk],
					Status: pref.Status,
				})
			}
		}

		resp = append(resp, &registryv1.Repository{
			Id:             repository.Id,
			Name:           repository.Name,
			Visibility:     proto.ReverseVisibilityMap[repository.Visibility],
			SdkPreferences: protoSdkPreferences,
		})
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if totalPages > math.MaxInt32 {
		return nil, connect.NewError(connect.CodeInternal, errors.New("total pages exceeds maximum value"))
	}
	nextPage := int32(0)
	if page < totalPages {
		if page+1 > math.MaxInt32 {
			return nil, connect.NewError(connect.CodeInternal, errors.New("page number exceeds maximum value"))
		}
		nextPage = int32(page + 1) // #nosec G115 -- bounds checked above
	}

	return &registryv1.GetRepositoriesResponse{
		Repositories: resp,
		NextPage:     nextPage,
		TotalPage:    int32(totalPages), // #nosec G115 -- bounds checked above
	}, nil
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

	repo.Name = req.GetName()
	repo.Visibility = proto.VisibilityMap[req.GetVisibility()]

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

	if err := errGroup.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateSdkPreferences(
	ctx context.Context,
	req *registryv1.UpdateSdkPreferencesRequest,
) error {
	repositoryId := req.GetId()

	repo, err := s.repository.GetRepositoryById(ctx, repositoryId)
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

	preferences := make([]SdkPreferencesDTO, 0, len(req.GetSdkPreferences()))
	for _, pref := range req.GetSdkPreferences() {
		preferences = append(preferences, SdkPreferencesDTO{
			Id:           uuid.NewString(),
			RepositoryId: repositoryId,
			Sdk:          SdkProtoToDbEnum[pref.GetSdk()],
			Status:       pref.GetStatus(),
		})
	}

	if err := s.repository.UpdateSdkPreferences(ctx, repositoryId, preferences); err != nil {
		return err
	}

	zap.L().Info("sdk preferences updated",
		zap.String("repositoryId", repositoryId),
		zap.Int("count", len(preferences)),
	)

	if err := s.enqueueSdkTriggerJob(ctx, repositoryId, repo.Path); err != nil {
		zap.L().Warn("failed to enqueue SDK trigger job",
			zap.String("repositoryId", repositoryId),
			zap.Error(err),
		)
	}

	return nil
}

func (s *service) enqueueSdkTriggerJob(ctx context.Context, repositoryId, repoPath string) error {
	if s.sdkQueue == nil {
		zap.L().Debug("SDK generation queue is not configured, skipping SDK trigger job")
		return nil
	}

	job := &SdkTriggerJobDTO{
		Id:           uuid.NewString(),
		RepositoryId: repositoryId,
		RepoPath:     repoPath,
		Status:       SdkGenerationJobStatusPending,
		Attempts:     0,
		MaxAttempts:  5,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.sdkQueue.EnqueueSdkTriggerJob(ctx, job); err != nil {
		return err
	}

	zap.L().Info("SDK trigger job enqueued",
		zap.String("repositoryId", repositoryId),
		zap.String("jobId", job.Id),
	)

	return nil
}

func (s *service) ProcessSdkTrigger(ctx context.Context, repositoryId, repoPath string) error {
	sdkPreferences, err := s.repository.GetSdkPreferences(ctx, repositoryId)
	if err != nil {
		return err
	}

	var enabledSdks []SdkPreferencesDTO
	for _, pref := range sdkPreferences {
		if pref.Status {
			enabledSdks = append(enabledSdks, pref)
		}
	}

	if len(enabledSdks) == 0 {
		zap.L().Debug("no SDK preferences enabled for repository, skipping SDK generation",
			zap.String("repositoryId", repositoryId))
		return nil
	}

	commits, _, err := s.repository.GetCommits(ctx, repoPath, 1, 10000)
	if err != nil {
		zap.L().Debug("no commits found in repository, skipping SDK generation",
			zap.String("repositoryId", repositoryId),
			zap.Error(err),
		)
		return nil
	}

	if len(commits) == 0 {
		zap.L().Debug("no commits in repository, skipping SDK generation",
			zap.String("repositoryId", repositoryId),
		)
		return nil
	}

	var jobs []*SdkGenerationJobDTO
	now := time.Now().UTC()

	for _, commit := range commits {
		for _, sdk := range enabledSdks {
			job := &SdkGenerationJobDTO{
				Id:           uuid.NewString(),
				RepositoryId: repositoryId,
				CommitHash:   commit.GetId(),
				Sdk:          sdk.Sdk,
				Status:       SdkGenerationJobStatusPending,
				Attempts:     0,
				MaxAttempts:  5,
				CreatedAt:    now,
			}
			jobs = append(jobs, job)
		}
	}

	if err := s.sdkQueue.EnqueueSdkGenerationJobs(ctx, jobs); err != nil {
		return err
	}

	zap.L().Info("SDK generation jobs created from trigger",
		zap.String("repositoryId", repositoryId),
		zap.Int("commitCount", len(commits)),
		zap.Int("jobCount", len(jobs)),
	)

	return nil
}

func (s *service) GetCommits(
	ctx context.Context,
	req *registryv1.GetCommitsRequest,
) (*registryv1.GetCommitsResponse, error) {
	userId, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	repoId := req.GetId()
	repo, err := s.repository.GetRepositoryById(ctx, repoId)
	if err != nil {
		return nil, err
	}

	if err := authorization.IsUserMember(ctx, s.orgRepo, repo.OrganizationId, userId); err != nil {
		return nil, err
	}

	page := 1
	pageSize := 10

	if req.Pagination != nil {
		if req.Pagination.GetPage() > 0 {
			page = int(req.Pagination.GetPage())
		}
		if req.Pagination.GetPageLimit() > 0 {
			pageSize = int(req.Pagination.GetPageLimit())
		}
	}

	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if page < 1 {
		page = 1
	}

	commits, totalCount, err := s.repository.GetCommits(ctx, repo.Path, page, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if totalPages > math.MaxInt32 {
		return nil, connect.NewError(connect.CodeInternal, errors.New("total pages exceeds maximum value"))
	}
	nextPage := int32(0)
	if page < totalPages {
		if page+1 > math.MaxInt32 {
			return nil, connect.NewError(connect.CodeInternal, errors.New("page number exceeds maximum value"))
		}
		nextPage = int32(page + 1) // #nosec G115 -- bounds checked above
	}

	return &registryv1.GetCommitsResponse{
		Commits:   commits,
		NextPage:  nextPage,
		TotalPage: int32(totalPages), // #nosec G115 -- bounds checked above
	}, nil
}

func (s *service) GetRecentCommit(
	ctx context.Context,
	req *registryv1.GetRecentCommitRequest,
) (*registryv1.Commit, error) {
	userId, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	repoId := req.GetRepositoryId()
	repo, err := s.repository.GetRepositoryById(ctx, repoId)
	if err != nil {
		return nil, err
	}

	if err := authorization.IsUserMember(ctx, s.orgRepo, repo.OrganizationId, userId); err != nil {
		return nil, err
	}

	commit, err := s.repository.GetRecentCommit(ctx, repo.Path)
	if err != nil {
		return nil, err
	}

	return commit, nil
}

func (s *service) GetFileTree(
	ctx context.Context,
	req *registryv1.GetFileTreeRequest,
) (*registryv1.GetFileTreeResponse, error) {
	userId, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	repoId := req.GetId()
	repo, err := s.repository.GetRepositoryById(ctx, repoId)
	if err != nil {
		return nil, err
	}

	if err := authorization.IsUserMember(ctx, s.orgRepo, repo.OrganizationId, userId); err != nil {
		return nil, err
	}

	var subPath *string
	if req.HasPath() {
		path := req.GetPath()
		subPath = &path
	}

	fileTree, err := s.repository.GetFileTree(ctx, repo.Path, subPath)
	if err != nil {
		return nil, err
	}

	return fileTree, nil
}

func (s *service) GetFilePreview(
	ctx context.Context,
	req *registryv1.GetFilePreviewRequest,
) (*registryv1.GetFilePreviewResponse, error) {
	userId, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	repoId := req.GetId()
	repo, err := s.repository.GetRepositoryById(ctx, repoId)
	if err != nil {
		return nil, err
	}

	if err := authorization.IsUserMember(ctx, s.orgRepo, repo.OrganizationId, userId); err != nil {
		return nil, err
	}

	filePath := req.GetPath()
	filePreview, err := s.repository.GetFilePreview(ctx, repo.Path, filePath)
	if err != nil {
		return nil, err
	}

	return filePreview, nil
}

func (s *service) ValidateSshAccess(
	ctx context.Context,
	userId, repoPath string,
	operation SshOperation,
) (bool, error) {
	repoId := filepath.Base(repoPath)
	repo, err := s.repository.GetRepositoryById(ctx, repoId)
	if err != nil {
		return false, err
	}

	role, err := s.orgRepo.GetMemberRole(ctx, repo.OrganizationId, userId)
	if err != nil {
		zap.L().Warn("SSH access denied: user not member of organization",
			zap.String("userId", userId),
			zap.String("repoPath", repoPath),
			zap.String("organizationId", repo.OrganizationId),
		)
		return false, nil
	}

	switch operation {
	case SshOperationWrite:
		if role == authorization.MemberRoleOwner || role == authorization.MemberRoleAuthor {
			zap.L().Info("SSH write access granted",
				zap.String("userId", userId),
				zap.String("repoPath", repoPath),
				zap.String("role", string(role)),
			)
			return true, nil
		}
		zap.L().Warn("SSH write access denied: insufficient role",
			zap.String("userId", userId),
			zap.String("repoPath", repoPath),
			zap.String("role", string(role)),
		)
		return false, nil
	case SshOperationRead:
		zap.L().Info("SSH read access granted",
			zap.String("userId", userId),
			zap.String("repoPath", repoPath),
			zap.String("role", string(role)),
		)
		return true, nil
	default:
		return false, errors.New("unknown SSH operation")
	}
}

func (s *service) HasProtoFiles(ctx context.Context, repoPath string) (bool, error) {
	protoFiles, err := s.findProtoFiles(repoPath)
	if err != nil {
		return false, err
	}
	return len(protoFiles) > 0, nil
}

func (s *service) TriggerSdkGeneration(ctx context.Context, repositoryId, commitHash string) error {
	if s.sdkQueue == nil {
		zap.L().Debug("SDK generation queue is not configured, skipping SDK generation")
		return nil
	}

	sdkPreferences, err := s.repository.GetSdkPreferences(ctx, repositoryId)
	if err != nil {
		return err
	}

	var jobs []*SdkGenerationJobDTO
	now := time.Now().UTC()

	for _, pref := range sdkPreferences {
		if !pref.Status {
			continue
		}

		job := &SdkGenerationJobDTO{
			Id:           uuid.NewString(),
			RepositoryId: repositoryId,
			CommitHash:   commitHash,
			Sdk:          pref.Sdk,
			Status:       SdkGenerationJobStatusPending,
			Attempts:     0,
			MaxAttempts:  5,
			CreatedAt:    now,
		}
		jobs = append(jobs, job)
	}

	if len(jobs) == 0 {
		zap.L().Debug("no SDK preferences enabled for repository, skipping SDK generation",
			zap.String("repositoryId", repositoryId))
		return nil
	}

	if err := s.sdkQueue.EnqueueSdkGenerationJobs(ctx, jobs); err != nil {
		return err
	}

	zap.L().Info("SDK generation jobs triggered",
		zap.String("repositoryId", repositoryId),
		zap.String("commitHash", commitHash),
		zap.Int("jobCount", len(jobs)))

	return nil
}

func (s *service) GenerateSDK(ctx context.Context, repositoryId, commitHash string, sdk SDK) error {
	repoFullPath := filepath.Join(s.rootPath, repositoryId)

	repo, err := s.repository.GetRepositoryById(ctx, repositoryId)
	if err != nil {
		return fmt.Errorf("failed to fetch repository: %w", err)
	}

	workDir, err := s.checkoutCommitToTempDir(ctx, repoFullPath, commitHash)
	if err != nil {
		return fmt.Errorf("failed to checkout commit: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(workDir); removeErr != nil {
			zap.L().Warn("failed to cleanup temp directory", zap.String("path", workDir), zap.Error(removeErr))
		}
	}()

	sdkDirName := sdkgenerator.SDK(sdk).DirName()
	outputPath := filepath.Join(s.sdkPath, repo.OrganizationId, repositoryId, commitHash, sdkDirName)
	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute output path: %w", err)
	}

	if err := os.MkdirAll(absOutputPath, 0o750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	protoFiles, err := s.findProtoFilesFromFilesystem(workDir)
	if err != nil {
		return fmt.Errorf("failed to find proto files: %w", err)
	}

	if len(protoFiles) == 0 {
		return errors.New("no proto files found in repository")
	}

	generator, err := s.sdkRegistry.Get(sdkgenerator.SDK(sdk))
	if err != nil {
		return fmt.Errorf("unsupported SDK type: %s", sdk)
	}

	input := sdkgenerator.GeneratorInput{
		RepoPath:   workDir,
		OutputPath: absOutputPath,
		ProtoFiles: protoFiles,
	}

	if _, err := generator.Generate(ctx, input); err != nil {
		zap.L().Error("SDK generation failed",
			zap.String("sdk", string(sdk)),
			zap.Error(err))
		return fmt.Errorf("SDK generation failed: %w", err)
	}

	zap.L().Info("SDK generated successfully",
		zap.String("repositoryId", repositoryId),
		zap.String("commitHash", commitHash),
		zap.String("sdk", string(sdk)),
		zap.String("outputPath", absOutputPath))

	if err := s.installSdkDependencies(ctx, absOutputPath, sdk); err != nil {
		zap.L().Warn("failed to install SDK dependencies, but SDK generation succeeded",
			zap.Error(err))
	}

	if err := s.commitSdkToRepo(ctx, repo.OrganizationId, repositoryId, commitHash, sdk); err != nil {
		zap.L().Warn("failed to commit SDK to git repo, but SDK generation succeeded",
			zap.Error(err))
	}

	if err := s.GenerateDocumentation(ctx, repositoryId, commitHash, workDir, repo.OrganizationId); err != nil {
		zap.L().Warn("documentation generation failed, but SDK generation succeeded",
			zap.Error(err))
	}

	return nil
}

func (s *service) checkoutCommitToTempDir(ctx context.Context, repoPath, commitHash string) (string, error) {
	tempDir, err := os.MkdirTemp("", "hasir-sdk-gen-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// #nosec G204 -- commitHash is validated as a git commit hash
	archiveCmd := exec.CommandContext(ctx, "git", "archive", "--format=tar", commitHash)
	archiveCmd.Dir = repoPath

	// #nosec G204 -- tempDir is created by os.MkdirTemp
	extractCmd := exec.CommandContext(ctx, "tar", "-xf", "-", "-C", tempDir)

	archiveOutput, err := archiveCmd.StdoutPipe()
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to create pipe: %w", err)
	}
	extractCmd.Stdin = archiveOutput

	if err := archiveCmd.Start(); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to start git archive: %w", err)
	}

	if err := extractCmd.Start(); err != nil {
		_ = archiveCmd.Wait()
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to start tar: %w", err)
	}

	if err := archiveCmd.Wait(); err != nil {
		_ = extractCmd.Wait()
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("git archive failed: %w", err)
	}

	if err := extractCmd.Wait(); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("tar extraction failed: %w", err)
	}

	return tempDir, nil
}

func (s *service) GenerateDocumentation(ctx context.Context, repositoryId, commitHash, repoPath, organizationId string) error {
	outputPath := filepath.Join(s.sdkPath, organizationId, repositoryId, commitHash, "docs")

	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute output path: %w", err)
	}

	if err := os.MkdirAll(absOutputPath, 0o750); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	protoFiles, err := s.findProtoFilesFromFilesystem(repoPath)
	if err != nil {
		return fmt.Errorf("failed to find proto files: %w", err)
	}

	if len(protoFiles) == 0 {
		return errors.New("no proto files found in repository")
	}

	validatedFiles, err := s.sanitizeProtocArgs(repoPath, absOutputPath, protoFiles)
	if err != nil {
		return fmt.Errorf("invalid proto arguments: %w", err)
	}

	args := []string{
		"--proto_path=" + filepath.Clean(repoPath),
		"--doc_out=" + filepath.Clean(absOutputPath),
		"--doc_opt=html,index.html",
	}
	args = append(args, validatedFiles...)

	// #nosec G204 -- Command is hardcoded "protoc", args are sanitized via sanitizeProtocArgs
	cmd := exec.CommandContext(ctx, "protoc", args...)
	cmd.Dir = filepath.Clean(repoPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		zap.L().Error("documentation generation failed",
			zap.String("output", string(output)),
			zap.Error(err))
		return fmt.Errorf("protoc-gen-doc failed: %w: %s", err, string(output))
	}

	zap.L().Info("documentation generated successfully",
		zap.String("repositoryId", repositoryId),
		zap.String("commitHash", commitHash),
		zap.String("outputPath", absOutputPath))

	return nil
}

func (s *service) findProtoFiles(repoPath string) ([]string, error) {
	// #nosec G204 -- repoPath is validated before calling this function
	cmd := exec.Command("git", "ls-tree", "-r", "HEAD", "--name-only")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return s.findProtoFilesFromFilesystem(repoPath)
	}

	var protoFiles []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.HasSuffix(line, ".proto") {
			protoFiles = append(protoFiles, line)
		}
	}

	return protoFiles, nil
}

func (s *service) findProtoFilesFromFilesystem(repoPath string) ([]string, error) {
	var protoFiles []string
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".proto" {
			relPath, err := filepath.Rel(repoPath, path)
			if err != nil {
				return err
			}
			protoFiles = append(protoFiles, relPath)
		}

		return nil
	})

	return protoFiles, err
}

func (s *service) sanitizeProtocArgs(repoPath, outputPath string, protoFiles []string) ([]string, error) {
	cleanRepoPath := filepath.Clean(repoPath)
	cleanOutputPath := filepath.Clean(outputPath)

	if strings.Contains(cleanRepoPath, "..") || strings.Contains(cleanOutputPath, "..") {
		return nil, errors.New("path traversal detected in repository or output path")
	}

	for _, file := range protoFiles {
		cleanFile := filepath.Clean(file)
		if strings.Contains(cleanFile, "..") || filepath.IsAbs(cleanFile) {
			return nil, fmt.Errorf("invalid proto file path: %s", file)
		}
		if !strings.HasSuffix(cleanFile, ".proto") {
			return nil, fmt.Errorf("invalid proto file extension: %s", file)
		}
	}

	return protoFiles, nil
}

func (s *service) commitSdkToRepo(ctx context.Context, orgId, repoId, commitHash string, sdk SDK) error {
	sdkDirName := sdkgenerator.SDK(sdk).DirName()
	sdkRepoPath := filepath.Join(s.sdkPath, orgId, repoId, commitHash, sdkDirName)
	absSdkRepoPath, err := filepath.Abs(sdkRepoPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute SDK repo path: %w", err)
	}

	if _, err = os.Stat(filepath.Join(absSdkRepoPath, ".git")); os.IsNotExist(err) {
		initCmd := exec.CommandContext(ctx, "git", "init")
		initCmd.Dir = absSdkRepoPath
		if output, err := initCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to init git repo: %w: %s", err, string(output))
		}

		configCmds := [][]string{
			{"git", "config", "user.email", "sdk@hasir.dev"},
			{"git", "config", "user.name", "Hasir SDK Generator"},
		}
		for _, args := range configCmds {
			// #nosec G204 -- args are hardcoded git config commands
			cmd := exec.CommandContext(ctx, args[0], args[1:]...)
			cmd.Dir = absSdkRepoPath
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to configure git: %w: %s", err, string(output))
			}
		}

		zap.L().Info("initialized SDK git repository",
			zap.String("path", absSdkRepoPath))
	}

	addCmd := exec.CommandContext(ctx, "git", "add", "-A")
	addCmd.Dir = absSdkRepoPath
	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to git add: %w: %s", err, string(output))
	}

	statusCmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	statusCmd.Dir = absSdkRepoPath
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if len(strings.TrimSpace(string(statusOutput))) == 0 {
		zap.L().Debug("no changes to commit in SDK repo",
			zap.String("repoId", repoId),
			zap.String("sdk", string(sdk)))
		return nil
	}

	commitMsg := fmt.Sprintf("SDK generated from commit %s", commitHash)
	// #nosec G204 -- commitMsg is a formatted string with validated commitHash
	commitCmd := exec.CommandContext(ctx, "git", "commit", "-m", commitMsg)
	commitCmd.Dir = absSdkRepoPath
	if output, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to git commit: %w: %s", err, string(output))
	}

	tagCmd := exec.CommandContext(ctx, "git", "tag", "-f", commitHash)
	tagCmd.Dir = absSdkRepoPath
	if output, err := tagCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to git tag: %w: %s", err, string(output))
	}

	zap.L().Info("SDK committed to git repository",
		zap.String("repoId", repoId),
		zap.String("commitHash", commitHash),
		zap.String("sdk", string(sdk)),
		zap.String("sdkRepoPath", absSdkRepoPath))

	return nil
}

func (s *service) installSdkDependencies(ctx context.Context, sdkRepoPath string, sdk SDK) error {
	absSdkRepoPath, err := filepath.Abs(sdkRepoPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute SDK repo path: %w", err)
	}

	sdkType := sdkgenerator.SDK(sdk)

	if sdkType.IsGo() {
		return s.installGoDependencies(ctx, absSdkRepoPath)
	}

	if sdkType.IsJs() {
		return s.installJsDependencies(ctx, absSdkRepoPath)
	}

	return nil
}

func (s *service) installGoDependencies(ctx context.Context, sdkRepoPath string) error {
	if _, err := os.Stat(filepath.Join(sdkRepoPath, "go.mod")); os.IsNotExist(err) {
		moduleName := "sdk"
		initCmd := exec.CommandContext(ctx, "go", "mod", "init", moduleName)
		initCmd.Dir = sdkRepoPath
		if output, err := initCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("go mod init failed: %w: %s", err, string(output))
		}
		zap.L().Debug("go.mod initialized", zap.String("path", sdkRepoPath))
	}

	getCmd := exec.CommandContext(ctx, "go", "get", "./...")
	getCmd.Dir = sdkRepoPath
	if output, err := getCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go get failed: %w: %s", err, string(output))
	}

	tidyCmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	tidyCmd.Dir = sdkRepoPath
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w: %s", err, string(output))
	}

	zap.L().Info("Go dependencies installed", zap.String("path", sdkRepoPath))
	return nil
}

func (s *service) installJsDependencies(ctx context.Context, sdkRepoPath string) error {
	if _, err := os.Stat(filepath.Join(sdkRepoPath, "package.json")); os.IsNotExist(err) {
		initCmd := exec.CommandContext(ctx, "npm", "init", "-y")
		initCmd.Dir = sdkRepoPath
		if output, err := initCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("npm init failed: %w: %s", err, string(output))
		}
		zap.L().Debug("package.json initialized", zap.String("path", sdkRepoPath))
	}

	depcheckCmd := exec.CommandContext(ctx, "npx", "depcheck", "--json")
	depcheckCmd.Dir = sdkRepoPath
	output, _ := depcheckCmd.CombinedOutput()

	var depResult struct {
		Missing map[string][]string `json:"missing"`
	}
	if err := json.Unmarshal(output, &depResult); err != nil {
		zap.L().Debug("depcheck output parsing failed, skipping dependency install",
			zap.String("output", string(output)),
			zap.Error(err))
		return nil
	}

	if len(depResult.Missing) > 0 {
		var packages []string
		for pkg := range depResult.Missing {
			packages = append(packages, pkg)
		}

		args := append([]string{"install", "--save"}, packages...)
		// #nosec G204 -- packages are npm package names from depcheck output
		installCmd := exec.CommandContext(ctx, "npm", args...)
		installCmd.Dir = sdkRepoPath
		if output, err := installCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("npm install failed: %w: %s", err, string(output))
		}
	}

	zap.L().Info("JS dependencies installed", zap.String("path", sdkRepoPath))
	return nil
}
