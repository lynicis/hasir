package repository

import (
	"context"
	"net/http"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/repository/v1/repositoryv1connect"
	repositoryv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/repository/v1"
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
	return repositoryv1connect.NewRepositoryServiceHandler(
		h,
		connect.WithInterceptors(h.interceptors...),
	)
}

func (h *handler) CreateRepository(
	ctx context.Context,
	req *connect.Request[repositoryv1.CreateRepositoryRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.service.CreateRepository(ctx, req.Msg); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) GetRepositories(
	ctx context.Context,
	req *connect.Request[repositoryv1.GetRepositoriesRequest],
) (*connect.Response[repositoryv1.GetRepositoriesResponse], error) {
	repositories, err := h.repository.GetRepositories(ctx)
	if err != nil {
		return nil, err
	}

	var resp []*repositoryv1.Repository
	for _, repository := range *repositories {
		resp = append(resp, &repositoryv1.Repository{
			Id:   repository.Id,
			Name: repository.Name,
		})
	}

	return connect.NewResponse(&repositoryv1.GetRepositoriesResponse{
		Repositories: resp,
	}), nil
}
