package organization

import (
	"context"
	"errors"
	"testing"
	"time"

	organizationv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/organization/v1"
	"buf.build/gen/go/hasir/hasir/protocolbuffers/go/shared"
	"connectrpc.com/connect"
	"go.uber.org/mock/gomock"

	"apps/api/pkg/email"
)

func TestCreateOrganization_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	req := &organizationv1.CreateOrganizationRequest{
		Name:       "test-org",
		Visibility: shared.Visibility_VISIBILITY_PRIVATE,
	}
	createdBy := "user-123"

	mockRepo.EXPECT().
		GetOrganizationByName(ctx, "test-org").
		Return(nil, ErrOrganizationNotFound)

	mockRepo.EXPECT().
		CreateOrganization(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, org *OrganizationDTO) error {
			if org.Name != "test-org" {
				t.Errorf("expected name 'test-org', got %s", org.Name)
			}
			if org.Visibility != VisibilityPrivate {
				t.Errorf("expected visibility 'private', got %s", org.Visibility)
			}
			if org.CreatedBy != createdBy {
				t.Errorf("expected createdBy '%s', got %s", createdBy, org.CreatedBy)
			}
			return nil
		})

	err := svc.CreateOrganization(ctx, req, createdBy)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateOrganization_WithInvites(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	req := &organizationv1.CreateOrganizationRequest{
		Name:         "test-org",
		Visibility:   shared.Visibility_VISIBILITY_PUBLIC,
		InviteEmails: []string{"friend1@example.com", "friend2@example.com"},
	}
	createdBy := "user-123"

	mockRepo.EXPECT().
		GetOrganizationByName(ctx, "test-org").
		Return(nil, ErrOrganizationNotFound)

	mockRepo.EXPECT().
		CreateOrganization(ctx, gomock.Any()).
		Return(nil)

	// Expect invites to be created for each email
	mockRepo.EXPECT().
		CreateInvite(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, invite *OrganizationInviteDTO) error {
			if invite.Email != "friend1@example.com" && invite.Email != "friend2@example.com" {
				t.Errorf("unexpected email: %s", invite.Email)
			}
			if invite.Status != InviteStatusPending {
				t.Errorf("expected status 'pending', got %s", invite.Status)
			}
			return nil
		}).Times(2)

	mockEmail.EXPECT().
		SendInvite(gomock.Any(), "test-org", gomock.Any()).
		Return(nil).Times(2)

	err := svc.CreateOrganization(ctx, req, createdBy)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateOrganization_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	req := &organizationv1.CreateOrganizationRequest{
		Name:       "existing-org",
		Visibility: shared.Visibility_VISIBILITY_PRIVATE,
	}
	createdBy := "user-123"

	existingOrg := &OrganizationDTO{
		Id:   "existing-id",
		Name: "existing-org",
	}

	mockRepo.EXPECT().
		GetOrganizationByName(ctx, "existing-org").
		Return(existingOrg, nil)

	err := svc.CreateOrganization(ctx, req, createdBy)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect.Error, got %T", err)
	}

	if connectErr.Code() != connect.CodeAlreadyExists {
		t.Errorf("expected CodeAlreadyExists, got %v", connectErr.Code())
	}
}

func TestCreateOrganization_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	req := &organizationv1.CreateOrganizationRequest{
		Name:       "test-org",
		Visibility: shared.Visibility_VISIBILITY_PRIVATE,
	}
	createdBy := "user-123"

	mockRepo.EXPECT().
		GetOrganizationByName(ctx, "test-org").
		Return(nil, ErrOrganizationNotFound)

	mockRepo.EXPECT().
		CreateOrganization(ctx, gomock.Any()).
		Return(connect.NewError(connect.CodeInternal, errors.New("database error")))

	err := svc.CreateOrganization(ctx, req, createdBy)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect.Error, got %T", err)
	}

	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("expected CodeInternal, got %v", connectErr.Code())
	}
}

func TestCreateOrganization_GetByNameError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	req := &organizationv1.CreateOrganizationRequest{
		Name:       "test-org",
		Visibility: shared.Visibility_VISIBILITY_PRIVATE,
	}
	createdBy := "user-123"

	mockRepo.EXPECT().
		GetOrganizationByName(ctx, "test-org").
		Return(nil, connect.NewError(connect.CodeInternal, errors.New("database error")))

	err := svc.CreateOrganization(ctx, req, createdBy)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateOrganization_InviteCreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	req := &organizationv1.CreateOrganizationRequest{
		Name:         "test-org",
		Visibility:   shared.Visibility_VISIBILITY_PRIVATE,
		InviteEmails: []string{"friend@example.com"},
	}
	createdBy := "user-123"

	mockRepo.EXPECT().
		GetOrganizationByName(ctx, "test-org").
		Return(nil, ErrOrganizationNotFound)

	mockRepo.EXPECT().
		CreateOrganization(ctx, gomock.Any()).
		Return(nil)

	// Invite creation fails - should continue without error (logged)
	mockRepo.EXPECT().
		CreateInvite(ctx, gomock.Any()).
		Return(connect.NewError(connect.CodeInternal, errors.New("invite creation failed")))

	// Organization creation should still succeed
	err := svc.CreateOrganization(ctx, req, createdBy)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateOrganization_EmailSendError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	req := &organizationv1.CreateOrganizationRequest{
		Name:         "test-org",
		Visibility:   shared.Visibility_VISIBILITY_PRIVATE,
		InviteEmails: []string{"friend@example.com"},
	}
	createdBy := "user-123"

	mockRepo.EXPECT().
		GetOrganizationByName(ctx, "test-org").
		Return(nil, ErrOrganizationNotFound)

	mockRepo.EXPECT().
		CreateOrganization(ctx, gomock.Any()).
		Return(nil)

	mockRepo.EXPECT().
		CreateInvite(ctx, gomock.Any()).
		Return(nil)

	// Email send fails - should continue without error (logged)
	mockEmail.EXPECT().
		SendInvite("friend@example.com", "test-org", gomock.Any()).
		Return(errors.New("SMTP error"))

	// Organization creation should still succeed
	err := svc.CreateOrganization(ctx, req, createdBy)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateOrganization_VisibilityMapping(t *testing.T) {
	tests := []struct {
		name               string
		protoVisibility    shared.Visibility
		expectedVisibility Visibility
	}{
		{
			name:               "public visibility",
			protoVisibility:    shared.Visibility_VISIBILITY_PUBLIC,
			expectedVisibility: VisibilityPublic,
		},
		{
			name:               "private visibility",
			protoVisibility:    shared.Visibility_VISIBILITY_PRIVATE,
			expectedVisibility: VisibilityPrivate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := NewMockRepository(ctrl)
			mockEmail := email.NewMockService(ctrl)

			svc := NewService(mockRepo, mockEmail)

			ctx := context.Background()
			req := &organizationv1.CreateOrganizationRequest{
				Name:       "test-org",
				Visibility: tt.protoVisibility,
			}
			createdBy := "user-123"

			mockRepo.EXPECT().
				GetOrganizationByName(ctx, "test-org").
				Return(nil, ErrOrganizationNotFound)

			mockRepo.EXPECT().
				CreateOrganization(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, org *OrganizationDTO) error {
					if org.Visibility != tt.expectedVisibility {
						t.Errorf("expected visibility %s, got %s", tt.expectedVisibility, org.Visibility)
					}
					return nil
				})

			err := svc.CreateOrganization(ctx, req, createdBy)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestGenerateInviteToken(t *testing.T) {
	token1, err := generateInviteToken()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if len(token1) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("expected token length 64, got %d", len(token1))
	}

	token2, err := generateInviteToken()
	if err != nil {
		t.Fatalf("failed to generate second token: %v", err)
	}

	if token1 == token2 {
		t.Error("expected unique tokens, got same token twice")
	}
}

func TestRespondToInvitation_AcceptSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	token := "valid-token-123"
	userId := "user-456"

	invite := &OrganizationInviteDTO{
		Id:             "invite-id",
		OrganizationId: "org-123",
		Email:          "user@example.com",
		Token:          token,
		InvitedBy:      "inviter-789",
		Status:         InviteStatusPending,
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().AddDate(0, 0, 7),
	}

	mockRepo.EXPECT().
		GetInviteByToken(ctx, token).
		Return(invite, nil)

	mockRepo.EXPECT().
		UpdateInviteStatus(ctx, invite.Id, InviteStatusAccepted, gomock.Any()).
		Return(nil)

	mockRepo.EXPECT().
		AddMember(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, member *OrganizationMemberDTO) error {
			if member.OrganizationId != invite.OrganizationId {
				t.Errorf("expected organizationId %s, got %s", invite.OrganizationId, member.OrganizationId)
			}
			if member.UserId != userId {
				t.Errorf("expected userId %s, got %s", userId, member.UserId)
			}
			if member.Role != MemberRoleMember {
				t.Errorf("expected role 'member', got %s", member.Role)
			}
			return nil
		})

	err := svc.RespondToInvitation(ctx, token, userId, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRespondToInvitation_RejectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	token := "valid-token-123"
	userId := "user-456"

	invite := &OrganizationInviteDTO{
		Id:             "invite-id",
		OrganizationId: "org-123",
		Email:          "user@example.com",
		Token:          token,
		InvitedBy:      "inviter-789",
		Status:         InviteStatusPending,
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().AddDate(0, 0, 7),
	}

	mockRepo.EXPECT().
		GetInviteByToken(ctx, token).
		Return(invite, nil)

	mockRepo.EXPECT().
		UpdateInviteStatus(ctx, invite.Id, InviteStatusCancelled, nil).
		Return(nil)

	err := svc.RespondToInvitation(ctx, token, userId, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRespondToInvitation_InviteNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	token := "invalid-token"
	userId := "user-456"

	mockRepo.EXPECT().
		GetInviteByToken(ctx, token).
		Return(nil, ErrInviteNotFound)

	err := svc.RespondToInvitation(ctx, token, userId, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInviteNotFound) {
		t.Errorf("expected ErrInviteNotFound, got %v", err)
	}
}

func TestRespondToInvitation_InviteAlreadyAccepted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	token := "valid-token-123"
	userId := "user-456"

	acceptedAt := time.Now().UTC()
	invite := &OrganizationInviteDTO{
		Id:             "invite-id",
		OrganizationId: "org-123",
		Email:          "user@example.com",
		Token:          token,
		InvitedBy:      "inviter-789",
		Status:         InviteStatusAccepted,
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().AddDate(0, 0, 7),
		AcceptedAt:     &acceptedAt,
	}

	mockRepo.EXPECT().
		GetInviteByToken(ctx, token).
		Return(invite, nil)

	err := svc.RespondToInvitation(ctx, token, userId, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect.Error, got %T", err)
	}

	if connectErr.Code() != connect.CodeFailedPrecondition {
		t.Errorf("expected CodeFailedPrecondition, got %v", connectErr.Code())
	}
}

func TestRespondToInvitation_InviteExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	token := "expired-token"
	userId := "user-456"

	invite := &OrganizationInviteDTO{
		Id:             "invite-id",
		OrganizationId: "org-123",
		Email:          "user@example.com",
		Token:          token,
		InvitedBy:      "inviter-789",
		Status:         InviteStatusPending,
		CreatedAt:      time.Now().UTC().AddDate(0, 0, -14),
		ExpiresAt:      time.Now().UTC().AddDate(0, 0, -7), // Expired 7 days ago
	}

	mockRepo.EXPECT().
		GetInviteByToken(ctx, token).
		Return(invite, nil)

	mockRepo.EXPECT().
		UpdateInviteStatus(ctx, invite.Id, InviteStatusExpired, nil).
		Return(nil)

	err := svc.RespondToInvitation(ctx, token, userId, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect.Error, got %T", err)
	}

	if connectErr.Code() != connect.CodeFailedPrecondition {
		t.Errorf("expected CodeFailedPrecondition, got %v", connectErr.Code())
	}
}

func TestRespondToInvitation_MemberAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	token := "valid-token-123"
	userId := "user-456"

	invite := &OrganizationInviteDTO{
		Id:             "invite-id",
		OrganizationId: "org-123",
		Email:          "user@example.com",
		Token:          token,
		InvitedBy:      "inviter-789",
		Status:         InviteStatusPending,
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().AddDate(0, 0, 7),
	}

	mockRepo.EXPECT().
		GetInviteByToken(ctx, token).
		Return(invite, nil)

	mockRepo.EXPECT().
		UpdateInviteStatus(ctx, invite.Id, InviteStatusAccepted, gomock.Any()).
		Return(nil)

	mockRepo.EXPECT().
		AddMember(ctx, gomock.Any()).
		Return(ErrMemberAlreadyExists)

	// Should succeed even if member already exists
	err := svc.RespondToInvitation(ctx, token, userId, true)
	if err != nil {
		t.Fatalf("expected no error when member already exists, got %v", err)
	}
}

func TestRespondToInvitation_UpdateStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	token := "valid-token-123"
	userId := "user-456"

	invite := &OrganizationInviteDTO{
		Id:             "invite-id",
		OrganizationId: "org-123",
		Email:          "user@example.com",
		Token:          token,
		InvitedBy:      "inviter-789",
		Status:         InviteStatusPending,
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().AddDate(0, 0, 7),
	}

	mockRepo.EXPECT().
		GetInviteByToken(ctx, token).
		Return(invite, nil)

	mockRepo.EXPECT().
		UpdateInviteStatus(ctx, invite.Id, InviteStatusAccepted, gomock.Any()).
		Return(connect.NewError(connect.CodeInternal, errors.New("database error")))

	err := svc.RespondToInvitation(ctx, token, userId, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect.Error, got %T", err)
	}

	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("expected CodeInternal, got %v", connectErr.Code())
	}
}

func TestRespondToInvitation_AddMemberError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	token := "valid-token-123"
	userId := "user-456"

	invite := &OrganizationInviteDTO{
		Id:             "invite-id",
		OrganizationId: "org-123",
		Email:          "user@example.com",
		Token:          token,
		InvitedBy:      "inviter-789",
		Status:         InviteStatusPending,
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().AddDate(0, 0, 7),
	}

	mockRepo.EXPECT().
		GetInviteByToken(ctx, token).
		Return(invite, nil)

	mockRepo.EXPECT().
		UpdateInviteStatus(ctx, invite.Id, InviteStatusAccepted, gomock.Any()).
		Return(nil)

	mockRepo.EXPECT().
		AddMember(ctx, gomock.Any()).
		Return(connect.NewError(connect.CodeInternal, errors.New("database error")))

	err := svc.RespondToInvitation(ctx, token, userId, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect.Error, got %T", err)
	}

	if connectErr.Code() != connect.CodeInternal {
		t.Errorf("expected CodeInternal, got %v", connectErr.Code())
	}
}

func TestRespondToInvitation_InviteCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockEmail := email.NewMockService(ctrl)

	svc := NewService(mockRepo, mockEmail)

	ctx := context.Background()
	token := "cancelled-token"
	userId := "user-456"

	invite := &OrganizationInviteDTO{
		Id:             "invite-id",
		OrganizationId: "org-123",
		Email:          "user@example.com",
		Token:          token,
		InvitedBy:      "inviter-789",
		Status:         InviteStatusCancelled,
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().AddDate(0, 0, 7),
	}

	mockRepo.EXPECT().
		GetInviteByToken(ctx, token).
		Return(invite, nil)

	err := svc.RespondToInvitation(ctx, token, userId, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect.Error, got %T", err)
	}

	if connectErr.Code() != connect.CodeFailedPrecondition {
		t.Errorf("expected CodeFailedPrecondition, got %v", connectErr.Code())
	}
}
