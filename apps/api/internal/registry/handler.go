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

type SshHandler struct {
	service   Service
	reposPath string
}
func NewSshHandler(service Service, reposPath string) *SshHandler {
	return &SshHandler{
		service:   service,
		reposPath: reposPath,
	}
}

func (h *SshHandler) HandleSession(session ssh.Session, userId string) error {
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
