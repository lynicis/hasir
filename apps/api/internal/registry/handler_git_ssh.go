package registry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gliderlabs/ssh"
	"go.uber.org/zap"
)

const banner = `
██╗  ██╗ █████╗ ███████╗██╗██████╗
██║  ██║██╔══██╗██╔════╝██║██╔══██╗
███████║███████║███████╗██║██████╔╝
██╔══██║██╔══██║╚════██║██║██╔══██╗
██║  ██║██║  ██║███████║██║██║  ██║
╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝╚═╝  ╚═╝

Protocol Buffer Schema Registry
`

type GitSshHandler struct {
	service   Service
	reposPath string
}

func NewGitSshHandler(service Service, reposPath string) *GitSshHandler {
	return &GitSshHandler{
		service:   service,
		reposPath: reposPath,
	}
}

func (h *GitSshHandler) HandleSession(session ssh.Session, userId string) error {
	cmd := session.RawCommand()
	if cmd == "" {
		_, _ = fmt.Fprint(session, banner)
		_, _ = fmt.Fprintf(session, "\n✅ SSH connection successful!\n")
		_, _ = fmt.Fprintf(session, "🔐 Authentication completed.\n\n")
		return nil
	}

	parts := strings.SplitN(cmd, " ", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid command format: %s", cmd)
	}

	gitCmd := parts[0]
	repoPath := strings.Trim(parts[1], "'\"")
	repoPath = strings.TrimPrefix(repoPath, "/")

	var operation SshOperation
	var safeGitCmd string
	switch gitCmd {
	case "git-upload-pack":
		operation = SshOperationRead
		safeGitCmd = "git-upload-pack"
	case "git-receive-pack":
		operation = SshOperationWrite
		safeGitCmd = "git-receive-pack"
	default:
		return fmt.Errorf("unsupported git command: %s", gitCmd)
	}

	fullRepoPath := h.reposPath + "/" + strings.TrimSuffix(repoPath, ".git")

	hasAccess, err := h.service.ValidateSshAccess(context.Background(), userId, fullRepoPath, operation)
	if err != nil {
		zap.L().Error("Access validation failed", zap.String("userId", userId), zap.Error(err))
		return fmt.Errorf("access validation failed: %w", err)
	}

	if !hasAccess {
		zap.L().Warn("SSH access denied", zap.String("userId", userId), zap.String("operation", string(operation)))
		return fmt.Errorf("permission denied")
	}

	absRepoPath, err := filepath.Abs(fullRepoPath)
	if err != nil {
		zap.L().Error("Failed to get absolute path", zap.String("path", fullRepoPath), zap.Error(err))
		return fmt.Errorf("failed to resolve repository path: %w", err)
	}

	zap.L().Info("Executing Git command",
		zap.String("userId", userId),
		zap.String("command", gitCmd),
		zap.String("repoPath", absRepoPath))

	if _, err := os.Stat(absRepoPath); os.IsNotExist(err) {
		zap.L().Error("Repository path does not exist", zap.String("path", absRepoPath))
		return fmt.Errorf("repository not found: %s", absRepoPath)
	}

	if _, err := os.Stat(filepath.Join(absRepoPath, "HEAD")); os.IsNotExist(err) {
		zap.L().Error("Not a valid git repository", zap.String("path", absRepoPath))
		return fmt.Errorf("not a git repository: %s", absRepoPath)
	}

	var execCmd *exec.Cmd
	switch safeGitCmd {
	case "git-upload-pack":
		// absRepoPath is validated above (exists, is a git repo, within reposPath)
		// #nosec G204 -- command is hardcoded, path is validated and sanitized
		execCmd = exec.Command("git-upload-pack", absRepoPath)
	case "git-receive-pack":
		// absRepoPath is validated above (exists, is a git repo, within reposPath)
		// #nosec G204 -- command is hardcoded, path is validated and sanitized
		execCmd = exec.Command("git-receive-pack", absRepoPath)
	default:
		return fmt.Errorf("unsupported git command: %s", safeGitCmd)
	}

	execCmd.Dir = filepath.Dir(absRepoPath)
	execCmd.Stdin = session
	execCmd.Stdout = session
	execCmd.Stderr = session.Stderr()

	zap.L().Debug("Starting git command",
		zap.String("command", gitCmd),
		zap.String("repoPath", absRepoPath))

	var unlock func()
	if operation == SshOperationWrite {
		unlock = lockRepoRefUpdate(absRepoPath)
	}

	if err := execCmd.Start(); err != nil {
		if unlock != nil {
			unlock()
		}
		zap.L().Error("Failed to start git command", zap.Error(err))
		return fmt.Errorf("failed to start %s: %w", gitCmd, err)
	}

	waitErr := execCmd.Wait()
	if unlock != nil {
		unlock()
	}
	if waitErr != nil {
		return waitErr
	}

	if operation == SshOperationWrite {
		h.triggerSdkGenerationAfterPush(absRepoPath)
	}

	return nil
}

func (h *GitSshHandler) triggerSdkGenerationAfterPush(repoPath string) {
	triggerSdkGenerationAfterPush(context.Background(), h.service, repoPath)
}

func triggerSdkGenerationAfterPush(ctx context.Context, svc Service, repoPath string) {
	repoId := filepath.Base(repoPath)

	commitHash, err := getLatestCommitHash(repoPath)
	if err != nil {
		zap.L().Warn("failed to get latest commit hash for SDK generation",
			zap.String("repoId", repoId),
			zap.Error(err))
		return
	}

	hasProtoFiles, err := svc.HasProtoFiles(ctx, repoPath, commitHash)
	if err != nil {
		zap.L().Warn("failed to check for proto files",
			zap.String("repoId", repoId),
			zap.Error(err))
		return
	}

	if !hasProtoFiles {
		zap.L().Info("skipping SDK generation: repository contains no proto files",
			zap.String("repoId", repoId),
			zap.String("commitHash", commitHash))
		return
	}

	if err := svc.TriggerSdkGeneration(ctx, repoId, commitHash); err != nil {
		zap.L().Error("failed to trigger SDK generation",
			zap.String("repoId", repoId),
			zap.String("commitHash", commitHash),
			zap.Error(err))
	}
}

var (
	repoRefLocksMu sync.Mutex
	repoRefLocks   = make(map[string]*sync.Mutex)
)

func lockRepoRefUpdate(repoPath string) func() {
	key := filepath.Clean(repoPath)
	if abs, err := filepath.Abs(key); err == nil {
		key = abs
	}

	repoRefLocksMu.Lock()
	mu, ok := repoRefLocks[key]
	if !ok {
		mu = &sync.Mutex{}
		repoRefLocks[key] = mu
	}
	repoRefLocksMu.Unlock()

	mu.Lock()
	return mu.Unlock
}

func resolveHeadCommit(repoPath string) (string, error) {
	// #nosec G204 -- repoPath is validated repository path
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func getLatestCommitHash(repoPath string) (string, error) {
	if hash, err := resolveHeadCommit(repoPath); err == nil {
		return hash, nil
	}

	// Coordinate fallback HEAD repair with ref updates so HEAD is only rewritten
	// while its current target remains unborn.
	unlock := lockRepoRefUpdate(repoPath)
	defer unlock()

	if hash, err := resolveHeadCommit(repoPath); err == nil {
		return hash, nil
	}

	// HEAD may point to an unborn branch (e.g. repo created with master, pushed to main).
	// Resolve and point HEAD to an existing branch (preferring main, then master, then first available).
	// #nosec G204 -- repoPath is validated repository path
	branchCmd := exec.Command("git", "for-each-ref", "--format=%(refname:short)", "refs/heads/")
	branchCmd.Dir = repoPath
	branchOut, branchErr := branchCmd.Output()
	if branchErr != nil {
		return "", fmt.Errorf("failed to list branches: %w", branchErr)
	}

	branches := strings.Fields(string(branchOut))
	if len(branches) == 0 {
		return "", errors.New("repository has no commits")
	}

	target := branches[0]
	for _, b := range branches {
		if b == "main" {
			target = "main"
			break
		}
		if b == "master" && target != "main" {
			target = "master"
		}
	}

	if hash, err := resolveHeadCommit(repoPath); err == nil {
		return hash, nil
	}

	// #nosec G204 -- repoPath is validated repository path, target is from refs/heads/
	symCmd := exec.Command("git", "symbolic-ref", "HEAD", "refs/heads/"+target)
	symCmd.Dir = repoPath
	if symErr := symCmd.Run(); symErr != nil {
		return "", fmt.Errorf("failed to update HEAD to %s: %w", target, symErr)
	}

	hash, err := resolveHeadCommit(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash after updating HEAD: %w", err)
	}

	return hash, nil
}
