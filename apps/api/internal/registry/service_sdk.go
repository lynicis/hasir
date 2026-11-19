package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"hasir/api/pkg/authentication"
	"hasir/api/pkg/authorization"
	"hasir/api/pkg/sdkgenerator"
	registryv1 "hasir/proto/gen/go/registry/v1"
)

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

	hasEnabledPreference := false
	for _, pref := range preferences {
		if pref.Status {
			hasEnabledPreference = true
			break
		}
	}

	if !hasEnabledPreference {
		zap.L().Debug("no enabled SDK preferences, skipping SDK trigger job",
			zap.String("repositoryId", repositoryId))
		return nil
	}

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
			jobs = append(jobs, newPendingSdkGenerationJob(repositoryId, commit.GetId(), sdk.Sdk, now))
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

func newPendingSdkGenerationJob(repositoryId, commitHash string, sdk SDK, now time.Time) *SdkGenerationJobDTO {
	return &SdkGenerationJobDTO{
		Id:           uuid.NewString(),
		RepositoryId: repositoryId,
		CommitHash:   commitHash,
		Sdk:          sdk,
		Status:       SdkGenerationJobStatusPending,
		Attempts:     0,
		MaxAttempts:  5,
		CreatedAt:    now,
	}
}

func (s *service) TriggerSdkGeneration(ctx context.Context, repositoryId, commitHash string) error {
	if s.sdkQueue == nil {
		zap.L().Debug("SDK generation queue is not configured, skipping SDK generation")
		return nil
	}

	repo, err := s.repository.GetRepositoryById(ctx, repositoryId)
	if err != nil {
		return fmt.Errorf("failed to fetch repository: %w", err)
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
		jobs = append(jobs, newPendingSdkGenerationJob(repositoryId, commitHash, pref.Sdk, now))
	}

	if len(jobs) == 0 && repo.ManagedByBuf {
		repoFullPath := filepath.Join(s.rootPath, repositoryId)
		if detectedSdks := s.detectSdksFromRepoCommit(ctx, repoFullPath, commitHash); len(detectedSdks) > 0 {
			zap.L().Info("auto-detected SDKs from buf.gen.yaml",
				zap.String("repositoryId", repositoryId),
				zap.String("commitHash", commitHash),
				zap.Int("sdkCount", len(detectedSdks)))

			for _, sdk := range detectedSdks {
				jobs = append(jobs, newPendingSdkGenerationJob(repositoryId, commitHash, sdk, now))
			}
		}
	}

	if len(jobs) == 0 {
		zap.L().Info("no SDK preferences enabled for repository, skipping SDK generation",
			zap.String("repositoryId", repositoryId),
			zap.String("commitHash", commitHash))
		s.triggerAsyncDocumentation(repositoryId, commitHash, repo.OrganizationId)
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

func (s *service) triggerAsyncDocumentation(repositoryId, commitHash, organizationId string) {
	// Trigger documentation generation asynchronously to not block the git client
	// #nosec G118 -- background goroutine needs to outlive the request context
	go func() {
		bgCtx := context.Background()
		repoFullPath := filepath.Join(s.rootPath, repositoryId)
		workDir, err := s.checkoutCommitToTempDir(bgCtx, repoFullPath, commitHash)
		if err != nil {
			zap.L().Error("failed to checkout commit for repo documentation generation",
				zap.String("repositoryId", repositoryId),
				zap.String("commitHash", commitHash),
				zap.Error(err))
			return
		}
		defer func() {
			// #nosec G703 -- workDir is created by os.MkdirTemp
			if removeErr := os.RemoveAll(workDir); removeErr != nil {
				zap.L().Warn("failed to cleanup temp directory", zap.String("path", workDir), zap.Error(removeErr))
			}
		}()

		if err := s.GenerateDocumentation(bgCtx, repositoryId, commitHash, workDir, organizationId); err != nil {
			zap.L().Error("failed to generate documentation for repo",
				zap.String("repositoryId", repositoryId),
				zap.String("commitHash", commitHash),
				zap.Error(err))
		}
	}()
}

func (s *service) detectSdksFromRepoCommit(ctx context.Context, repoPath, commitHash string) []SDK {
	if commitHash == "" {
		commitHash = "HEAD"
	}

	runner := s.commandRunner()
	output, err := runner.Run(ctx, "git", []string{"ls-tree", "-r", "--name-only", commitHash}, repoPath)
	if err != nil {
		return nil
	}

	var bufGenPaths []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if filepath.Base(line) == "buf.gen.yaml" {
			bufGenPaths = append(bufGenPaths, line)
		}
	}

	if len(bufGenPaths) == 0 {
		return nil
	}

	var allSdks []SDK
	seen := make(map[SDK]bool)

	for _, p := range bufGenPaths {
		content, showErr := runner.Run(ctx, "git", []string{"show", fmt.Sprintf("%s:%s", commitHash, p)}, repoPath)
		if showErr != nil {
			continue
		}

		for _, sdk := range detectSdksFromBufGen(content) {
			if !seen[sdk] {
				seen[sdk] = true
				allSdks = append(allSdks, sdk)
			}
		}
	}

	return allSdks
}

type bufGenSdkRule struct {
	sdk      SDK
	patterns []string
}

var bufGenSdkRules = []bufGenSdkRule{
	{sdk: SdkGoConnectRpc, patterns: []string{"connect-go", "connectrpc/go"}},
	{sdk: SdkGoProtobuf, patterns: []string{"protoc-gen-go", "protocolbuffers/go"}},
	{sdk: SdkGoGrpc, patterns: []string{"grpc-go", "go-grpc", "grpc/go"}},
	{sdk: SdkJsConnectrpc, patterns: []string{"connect-query", "connectrpc/es"}},
	{sdk: SdkJsBufbuildEs, patterns: []string{"bufbuild/es", "protoc-gen-es"}},
	{sdk: SdkJsProtobuf, patterns: []string{"protobuf/js", "protocolbuffers/js"}},
	{sdk: SdkRustProtobuf, patterns: []string{"prost"}},
	{sdk: SdkRustGrpc, patterns: []string{"tonic"}},
	{sdk: SdkJavaProtobuf, patterns: []string{"protocolbuffers/java"}},
	{sdk: SdkJavaGrpc, patterns: []string{"grpc/java"}},
	{sdk: SdkCsharpProtobuf, patterns: []string{"protocolbuffers/csharp"}},
	{sdk: SdkCsharpGrpc, patterns: []string{"grpc/csharp"}},
}

func detectSdksFromBufGen(content []byte) []SDK {
	text := string(content)
	var sdks []SDK

	for _, rule := range bufGenSdkRules {
		for _, pattern := range rule.patterns {
			if strings.Contains(text, pattern) {
				sdks = append(sdks, rule.sdk)
				break
			}
		}
	}

	return sdks
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

	registry := s.sdkRegistry
	if repo.ManagedByBuf {
		registry = s.bufRegistry
	}

	generator, err := registry.Get(sdkgenerator.SDK(sdk))
	if err != nil {
		return fmt.Errorf("unsupported SDK type: %s", sdk)
	}

	sdkDirName := generator.DirName()
	outputPath := filepath.Join(s.sdkPath, repo.OrganizationId, repositoryId, commitHash, sdkDirName)

	var goPackagePrefix string
	if generator.SDK().IsGo() {
		moduleBasePath := "localhost"
		if s.cfg != nil {
			moduleBasePath = s.cfg.SdkGeneration.GetModuleBasePath()
		}
		goPackagePrefix = fmt.Sprintf("%s/sdk/%s/%s/%s/%s", moduleBasePath, repo.OrganizationId, repositoryId, commitHash, sdkDirName)
	}

	output, err := sdkgenerator.GenerateFromRepo(ctx, generator, workDir, outputPath, goPackagePrefix)
	if err != nil {
		zap.L().Error("SDK generation failed",
			zap.String("sdk", string(sdk)),
			zap.Error(err))
		return fmt.Errorf("SDK generation failed: %w", err)
	}

	zap.L().Info("SDK generated successfully",
		zap.String("repositoryId", repositoryId),
		zap.String("commitHash", commitHash),
		zap.String("sdk", string(sdk)),
		zap.Bool("managedByBuf", repo.ManagedByBuf),
		zap.String("outputPath", output.OutputPath))

	if err := s.installSdkDependencies(ctx, output.OutputPath, sdk, repo.OrganizationId, repositoryId, commitHash); err != nil {
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

func (s *service) GenerateDocumentation(ctx context.Context, repositoryId, commitHash, repoPath, organizationId string) error {
	outputPath := filepath.Join(s.sdkPath, organizationId, repositoryId, commitHash, "docs")

	if _, err := sdkgenerator.GenerateFromRepo(ctx, s.docGenerator, repoPath, outputPath, ""); err != nil {
		zap.L().Error("documentation generation failed",
			zap.String("repositoryId", repositoryId),
			zap.String("commitHash", commitHash),
			zap.Error(err))
		return fmt.Errorf("documentation generation failed: %w", err)
	}

	zap.L().Info("documentation generated successfully",
		zap.String("repositoryId", repositoryId),
		zap.String("commitHash", commitHash),
		zap.String("outputPath", outputPath))

	return nil
}
