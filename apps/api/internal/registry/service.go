package registry

import (
	"context"
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
)

const DefaultReposPath = "./repos"

type Service interface {
	CreateRepository(ctx context.Context, req *registryv1.CreateRepositoryRequest) error
	GetRepository(ctx context.Context, req *registryv1.GetRepositoryRequest) (*registryv1.Repository, error)
	GetRepositories(ctx context.Context, organizationId *string, page, pageSize int) (*registryv1.GetRepositoriesResponse, error)
	UpdateRepository(ctx context.Context, req *registryv1.UpdateRepositoryRequest) error
	DeleteRepository(ctx context.Context, req *registryv1.DeleteRepositoryRequest) error
	DeleteRepositoriesByOrganization(ctx context.Context, organizationId string) error
	UpdateSdkPreferences(ctx context.Context, req *registryv1.UpdateSdkPreferencesRequest) error
	GetCommits(ctx context.Context, req *registryv1.GetCommitsRequest) (*registryv1.GetCommitsResponse, error)
	GetFileTree(ctx context.Context, req *registryv1.GetFileTreeRequest) (*registryv1.GetFileTreeResponse, error)
	GetFilePreview(ctx context.Context, req *registryv1.GetFilePreviewRequest) (*registryv1.GetFilePreviewResponse, error)
	ValidateSshAccess(ctx context.Context, userId, repoPath string, operation SshOperation) (bool, error)
	HasProtoFiles(ctx context.Context, repoPath string) (bool, error)
	TriggerSdkGeneration(ctx context.Context, repositoryId, commitHash string) error
	GenerateSDK(ctx context.Context, repositoryId, commitHash string, sdk SDK) error
}

type service struct {
	rootPath      string
	repository    Repository
	orgRepo       authorization.MemberRoleChecker
	sdkQueue      SdkGenerationQueue
	cfg           *config.Config
	sdkPath       string
	sdkGenerators map[SDK]func(context.Context, string, string, []string) *exec.Cmd
	sdkDirNames   map[SDK]string
}

func NewService(repository Repository, orgRepo authorization.MemberRoleChecker, sdkQueue SdkGenerationQueue, cfg *config.Config) Service {
	sdkPath := "./sdk"
	if cfg != nil && cfg.SdkGeneration.OutputPath != "" {
		sdkPath = cfg.SdkGeneration.OutputPath
	}

	svc := &service{
		rootPath:   DefaultReposPath,
		repository: repository,
		orgRepo:    orgRepo,
		sdkQueue:   sdkQueue,
		cfg:        cfg,
		sdkPath:    sdkPath,
	}

	svc.sdkGenerators = map[SDK]func(context.Context, string, string, []string) *exec.Cmd{
		SdkGoProtobuf:   svc.generateGoProtobuf,
		SdkGoConnectRpc: svc.generateGoConnectRpc,
		SdkGoGrpc:       svc.generateGoGrpc,
		SdkJsBufbuildEs: svc.generateJsBufbuildEs,
		SdkJsProtobuf:   svc.generateJsProtobuf,
		SdkJsConnectrpc: svc.generateJsConnectrpc,
	}

	svc.sdkDirNames = map[SDK]string{
		SdkGoProtobuf:   "go-protobuf",
		SdkGoConnectRpc: "go-connectrpc",
		SdkGoGrpc:       "go-grpc",
		SdkJsBufbuildEs: "js-bufbuild-es",
		SdkJsProtobuf:   "js-protobuf",
		SdkJsConnectrpc: "js-connectrpc",
	}

	return svc
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

	commits, err := s.repository.GetCommits(ctx, repo.Path)
	if err != nil {
		return nil, err
	}

	return commits, nil
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

	outputPath := filepath.Join(s.sdkPath, repo.OrganizationId, repositoryId, commitHash, s.getSdkDirName(sdk))
	if err := os.MkdirAll(outputPath, 0o750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	protoFiles, err := s.findProtoFiles(repoFullPath)
	if err != nil {
		return fmt.Errorf("failed to find proto files: %w", err)
	}

	if len(protoFiles) == 0 {
		return errors.New("no proto files found in repository")
	}

	generatorFunc, ok := s.sdkGenerators[sdk]
	if !ok {
		return fmt.Errorf("unsupported SDK type: %s", sdk)
	}

	cmd := generatorFunc(ctx, repoFullPath, outputPath, protoFiles)

	output, err := cmd.CombinedOutput()
	if err != nil {
		zap.L().Error("SDK generation failed",
			zap.String("sdk", string(sdk)),
			zap.String("output", string(output)),
			zap.Error(err))
		return fmt.Errorf("protoc failed: %w: %s", err, string(output))
	}

	zap.L().Info("SDK generated successfully",
		zap.String("repositoryId", repositoryId),
		zap.String("commitHash", commitHash),
		zap.String("sdk", string(sdk)),
		zap.String("outputPath", outputPath))

	if err := s.GenerateDocumentation(ctx, repositoryId, commitHash, repoFullPath, repo.OrganizationId); err != nil {
		zap.L().Warn("documentation generation failed, but SDK generation succeeded",
			zap.Error(err))
	}

	return nil
}

func (s *service) GenerateDocumentation(ctx context.Context, repositoryId, commitHash, repoPath, organizationId string) error {
	outputPath := filepath.Join(s.sdkPath, organizationId, repositoryId, commitHash, "docs")
	if err := os.MkdirAll(outputPath, 0o750); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	protoFiles, err := s.findProtoFiles(repoPath)
	if err != nil {
		return fmt.Errorf("failed to find proto files: %w", err)
	}

	if len(protoFiles) == 0 {
		return errors.New("no proto files found in repository")
	}

	validatedFiles, err := s.sanitizeProtocArgs(repoPath, outputPath, protoFiles)
	if err != nil {
		return fmt.Errorf("invalid proto arguments: %w", err)
	}

	args := []string{
		"--proto_path=" + filepath.Clean(repoPath),
		"--doc_out=" + filepath.Clean(outputPath),
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
		zap.String("outputPath", outputPath))

	return nil
}

func (s *service) generateGoProtobuf(ctx context.Context, repoPath, outputPath string, protoFiles []string) *exec.Cmd {
	validatedFiles, err := s.sanitizeProtocArgs(repoPath, outputPath, protoFiles)
	if err != nil {
		return exec.CommandContext(ctx, "false")
	}

	args := []string{
		"--proto_path=" + filepath.Clean(repoPath),
		"--go_out=" + filepath.Clean(outputPath),
		"--go_opt=paths=source_relative",
	}
	args = append(args, validatedFiles...)

	// #nosec G204 -- Command is hardcoded "protoc", args are sanitized via sanitizeProtocArgs
	cmd := exec.CommandContext(ctx, "protoc", args...)
	cmd.Dir = filepath.Clean(repoPath)
	return cmd
}

func (s *service) generateGoConnectRpc(ctx context.Context, repoPath, outputPath string, protoFiles []string) *exec.Cmd {
	validatedFiles, err := s.sanitizeProtocArgs(repoPath, outputPath, protoFiles)
	if err != nil {
		return exec.CommandContext(ctx, "false")
	}

	args := []string{
		"--proto_path=" + filepath.Clean(repoPath),
		"--go_out=" + filepath.Clean(outputPath),
		"--go_opt=paths=source_relative",
		"--connect-go_out=" + filepath.Clean(outputPath),
		"--connect-go_opt=paths=source_relative",
	}
	args = append(args, validatedFiles...)

	// #nosec G204 -- Command is hardcoded "protoc", args are sanitized via sanitizeProtocArgs
	cmd := exec.CommandContext(ctx, "protoc", args...)
	cmd.Dir = filepath.Clean(repoPath)
	return cmd
}

func (s *service) generateGoGrpc(ctx context.Context, repoPath, outputPath string, protoFiles []string) *exec.Cmd {
	validatedFiles, err := s.sanitizeProtocArgs(repoPath, outputPath, protoFiles)
	if err != nil {
		return exec.CommandContext(ctx, "false")
	}

	args := []string{
		"--proto_path=" + filepath.Clean(repoPath),
		"--go_out=" + filepath.Clean(outputPath),
		"--go_opt=paths=source_relative",
		"--go-grpc_out=" + filepath.Clean(outputPath),
		"--go-grpc_opt=paths=source_relative",
	}
	args = append(args, validatedFiles...)

	// #nosec G204 -- Command is hardcoded "protoc", args are sanitized via sanitizeProtocArgs
	cmd := exec.CommandContext(ctx, "protoc", args...)
	cmd.Dir = filepath.Clean(repoPath)
	return cmd
}

func (s *service) generateJsBufbuildEs(ctx context.Context, repoPath, outputPath string, protoFiles []string) *exec.Cmd {
	validatedFiles, err := s.sanitizeProtocArgs(repoPath, outputPath, protoFiles)
	if err != nil {
		return exec.CommandContext(ctx, "false")
	}

	args := []string{
		"--proto_path=" + filepath.Clean(repoPath),
		"--es_out=" + filepath.Clean(outputPath),
		"--es_opt=target=ts",
	}
	args = append(args, validatedFiles...)

	// #nosec G204 -- Command is hardcoded "protoc", args are sanitized via sanitizeProtocArgs
	cmd := exec.CommandContext(ctx, "protoc", args...)
	cmd.Dir = filepath.Clean(repoPath)
	return cmd
}

func (s *service) generateJsProtobuf(ctx context.Context, repoPath, outputPath string, protoFiles []string) *exec.Cmd {
	validatedFiles, err := s.sanitizeProtocArgs(repoPath, outputPath, protoFiles)
	if err != nil {
		return exec.CommandContext(ctx, "false")
	}

	args := []string{
		"--proto_path=" + filepath.Clean(repoPath),
		"--js_out=import_style=commonjs,binary:" + filepath.Clean(outputPath),
	}
	args = append(args, validatedFiles...)

	// #nosec G204 -- Command is hardcoded "protoc", args are sanitized via sanitizeProtocArgs
	cmd := exec.CommandContext(ctx, "protoc", args...)
	cmd.Dir = filepath.Clean(repoPath)
	return cmd
}

func (s *service) generateJsConnectrpc(ctx context.Context, repoPath, outputPath string, protoFiles []string) *exec.Cmd {
	validatedFiles, err := s.sanitizeProtocArgs(repoPath, outputPath, protoFiles)
	if err != nil {
		return exec.CommandContext(ctx, "false")
	}

	args := []string{
		"--proto_path=" + filepath.Clean(repoPath),
		"--es_out=" + filepath.Clean(outputPath),
		"--es_opt=target=ts",
		"--connect-es_out=" + filepath.Clean(outputPath),
		"--connect-es_opt=target=ts",
	}
	args = append(args, validatedFiles...)

	// #nosec G204 -- Command is hardcoded "protoc", args are sanitized via sanitizeProtocArgs
	cmd := exec.CommandContext(ctx, "protoc", args...)
	cmd.Dir = filepath.Clean(repoPath)
	return cmd
}

func (s *service) findProtoFiles(repoPath string) ([]string, error) {
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

func (s *service) getSdkDirName(sdk SDK) string {
	if dirName, ok := s.sdkDirNames[sdk]; ok {
		return dirName
	}
	return "unknown"
}
