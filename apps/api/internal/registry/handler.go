package registry

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/registry/v1/registryv1connect"
	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
	"connectrpc.com/connect"
	"github.com/gliderlabs/ssh"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"

	"hasir-api/internal/user"
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

type handler struct {
	interceptors []connect.Interceptor
	service      Service
	repository   Repository
}

func NewHandler(service Service, repository Repository, interceptors ...connect.Interceptor) *handler {
	return &handler{
		interceptors: interceptors,
		service:      service,
		repository:   repository,
	}
}

func (h *handler) RegisterRoutes() (string, http.Handler) {
	return registryv1connect.NewRegistryServiceHandler(
		h,
		connect.WithInterceptors(h.interceptors...),
	)
}

func (h *handler) CreateRepository(
	ctx context.Context,
	req *connect.Request[registryv1.CreateRepositoryRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.service.CreateRepository(ctx, req.Msg); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) GetRepository(
	ctx context.Context,
	req *connect.Request[registryv1.GetRepositoryRequest],
) (*connect.Response[registryv1.Repository], error) {
	repo, err := h.service.GetRepository(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(repo), nil
}

func (h *handler) GetRepositories(
	ctx context.Context,
	req *connect.Request[registryv1.GetRepositoriesRequest],
) (*connect.Response[registryv1.GetRepositoriesResponse], error) {
	page := 1
	pageSize := 10

	if req.Msg.Pagination.GetPage() > 0 {
		page = int(req.Msg.Pagination.GetPage())
	}
	if req.Msg.Pagination.GetPageLimit() > 0 {
		pageSize = int(req.Msg.Pagination.GetPageLimit())
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

	var organizationId *string
	if req.Msg.HasOrganizationId() {
		orgId := req.Msg.GetOrganizationId()
		organizationId = &orgId
	}

	resp, err := h.service.GetRepositories(ctx, organizationId, page, pageSize)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func (h *handler) UpdateRepository(
	ctx context.Context,
	req *connect.Request[registryv1.UpdateRepositoryRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.service.UpdateRepository(ctx, req.Msg); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) DeleteRepository(
	ctx context.Context,
	req *connect.Request[registryv1.DeleteRepositoryRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.service.DeleteRepository(ctx, req.Msg); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) UpdateSdkPreferences(
	ctx context.Context,
	req *connect.Request[registryv1.UpdateSdkPreferencesRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.service.UpdateSdkPreferences(ctx, req.Msg); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) GetCommits(
	ctx context.Context,
	req *connect.Request[registryv1.GetCommitsRequest],
) (*connect.Response[registryv1.GetCommitsResponse], error) {
	commits, err := h.service.GetCommits(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(commits), nil
}

func (h *handler) GetFileTree(
	ctx context.Context,
	req *connect.Request[registryv1.GetFileTreeRequest],
) (*connect.Response[registryv1.GetFileTreeResponse], error) {
	fileTree, err := h.service.GetFileTree(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(fileTree), nil
}

func (h *handler) GetFilePreview(
	ctx context.Context,
	req *connect.Request[registryv1.GetFilePreviewRequest],
) (*connect.Response[registryv1.GetFilePreviewResponse], error) {
	filePreview, err := h.service.GetFilePreview(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(filePreview), nil
}

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

	if err := execCmd.Start(); err != nil {
		zap.L().Error("Failed to start git command", zap.Error(err))
		return fmt.Errorf("failed to start %s: %w", gitCmd, err)
	}

	if err := execCmd.Wait(); err != nil {
		return err
	}

	if operation == SshOperationWrite {
		h.triggerSdkGenerationAfterPush(absRepoPath)
	}

	return nil
}

func (h *GitSshHandler) triggerSdkGenerationAfterPush(repoPath string) {
	ctx := context.Background()
	repoId := filepath.Base(repoPath)

	commitHash, err := h.getLatestCommitHash(repoPath)
	if err != nil {
		zap.L().Warn("failed to get latest commit hash for SDK generation",
			zap.String("repoId", repoId),
			zap.Error(err))
		return
	}

	hasProtoFiles, err := h.service.HasProtoFiles(ctx, repoPath)
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

	if err := h.service.TriggerSdkGeneration(ctx, repoId, commitHash); err != nil {
		zap.L().Error("failed to trigger SDK generation",
			zap.String("repoId", repoId),
			zap.String("commitHash", commitHash),
			zap.Error(err))
	}
}

func (h *GitSshHandler) getLatestCommitHash(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

type GitHttpHandler struct {
	service   Service
	userRepo  user.Repository
	reposPath string
}

func NewGitHttpHandler(service Service, userRepo user.Repository, reposPath string) *GitHttpHandler {
	return &GitHttpHandler{
		service:   service,
		userRepo:  userRepo,
		reposPath: reposPath,
	}
}

func (h *GitHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userId, err := h.authenticate(r)
	if err != nil {
		h.requireAuth(w)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/git/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	repoUUID := strings.TrimSuffix(parts[0], ".git")
	repoPath := h.reposPath + "/" + repoUUID
	subPath := ""
	if len(parts) > 1 {
		subPath = parts[1]
	}

	var operation SshOperation
	serviceName := r.URL.Query().Get("service")
	if serviceName == "git-receive-pack" || subPath == "git-receive-pack" {
		operation = SshOperationWrite
	} else {
		operation = SshOperationRead
	}

	hasAccess, err := h.service.ValidateSshAccess(r.Context(), userId, repoPath, operation)
	if err != nil {
		zap.L().Error("Access validation failed", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !hasAccess {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	switch {
	case subPath == "info/refs":
		h.handleInfoRefs(w, r, repoPath)
	case subPath == "git-upload-pack" && r.Method == http.MethodPost:
		h.handleUploadPack(w, r, repoPath)
	case subPath == "git-receive-pack" && r.Method == http.MethodPost:
		h.handleReceivePack(w, r, repoPath)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (h *GitHttpHandler) requireAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Git Repository"`)
	http.Error(w, "Authentication required", http.StatusUnauthorized)
}

func (h *GitHttpHandler) authenticate(r *http.Request) (string, error) {
	_, apiKey, ok := r.BasicAuth()
	if !ok || apiKey == "" {
		return "", fmt.Errorf("missing credentials")
	}

	userDTO, err := h.userRepo.GetUserByApiKey(r.Context(), apiKey)
	if err != nil {
		return "", err
	}

	return userDTO.Id, nil
}

func (h *GitHttpHandler) handleInfoRefs(w http.ResponseWriter, r *http.Request, repoPath string) {
	serviceName := r.URL.Query().Get("service")

	var gitCommand string
	switch serviceName {
	case "git-upload-pack":
		gitCommand = "git-upload-pack"
	case "git-receive-pack":
		gitCommand = "git-receive-pack"
	default:
		http.Error(w, "Invalid service", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", serviceName))
	w.Header().Set("Cache-Control", "no-cache")

	pktLine := fmt.Sprintf("# service=%s\n", serviceName)
	_, _ = fmt.Fprintf(w, "%04x%s", len(pktLine)+4, pktLine)
	_, _ = fmt.Fprint(w, "0000")

	var cmd *exec.Cmd
	switch gitCommand {
	case "git-upload-pack":
		cmd = exec.Command("git-upload-pack", "--stateless-rpc", "--advertise-refs", repoPath)
	case "git-receive-pack":
		cmd = exec.Command("git-receive-pack", "--stateless-rpc", "--advertise-refs", repoPath)
	default:
		http.Error(w, "Invalid service", http.StatusBadRequest)
		return
	}

	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		zap.L().Error("Failed to run git command", zap.String("service", serviceName), zap.Error(err))
	}
}

func (h *GitHttpHandler) handleUploadPack(w http.ResponseWriter, r *http.Request, repoPath string) {
	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	w.Header().Set("Cache-Control", "no-cache")

	cmd := exec.Command("git-upload-pack", "--stateless-rpc", repoPath)
	cmd.Stdin = r.Body
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		zap.L().Error("git-upload-pack failed", zap.Error(err))
	}
}

func (h *GitHttpHandler) handleReceivePack(w http.ResponseWriter, r *http.Request, repoPath string) {
	w.Header().Set("Content-Type", "application/x-git-receive-pack-result")
	w.Header().Set("Cache-Control", "no-cache")

	message := fmt.Sprintf("\x02%s", banner)
	pktLine := fmt.Sprintf("%04x%s", len(message)+4, message)
	_, _ = w.Write([]byte(pktLine))

	cmd := exec.Command("git-receive-pack", "--stateless-rpc", repoPath)
	cmd.Stdin = r.Body
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		zap.L().Error("git-receive-pack failed", zap.Error(err))
		return
	}

	repoId := filepath.Base(repoPath)
	commitHash, err := h.getLatestCommitHash(repoPath)
	if err != nil {
		zap.L().Warn("failed to get latest commit hash for SDK generation",
			zap.String("repoId", repoId),
			zap.Error(err))
		return
	}

	hasProtoFiles, err := h.service.HasProtoFiles(r.Context(), repoPath)
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

	if err := h.service.TriggerSdkGeneration(r.Context(), repoId, commitHash); err != nil {
		zap.L().Error("failed to trigger SDK generation",
			zap.String("repoId", repoId),
			zap.String("commitHash", commitHash),
			zap.Error(err))
	}
}

func (h *GitHttpHandler) getLatestCommitHash(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

type SdkHttpHandler struct {
	sdkReposPath string
}

func NewSdkHttpHandler(sdkReposPath string) *SdkHttpHandler {
	return &SdkHttpHandler{
		sdkReposPath: sdkReposPath,
	}
}

func isValidPathComponent(component string) bool {
	if strings.Contains(component, "..") {
		return false
	}

	if filepath.IsAbs(component) {
		return false
	}

	if strings.ContainsAny(component, "/\\") {
		return false
	}

	cleaned := filepath.Clean(component)
	if cleaned != component || cleaned == "." {
		return false
	}

	return true
}

func (h *SdkHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/sdk/")
	parts := strings.SplitN(path, "/", 4)

	if len(parts) < 3 {
		http.Error(w, "Invalid SDK path. Format: /sdk/{orgId}/{repoId}/{sdkType}/", http.StatusNotFound)
		return
	}

	orgId := parts[0]
	repoId := strings.TrimSuffix(parts[1], ".git")
	sdkType := strings.TrimSuffix(parts[2], ".git")
	if !isValidPathComponent(orgId) || !isValidPathComponent(repoId) || !isValidPathComponent(sdkType) {
		http.Error(w, "Invalid path component", http.StatusBadRequest)
		return
	}

	repoPath := filepath.Join(h.sdkReposPath, orgId, repoId, sdkType)
	absRepoPath, err := filepath.Abs(repoPath)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	absSdkReposPath, err := filepath.Abs(h.sdkReposPath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !strings.HasPrefix(absRepoPath, absSdkReposPath+string(filepath.Separator)) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		http.Error(w, "SDK repository not found", http.StatusNotFound)
		return
	}

	subPath := ""
	if len(parts) > 3 {
		subPath = parts[3]
	}

	serviceName := r.URL.Query().Get("service")
	if serviceName == "git-receive-pack" || subPath == "git-receive-pack" {
		http.Error(w, "SDK repositories are read-only", http.StatusForbidden)
		return
	}

	switch {
	case subPath == "info/refs":
		h.handleInfoRefs(w, r, repoPath)
	case subPath == "git-upload-pack" && r.Method == http.MethodPost:
		h.handleUploadPack(w, r, repoPath)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (h *SdkHttpHandler) handleInfoRefs(w http.ResponseWriter, r *http.Request, repoPath string) {
	serviceName := r.URL.Query().Get("service")
	if serviceName != "git-upload-pack" {
		http.Error(w, "Only git-upload-pack is supported for SDK repos", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", serviceName))
	w.Header().Set("Cache-Control", "no-cache")

	pktLine := fmt.Sprintf("# service=%s\n", serviceName)
	_, _ = fmt.Fprintf(w, "%04x%s", len(pktLine)+4, pktLine)
	_, _ = fmt.Fprint(w, "0000")

	cmd := exec.Command("git-upload-pack", "--stateless-rpc", "--advertise-refs", repoPath)
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		zap.L().Error("Failed to run git command for SDK", zap.String("service", serviceName), zap.Error(err))
	}
}

func (h *SdkHttpHandler) handleUploadPack(w http.ResponseWriter, r *http.Request, repoPath string) {
	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	w.Header().Set("Cache-Control", "no-cache")

	cmd := exec.Command("git-upload-pack", "--stateless-rpc", repoPath)
	cmd.Stdin = r.Body
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		zap.L().Error("git-upload-pack failed for SDK", zap.Error(err))
	}
}

type SdkSshHandler struct {
	sdkReposPath string
}

func NewSdkSshHandler(sdkReposPath string) *SdkSshHandler {
	return &SdkSshHandler{
		sdkReposPath: sdkReposPath,
	}
}

func (h *SdkSshHandler) HandleSession(session ssh.Session, userId string) error {
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

	if gitCmd != "git-upload-pack" {
		if gitCmd == "git-receive-pack" {
			return fmt.Errorf("SDK repositories are read-only")
		}
		return fmt.Errorf("unsupported git command: %s", gitCmd)
	}

	if !strings.HasPrefix(repoPath, "sdk/") {
		return fmt.Errorf("invalid SDK path format, must start with 'sdk/'")
	}

	path := strings.TrimPrefix(repoPath, "sdk/")
	pathParts := strings.SplitN(path, "/", 4)
	if len(pathParts) < 3 {
		return fmt.Errorf("invalid SDK path format: sdk/{orgId}/{repoId}/{sdkType}")
	}

	orgId := pathParts[0]
	repoId := strings.TrimSuffix(pathParts[1], ".git")
	sdkType := strings.TrimSuffix(pathParts[2], ".git")
	if !isValidPathComponent(orgId) || !isValidPathComponent(repoId) || !isValidPathComponent(sdkType) {
		return fmt.Errorf("invalid path component")
	}

	fullRepoPath := filepath.Join(h.sdkReposPath, orgId, repoId, sdkType)

	absRepoPath, err := filepath.Abs(fullRepoPath)
	if err != nil {
		return fmt.Errorf("invalid path")
	}
	absSdkReposPath, err := filepath.Abs(h.sdkReposPath)
	if err != nil {
		return fmt.Errorf("internal server error")
	}
	if !strings.HasPrefix(absRepoPath, absSdkReposPath+string(filepath.Separator)) {
		return fmt.Errorf("invalid path")
	}

	zap.L().Info("Executing Git command for SDK",
		zap.String("userId", userId),
		zap.String("command", gitCmd),
		zap.String("repoPath", absRepoPath))

	if _, err := os.Stat(absRepoPath); os.IsNotExist(err) {
		zap.L().Error("SDK repository path does not exist", zap.String("path", absRepoPath))
		return fmt.Errorf("SDK repository not found: %s", absRepoPath)
	}

	if _, err := os.Stat(filepath.Join(absRepoPath, ".git")); os.IsNotExist(err) {
		zap.L().Error("Not a valid git repository", zap.String("path", absRepoPath))
		return fmt.Errorf("not a git repository: %s", absRepoPath)
	}

	// absRepoPath is validated above and constrained to be within sdkReposPath
	// #nosec G204 -- command is hardcoded, path is validated and sanitized
	execCmd := exec.Command("git-upload-pack", absRepoPath)
	execCmd.Dir = filepath.Dir(absRepoPath)
	execCmd.Stdin = session
	execCmd.Stdout = session
	execCmd.Stderr = session.Stderr()

	zap.L().Debug("Starting git command for SDK",
		zap.String("command", gitCmd),
		zap.String("repoPath", absRepoPath))

	if err := execCmd.Start(); err != nil {
		zap.L().Error("Failed to start git command", zap.Error(err))
		return fmt.Errorf("failed to start %s: %w", gitCmd, err)
	}

	return execCmd.Wait()
}
