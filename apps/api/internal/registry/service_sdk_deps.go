package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-json"
	"go.uber.org/zap"

	"hasir/api/pkg/sdkgenerator"
)

func (s *service) checkoutCommitToTempDir(ctx context.Context, repoPath, commitHash string) (string, error) {
	tempDir, err := os.MkdirTemp("", "hasir-sdk-gen-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	err = s.commandRunner().Pipe(
		ctx,
		"git", []string{"archive", "--format=tar", commitHash}, repoPath,
		"tar", []string{"-xf", "-", "-C", tempDir}, "",
	)
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return "", err
	}

	return tempDir, nil
}

func (s *service) commitSdkToRepo(ctx context.Context, orgId, repoId, commitHash string, sdk SDK) error {
	generator, err := s.sdkRegistry.Get(sdkgenerator.SDK(sdk))
	if err != nil {
		return fmt.Errorf("unsupported SDK type: %s", sdk)
	}
	sdkDirName := generator.DirName()
	sdkRepoPath := filepath.Join(s.sdkPath, orgId, repoId, commitHash, sdkDirName)
	absSdkRepoPath, err := filepath.Abs(sdkRepoPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute SDK repo path: %w", err)
	}

	runner := s.commandRunner()

	if _, err = os.Stat(filepath.Join(absSdkRepoPath, ".git")); os.IsNotExist(err) {
		if output, err := runner.RunCombined(ctx, "git", []string{"init"}, absSdkRepoPath); err != nil {
			return fmt.Errorf("failed to init git repo: %w: %s", err, string(output))
		}

		configCmds := [][]string{
			{"config", "user.email", "sdk@hasir.dev"},
			{"config", "user.name", "Hasir SDK Generator"},
		}
		for _, args := range configCmds {
			if output, err := runner.RunCombined(ctx, "git", args, absSdkRepoPath); err != nil {
				return fmt.Errorf("failed to configure git: %w: %s", err, string(output))
			}
		}

		zap.L().Info("initialized SDK git repository",
			zap.String("path", absSdkRepoPath))
	}

	if output, err := runner.RunCombined(ctx, "git", []string{"add", "-A"}, absSdkRepoPath); err != nil {
		return fmt.Errorf("failed to git add: %w: %s", err, string(output))
	}

	statusOutput, err := runner.Run(ctx, "git", []string{"status", "--porcelain"}, absSdkRepoPath)
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
	if output, err := runner.RunCombined(ctx, "git", []string{"commit", "-m", commitMsg}, absSdkRepoPath); err != nil {
		return fmt.Errorf("failed to git commit: %w: %s", err, string(output))
	}

	if output, err := runner.RunCombined(ctx, "git", []string{"tag", "-f", commitHash}, absSdkRepoPath); err != nil {
		return fmt.Errorf("failed to git tag: %w: %s", err, string(output))
	}

	zap.L().Info("SDK committed to git repository",
		zap.String("repoId", repoId),
		zap.String("commitHash", commitHash),
		zap.String("sdk", string(sdk)),
		zap.String("sdkRepoPath", absSdkRepoPath))

	return nil
}

func (s *service) installSdkDependencies(ctx context.Context, sdkRepoPath string, sdk SDK, organizationId, repositoryId, commitHash string) error {
	absSdkRepoPath, err := filepath.Abs(sdkRepoPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute SDK repo path: %w", err)
	}

	sdkType := sdkgenerator.SDK(sdk)

	if sdkType.IsGo() {
		return s.installGoDependencies(ctx, absSdkRepoPath, organizationId, repositoryId, commitHash, sdkType.DirName())
	}

	if sdkType.IsJs() {
		return s.installJsDependencies(ctx, absSdkRepoPath)
	}

	return nil
}

func (s *service) installGoDependencies(ctx context.Context, sdkRepoPath string, organizationId, repositoryId, commitHash, sdkDirName string) error {
	runner := s.commandRunner()

	if _, err := os.Stat(filepath.Join(sdkRepoPath, "go.mod")); os.IsNotExist(err) {
		moduleBasePath := "localhost"
		if s.cfg != nil {
			moduleBasePath = s.cfg.SdkGeneration.GetModuleBasePath()
		}
		moduleName := fmt.Sprintf("%s/sdk/%s/%s/%s/%s", moduleBasePath, organizationId, repositoryId, commitHash, sdkDirName)
		if output, err := runner.RunCombined(ctx, "go", []string{"mod", "init", moduleName}, sdkRepoPath); err != nil {
			return fmt.Errorf("go mod init failed: %w: %s", err, string(output))
		}

		zap.L().Debug("go.mod initialized", zap.String("path", sdkRepoPath), zap.String("module", moduleName))
	}

	if output, err := runner.RunCombined(ctx, "go", []string{"mod", "tidy"}, sdkRepoPath); err != nil {
		return fmt.Errorf("go mod tidy failed: %w: %s", err, string(output))
	}

	zap.L().Info("Go dependencies installed", zap.String("path", sdkRepoPath))
	return nil
}

func (s *service) installJsDependencies(ctx context.Context, sdkRepoPath string) error {
	runner := s.commandRunner()

	pkgJsonPath := filepath.Join(sdkRepoPath, "package.json")
	if _, err := os.Stat(pkgJsonPath); os.IsNotExist(err) {
		// #nosec G306 -- generated SDK package.json must be readable by npm/git
		if err := os.WriteFile(pkgJsonPath, []byte("{\"name\":\"sdk\",\"version\":\"1.0.0\",\"private\":true}\n"), 0o644); err != nil {
			return fmt.Errorf("failed to write package.json: %w", err)
		}
		zap.L().Debug("package.json initialized", zap.String("path", sdkRepoPath))
	}

	output, depErr := runner.RunCombined(ctx, "npx", []string{"depcheck", "--json"}, sdkRepoPath)

	var depResult struct {
		Missing map[string][]string `json:"missing"`
	}
	if err := json.Unmarshal(output, &depResult); err != nil {
		if depErr != nil {
			return fmt.Errorf("depcheck failed: %w: %s", depErr, string(output))
		}
		return fmt.Errorf("failed to parse depcheck output: %w: %s", err, string(output))
	}

	if len(depResult.Missing) > 0 {
		packages := make([]string, 0, len(depResult.Missing))
		for pkg := range depResult.Missing {
			packages = append(packages, pkg)
		}

		args := append([]string{"install", "--save"}, packages...)
		if output, err := runner.RunCombined(ctx, "npm", args, sdkRepoPath); err != nil {
			return fmt.Errorf("npm install failed: %w: %s", err, string(output))
		}
	}

	zap.L().Info("JS dependencies installed", zap.String("path", sdkRepoPath))
	return nil
}
