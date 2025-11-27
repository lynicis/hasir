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

	"apps/api/pkg/auth"
)

func testAuthInterceptor(userID string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			ctx = context.WithValue(ctx, auth.UserIDKey, userID)
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

		// No auth interceptor - should fail with unauthenticated
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

		orgs := &[]OrganizationDTO{
			{Id: "org-1", Name: "first-org", Visibility: VisibilityPrivate},
			{Id: "org-2", Name: "second-org", Visibility: VisibilityPublic},
		}

		mockRepository.EXPECT().
			GetOrganizations(gomock.Any()).
			Return(orgs, nil)

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

		orgs := &[]OrganizationDTO{}

		mockRepository.EXPECT().
			GetOrganizations(gomock.Any()).
			Return(orgs, nil)

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

		resp, err := client.GetOrganizations(context.Background(), connect.NewRequest(&organizationv1.GetOrganizationsRequest{}))
		require.NoError(t, err)
		require.Empty(t, resp.Msg.GetOrganizations())
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockService := NewMockService(ctrl)
		mockRepository := NewMockRepository(ctrl)

		mockRepository.EXPECT().
			GetOrganizations(gomock.Any()).
			Return(nil, connect.NewError(connect.CodeInternal, errors.New("database error")))

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
		require.Equal(t, connect.CodeInternal, connectErr.Code())
	})
}
