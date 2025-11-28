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
	"hasir-api/pkg/auth"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/user/v1/userv1connect"
	userv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/user/v1"
)

func testAuthInterceptor(userID string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			ctx = context.WithValue(ctx, auth.UserIDKey, userID)
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

				// Service might be called if validation doesn't catch the error
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
		require.NotNil(t, resp, "response should not be nil")
		require.NotNil(t, resp.Msg, "response message should not be nil")
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

				// Service might be called if validation doesn't catch the error
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
		require.NotNil(t, resp, "response should not be nil")
		require.NotNil(t, resp.Msg, "response message should not be nil")
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

				// Service might be called if validation doesn't catch the error
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
