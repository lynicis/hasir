package organization

import (
	"context"
	"net/http"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/organization/v1/organizationv1connect"
	organizationv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/organization/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"apps/api/pkg/auth"
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
	return organizationv1connect.NewOrganizationServiceHandler(
		h,
		connect.WithInterceptors(h.interceptors...),
	)
}

func (h *handler) CreateOrganization(
	ctx context.Context,
	req *connect.Request[organizationv1.CreateOrganizationRequest],
) (*connect.Response[emptypb.Empty], error) {
	createdBy, err := auth.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.CreateOrganization(ctx, req.Msg, createdBy); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) GetOrganizations(
	ctx context.Context,
	req *connect.Request[organizationv1.GetOrganizationsRequest],
) (*connect.Response[organizationv1.GetRepositoriesResponse], error) {
	organizations, err := h.repository.GetOrganizations(ctx)
	if err != nil {
		return nil, err
	}

	var resp []*organizationv1.Organization
	for _, org := range *organizations {
		resp = append(resp, &organizationv1.Organization{
			Id:   org.Id,
			Name: org.Name,
		})
	}

	return connect.NewResponse(&organizationv1.GetRepositoriesResponse{
		Organizations: resp,
	}), nil
}

func (h *handler) RespondToInvitation(
	ctx context.Context,
	req *connect.Request[organizationv1.RespondToInvitationRequest],
) (*connect.Response[emptypb.Empty], error) {
	userId, err := auth.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.RespondToInvitation(
		ctx,
		req.Msg.GetInvitationId(),
		userId,
		req.Msg.GetAccept(),
	); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}
