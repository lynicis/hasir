package registry

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"hasir/api/pkg/pagination"
	registryv1 "hasir/proto/gen/go/registry/v1"
	"hasir/proto/gen/go/registry/v1/registryv1connect"
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
	page, pageSize := pagination.Normalize(req.Msg.Pagination)

	var organizationId *string
	if req.Msg.OrganizationId != nil {
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

func (h *handler) GetRecentCommit(
	ctx context.Context,
	req *connect.Request[registryv1.GetRecentCommitRequest],
) (*connect.Response[registryv1.Commit], error) {
	commit, err := h.service.GetRecentCommit(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(commit), nil
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
