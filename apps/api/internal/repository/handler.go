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
}

func NewHandler(service Service, interceptors ...connect.Interceptor) *handler {
	return &handler{
		interceptors: interceptors,
		service:      service,
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
