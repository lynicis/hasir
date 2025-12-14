package user

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"connectrpc.com/validate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"

	"hasir-api/internal"
	"hasir-api/pkg/authentication"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/user/v1/userv1connect"
	"buf.build/gen/go/hasir/hasir/protocolbuffers/go/shared"
	userv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/user/v1"
)

func testAuthInterceptor(userID string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			ctx = context.WithValue(ctx, authentication.UserIDKey, userID)
			return next(ctx, req)
		}
	}
}

func setupTestServer(t *testing.T, h internal.GlobalHandler) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	path, handler := h.RegisterRoutes()
	mux.Handle(path, handler)
	return httptest.NewServer(mux)
}

func TestNewHandler(t *testing.T) {
	h := NewHandler(nil, nil)
	assert.Implements(t, (*internal.GlobalHandler)(nil), h)
}

func TestHandler_RegisterRoutes(t *testing.T) {
	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)

	h := NewHandler(nil, nil, validateInterceptor, otelInterceptor)
	routes, handler := h.RegisterRoutes()
	assert.NotNil(t, routes)
	assert.NotNil(t, handler)
}

func TestHandler_Register(t *testing.T) {
	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)
	interceptors := []connect.Interceptor{validateInterceptor, otelInterceptor}

	t.Run("happy path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		mockUserService.
			EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.Register(context.Background(), connect.NewRequest(&userv1.RegisterRequest{
			Email:    "test@mail.com",
			Username: "test",
			Password: "Asdfg123456_",
		}))

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.IsType(t, new(emptypb.Empty), resp.Msg)
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			request *userv1.RegisterRequest
		}{
			{
				name: "invalid email",
				request: &userv1.RegisterRequest{
					Email:    "invalid@com",
					Username: "TestUser",
					Password: "Asdfg1235_",
				},
			},
			{
				name: "empty username",
				request: &userv1.RegisterRequest{
					Email:    "test@mail.com",
					Username: "",
					Password: "Asdfg123456_",
				},
			},
			{
				name: "too short password",
				request: &userv1.RegisterRequest{
					Email:    "test@mail.com",
					Username: "TestUser",
					Password: "Asdf123",
				},
			},
			{
				name: "too long password",
				request: &userv1.RegisterRequest{
					Email:    "test@mail.com",
					Username: "TestUser",
					Password: "Asdf123456789123456789123456789123456789123456789123456789123456_",
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				mockUserService := NewMockService(ctrl)
				mockUserRepository := NewMockRepository(ctrl)

				mockUserService.
					EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(connect.NewError(connect.CodeInvalidArgument, errors.New("validation error"))).
					AnyTimes()

				h := NewHandler(mockUserService, mockUserRepository, interceptors...)
				server := setupTestServer(t, h)
				defer server.Close()

				client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
				resp, err := client.Register(context.Background(), connect.NewRequest(tc.request))

				assert.Error(t, err)
				assert.Nil(t, resp)
			})
		}
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		mockUserService.
			EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return(errors.New("something went wrong")).
			Times(1)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.Register(context.Background(), connect.NewRequest(&userv1.RegisterRequest{
			Email:    "test@mail.com",
			Username: "test",
			Password: "Asdfg123456_",
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestHandler_Login(t *testing.T) {
	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)
	interceptors := []connect.Interceptor{validateInterceptor, otelInterceptor}

	t.Run("happy path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		mockRespBody := &userv1.TokenEnvelope{
			AccessToken:  "abcd.abcd.abcd",
			RefreshToken: "abcd.abcd.abcd",
		}

		mockUserService.
			EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return(mockRespBody, nil).
			Times(1)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.Login(context.Background(), connect.NewRequest(&userv1.LoginRequest{
			Email:    "test@mail.com",
			Password: "Asdfg123456_",
		}))

		assert.NoError(t, err)
		assert.NotNil(t, resp, "response should not be nil")
		assert.NotNil(t, resp.Msg, "response message should not be nil")
		assert.Equal(t, mockRespBody.AccessToken, resp.Msg.AccessToken)
		assert.Equal(t, mockRespBody.RefreshToken, resp.Msg.RefreshToken)
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			request *userv1.LoginRequest
		}{
			{
				name: "invalid email",
				request: &userv1.LoginRequest{
					Email:    "invalid@mail.com",
					Password: "Asdfg123456_",
				},
			},
			{
				name: "too short password",
				request: &userv1.LoginRequest{
					Email:    "test@mail.com",
					Password: "Asdfg12_",
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				mockUserService := NewMockService(ctrl)
				mockUserRepository := NewMockRepository(ctrl)

				mockUserService.
					EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(nil, connect.NewError(connect.CodeInvalidArgument, errors.New("validation error"))).
					AnyTimes()

				h := NewHandler(mockUserService, mockUserRepository, interceptors...)
				server := setupTestServer(t, h)
				defer server.Close()

				client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
				resp, err := client.Login(context.Background(), connect.NewRequest(tc.request))

				assert.Error(t, err)
				assert.Nil(t, resp)
			})
		}
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		mockUserService.
			EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("something went wrong")).
			Times(1)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.Login(context.Background(), connect.NewRequest(&userv1.LoginRequest{
			Email:    "test@mail.com",
			Password: "Asdfg123456_",
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestHandler_RenewTokens(t *testing.T) {
	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)
	interceptors := []connect.Interceptor{validateInterceptor, otelInterceptor}

	t.Run("happy path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		mockRespBody := &userv1.RenewTokensResponse{
			AccessToken: "new.access.token",
		}

		mockUserService.
			EXPECT().
			RenewTokens(gomock.Any(), gomock.Any()).
			Return(mockRespBody, nil).
			Times(1)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.RenewTokens(context.Background(), connect.NewRequest(&userv1.RenewTokensRequest{
			RefreshToken: "old.refresh.token",
		}))

		assert.NoError(t, err)
		assert.NotNil(t, resp, "response should not be nil")
		assert.NotNil(t, resp.Msg, "response message should not be nil")
		assert.Equal(t, mockRespBody.AccessToken, resp.Msg.AccessToken)
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		mockUserService.
			EXPECT().
			RenewTokens(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("something went wrong")).
			Times(1)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.RenewTokens(context.Background(), connect.NewRequest(&userv1.RenewTokensRequest{
			RefreshToken: "old.refresh.token",
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestHandler_UpdateUser(t *testing.T) {
	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)
	interceptors := []connect.Interceptor{validateInterceptor, otelInterceptor}

	t.Run("happy path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		mockRespBody := &userv1.TokenEnvelope{
			AccessToken:  "abcd.abcd.abcd",
			RefreshToken: "abcd.abcd.abcd",
		}

		mockUserService.
			EXPECT().
			UpdateUser(gomock.Any(), gomock.Any()).
			Return(mockRespBody, nil).
			Times(1)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		email := "newemail@mail.com"
		password := "OldPassword123_"
		newPassword := "NewPassword123_"
		resp, err := client.UpdateUser(context.Background(), connect.NewRequest(&userv1.UpdateUserRequest{
			Email:       &email,
			Password:    password,
			NewPassword: &newPassword,
		}))

		assert.NoError(t, err)
		assert.NotNil(t, resp, "response should not be nil")
		assert.NotNil(t, resp.Msg, "response message should not be nil")
		assert.Equal(t, mockRespBody.AccessToken, resp.Msg.AccessToken)
		assert.Equal(t, mockRespBody.RefreshToken, resp.Msg.RefreshToken)
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			request *userv1.UpdateUserRequest
		}{
			{
				name: "invalid email",
				request: &userv1.UpdateUserRequest{
					Email: func() *string { s := "invalid@com"; return &s }(),
				},
			},
			{
				name: "empty username",
				request: &userv1.UpdateUserRequest{
					Email: func() *string { s := "newemail@mail.com"; return &s }(),
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				mockUserService := NewMockService(ctrl)
				mockUserRepository := NewMockRepository(ctrl)

				mockUserService.
					EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Return(nil, connect.NewError(connect.CodeInvalidArgument, errors.New("validation error"))).
					AnyTimes()

				h := NewHandler(mockUserService, mockUserRepository, interceptors...)
				server := setupTestServer(t, h)
				defer server.Close()

				client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
				resp, err := client.UpdateUser(context.Background(), connect.NewRequest(tc.request))

				assert.Error(t, err)
				assert.Nil(t, resp)
			})
		}
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		mockUserService.
			EXPECT().
			UpdateUser(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("something went wrong")).
			Times(1)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		email := "newemail@mail.com"
		password := "OldPassword123_"
		newPassword := "NewPassword123_"
		resp, err := client.UpdateUser(context.Background(), connect.NewRequest(&userv1.UpdateUserRequest{
			Email:       &email,
			Password:    password,
			NewPassword: &newPassword,
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestHandler_DeleteAccount(t *testing.T) {
	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)
	interceptors := []connect.Interceptor{validateInterceptor, otelInterceptor}

	t.Run("happy path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		testUserID := "test-user-id"

		mockUserRepository.
			EXPECT().
			DeleteUser(gomock.Any(), testUserID).
			Return(nil).
			Times(1)

		allInterceptors := append(interceptors, testAuthInterceptor(testUserID))
		h := NewHandler(mockUserService, mockUserRepository, allInterceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.DeleteAccount(context.Background(), connect.NewRequest(new(emptypb.Empty)))

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.IsType(t, new(emptypb.Empty), resp.Msg)
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			request *emptypb.Empty
		}{
			{
				name:    "empty request",
				request: new(emptypb.Empty),
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				mockUserService := NewMockService(ctrl)
				mockUserRepository := NewMockRepository(ctrl)

				h := NewHandler(mockUserService, mockUserRepository, interceptors...)
				server := setupTestServer(t, h)
				defer server.Close()

				client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
				resp, err := client.DeleteAccount(context.Background(), connect.NewRequest(tc.request))

				assert.Error(t, err)
				assert.Nil(t, resp)
			})
		}
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		testUserID := "test-user-id"

		mockUserRepository.
			EXPECT().
			DeleteUser(gomock.Any(), testUserID).
			Return(errors.New("something went wrong")).
			Times(1)

		allInterceptors := append(interceptors, testAuthInterceptor(testUserID))
		h := NewHandler(mockUserService, mockUserRepository, allInterceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.DeleteAccount(context.Background(), connect.NewRequest(new(emptypb.Empty)))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestHandler_CreateApiKey(t *testing.T) {
	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)
	interceptors := []connect.Interceptor{validateInterceptor, otelInterceptor}

	t.Run("happy path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		testUserID := "test-user-id"

		mockUserRepository.
			EXPECT().
			CreateApiKey(gomock.Any(), testUserID, gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		allInterceptors := append(interceptors, testAuthInterceptor(testUserID))
		h := NewHandler(mockUserService, mockUserRepository, allInterceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.CreateApiKey(context.Background(), connect.NewRequest(&userv1.CreateApiKeyRequest{
			Name: "test-key",
		}))

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Msg)
		assert.NotEmpty(t, resp.Msg.Key)
	})

	t.Run("auth error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.CreateApiKey(context.Background(), connect.NewRequest(&userv1.CreateApiKeyRequest{
			Name: "test-key",
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		testUserID := "test-user-id"

		mockUserRepository.
			EXPECT().
			CreateApiKey(gomock.Any(), testUserID, gomock.Any(), gomock.Any()).
			Return(errors.New("something went wrong")).
			Times(1)

		allInterceptors := append(interceptors, testAuthInterceptor(testUserID))
		h := NewHandler(mockUserService, mockUserRepository, allInterceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.CreateApiKey(context.Background(), connect.NewRequest(&userv1.CreateApiKeyRequest{
			Name: "test-key",
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestHandler_GetApiKeys(t *testing.T) {
	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)
	interceptors := []connect.Interceptor{validateInterceptor, otelInterceptor}

	t.Run("happy path with pagination", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		testUserID := "test-user-id"

		apiKeys := []ApiKeyDTO{
			{Id: "id-1", Name: "key-1"},
			{Id: "id-2", Name: "key-2"},
		}

		mockUserRepository.
			EXPECT().
			GetApiKeysCount(gomock.Any(), testUserID).
			Return(5, nil).
			Times(1)

		mockUserRepository.
			EXPECT().
			GetApiKeys(gomock.Any(), testUserID, 1, 10).
			Return(&apiKeys, nil).
			Times(1)

		allInterceptors := append(interceptors, testAuthInterceptor(testUserID))
		h := NewHandler(mockUserService, mockUserRepository, allInterceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.GetApiKeys(context.Background(), connect.NewRequest(&shared.Pagination{
			Page:      1,
			PageLimit: 10,
		}))

		assert.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Msg)
		require.Len(t, resp.Msg.Keys, 2)
		assert.Equal(t, int32(1), resp.Msg.TotalPage)
		assert.Equal(t, int32(0), resp.Msg.NextPage)

		var names []string
		for _, k := range resp.Msg.Keys {
			names = append(names, k.Name)
		}
		assert.ElementsMatch(t, []string{"key-1", "key-2"}, names)
	})

	t.Run("auth error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.GetApiKeys(context.Background(), connect.NewRequest(&shared.Pagination{
			Page:      1,
			PageLimit: 10,
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("count error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		testUserID := "test-user-id"

		mockUserRepository.
			EXPECT().
			GetApiKeysCount(gomock.Any(), testUserID).
			Return(0, errors.New("something went wrong")).
			Times(1)

		allInterceptors := append(interceptors, testAuthInterceptor(testUserID))
		h := NewHandler(mockUserService, mockUserRepository, allInterceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.GetApiKeys(context.Background(), connect.NewRequest(&shared.Pagination{
			Page:      1,
			PageLimit: 10,
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		testUserID := "test-user-id"

		mockUserRepository.
			EXPECT().
			GetApiKeysCount(gomock.Any(), testUserID).
			Return(5, nil).
			Times(1)

		mockUserRepository.
			EXPECT().
			GetApiKeys(gomock.Any(), testUserID, 1, 10).
			Return(nil, errors.New("something went wrong")).
			Times(1)

		allInterceptors := append(interceptors, testAuthInterceptor(testUserID))
		h := NewHandler(mockUserService, mockUserRepository, allInterceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.GetApiKeys(context.Background(), connect.NewRequest(&shared.Pagination{
			Page:      1,
			PageLimit: 10,
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestHandler_CreateSshKey(t *testing.T) {
	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)
	interceptors := []connect.Interceptor{validateInterceptor, otelInterceptor}

	t.Run("happy path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		testUserID := "test-user-id"

		publicKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBhVLF+dcZbEWbWr1A+8YYLBxDGgmBdwk6IB/+W5v/Wh test@example.com"
		expectedNormalizedKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBhVLF+dcZbEWbWr1A+8YYLBxDGgmBdwk6IB/+W5v/Wh"

		mockUserRepository.
			EXPECT().
			CreateSshKey(gomock.Any(), testUserID, gomock.Any(), expectedNormalizedKey).
			Return(nil).
			Times(1)

		allInterceptors := append(interceptors, testAuthInterceptor(testUserID))
		h := NewHandler(mockUserService, mockUserRepository, allInterceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.CreateSshKey(context.Background(), connect.NewRequest(&userv1.CreateSshKeyRequest{
			Name:      "test-ssh-key",
			PublicKey: publicKey,
		}))

		assert.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Msg)
	})

	t.Run("auth error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		publicKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBhVLF+dcZbEWbWr1A+8YYLBxDGgmBdwk6IB/+W5v/Wh test@example.com"

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.CreateSshKey(context.Background(), connect.NewRequest(&userv1.CreateSshKeyRequest{
			Name:      "test-ssh-key",
			PublicKey: publicKey,
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		testUserID := "test-user-id"

		publicKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBhVLF+dcZbEWbWr1A+8YYLBxDGgmBdwk6IB/+W5v/Wh test@example.com"
		expectedNormalizedKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBhVLF+dcZbEWbWr1A+8YYLBxDGgmBdwk6IB/+W5v/Wh"

		mockUserRepository.
			EXPECT().
			CreateSshKey(gomock.Any(), testUserID, gomock.Any(), expectedNormalizedKey).
			Return(errors.New("something went wrong")).
			Times(1)

		allInterceptors := append(interceptors, testAuthInterceptor(testUserID))
		h := NewHandler(mockUserService, mockUserRepository, allInterceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.CreateSshKey(context.Background(), connect.NewRequest(&userv1.CreateSshKeyRequest{
			Name:      "test-ssh-key",
			PublicKey: publicKey,
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestHandler_RevokeApiKey(t *testing.T) {
	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)
	interceptors := []connect.Interceptor{validateInterceptor, otelInterceptor}

	t.Run("successful key revocation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		testUserID := "test-user-id"
		keyID := "test-key-id"

		req := &userv1.RevokeKeyRequest{Id: keyID}

		mockUserRepository.
			EXPECT().
			RevokeApiKey(gomock.Any(), testUserID, keyID).
			Return(nil).
			Times(1)

		allInterceptors := append(interceptors, testAuthInterceptor(testUserID))
		h := NewHandler(mockUserService, mockUserRepository, allInterceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		_, err := client.RevokeApiKey(context.Background(), connect.NewRequest(req))

		assert.NoError(t, err)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.RevokeApiKey(context.Background(), connect.NewRequest(&userv1.RevokeKeyRequest{Id: "key-id"}))

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		testUserID := "test-user-id"
		keyID := "test-key-id"

		req := &userv1.RevokeKeyRequest{Id: keyID}

		expectedErr := errors.New("database error")
		mockUserRepository.
			EXPECT().
			RevokeApiKey(gomock.Any(), testUserID, keyID).
			Return(expectedErr).
			Times(1)

		allInterceptors := append(interceptors, testAuthInterceptor(testUserID))
		h := NewHandler(mockUserService, mockUserRepository, allInterceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.RevokeApiKey(context.Background(), connect.NewRequest(req))

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), expectedErr.Error())
	})
}

func TestHandler_GetSshKeys(t *testing.T) {
	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)
	interceptors := []connect.Interceptor{validateInterceptor, otelInterceptor}

	t.Run("happy path with pagination", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		testUserID := "test-user-id"

		sshKeys := []SshKeyDTO{
			{Id: "id-1", Name: "ssh-key-1"},
			{Id: "id-2", Name: "ssh-key-2"},
		}

		mockUserRepository.
			EXPECT().
			GetSshKeysCount(gomock.Any(), testUserID).
			Return(5, nil).
			Times(1)

		mockUserRepository.
			EXPECT().
			GetSshKeys(gomock.Any(), testUserID, 1, 10).
			Return(&sshKeys, nil).
			Times(1)

		allInterceptors := append(interceptors, testAuthInterceptor(testUserID))
		h := NewHandler(mockUserService, mockUserRepository, allInterceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.GetSshKeys(context.Background(), connect.NewRequest(&shared.Pagination{
			Page:      1,
			PageLimit: 10,
		}))

		assert.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Msg)
		require.Len(t, resp.Msg.Keys, 2)
		assert.Equal(t, int32(1), resp.Msg.TotalPage)
		assert.Equal(t, int32(0), resp.Msg.NextPage)

		var names []string
		for _, k := range resp.Msg.Keys {
			names = append(names, k.Name)
		}
		assert.ElementsMatch(t, []string{"ssh-key-1", "ssh-key-2"}, names)
	})

	t.Run("auth error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.GetSshKeys(context.Background(), connect.NewRequest(&shared.Pagination{
			Page:      1,
			PageLimit: 10,
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("count error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		testUserID := "test-user-id"

		mockUserRepository.
			EXPECT().
			GetSshKeysCount(gomock.Any(), testUserID).
			Return(0, errors.New("something went wrong")).
			Times(1)

		allInterceptors := append(interceptors, testAuthInterceptor(testUserID))
		h := NewHandler(mockUserService, mockUserRepository, allInterceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.GetSshKeys(context.Background(), connect.NewRequest(&shared.Pagination{
			Page:      1,
			PageLimit: 10,
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		testUserID := "test-user-id"

		mockUserRepository.
			EXPECT().
			GetSshKeysCount(gomock.Any(), testUserID).
			Return(5, nil).
			Times(1)

		mockUserRepository.
			EXPECT().
			GetSshKeys(gomock.Any(), testUserID, 1, 10).
			Return(nil, errors.New("something went wrong")).
			Times(1)

		allInterceptors := append(interceptors, testAuthInterceptor(testUserID))
		h := NewHandler(mockUserService, mockUserRepository, allInterceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.GetSshKeys(context.Background(), connect.NewRequest(&shared.Pagination{
			Page:      1,
			PageLimit: 10,
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestHandler_ForgotPassword(t *testing.T) {
	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)
	interceptors := []connect.Interceptor{validateInterceptor, otelInterceptor}

	t.Run("happy path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		mockUserService.
			EXPECT().
			ForgotPassword(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.ForgotPassword(context.Background(), connect.NewRequest(&userv1.ForgotPasswordRequest{
			Email: "test@mail.com",
		}))

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.IsType(t, new(emptypb.Empty), resp.Msg)
	})

	t.Run("validation error - invalid email", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.ForgotPassword(context.Background(), connect.NewRequest(&userv1.ForgotPasswordRequest{
			Email: "invalid-email",
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		mockUserService.
			EXPECT().
			ForgotPassword(gomock.Any(), gomock.Any()).
			Return(errors.New("service error")).
			Times(1)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.ForgotPassword(context.Background(), connect.NewRequest(&userv1.ForgotPasswordRequest{
			Email: "test@mail.com",
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestHandler_ResetPassword(t *testing.T) {
	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)
	interceptors := []connect.Interceptor{validateInterceptor, otelInterceptor}

	t.Run("happy path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		mockUserService.
			EXPECT().
			ResetPassword(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.ResetPassword(context.Background(), connect.NewRequest(&userv1.ResetPasswordRequest{
			Token:       "reset-token-123",
			NewPassword: "NewPassword123!",
		}))

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.IsType(t, new(emptypb.Empty), resp.Msg)
	})

	t.Run("validation error - empty token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.ResetPassword(context.Background(), connect.NewRequest(&userv1.ResetPasswordRequest{
			Token:       "",
			NewPassword: "NewPassword123!",
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("validation error - weak password", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.ResetPassword(context.Background(), connect.NewRequest(&userv1.ResetPasswordRequest{
			Token:       "reset-token-123",
			NewPassword: "weak",
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockUserService := NewMockService(ctrl)
		mockUserRepository := NewMockRepository(ctrl)

		mockUserService.
			EXPECT().
			ResetPassword(gomock.Any(), gomock.Any()).
			Return(errors.New("service error")).
			Times(1)

		h := NewHandler(mockUserService, mockUserRepository, interceptors...)
		server := setupTestServer(t, h)
		defer server.Close()

		client := userv1connect.NewUserServiceClient(http.DefaultClient, server.URL)
		resp, err := client.ResetPassword(context.Background(), connect.NewRequest(&userv1.ResetPasswordRequest{
			Token:       "reset-token-123",
			NewPassword: "NewPassword123!",
		}))

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
