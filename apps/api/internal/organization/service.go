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

	"hasir-api/pkg/email"
)

type Service interface {
	CreateOrganization(
		ctx context.Context,
		req *organizationv1.CreateOrganizationRequest,
		createdBy string,
	) error
	RespondToInvitation(
		ctx context.Context,
		token string,
		userId string,
		accept bool,
	) error
}

type service struct {
	repository   Repository
	emailService email.Service
}

func NewService(repository Repository, emailService email.Service) Service {
	return &service{
		repository:   repository,
		emailService: emailService,
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

	visibility := protoVisibilityMap[req.GetVisibility()]
	org := &OrganizationDTO{
		Id:         uuid.NewString(),
		Name:       req.GetName(),
		Visibility: visibility,
		CreatedBy:  createdBy,
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.repository.CreateOrganization(ctx, org); err != nil {
		return err
	}

	inviteEmails := req.GetInviteEmails()
	if len(inviteEmails) > 0 {
		if err := s.sendInvites(ctx, org.Id, org.Name, createdBy, inviteEmails); err != nil {
			zap.L().Error("failed to send invites", zap.Error(err), zap.String("organizationId", org.Id))
		}
	}

	return nil
}

func (s *service) sendInvites(ctx context.Context, orgId, orgName, invitedBy string, emails []string) error {
	for _, emailAddr := range emails {
		token, err := generateInviteToken()
		if err != nil {
			zap.L().Error("failed to generate invite token", zap.Error(err), zap.String("email", emailAddr))
			continue
		}

		now := time.Now().UTC()
		invite := &OrganizationInviteDTO{
			Id:             uuid.NewString(),
			OrganizationId: orgId,
			Email:          emailAddr,
			Token:          token,
			InvitedBy:      invitedBy,
			Status:         InviteStatusPending,
			CreatedAt:      now,
			ExpiresAt:      now.AddDate(0, 0, 7),
		}

		if err := s.repository.CreateInvite(ctx, invite); err != nil {
			zap.L().Error("failed to create invite", zap.Error(err), zap.String("email", emailAddr))
			continue
		}

		if err := s.emailService.SendInvite(emailAddr, orgName, token); err != nil {
			zap.L().Error("failed to send invite email", zap.Error(err), zap.String("email", emailAddr))
			continue
		}

		zap.L().Info("invite sent successfully", zap.String("email", emailAddr), zap.String("organizationId", orgId))
	}

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
		Role:           MemberRoleMember,
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
