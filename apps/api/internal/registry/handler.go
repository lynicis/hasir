package registry

import (
	"context"
	"net/http"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/registry/v1/registryv1connect"
	registryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/registry/v1"
	"connectrpc.com/connect"
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

	totalCount, err := h.repository.GetRepositoriesCount(ctx)
	if err != nil {
		return nil, err
	}

	repositories, err := h.repository.GetRepositories(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	var resp []*registryv1.Repository
	for _, repository := range *repositories {
		resp = append(resp, &registryv1.Repository{
			Id:   repository.Id,
			Name: repository.Name,
		})
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	nextPage := int32(page + 1)
	if page >= totalPages {
		nextPage = 0
	}

	return connect.NewResponse(&registryv1.GetRepositoriesResponse{
		Repositories: resp,
		NextPage:     nextPage,
		TotalPage:    int32(totalPages),
	}), nil
}
