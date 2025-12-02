package organization

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"go.uber.org/zap"

	organizationv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/organization/v1"

	"hasir-api/internal/registry"
	"hasir-api/pkg/email"
	"hasir-api/pkg/proto"
)

type Service interface {
	CreateOrganization(
		ctx context.Context,
		req *organizationv1.CreateOrganizationRequest,
		createdBy string,
	) error
	UpdateOrganization(
		ctx context.Context,
		req *organizationv1.UpdateOrganizationRequest,
		userId string,
	) error
	DeleteOrganization(
		ctx context.Context,
		organizationId string,
		userId string,
	) error
	InviteUser(
		ctx context.Context,
		req *organizationv1.InviteMemberRequest,
		invitedBy string,
	) error
	RespondToInvitation(
		ctx context.Context,
		token string,
		userId string,
		accept bool,
	) error
	UpdateMemberRole(
		ctx context.Context,
		req *organizationv1.UpdateMemberRoleRequest,
		updatedBy string,
	) error
	DeleteMember(
		ctx context.Context,
		req *organizationv1.DeleteMemberRequest,
		deletedBy string,
	) error
}

type inviteInfo struct {
	email string
	role  MemberRole
}

type service struct {
	repository      Repository
	emailService    email.Service
	registryService registry.Service
}

func NewService(repository Repository, registryService registry.Service, emailService email.Service) Service {
	return &service{
		repository:      repository,
		emailService:    emailService,
		registryService: registryService,
	}
}

func (s *service) CreateOrganization(
	ctx context.Context,
	req *organizationv1.CreateOrganizationRequest,
	createdBy string,
) error {
	existingOrg, err := s.repository.GetOrganizationByName(ctx, req.GetName())
	if err != nil && !errors.Is(err, ErrOrganizationNotFound) {
		return err
	}

	if existingOrg != nil {
		return connect.NewError(connect.CodeAlreadyExists, errors.New("organization already exists"))
	}

	org := &OrganizationDTO{
		Id:         uuid.NewString(),
		Name:       req.GetName(),
		Visibility: proto.VisibilityMap[req.GetVisibility()],
		CreatedBy:  createdBy,
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.repository.CreateOrganization(ctx, org); err != nil {
		return err
	}

	ownerMember := &OrganizationMemberDTO{
		Id:             uuid.NewString(),
		OrganizationId: org.Id,
		UserId:         createdBy,
		Role:           MemberRoleOwner,
		JoinedAt:       time.Now().UTC(),
	}

	if err := s.repository.AddMember(ctx, ownerMember); err != nil {
		zap.L().Error("failed to add creator as owner", zap.Error(err), zap.String("organizationId", org.Id))
		return err
	}

	var invites []inviteInfo
	for _, member := range req.GetMembers() {
		email := member.GetEmail()
		role := SharedRoleToMemberRoleMap[member.GetRole()]
		invites = append(invites, inviteInfo{email: email, role: role})
	}

	if len(invites) > 0 {
		if err := s.sendInvites(ctx, org.Id, org.Name, createdBy, invites); err != nil {
			zap.L().Error("failed to send invites", zap.Error(err), zap.String("organizationId", org.Id))
		}
	}

	return nil
}

func (s *service) InviteUser(
	ctx context.Context,
	req *organizationv1.InviteMemberRequest,
	invitedBy string,
) error {
	org, err := s.repository.GetOrganizationById(ctx, req.GetId())
	if err != nil {
		return err
	}

	role, err := s.repository.GetMemberRole(ctx, req.GetId(), invitedBy)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return connect.NewError(connect.CodePermissionDenied, errors.New("you are not a member of this organization"))
		}
		return err
	}

	if role != MemberRoleOwner {
		return connect.NewError(connect.CodePermissionDenied, errors.New("only organization owners can invite users"))
	}

	email := req.GetEmail()
	if email == "" {
		return nil
	}

	invites := []inviteInfo{
		{email: email, role: MemberRoleAuthor},
	}

	if err := s.sendInvites(ctx, org.Id, org.Name, invitedBy, invites); err != nil {
		zap.L().Error("failed to send invites", zap.Error(err), zap.String("organizationId", org.Id))
	}

	return nil
}

func (s *service) sendInvites(ctx context.Context, orgId, orgName, invitedBy string, invites []inviteInfo) error {
	var organizationInvites []*OrganizationInviteDTO
	var emailJobs []*EmailJobDTO
	now := time.Now().UTC()

	for _, inviteData := range invites {
		token, err := generateInviteToken()
		if err != nil {
			zap.L().Error("failed to generate invite token", zap.Error(err), zap.String("email", inviteData.email))
			continue
		}

		invite := &OrganizationInviteDTO{
			Id:             uuid.NewString(),
			OrganizationId: orgId,
			Email:          inviteData.email,
			Token:          token,
			InvitedBy:      invitedBy,
			Role:           inviteData.role,
			Status:         InviteStatusPending,
			CreatedAt:      now,
			ExpiresAt:      now.AddDate(0, 0, 7),
		}
		organizationInvites = append(organizationInvites, invite)

		emailJob := &EmailJobDTO{
			Id:               uuid.NewString(),
			InviteId:         invite.Id,
			OrganizationId:   orgId,
			Email:            inviteData.email,
			OrganizationName: orgName,
			InviteToken:      token,
			Status:           EmailJobStatusPending,
			Attempts:         0,
			MaxAttempts:      3,
			CreatedAt:        now,
		}
		emailJobs = append(emailJobs, emailJob)
	}

	if err := s.repository.CreateInvites(ctx, organizationInvites); err != nil {
		zap.L().Error("failed to create invites", zap.Error(err), zap.String("organizationId", orgId))
		return err
	}

	if err := s.repository.EnqueueEmailJobs(ctx, emailJobs); err != nil {
		zap.L().Error("failed to enqueue email jobs", zap.Error(err), zap.String("organizationId", orgId))
		return err
	}
	zap.L().Info("enqueued email jobs for batch processing",
		zap.Int("count", len(emailJobs)),
		zap.String("organizationId", orgId))

	return nil
}

func (s *service) UpdateOrganization(
	ctx context.Context,
	req *organizationv1.UpdateOrganizationRequest,
	userId string,
) error {
	org, err := s.repository.GetOrganizationById(ctx, req.GetId())
	if err != nil {
		return err
	}

	role, err := s.repository.GetMemberRole(ctx, req.GetId(), userId)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return connect.NewError(connect.CodePermissionDenied, errors.New("you are not a member of this organization"))
		}
		return err
	}

	if role != MemberRoleOwner {
		return connect.NewError(connect.CodePermissionDenied, errors.New("only organization owners can update the organization"))
	}

	org.Name = req.GetName()
	org.Visibility = proto.VisibilityMap[req.GetVisibility()]
	if err := s.repository.UpdateOrganization(ctx, org); err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteOrganization(
	ctx context.Context,
	organizationId string,
	userId string,
) error {
	role, err := s.repository.GetMemberRole(ctx, organizationId, userId)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return connect.NewError(connect.CodePermissionDenied, errors.New("you are not a member of this organization"))
		}
		return err
	}

	if role != MemberRoleOwner {
		return connect.NewError(connect.CodePermissionDenied, errors.New("only organization owners can delete the organization"))
	}

	if err := s.registryService.DeleteRepositoriesByOrganization(ctx, organizationId); err != nil {
		return err
	}

	if err := s.repository.DeleteOrganization(ctx, organizationId); err != nil {
		return err
	}

	zap.L().Info("organization deleted",
		zap.String("organizationId", organizationId),
		zap.String("userId", userId),
	)

	return nil
}

func generateInviteToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func (s *service) RespondToInvitation(
	ctx context.Context,
	token string,
	userId string,
	accept bool,
) error {
	invite, err := s.repository.GetInviteByToken(ctx, token)
	if err != nil {
		return err
	}

	if invite.Status != InviteStatusPending {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("invite is no longer pending"))
	}

	if time.Now().UTC().After(invite.ExpiresAt) {
		if err := s.repository.UpdateInviteStatus(ctx, invite.Id, InviteStatusExpired, nil); err != nil {
			zap.L().Error("failed to update expired invite status", zap.Error(err), zap.String("inviteId", invite.Id))
		}
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("invite has expired"))
	}

	if !accept {
		if err := s.repository.UpdateInviteStatus(ctx, invite.Id, InviteStatusCancelled, nil); err != nil {
			return err
		}
		zap.L().Info("invite rejected", zap.String("inviteId", invite.Id), zap.String("userId", userId))
		return nil
	}

	now := time.Now().UTC()
	if err := s.repository.UpdateInviteStatus(ctx, invite.Id, InviteStatusAccepted, &now); err != nil {
		return err
	}

	member := &OrganizationMemberDTO{
		Id:             uuid.NewString(),
		OrganizationId: invite.OrganizationId,
		UserId:         userId,
		Role:           invite.Role,
		JoinedAt:       now,
	}

	if err := s.repository.AddMember(ctx, member); err != nil {
		if errors.Is(err, ErrMemberAlreadyExists) {
			zap.L().Warn("user is already a member", zap.String("userId", userId), zap.String("organizationId", invite.OrganizationId))
			return nil
		}
		return err
	}

	zap.L().Info("invite accepted and member added",
		zap.String("inviteId", invite.Id),
		zap.String("userId", userId),
		zap.String("organizationId", invite.OrganizationId),
	)

	return nil
}

func (s *service) UpdateMemberRole(
	ctx context.Context,
	req *organizationv1.UpdateMemberRoleRequest,
	updatedBy string,
) error {
	organizationId := req.GetOrganizationId()
	memberUserId := req.GetMemberId()
	newRole := SharedRoleToMemberRoleMap[req.GetRole()]

	_, err := s.repository.GetOrganizationById(ctx, organizationId)
	if err != nil {
		return err
	}

	updaterRole, err := s.repository.GetMemberRole(ctx, organizationId, updatedBy)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return connect.NewError(connect.CodePermissionDenied, errors.New("you are not a member of this organization"))
		}
		return err
	}

	if updaterRole != MemberRoleOwner {
		return connect.NewError(connect.CodePermissionDenied, errors.New("only organization owners can update member roles"))
	}

	currentMemberRole, err := s.repository.GetMemberRole(ctx, organizationId, memberUserId)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return connect.NewError(connect.CodeNotFound, errors.New("member not found"))
		}
		return err
	}

	if updatedBy == memberUserId && currentMemberRole == MemberRoleOwner && newRole != MemberRoleOwner {
		return connect.NewError(connect.CodePermissionDenied, errors.New("owners cannot decrease their own role"))
	}

	if newRole != MemberRoleOwner {
		members, _, _, err := s.repository.GetMembers(ctx, organizationId)
		if err != nil {
			return err
		}

		ownerCount := 0
		for _, member := range members {
			if member.Role == MemberRoleOwner {
				ownerCount++
			}
		}

		if currentMemberRole == MemberRoleOwner && ownerCount == 1 {
			return connect.NewError(connect.CodeFailedPrecondition, errors.New("cannot change role of the last owner"))
		}
	}

	if err := s.repository.UpdateMemberRole(ctx, organizationId, memberUserId, newRole); err != nil {
		return err
	}

	zap.L().Info("member role updated",
		zap.String("organizationId", organizationId),
		zap.String("memberUserId", memberUserId),
		zap.String("newRole", string(newRole)),
		zap.String("updatedBy", updatedBy),
	)

	return nil
}

func (s *service) DeleteMember(
	ctx context.Context,
	req *organizationv1.DeleteMemberRequest,
	deletedBy string,
) error {
	organizationId := req.GetOrganizationId()
	memberUserId := req.GetMemberId()

	_, err := s.repository.GetOrganizationById(ctx, organizationId)
	if err != nil {
		return err
	}

	deleterRole, err := s.repository.GetMemberRole(ctx, organizationId, deletedBy)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return connect.NewError(connect.CodePermissionDenied, errors.New("you are not a member of this organization"))
		}
		return err
	}

	if deleterRole != MemberRoleOwner {
		return connect.NewError(connect.CodePermissionDenied, errors.New("only organization owners can delete members"))
	}

	memberRole, err := s.repository.GetMemberRole(ctx, organizationId, memberUserId)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return connect.NewError(connect.CodeNotFound, errors.New("member not found"))
		}
		return err
	}

	// Prevent deleting the last owner
	if memberRole == MemberRoleOwner {
		members, _, _, err := s.repository.GetMembers(ctx, organizationId)
		if err != nil {
			return err
		}

		ownerCount := 0
		for _, member := range members {
			if member.Role == MemberRoleOwner {
				ownerCount++
			}
		}

		if ownerCount == 1 {
			return connect.NewError(connect.CodeFailedPrecondition, errors.New("cannot delete the last owner"))
		}
	}

	if err := s.repository.DeleteMember(ctx, organizationId, memberUserId); err != nil {
		return err
	}

	zap.L().Info("member deleted",
		zap.String("organizationId", organizationId),
		zap.String("memberUserId", memberUserId),
		zap.String("deletedBy", deletedBy),
	)

	return nil
}
