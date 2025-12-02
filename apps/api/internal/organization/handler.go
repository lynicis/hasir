package organization

import (
	"context"
	"net/http"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/organization/v1/organizationv1connect"
	organizationv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/organization/v1"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"hasir-api/pkg/auth"
	"hasir-api/pkg/proto"
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
) (*connect.Response[organizationv1.GetOrganizationsResponse], error) {
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

	totalCount, err := h.repository.GetOrganizationsCount(ctx)
	if err != nil {
		return nil, err
	}

	organizations, err := h.repository.GetOrganizations(ctx, page, pageSize)
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

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	nextPage := int32(page + 1)
	if page >= totalPages {
		nextPage = 0
	}

	return connect.NewResponse(&organizationv1.GetOrganizationsResponse{
		Organizations: resp,
		NextPage:      nextPage,
		TotalPage:     int32(totalPages),
	}), nil
}

func (h *handler) GetOrganization(
	ctx context.Context,
	req *connect.Request[organizationv1.GetOrganizationRequest],
) (*connect.Response[organizationv1.GetOrganizationResponse], error) {
	org, err := h.repository.GetOrganizationById(ctx, req.Msg.GetId())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&organizationv1.GetOrganizationResponse{
		Organization: &organizationv1.Organization{
			Id:         org.Id,
			Name:       org.Name,
			Visibility: proto.ReverseVisibilityMap[org.Visibility],
		},
	}), nil
}

func (h *handler) UpdateOrganization(
	ctx context.Context,
	req *connect.Request[organizationv1.UpdateOrganizationRequest],
) (*connect.Response[emptypb.Empty], error) {
	userId, err := auth.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.UpdateOrganization(ctx, req.Msg, userId); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) DeleteOrganization(
	ctx context.Context,
	req *connect.Request[organizationv1.DeleteOrganizationRequest],
) (*connect.Response[emptypb.Empty], error) {
	userId, err := auth.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.DeleteOrganization(ctx, req.Msg.GetId(), userId); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) InviteMember(
	ctx context.Context,
	req *connect.Request[organizationv1.InviteMemberRequest],
) (*connect.Response[emptypb.Empty], error) {
	userId, err := auth.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.InviteUser(ctx, req.Msg, userId); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
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
		req.Msg.GetStatus(),
	); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) IsInvitationValid(
	ctx context.Context,
	req *connect.Request[organizationv1.IsInvitationValidRequest],
) (*connect.Response[emptypb.Empty], error) {
	if _, err := h.repository.GetInviteByToken(
		ctx,
		req.Msg.GetToken(),
	); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) GetMembers(
	ctx context.Context,
	req *connect.Request[organizationv1.GetMembersRequest],
) (*connect.Response[organizationv1.GetMembersResponse], error) {
	members, usernames, emails, err := h.repository.GetMembers(ctx, req.Msg.GetId())
	if err != nil {
		return nil, err
	}

	var resp []*organizationv1.Member
	for i, member := range members {
		role := MemberRoleToSharedRoleMap[member.Role]
		resp = append(resp, &organizationv1.Member{
			Id:       member.UserId,
			Username: usernames[i],
			Email:    emails[i],
			Role:     role,
		})
	}

	return connect.NewResponse(&organizationv1.GetMembersResponse{
		Members: resp,
	}), nil
}

func (h *handler) UpdateMemberRole(
	ctx context.Context,
	req *connect.Request[organizationv1.UpdateMemberRoleRequest],
) (*connect.Response[emptypb.Empty], error) {
	userId, err := auth.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.UpdateMemberRole(ctx, req.Msg, userId); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}

func (h *handler) DeleteMember(
	ctx context.Context,
	req *connect.Request[organizationv1.DeleteMemberRequest],
) (*connect.Response[emptypb.Empty], error) {
	userId, err := auth.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.DeleteMember(ctx, req.Msg, userId); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(emptypb.Empty)), nil
}
