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

	"hasir-api/internal/registry"
	"hasir-api/internal/user"
	"hasir-api/pkg/email"
	"hasir-api/pkg/proto"
)

var (
	ErrOrganizationAlreadyExists = connect.NewError(connect.CodeAlreadyExists, errors.New("organization already exists"))
	ErrOrganizationNotFound      = connect.NewError(connect.CodeNotFound, errors.New("organization not found"))
	ErrInviteNotFound            = connect.NewError(connect.CodeNotFound, errors.New("invite not found"))
	ErrMemberAlreadyExists       = connect.NewError(connect.CodeAlreadyExists, errors.New("member already exists"))
	ErrMemberNotFound            = connect.NewError(connect.CodeNotFound, errors.New("member not found"))
)

func newTestService(t *testing.T) (Service, *MockRepository, *MockQueue, *registry.MockService, *email.MockService, *user.MockRepository, context.Context) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockRepo := NewMockRepository(ctrl)
	mockQueue := NewMockQueue(ctrl)
	mockRegistry := registry.NewMockService(ctrl)
	mockEmail := email.NewMockService(ctrl)
	mockUserRepo := user.NewMockRepository(ctrl)

	svc := NewService(mockRepo, mockQueue, mockRegistry, mockEmail, mockUserRepo)

	return svc, mockRepo, mockQueue, mockRegistry, mockEmail, mockUserRepo, context.Background()
}

func TestCreateOrganization(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
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
				if org.Visibility != proto.VisibilityPrivate {
					t.Errorf("expected visibility 'private', got %s", org.Visibility)
				}
				if org.CreatedBy != createdBy {
					t.Errorf("expected createdBy '%s', got %s", createdBy, org.CreatedBy)
				}
				return nil
			})

		mockRepo.EXPECT().
			AddMember(ctx, gomock.Any()).
			Return(nil)

		err := svc.CreateOrganization(ctx, req, createdBy)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("with invites", func(t *testing.T) {
		svc, mockRepo, mockQueue, _, _, mockUserRepo, ctx := newTestService(t)
		req := &organizationv1.CreateOrganizationRequest{
			Name:       "test-org",
			Visibility: shared.Visibility_VISIBILITY_PUBLIC,
			Members: []*organizationv1.InvitationMember{
				{Email: "friend1@example.com", Role: shared.Role_ROLE_AUTHOR},
				{Email: "friend2@example.com", Role: shared.Role_ROLE_AUTHOR},
			},
		}
		createdBy := "user-123"

		mockRepo.EXPECT().
			GetOrganizationByName(ctx, "test-org").
			Return(nil, ErrOrganizationNotFound)

		mockRepo.EXPECT().
			CreateOrganization(ctx, gomock.Any()).
			Return(nil)

		mockRepo.EXPECT().
			AddMember(ctx, gomock.Any()).
			Return(nil)

		mockUserRepo.EXPECT().
			GetUsersByEmails(ctx, []string{"friend1@example.com", "friend2@example.com"}).
			Return(map[string]*user.UserDTO{
				"friend1@example.com": {},
				"friend2@example.com": {},
			}, nil)

		mockRepo.EXPECT().
			CreateInvites(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, invites []*OrganizationInviteDTO) error {
				if len(invites) != 2 {
					t.Errorf("expected 2 invites, got %d", len(invites))
				}
				for _, invite := range invites {
					if invite.Email != "friend1@example.com" && invite.Email != "friend2@example.com" {
						t.Errorf("unexpected email: %s", invite.Email)
					}
					if invite.Status != InviteStatusPending {
						t.Errorf("expected status 'pending', got %s", invite.Status)
					}
				}
				return nil
			})

		mockQueue.EXPECT().
			EnqueueEmailJobs(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, jobs []*EmailJobDTO) error {
				if len(jobs) != 2 {
					t.Errorf("expected 2 email jobs, got %d", len(jobs))
				}
				for _, job := range jobs {
					if job.Email != "friend1@example.com" && job.Email != "friend2@example.com" {
						t.Errorf("unexpected email in job: %s", job.Email)
					}
					if job.OrganizationName != "test-org" {
						t.Errorf("expected organization name 'test-org', got %s", job.OrganizationName)
					}
					if job.Status != EmailJobStatusPending {
						t.Errorf("expected job status 'pending', got %s", job.Status)
					}
				}
				return nil
			})

		err := svc.CreateOrganization(ctx, req, createdBy)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("already exists", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
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
	})

	t.Run("repository error", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
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
	})

	t.Run("get by name error", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
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
	})

	t.Run("invite create error", func(t *testing.T) {
		svc, mockRepo, _, _, _, mockUserRepo, ctx := newTestService(t)
		req := &organizationv1.CreateOrganizationRequest{
			Name:       "test-org",
			Visibility: shared.Visibility_VISIBILITY_PRIVATE,
			Members: []*organizationv1.InvitationMember{
				{Email: "friend@example.com", Role: shared.Role_ROLE_AUTHOR},
			},
		}
		createdBy := "user-123"

		mockRepo.EXPECT().
			GetOrganizationByName(ctx, "test-org").
			Return(nil, ErrOrganizationNotFound)

		mockRepo.EXPECT().
			CreateOrganization(ctx, gomock.Any()).
			Return(nil)

		mockRepo.EXPECT().
			AddMember(ctx, gomock.Any()).
			Return(nil)

		mockUserRepo.EXPECT().
			GetUsersByEmails(ctx, []string{"friend@example.com"}).
			Return(map[string]*user.UserDTO{
				"friend@example.com": {},
			}, nil)

		mockRepo.EXPECT().
			CreateInvites(ctx, gomock.Any()).
			Return(connect.NewError(connect.CodeInternal, errors.New("invite creation failed")))

		err := svc.CreateOrganization(ctx, req, createdBy)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("email send error", func(t *testing.T) {
		svc, mockRepo, mockQueue, _, _, mockUserRepo, ctx := newTestService(t)
		req := &organizationv1.CreateOrganizationRequest{
			Name:       "test-org",
			Visibility: shared.Visibility_VISIBILITY_PRIVATE,
			Members: []*organizationv1.InvitationMember{
				{Email: "friend@example.com", Role: shared.Role_ROLE_AUTHOR},
			},
		}
		createdBy := "user-123"

		mockRepo.EXPECT().
			GetOrganizationByName(ctx, "test-org").
			Return(nil, ErrOrganizationNotFound)

		mockRepo.EXPECT().
			CreateOrganization(ctx, gomock.Any()).
			Return(nil)

		mockRepo.EXPECT().
			AddMember(ctx, gomock.Any()).
			Return(nil)

		mockUserRepo.EXPECT().
			GetUsersByEmails(ctx, []string{"friend@example.com"}).
			Return(map[string]*user.UserDTO{
				"friend@example.com": {},
			}, nil)

		mockRepo.EXPECT().
			CreateInvites(ctx, gomock.Any()).
			Return(nil)

		mockQueue.EXPECT().
			EnqueueEmailJobs(ctx, gomock.Any()).
			Return(errors.New("queue error"))

		err := svc.CreateOrganization(ctx, req, createdBy)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("visibility mapping", func(t *testing.T) {
		tests := []struct {
			name               string
			protoVisibility    shared.Visibility
			expectedVisibility proto.Visibility
		}{
			{
				name:               "public visibility",
				protoVisibility:    shared.Visibility_VISIBILITY_PUBLIC,
				expectedVisibility: proto.VisibilityPublic,
			},
			{
				name:               "private visibility",
				protoVisibility:    shared.Visibility_VISIBILITY_PRIVATE,
				expectedVisibility: proto.VisibilityPrivate,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				svc, mockRepo, _, _, _, _, ctx := newTestService(t)
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

				mockRepo.EXPECT().
					AddMember(ctx, gomock.Any()).
					Return(nil)

				err := svc.CreateOrganization(ctx, req, createdBy)
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			})
		}
	})
}

func TestInviteUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, mockRepo, mockQueue, _, _, mockUserRepo, ctx := newTestService(t)
		req := &organizationv1.InviteMemberRequest{
			Id:    "org-123",
			Email: "friend1@example.com",
			Role:  shared.Role_ROLE_READER,
		}
		invitedBy := "user-123"

		org := &OrganizationDTO{
			Id:        "org-123",
			Name:      "test-org",
			CreatedBy: invitedBy,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, "org-123").
			Return(org, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, "org-123", invitedBy).
			Return(MemberRoleOwner, nil)

		targetUser := &user.UserDTO{Id: "target-user-id"}
		mockUserRepo.EXPECT().
			GetUserByEmail(ctx, "friend1@example.com").
			Return(targetUser, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, "org-123", targetUser.Id).
			Return(MemberRole(""), ErrMemberNotFound)

		mockRepo.EXPECT().
			CreateInvites(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, invites []*OrganizationInviteDTO) error {
				if len(invites) != 1 {
					t.Errorf("expected 1 invite, got %d", len(invites))
				}
				if invites[0].Email != "friend1@example.com" {
					t.Errorf("expected email 'friend1@example.com', got %s", invites[0].Email)
				}
				if invites[0].Role != MemberRoleReader {
					t.Errorf("expected role 'reader', got %s", invites[0].Role)
				}
				return nil
			})

		mockQueue.EXPECT().
			EnqueueEmailJobs(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, jobs []*EmailJobDTO) error {
				if len(jobs) != 1 {
					t.Errorf("expected 1 email job, got %d", len(jobs))
				}
				return nil
			})

		err := svc.InviteUser(ctx, req, invitedBy)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("permission denied when not creator", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		req := &organizationv1.InviteMemberRequest{
			Id:    "org-123",
			Email: "friend@example.com",
		}
		invitedBy := "user-123"

		org := &OrganizationDTO{
			Id:        "org-123",
			Name:      "test-org",
			CreatedBy: "other-user",
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, "org-123").
			Return(org, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, "org-123", invitedBy).
			Return(MemberRoleAuthor, nil)

		err := svc.InviteUser(ctx, req, invitedBy)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected connect.Error, got %T", err)
		}

		if connectErr.Code() != connect.CodePermissionDenied {
			t.Errorf("expected CodePermissionDenied, got %v", connectErr.Code())
		}
	})

	t.Run("user not found by email", func(t *testing.T) {
		svc, mockRepo, _, _, _, mockUserRepo, ctx := newTestService(t)
		req := &organizationv1.InviteMemberRequest{
			Id:    "org-123",
			Email: "unknown@example.com",
		}
		invitedBy := "user-123"

		org := &OrganizationDTO{
			Id:        "org-123",
			Name:      "test-org",
			CreatedBy: invitedBy,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, "org-123").
			Return(org, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, "org-123", invitedBy).
			Return(MemberRoleOwner, nil)

		mockUserRepo.EXPECT().
			GetUserByEmail(ctx, "unknown@example.com").
			Return(nil, connect.NewError(connect.CodeNotFound, errors.New("user not found")))

		err := svc.InviteUser(ctx, req, invitedBy)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected connect.Error, got %T", err)
		}

		if connectErr.Code() != connect.CodeNotFound {
			t.Errorf("expected CodeNotFound, got %v", connectErr.Code())
		}
	})

	t.Run("user already a member", func(t *testing.T) {
		svc, mockRepo, _, _, _, mockUserRepo, ctx := newTestService(t)
		req := &organizationv1.InviteMemberRequest{
			Id:    "org-123",
			Email: "member@example.com",
		}
		invitedBy := "user-123"

		org := &OrganizationDTO{
			Id:        "org-123",
			Name:      "test-org",
			CreatedBy: invitedBy,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, "org-123").
			Return(org, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, "org-123", invitedBy).
			Return(MemberRoleOwner, nil)

		targetUser := &user.UserDTO{Id: "member-user-id"}
		mockUserRepo.EXPECT().
			GetUserByEmail(ctx, "member@example.com").
			Return(targetUser, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, "org-123", targetUser.Id).
			Return(MemberRoleAuthor, nil)

		err := svc.InviteUser(ctx, req, invitedBy)
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
	})
}

func TestGenerateInviteToken(t *testing.T) {
	token1, err := generateInviteToken()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if len(token1) != 64 {
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

func TestRespondToInvitation(t *testing.T) {
	t.Run("accept success", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		token := "valid-token-123"
		userId := "user-456"
		userEmail := "user@example.com"

		invite := &OrganizationInviteDTO{
			Id:             "invite-id",
			OrganizationId: "org-123",
			Email:          userEmail,
			Token:          token,
			InvitedBy:      "inviter-789",
			Role:           MemberRoleAuthor,
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
				if member.Role != invite.Role {
					t.Errorf("expected role %s, got %s", invite.Role, member.Role)
				}
				return nil
			})

		err := svc.RespondToInvitation(ctx, token, userId, userEmail, true)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("reject when email does not match", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		token := "valid-token-123"
		userId := "user-456"
		userEmail := "different@example.com"

		invite := &OrganizationInviteDTO{
			Id:             "invite-id",
			OrganizationId: "org-123",
			Email:          "user@example.com",
			Token:          token,
			InvitedBy:      "inviter-789",
			Role:           MemberRoleAuthor,
			Status:         InviteStatusPending,
			CreatedAt:      time.Now().UTC(),
			ExpiresAt:      time.Now().UTC().AddDate(0, 0, 7),
		}

		mockRepo.EXPECT().
			GetInviteByToken(ctx, token).
			Return(invite, nil)

		err := svc.RespondToInvitation(ctx, token, userId, userEmail, true)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected connect.Error, got %T", err)
		}

		if connectErr.Code() != connect.CodePermissionDenied {
			t.Errorf("expected CodePermissionDenied, got %v", connectErr.Code())
		}

		if !errors.Is(err, connect.NewError(connect.CodePermissionDenied, errors.New("this invitation is not for your email address"))) {
			expectedMsg := "this invitation is not for your email address"
			if connectErr.Message() != expectedMsg {
				t.Errorf("expected message '%s', got '%s'", expectedMsg, connectErr.Message())
			}
		}
	})

	t.Run("reject success", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		token := "valid-token-123"
		userId := "user-456"
		userEmail := "user@example.com"

		invite := &OrganizationInviteDTO{
			Id:             "invite-id",
			OrganizationId: "org-123",
			Email:          userEmail,
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

		err := svc.RespondToInvitation(ctx, token, userId, userEmail, false)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("invite not found", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		token := "invalid-token"
		userId := "user-456"
		userEmail := "user@example.com"

		mockRepo.EXPECT().
			GetInviteByToken(ctx, token).
			Return(nil, ErrInviteNotFound)

		err := svc.RespondToInvitation(ctx, token, userId, userEmail, true)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrInviteNotFound) {
			t.Errorf("expected ErrInviteNotFound, got %v", err)
		}
	})

	t.Run("invite already accepted", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		token := "valid-token-123"
		userId := "user-456"
		userEmail := "user@example.com"

		acceptedAt := time.Now().UTC()
		invite := &OrganizationInviteDTO{
			Id:             "invite-id",
			OrganizationId: "org-123",
			Email:          userEmail,
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

		err := svc.RespondToInvitation(ctx, token, userId, userEmail, true)
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
	})

	t.Run("invite expired", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		token := "expired-token"
		userId := "user-456"
		userEmail := "user@example.com"

		invite := &OrganizationInviteDTO{
			Id:             "invite-id",
			OrganizationId: "org-123",
			Email:          userEmail,
			Token:          token,
			InvitedBy:      "inviter-789",
			Status:         InviteStatusPending,
			CreatedAt:      time.Now().UTC().AddDate(0, 0, -14),
			ExpiresAt:      time.Now().UTC().AddDate(0, 0, -7),
		}

		mockRepo.EXPECT().
			GetInviteByToken(ctx, token).
			Return(invite, nil)

		mockRepo.EXPECT().
			UpdateInviteStatus(ctx, invite.Id, InviteStatusExpired, nil).
			Return(nil)

		err := svc.RespondToInvitation(ctx, token, userId, userEmail, true)
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
	})

	t.Run("member already exists", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		token := "valid-token-123"
		userId := "user-456"
		userEmail := "user@example.com"

		invite := &OrganizationInviteDTO{
			Id:             "invite-id",
			OrganizationId: "org-123",
			Email:          userEmail,
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

		err := svc.RespondToInvitation(ctx, token, userId, userEmail, true)
		if err != nil {
			t.Fatalf("expected no error when member already exists, got %v", err)
		}
	})

	t.Run("update status error", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		token := "valid-token-123"
		userId := "user-456"
		userEmail := "user@example.com"

		invite := &OrganizationInviteDTO{
			Id:             "invite-id",
			OrganizationId: "org-123",
			Email:          userEmail,
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

		err := svc.RespondToInvitation(ctx, token, userId, userEmail, true)
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
	})

	t.Run("add member error", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		token := "valid-token-123"
		userId := "user-456"
		userEmail := "user@example.com"

		invite := &OrganizationInviteDTO{
			Id:             "invite-id",
			OrganizationId: "org-123",
			Email:          userEmail,
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

		err := svc.RespondToInvitation(ctx, token, userId, userEmail, true)
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
	})

	t.Run("invite cancelled", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		token := "cancelled-token"
		userId := "user-456"
		userEmail := "user@example.com"

		invite := &OrganizationInviteDTO{
			Id:             "invite-id",
			OrganizationId: "org-123",
			Email:          userEmail,
			Token:          token,
			InvitedBy:      "inviter-789",
			Status:         InviteStatusCancelled,
			CreatedAt:      time.Now().UTC(),
			ExpiresAt:      time.Now().UTC().AddDate(0, 0, 7),
		}

		mockRepo.EXPECT().
			GetInviteByToken(ctx, token).
			Return(invite, nil)

		err := svc.RespondToInvitation(ctx, token, userId, userEmail, true)
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
	})
}

func TestDeleteOrganization(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, mockRepo, _, mockRegistry, _, _, ctx := newTestService(t)
		orgID := "org-123"
		userID := "user-123"

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(MemberRoleOwner, nil)

		mockRegistry.EXPECT().
			DeleteRepositoriesByOrganization(ctx, orgID).
			Return(nil)

		mockRepo.EXPECT().
			DeleteOrganization(ctx, orgID).
			Return(nil)

		err := svc.DeleteOrganization(ctx, orgID, userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("organization not found", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "non-existent-org"
		userID := "user-123"

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(MemberRole(""), ErrMemberNotFound)

		err := svc.DeleteOrganization(ctx, orgID, userID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected connect.Error, got %T", err)
		}

		if connectErr.Code() != connect.CodePermissionDenied {
			t.Errorf("expected CodeNotFound, got %v", connectErr.Code())
		}
	})

	t.Run("permission denied", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		userID := "user-123"

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(MemberRoleReader, nil)

		err := svc.DeleteOrganization(ctx, orgID, userID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected connect.Error, got %T", err)
		}

		if connectErr.Code() != connect.CodePermissionDenied {
			t.Errorf("expected CodePermissionDenied, got %v", connectErr.Code())
		}
	})

	t.Run("repository error", func(t *testing.T) {
		svc, mockRepo, _, mockRegistry, _, _, ctx := newTestService(t)
		orgID := "org-123"
		userID := "user-123"

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(MemberRoleOwner, nil)

		mockRegistry.EXPECT().
			DeleteRepositoriesByOrganization(ctx, orgID).
			Return(nil)

		mockRepo.EXPECT().
			DeleteOrganization(ctx, orgID).
			Return(connect.NewError(connect.CodeInternal, errors.New("database error")))

		err := svc.DeleteOrganization(ctx, orgID, userID)
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
	})
}

func TestUpdateOrganization(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		userID := "user-123"

		existingOrg := &OrganizationDTO{
			Id:         orgID,
			Name:       "old-name",
			Visibility: proto.VisibilityPrivate,
			CreatedBy:  userID,
		}

		req := &organizationv1.UpdateOrganizationRequest{
			Id:         orgID,
			Name:       "new-name",
			Visibility: shared.Visibility_VISIBILITY_PUBLIC,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			UpdateOrganization(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, org *OrganizationDTO) error {
				if org.Id != orgID {
					t.Errorf("expected id %s, got %s", orgID, org.Id)
				}
				if org.Name != "new-name" {
					t.Errorf("expected name 'new-name', got %s", org.Name)
				}
				if org.Visibility != proto.VisibilityPublic {
					t.Errorf("expected visibility 'public', got %s", org.Visibility)
				}
				return nil
			})

		err := svc.UpdateOrganization(ctx, req, userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("permission denied when not creator", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		creatorID := "creator-123"
		otherUserID := "user-456"

		existingOrg := &OrganizationDTO{
			Id:         orgID,
			Name:       "org-name",
			Visibility: proto.VisibilityPrivate,
			CreatedBy:  creatorID,
		}

		req := &organizationv1.UpdateOrganizationRequest{
			Id:   orgID,
			Name: "updated-name",
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, otherUserID).
			Return(MemberRoleReader, nil)

		err := svc.UpdateOrganization(ctx, req, otherUserID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected connect.Error, got %T", err)
		}

		if connectErr.Code() != connect.CodePermissionDenied {
			t.Errorf("expected CodePermissionDenied, got %v", connectErr.Code())
		}
	})

	t.Run("name already exists", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		userID := "user-123"

		existingOrg := &OrganizationDTO{
			Id:         orgID,
			Name:       "old-name",
			Visibility: proto.VisibilityPrivate,
			CreatedBy:  userID,
		}

		req := &organizationv1.UpdateOrganizationRequest{
			Id:   orgID,
			Name: "conflict-name",
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, userID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			UpdateOrganization(ctx, gomock.Any()).
			Return(ErrOrganizationAlreadyExists)

		err := svc.UpdateOrganization(ctx, req, userID)
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
	})
}

func TestUpdateMemberRole(t *testing.T) {
	t.Run("success - owner updating another member's role", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		ownerID := "owner-123"
		memberID := "member-456"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.UpdateMemberRoleRequest{
			OrganizationId: orgID,
			MemberId:       memberID,
			Role:           shared.Role_ROLE_AUTHOR,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, ownerID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, memberID).
			Return(MemberRoleReader, nil)
		mockRepo.EXPECT().
			UpdateMemberRole(ctx, orgID, memberID, MemberRoleAuthor).
			Return(nil)

		err := svc.UpdateMemberRole(ctx, req, ownerID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("permission denied - non-owner trying to update", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		authorID := "author-123"
		memberID := "member-456"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.UpdateMemberRoleRequest{
			OrganizationId: orgID,
			MemberId:       memberID,
			Role:           shared.Role_ROLE_READER,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, authorID).
			Return(MemberRoleAuthor, nil)

		err := svc.UpdateMemberRole(ctx, req, authorID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected connect.Error, got %T", err)
		}

		if connectErr.Code() != connect.CodePermissionDenied {
			t.Errorf("expected CodePermissionDenied, got %v", connectErr.Code())
		}
	})

	t.Run("success - owner can decrease their own role when multiple owners exist", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		ownerID := "owner-123"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.UpdateMemberRoleRequest{
			OrganizationId: orgID,
			MemberId:       ownerID,
			Role:           shared.Role_ROLE_AUTHOR,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, ownerID).
			Return(MemberRoleOwner, nil).
			Times(2)

		mockRepo.EXPECT().
			GetOwnerCount(ctx, orgID).
			Return(2, nil)

		mockRepo.EXPECT().
			UpdateMemberRole(ctx, orgID, ownerID, MemberRoleAuthor).
			Return(nil)

		err := svc.UpdateMemberRole(ctx, req, ownerID)
		if err != nil {
			t.Fatalf("expected no error when multiple owners exist, got %v", err)
		}
	})

	t.Run("failed precondition - last owner cannot decrease their own role", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		ownerID := "owner-123"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.UpdateMemberRoleRequest{
			OrganizationId: orgID,
			MemberId:       ownerID,
			Role:           shared.Role_ROLE_AUTHOR,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, ownerID).
			Return(MemberRoleOwner, nil).
			Times(2)

		mockRepo.EXPECT().
			GetOwnerCount(ctx, orgID).
			Return(1, nil)

		err := svc.UpdateMemberRole(ctx, req, ownerID)
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

		if connectErr.Message() != errCannotChangeLastOwner {
			t.Errorf("expected error message '%s', got '%s'", errCannotChangeLastOwner, connectErr.Message())
		}
	})

	t.Run("success - multiple owners can change another owner's role", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		ownerID := "owner-123"
		memberID := "member-789"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.UpdateMemberRoleRequest{
			OrganizationId: orgID,
			MemberId:       memberID,
			Role:           shared.Role_ROLE_READER,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, ownerID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, memberID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			GetOwnerCount(ctx, orgID).
			Return(3, nil)

		mockRepo.EXPECT().
			UpdateMemberRole(ctx, orgID, memberID, MemberRoleReader).
			Return(nil)

		err := svc.UpdateMemberRole(ctx, req, ownerID)
		if err != nil {
			t.Fatalf("expected no error when multiple owners exist, got %v", err)
		}
	})

	t.Run("failed precondition - trying to change only owner's role", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		ownerID := "owner-123"
		onlyOwnerID := "only-owner-456"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.UpdateMemberRoleRequest{
			OrganizationId: orgID,
			MemberId:       onlyOwnerID,
			Role:           shared.Role_ROLE_READER,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, ownerID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, onlyOwnerID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			GetOwnerCount(ctx, orgID).
			Return(1, nil)

		err := svc.UpdateMemberRole(ctx, req, ownerID)
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
	})

	t.Run("member not found", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		ownerID := "owner-123"
		nonExistentMemberID := "non-existent-456"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.UpdateMemberRoleRequest{
			OrganizationId: orgID,
			MemberId:       nonExistentMemberID,
			Role:           shared.Role_ROLE_AUTHOR,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, ownerID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, nonExistentMemberID).
			Return(MemberRole(""), ErrMemberNotFound)

		err := svc.UpdateMemberRole(ctx, req, ownerID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected connect.Error, got %T", err)
		}

		if connectErr.Code() != connect.CodeNotFound {
			t.Errorf("expected CodeNotFound, got %v", connectErr.Code())
		}
	})

	t.Run("organization not found", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "non-existent-org"
		ownerID := "owner-123"
		memberID := "member-456"

		req := &organizationv1.UpdateMemberRoleRequest{
			OrganizationId: orgID,
			MemberId:       memberID,
			Role:           shared.Role_ROLE_AUTHOR,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(nil, ErrOrganizationNotFound)

		err := svc.UpdateMemberRole(ctx, req, ownerID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected connect.Error, got %T", err)
		}

		if connectErr.Code() != connect.CodeNotFound {
			t.Errorf("expected CodeNotFound, got %v", connectErr.Code())
		}
	})

	t.Run("updater not a member", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		nonMemberID := "non-member-123"
		memberID := "member-456"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.UpdateMemberRoleRequest{
			OrganizationId: orgID,
			MemberId:       memberID,
			Role:           shared.Role_ROLE_AUTHOR,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, nonMemberID).
			Return(MemberRole(""), ErrMemberNotFound)

		err := svc.UpdateMemberRole(ctx, req, nonMemberID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected connect.Error, got %T", err)
		}

		if connectErr.Code() != connect.CodePermissionDenied {
			t.Errorf("expected CodePermissionDenied, got %v", connectErr.Code())
		}
	})

	t.Run("success - owner can promote another member to owner", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		ownerID := "owner-123"
		memberID := "member-456"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.UpdateMemberRoleRequest{
			OrganizationId: orgID,
			MemberId:       memberID,
			Role:           shared.Role_ROLE_OWNER,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, ownerID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, memberID).
			Return(MemberRoleAuthor, nil)

		mockRepo.EXPECT().
			UpdateMemberRole(ctx, orgID, memberID, MemberRoleOwner).
			Return(nil)

		err := svc.UpdateMemberRole(ctx, req, ownerID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestDeleteMember(t *testing.T) {
	t.Run("success - owner deleting non-owner member", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		ownerID := "owner-123"
		memberID := "member-456"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.DeleteMemberRequest{
			OrganizationId: orgID,
			MemberId:       memberID,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, ownerID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, memberID).
			Return(MemberRoleAuthor, nil)

		mockRepo.EXPECT().
			DeleteMember(ctx, orgID, memberID).
			Return(nil)

		err := svc.DeleteMember(ctx, req, ownerID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("success - owner deleting another owner when multiple owners exist", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		ownerID := "owner-123"
		otherOwnerID := "owner-456"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.DeleteMemberRequest{
			OrganizationId: orgID,
			MemberId:       otherOwnerID,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, ownerID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, otherOwnerID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			GetOwnerCount(ctx, orgID).
			Return(2, nil)

		mockRepo.EXPECT().
			DeleteMember(ctx, orgID, otherOwnerID).
			Return(nil)

		err := svc.DeleteMember(ctx, req, ownerID)
		if err != nil {
			t.Fatalf("expected no error when multiple owners exist, got %v", err)
		}
	})

	t.Run("permission denied - non-owner trying to delete", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		authorID := "author-123"
		memberID := "member-456"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.DeleteMemberRequest{
			OrganizationId: orgID,
			MemberId:       memberID,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, authorID).
			Return(MemberRoleAuthor, nil)

		err := svc.DeleteMember(ctx, req, authorID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected connect.Error, got %T", err)
		}

		if connectErr.Code() != connect.CodePermissionDenied {
			t.Errorf("expected CodePermissionDenied, got %v", connectErr.Code())
		}
	})

	t.Run("permission denied - user not a member", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		nonMemberID := "non-member-123"
		memberID := "member-456"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.DeleteMemberRequest{
			OrganizationId: orgID,
			MemberId:       memberID,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, nonMemberID).
			Return(MemberRole(""), ErrMemberNotFound)

		err := svc.DeleteMember(ctx, req, nonMemberID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected connect.Error, got %T", err)
		}

		if connectErr.Code() != connect.CodePermissionDenied {
			t.Errorf("expected CodePermissionDenied, got %v", connectErr.Code())
		}
	})

	t.Run("failed precondition - trying to delete the last owner", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		ownerID := "owner-123"
		lastOwnerID := "last-owner-456"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.DeleteMemberRequest{
			OrganizationId: orgID,
			MemberId:       lastOwnerID,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, ownerID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, lastOwnerID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			GetOwnerCount(ctx, orgID).
			Return(1, nil)

		err := svc.DeleteMember(ctx, req, ownerID)
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
	})

	t.Run("not found - member does not exist", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "org-123"
		ownerID := "owner-123"
		nonExistentMemberID := "non-existent-456"

		existingOrg := &OrganizationDTO{
			Id:   orgID,
			Name: "test-org",
		}

		req := &organizationv1.DeleteMemberRequest{
			OrganizationId: orgID,
			MemberId:       nonExistentMemberID,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(existingOrg, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, ownerID).
			Return(MemberRoleOwner, nil)

		mockRepo.EXPECT().
			GetMemberRole(ctx, orgID, nonExistentMemberID).
			Return(MemberRole(""), ErrMemberNotFound)

		err := svc.DeleteMember(ctx, req, ownerID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected connect.Error, got %T", err)
		}

		if connectErr.Code() != connect.CodeNotFound {
			t.Errorf("expected CodeNotFound, got %v", connectErr.Code())
		}
	})

	t.Run("organization not found", func(t *testing.T) {
		svc, mockRepo, _, _, _, _, ctx := newTestService(t)
		orgID := "non-existent-org"
		ownerID := "owner-123"
		memberID := "member-456"

		req := &organizationv1.DeleteMemberRequest{
			OrganizationId: orgID,
			MemberId:       memberID,
		}

		mockRepo.EXPECT().
			GetOrganizationById(ctx, orgID).
			Return(nil, ErrOrganizationNotFound)

		err := svc.DeleteMember(ctx, req, ownerID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var connectErr *connect.Error
		if !errors.As(err, &connectErr) {
			t.Fatalf("expected connect.Error, got %T", err)
		}

		if connectErr.Code() != connect.CodeNotFound {
			t.Errorf("expected CodeNotFound, got %v", connectErr.Code())
		}
	})
}
