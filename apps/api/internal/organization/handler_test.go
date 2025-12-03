package organization

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/organization/v1/organizationv1connect"
	organizationv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/organization/v1"
	"buf.build/gen/go/hasir/hasir/protocolbuffers/go/shared"

	"hasir-api/pkg/authentication"
	"hasir-api/pkg/proto"
)

func testAuthInterceptor(userID string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			ctx = context.WithValue(ctx, authentication.UserIDKey, userID)
			return next(ctx, req)
		}
	}
}

func TestNewHandler(t *testing.T) {
	t.Run("creates handler with service and repository", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		h := NewHandler(mockService, mockRepository)

		require.NotNil(t, h)
		require.Equal(t, mockService, h.service)
		require.Equal(t, mockRepository, h.repository)
		require.Empty(t, h.interceptors)
	})

	t.Run("creates handler with interceptors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		interceptor1 := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
			return next
		})
		interceptor2 := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
			return next
		})

		h := NewHandler(mockService, mockRepository, interceptor1, interceptor2)

		require.NotNil(t, h)
		require.Len(t, h.interceptors, 2)
	})
}

func TestHandler_RegisterRoutes(t *testing.T) {
	t.Run("returns path and handler", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		h := NewHandler(mockService, mockRepository)
		path, httpHandler := h.RegisterRoutes()

		require.Equal(t, "/"+organizationv1connect.OrganizationServiceName+"/", path)
		require.NotNil(t, httpHandler)
	})
}

func TestHandler_CreateOrganization(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-123"

		mockService.EXPECT().
			CreateOrganization(gomock.Any(), gomock.Any(), testUserID).
			DoAndReturn(func(_ context.Context, req *organizationv1.CreateOrganizationRequest, createdBy string) error {
				require.Equal(t, "test-org", req.GetName())
				require.Equal(t, shared.Visibility_VISIBILITY_PRIVATE, req.GetVisibility())
				require.Equal(t, testUserID, createdBy)
				return nil
			})

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.CreateOrganization(context.Background(), connect.NewRequest(&organizationv1.CreateOrganizationRequest{
			Name:       "test-org",
			Visibility: shared.Visibility_VISIBILITY_PRIVATE,
		}))
		require.NoError(t, err)
	})

	t.Run("success with public visibility", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-456"

		mockService.EXPECT().
			CreateOrganization(gomock.Any(), gomock.Any(), testUserID).
			DoAndReturn(func(_ context.Context, req *organizationv1.CreateOrganizationRequest, createdBy string) error {
				require.Equal(t, "public-org", req.GetName())
				require.Equal(t, shared.Visibility_VISIBILITY_PUBLIC, req.GetVisibility())
				return nil
			})

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.CreateOrganization(context.Background(), connect.NewRequest(&organizationv1.CreateOrganizationRequest{
			Name:       "public-org",
			Visibility: shared.Visibility_VISIBILITY_PUBLIC,
		}))
		require.NoError(t, err)
	})

	t.Run("service error - already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-789"

		mockService.EXPECT().
			CreateOrganization(gomock.Any(), gomock.Any(), testUserID).
			Return(connect.NewError(connect.CodeAlreadyExists, errors.New("organization already exists")))

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.CreateOrganization(context.Background(), connect.NewRequest(&organizationv1.CreateOrganizationRequest{
			Name:       "existing-org",
			Visibility: shared.Visibility_VISIBILITY_PRIVATE,
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeAlreadyExists, connectErr.Code())
	})

	t.Run("unauthenticated - missing user ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		h := NewHandler(mockService, mockRepository)
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.CreateOrganization(context.Background(), connect.NewRequest(&organizationv1.CreateOrganizationRequest{
			Name:       "test-org",
			Visibility: shared.Visibility_VISIBILITY_PRIVATE,
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})
}

func TestHandler_GetOrganizations(t *testing.T) {
	t.Run("success with organizations", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-123"

		orgs := &[]OrganizationDTO{
			{Id: "org-1", Name: "first-org", Visibility: proto.VisibilityPrivate},
			{Id: "org-2", Name: "second-org", Visibility: proto.VisibilityPublic},
		}

		mockRepository.EXPECT().
			GetUserOrganizationsCount(gomock.Any(), testUserID).
			Return(2, nil)
		mockRepository.EXPECT().
			GetUserOrganizations(gomock.Any(), testUserID, 1, 10).
			Return(orgs, nil)

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		resp, err := client.GetOrganizations(context.Background(), connect.NewRequest(&organizationv1.GetOrganizationsRequest{}))
		require.NoError(t, err)
		require.Len(t, resp.Msg.GetOrganizations(), 2)
		require.Equal(t, "org-1", resp.Msg.GetOrganizations()[0].GetId())
		require.Equal(t, "first-org", resp.Msg.GetOrganizations()[0].GetName())
		require.Equal(t, "org-2", resp.Msg.GetOrganizations()[1].GetId())
		require.Equal(t, "second-org", resp.Msg.GetOrganizations()[1].GetName())
	})

	t.Run("success with empty organizations", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-456"

		orgs := &[]OrganizationDTO{}

		mockRepository.EXPECT().
			GetUserOrganizationsCount(gomock.Any(), testUserID).
			Return(0, nil)
		mockRepository.EXPECT().
			GetUserOrganizations(gomock.Any(), testUserID, 1, 10).
			Return(orgs, nil)

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		resp, err := client.GetOrganizations(context.Background(), connect.NewRequest(&organizationv1.GetOrganizationsRequest{}))
		require.NoError(t, err)
		require.Empty(t, resp.Msg.GetOrganizations())
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)
		testUserID := "test-user-789"

		mockRepository.EXPECT().
			GetUserOrganizationsCount(gomock.Any(), testUserID).
			Return(0, connect.NewError(connect.CodeInternal, errors.New("database error")))

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.GetOrganizations(context.Background(), connect.NewRequest(&organizationv1.GetOrganizationsRequest{}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeInternal, connectErr.Code())
	})

	t.Run("unauthenticated - missing user ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		h := NewHandler(mockService, mockRepository)
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.GetOrganizations(context.Background(), connect.NewRequest(&organizationv1.GetOrganizationsRequest{}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})
}

func TestHandler_DeleteOrganization(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-123"
		orgID := "org-123"

		mockService.EXPECT().
			DeleteOrganization(gomock.Any(), orgID, testUserID).
			Return(nil)

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.DeleteOrganization(context.Background(), connect.NewRequest(&organizationv1.DeleteOrganizationRequest{
			Id: orgID,
		}))
		require.NoError(t, err)
	})

	t.Run("service error - organization not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-456"
		orgID := "non-existent-org"

		mockService.EXPECT().
			DeleteOrganization(gomock.Any(), orgID, testUserID).
			Return(ErrOrganizationNotFound)

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.DeleteOrganization(context.Background(), connect.NewRequest(&organizationv1.DeleteOrganizationRequest{
			Id: orgID,
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeNotFound, connectErr.Code())
	})

	t.Run("service error - permission denied", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-789"
		orgID := "org-456"

		mockService.EXPECT().
			DeleteOrganization(gomock.Any(), orgID, testUserID).
			Return(connect.NewError(connect.CodePermissionDenied, errors.New("only the organization creator can delete it")))

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.DeleteOrganization(context.Background(), connect.NewRequest(&organizationv1.DeleteOrganizationRequest{
			Id: orgID,
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodePermissionDenied, connectErr.Code())
	})

	t.Run("unauthenticated - missing user ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		h := NewHandler(mockService, mockRepository)
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.DeleteOrganization(context.Background(), connect.NewRequest(&organizationv1.DeleteOrganizationRequest{
			Id: "org-123",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})
}

func TestHandler_UpdateOrganization(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-123"
		orgID := "org-123"

		mockService.EXPECT().
			UpdateOrganization(gomock.Any(), gomock.Any(), testUserID).
			DoAndReturn(func(_ context.Context, req *organizationv1.UpdateOrganizationRequest, userId string) error {
				require.Equal(t, orgID, req.GetId())
				require.Equal(t, "updated-name", req.GetName())
				require.Equal(t, shared.Visibility_VISIBILITY_PUBLIC, req.GetVisibility())
				require.Equal(t, testUserID, userId)
				return nil
			})

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.UpdateOrganization(context.Background(), connect.NewRequest(&organizationv1.UpdateOrganizationRequest{
			Id:         orgID,
			Name:       "updated-name",
			Visibility: shared.Visibility_VISIBILITY_PUBLIC,
		}))
		require.NoError(t, err)
	})

	t.Run("service error - permission denied", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-456"
		orgID := "org-456"

		mockService.EXPECT().
			UpdateOrganization(gomock.Any(), gomock.Any(), testUserID).
			Return(connect.NewError(connect.CodePermissionDenied, errors.New("only the organization creator can update it")))

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.UpdateOrganization(context.Background(), connect.NewRequest(&organizationv1.UpdateOrganizationRequest{
			Id:   orgID,
			Name: "updated-name",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodePermissionDenied, connectErr.Code())
	})

	t.Run("unauthenticated - missing user ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		h := NewHandler(mockService, mockRepository)
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.UpdateOrganization(context.Background(), connect.NewRequest(&organizationv1.UpdateOrganizationRequest{
			Id:   "org-123",
			Name: "updated-name",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})
}

func TestHandler_InviteMember(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-123"
		orgID := "org-123"
		email := "user1@example.com"

		mockService.EXPECT().
			InviteUser(gomock.Any(), gomock.Any(), testUserID).
			DoAndReturn(func(_ context.Context, req *organizationv1.InviteMemberRequest, invitedBy string) error {
				require.Equal(t, orgID, req.GetId())
				require.Equal(t, email, req.GetEmail())
				require.Equal(t, testUserID, invitedBy)
				return nil
			})

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.InviteMember(context.Background(), connect.NewRequest(&organizationv1.InviteMemberRequest{
			Id:    orgID,
			Email: email,
		}))
		require.NoError(t, err)
	})

	t.Run("unauthenticated - missing user ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		h := NewHandler(mockService, mockRepository)
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.InviteMember(context.Background(), connect.NewRequest(&organizationv1.InviteMemberRequest{
			Id:    "org-123",
			Email: "user@example.com",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})
}

func TestHandler_UpdateMemberRole(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-123"
		orgID := "org-123"
		memberID := "member-456"

		mockService.EXPECT().
			UpdateMemberRole(gomock.Any(), gomock.Any(), testUserID).
			DoAndReturn(func(_ context.Context, req *organizationv1.UpdateMemberRoleRequest, userId string) error {
				require.Equal(t, orgID, req.GetOrganizationId())
				require.Equal(t, memberID, req.GetMemberId())
				require.Equal(t, shared.Role_ROLE_AUTHOR, req.GetRole())
				require.Equal(t, testUserID, userId)
				return nil
			})

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.UpdateMemberRole(context.Background(), connect.NewRequest(&organizationv1.UpdateMemberRoleRequest{
			OrganizationId: orgID,
			MemberId:       memberID,
			Role:           shared.Role_ROLE_AUTHOR,
		}))
		require.NoError(t, err)
	})

	t.Run("service error - permission denied", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-456"
		orgID := "org-456"
		memberID := "member-789"

		mockService.EXPECT().
			UpdateMemberRole(gomock.Any(), gomock.Any(), testUserID).
			Return(connect.NewError(connect.CodePermissionDenied, errors.New("only organization owners can update member roles")))

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.UpdateMemberRole(context.Background(), connect.NewRequest(&organizationv1.UpdateMemberRoleRequest{
			OrganizationId: orgID,
			MemberId:       memberID,
			Role:           shared.Role_ROLE_AUTHOR,
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodePermissionDenied, connectErr.Code())
	})

	t.Run("service error - owner cannot decrease own role", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-123"
		orgID := "org-123"

		mockService.EXPECT().
			UpdateMemberRole(gomock.Any(), gomock.Any(), testUserID).
			Return(connect.NewError(connect.CodePermissionDenied, errors.New("owners cannot decrease their own role")))

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.UpdateMemberRole(context.Background(), connect.NewRequest(&organizationv1.UpdateMemberRoleRequest{
			OrganizationId: orgID,
			MemberId:       testUserID,
			Role:           shared.Role_ROLE_AUTHOR,
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodePermissionDenied, connectErr.Code())
	})

	t.Run("service error - member not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-123"
		orgID := "org-123"
		nonExistentMemberID := "non-existent-456"

		mockService.EXPECT().
			UpdateMemberRole(gomock.Any(), gomock.Any(), testUserID).
			Return(connect.NewError(connect.CodeNotFound, errors.New("member not found")))

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.UpdateMemberRole(context.Background(), connect.NewRequest(&organizationv1.UpdateMemberRoleRequest{
			OrganizationId: orgID,
			MemberId:       nonExistentMemberID,
			Role:           shared.Role_ROLE_AUTHOR,
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeNotFound, connectErr.Code())
	})

	t.Run("unauthenticated - missing user ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		h := NewHandler(mockService, mockRepository)
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.UpdateMemberRole(context.Background(), connect.NewRequest(&organizationv1.UpdateMemberRoleRequest{
			OrganizationId: "org-123",
			MemberId:       "member-456",
			Role:           shared.Role_ROLE_AUTHOR,
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})
}

func TestHandler_DeleteMember(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-123"
		orgID := "org-123"
		memberID := "member-456"

		mockService.EXPECT().
			DeleteMember(gomock.Any(), gomock.Any(), testUserID).
			DoAndReturn(func(_ context.Context, req *organizationv1.DeleteMemberRequest, userId string) error {
				require.Equal(t, orgID, req.GetOrganizationId())
				require.Equal(t, memberID, req.GetMemberId())
				require.Equal(t, testUserID, userId)
				return nil
			})

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.DeleteMember(context.Background(), connect.NewRequest(&organizationv1.DeleteMemberRequest{
			OrganizationId: orgID,
			MemberId:       memberID,
		}))
		require.NoError(t, err)
	})

	t.Run("service error - permission denied", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-456"
		orgID := "org-456"
		memberID := "member-789"

		mockService.EXPECT().
			DeleteMember(gomock.Any(), gomock.Any(), testUserID).
			Return(connect.NewError(connect.CodePermissionDenied, errors.New("only organization owners can delete members")))

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.DeleteMember(context.Background(), connect.NewRequest(&organizationv1.DeleteMemberRequest{
			OrganizationId: orgID,
			MemberId:       memberID,
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodePermissionDenied, connectErr.Code())
	})

	t.Run("service error - cannot delete last owner", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-123"
		orgID := "org-123"
		lastOwnerID := "last-owner-456"

		mockService.EXPECT().
			DeleteMember(gomock.Any(), gomock.Any(), testUserID).
			Return(connect.NewError(connect.CodeFailedPrecondition, errors.New("cannot delete the last owner")))

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.DeleteMember(context.Background(), connect.NewRequest(&organizationv1.DeleteMemberRequest{
			OrganizationId: orgID,
			MemberId:       lastOwnerID,
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeFailedPrecondition, connectErr.Code())
	})

	t.Run("service error - member not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		testUserID := "test-user-123"
		orgID := "org-123"
		nonExistentMemberID := "non-existent-456"

		mockService.EXPECT().
			DeleteMember(gomock.Any(), gomock.Any(), testUserID).
			Return(connect.NewError(connect.CodeNotFound, errors.New("member not found")))

		h := NewHandler(mockService, mockRepository, testAuthInterceptor(testUserID))
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.DeleteMember(context.Background(), connect.NewRequest(&organizationv1.DeleteMemberRequest{
			OrganizationId: orgID,
			MemberId:       nonExistentMemberID,
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeNotFound, connectErr.Code())
	})

	t.Run("unauthenticated - missing user ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		h := NewHandler(mockService, mockRepository)
		mux := http.NewServeMux()
		path, handler := h.RegisterRoutes()
		mux.Handle(path, handler)

		server := httptest.NewServer(mux)
		defer server.Close()

		client := organizationv1connect.NewOrganizationServiceClient(
			http.DefaultClient,
			server.URL,
		)

		_, err := client.DeleteMember(context.Background(), connect.NewRequest(&organizationv1.DeleteMemberRequest{
			OrganizationId: "org-123",
			MemberId:       "member-456",
		}))
		require.Error(t, err)

		var connectErr *connect.Error
		require.True(t, errors.As(err, &connectErr))
		require.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})
}
