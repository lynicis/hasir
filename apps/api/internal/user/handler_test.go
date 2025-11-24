package user

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"connectrpc.com/validate"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"

	"apps/api/internal"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/user/v1/userv1connect"
	userv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/user/v1"
)

func TestNewHandler(t *testing.T) {
	h := NewHandler(nil, nil, nil)
	assert.Implements(t, (*internal.GlobalHandler)(nil), h)
}

func TestHandler_RegisterRoutes(t *testing.T) {
	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)

	h := NewHandler(validateInterceptor, otelInterceptor, nil)
	routes, handler := h.RegisterRoutes()
	assert.NotNil(t, routes)
	assert.NotNil(t, handler)
}

func TestHandler_Register(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)

	t.Run("happy path", func(t *testing.T) {
		mockUserService := NewMockService(mockController)
		mockUserService.
			EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		h := NewHandler(validateInterceptor, otelInterceptor, mockUserService)

		port, err := freeport.GetFreePort()
		require.NoError(t, err)

		urn := fmt.Sprintf("127.0.0.1:%d", port)
		go func() {
			_ = setupServer(urn, h)
		}()

		err = waitForServer(urn, 5*time.Second)
		require.NoError(t, err, "server should be ready")

		url := fmt.Sprintf("http://127.0.0.1:%d", port)
		client := userv1connect.NewUserServiceClient(http.DefaultClient, url)
		resp, err := client.Register(t.Context(), &connect.Request[userv1.RegisterRequest]{
			Msg: &userv1.RegisterRequest{
				Email:    "test@mail.com",
				Username: "test",
				Password: "Asdfg123456_",
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.IsType(t, new(emptypb.Empty), resp.Msg)
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			reqBody connect.Request[userv1.RegisterRequest]
		}{
			{
				name: "invalid email",
				reqBody: connect.Request[userv1.RegisterRequest]{
					Msg: &userv1.RegisterRequest{
						Email:    "invalid@com",
						Username: "TestUser",
						Password: "Asdfg1235_",
					},
				},
			},
			{
				name: "empty username",
				reqBody: connect.Request[userv1.RegisterRequest]{
					Msg: &userv1.RegisterRequest{
						Email:    "test@mail.com",
						Username: "",
						Password: "Asdfg123456_",
					},
				},
			},
			{
				name: "too short password",
				reqBody: connect.Request[userv1.RegisterRequest]{
					Msg: &userv1.RegisterRequest{
						Email:    "test@mail.com",
						Username: "TestUser",
						Password: "Asdf123",
					},
				},
			},
			{
				name: "too long password",
				reqBody: connect.Request[userv1.RegisterRequest]{
					Msg: &userv1.RegisterRequest{
						Email:    "test@mail.com",
						Username: "TestUser",
						Password: "Asdf123456789123456789123456789123456789123456789123456789123456_",
					},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				port, err := freeport.GetFreePort()
				require.NoError(t, err)

				h := NewHandler(validateInterceptor, otelInterceptor, nil)

				urn := fmt.Sprintf("127.0.0.1:%d", port)
				go func() {
					_ = setupServer(urn, h)
				}()

				err = waitForServer(urn, 5*time.Second)
				require.NoError(t, err, "server should be ready")

				url := fmt.Sprintf("http://127.0.0.1:%d", port)
				client := userv1connect.NewUserServiceClient(http.DefaultClient, url)
				resp, err := client.Register(t.Context(), &test.reqBody)

				assert.Error(t, err)
				assert.Nil(t, resp)
			})
		}
	})

	t.Run("service error", func(t *testing.T) {
		mockUserService := NewMockService(mockController)
		mockUserService.
			EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return(errors.New("something went wrong")).
			Times(1)

		h := NewHandler(validateInterceptor, otelInterceptor, mockUserService)

		port, err := freeport.GetFreePort()
		require.NoError(t, err)

		urn := fmt.Sprintf("127.0.0.1:%d", port)
		go func() {
			_ = setupServer(urn, h)
		}()

		err = waitForServer(urn, 5*time.Second)
		require.NoError(t, err, "server should be ready")

		url := fmt.Sprintf("http://127.0.0.1:%d", port)
		client := userv1connect.NewUserServiceClient(http.DefaultClient, url)
		resp, err := client.Register(t.Context(), &connect.Request[userv1.RegisterRequest]{
			Msg: &userv1.RegisterRequest{
				Email:    "test@mail.com",
				Username: "test",
				Password: "Asdfg123456_",
			},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestHandler_Login(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	validateInterceptor := validate.NewInterceptor()
	otelInterceptor, err := otelconnect.NewInterceptor()
	require.NoError(t, err)

	t.Run("happy path", func(t *testing.T) {
		mockUserService := NewMockService(mockController)

		mockRespBody := &userv1.LoginResponse{
			AccessToken:  "abcd.abcd.abcd",
			RefreshToken: "abcd.abcd.abcd",
		}

		mockUserService.
			EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return(mockRespBody, nil).
			Times(1)

		h := NewHandler(validateInterceptor, otelInterceptor, mockUserService)

		port, err := freeport.GetFreePort()
		require.NoError(t, err)

		urn := fmt.Sprintf("127.0.0.1:%d", port)
		go func() {
			_ = setupServer(urn, h)
		}()

		err = waitForServer(urn, 5*time.Second)
		require.NoError(t, err, "server should be ready")

		url := fmt.Sprintf("http://127.0.0.1:%d", port)
		client := userv1connect.NewUserServiceClient(http.DefaultClient, url)
		resp, err := client.Login(t.Context(), &connect.Request[userv1.LoginRequest]{
			Msg: &userv1.LoginRequest{
				Email:    "test@mail.com",
				Password: "Asdfg123456_",
			},
		})

		assert.NoError(t, err)
		require.NotNil(t, resp, "response should not be nil")
		require.NotNil(t, resp.Msg, "response message should not be nil")
		assert.Equal(t, mockRespBody.AccessToken, resp.Msg.AccessToken)
		assert.Equal(t, mockRespBody.RefreshToken, resp.Msg.RefreshToken)
	})

	t.Run("validation errors", func(t *testing.T) {
		testTable := []struct {
			name    string
			reqBody connect.Request[userv1.LoginRequest]
		}{
			{
				name: "invalid email",
				reqBody: connect.Request[userv1.LoginRequest]{
					Msg: &userv1.LoginRequest{
						Email:    "invalid@mail.com",
						Password: "Asdfg123456_",
					},
				},
			},
			{
				name: "too short password",
				reqBody: connect.Request[userv1.LoginRequest]{
					Msg: &userv1.LoginRequest{
						Email:    "test@mail.com",
						Password: "Asdfg12_",
					},
				},
			},
		}

		for _, test := range testTable {
			t.Run(test.name, func(t *testing.T) {
				h := NewHandler(validateInterceptor, otelInterceptor, nil)

				port, err := freeport.GetFreePort()
				require.NoError(t, err)

				urn := fmt.Sprintf("127.0.0.1:%d", port)
				go func() {
					_ = setupServer(urn, h)
				}()

				err = waitForServer(urn, 5*time.Second)
				require.NoError(t, err, "server should be ready")

				url := fmt.Sprintf("http://127.0.0.1:%d", port)
				client := userv1connect.NewUserServiceClient(http.DefaultClient, url)
				resp, err := client.Login(t.Context(), &test.reqBody)

				assert.Error(t, err)
				assert.Nil(t, resp)
			})
		}
	})

	t.Run("service error", func(t *testing.T) {
		mockUserService := NewMockService(mockController)
		mockUserService.
			EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("something went wrong")).
			Times(1)

		h := NewHandler(validateInterceptor, otelInterceptor, mockUserService)

		port, err := freeport.GetFreePort()
		require.NoError(t, err)

		urn := fmt.Sprintf("127.0.0.1:%d", port)
		go func() {
			_ = setupServer(urn, h)
		}()

		err = waitForServer(urn, 5*time.Second)
		require.NoError(t, err, "server should be ready")

		url := fmt.Sprintf("http://127.0.0.1:%d", port)
		client := userv1connect.NewUserServiceClient(http.DefaultClient, url)
		resp, err := client.Login(t.Context(), &connect.Request[userv1.LoginRequest]{
			Msg: &userv1.LoginRequest{
				Email:    "test@mail.com",
				Password: "Asdfg123456_",
			},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func setupServer(urn string, h internal.GlobalHandler) error {
	mux := http.NewServeMux()
	path, handler := h.RegisterRoutes()
	mux.Handle(path, handler)

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)
	server := &http.Server{
		Addr:      urn,
		Handler:   mux,
		Protocols: protocols,
	}

	return server.ListenAndServe()
}

func waitForServer(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("server at %s did not become ready within %v", addr, timeout)
}
