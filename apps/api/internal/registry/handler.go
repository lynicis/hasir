package registry

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/registry/v1/registryv1connect"
	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
	"connectrpc.com/connect"
	"github.com/gliderlabs/ssh"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"

	"hasir-api/internal/user"
)

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

	resp, err := h.service.GetRepositories(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
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
		return fmt.Errorf("interactive shell access is not supported, use git commands")
	}

	parts := strings.SplitN(cmd, " ", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid command format: %s", cmd)
	}

	gitCmd := parts[0]
	repoPath := strings.Trim(parts[1], "'\"")
	repoPath = strings.TrimPrefix(repoPath, "/")

	var operation SshOperation
	switch gitCmd {
	case "git-upload-pack":
		operation = SshOperationRead
	case "git-receive-pack":
		operation = SshOperationWrite
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

	zap.L().Info("Executing Git command", zap.String("userId", userId), zap.String("command", gitCmd))

	execCmd := exec.Command(gitCmd, fullRepoPath)
	execCmd.Dir = fullRepoPath
	execCmd.Stdin = session
	execCmd.Stdout = session
	execCmd.Stderr = session.Stderr()

	if err := execCmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", gitCmd, err)
	}

	return execCmd.Wait()
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
	path := strings.TrimPrefix(r.URL.Path, "/git/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	repoUUID := parts[0]
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

	userId, err := h.authenticate(r)
	if err != nil {
		h.requireAuth(w)
		return
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
	_, password, ok := r.BasicAuth()
	if !ok || password == "" {
		return "", fmt.Errorf("missing credentials")
	}

	userDTO, err := h.userRepo.GetUserByApiKey(r.Context(), password)
	if err != nil {
		return "", err
	}

	return userDTO.Id, nil
}

func (h *GitHttpHandler) handleInfoRefs(w http.ResponseWriter, r *http.Request, repoPath string) {
	serviceName := r.URL.Query().Get("service")
	if serviceName != "git-upload-pack" && serviceName != "git-receive-pack" {
		http.Error(w, "Invalid service", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", serviceName))
	w.Header().Set("Cache-Control", "no-cache")

	pktLine := fmt.Sprintf("# service=%s\n", serviceName)
	fmt.Fprintf(w, "%04x%s", len(pktLine)+4, pktLine)
	fmt.Fprint(w, "0000")

	cmd := exec.Command(serviceName, "--stateless-rpc", "--advertise-refs", repoPath)
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

	cmd := exec.Command("git-receive-pack", "--stateless-rpc", repoPath)
	cmd.Stdin = r.Body
	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Run(); err != nil {
		zap.L().Error("git-receive-pack failed", zap.Error(err))
	}
}
