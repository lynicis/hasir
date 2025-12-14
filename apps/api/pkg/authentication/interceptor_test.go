package authentication

import (
	"context"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

var testSecret = []byte("test-secret-key-for-testing-only")

func generateTestToken(t *testing.T, userID string, email string, expiresAt time.Time) string {
	t.Helper()

	claims := &JwtClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(testSecret)
	require.NoError(t, err)

	return signedToken
}

func TestNewAuthInterceptor(t *testing.T) {
	interceptor := NewAuthInterceptor(testSecret)
	require.NotNil(t, interceptor)
	assert.Equal(t, testSecret, interceptor.jwtSecret)
}

func TestGetUserID(t *testing.T) {
	t.Run("returns user ID when present", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIDKey, "user-123")
		userID, ok := GetUserID(ctx)
		assert.True(t, ok)
		assert.Equal(t, "user-123", userID)
	})

	t.Run("returns false when not present", func(t *testing.T) {
		ctx := context.Background()
		userID, ok := GetUserID(ctx)
		assert.False(t, ok)
		assert.Empty(t, userID)
	})

	t.Run("returns false for wrong type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIDKey, 123)
		userID, ok := GetUserID(ctx)
		assert.False(t, ok)
		assert.Empty(t, userID)
	})
}

func TestGetUserEmail(t *testing.T) {
	t.Run("returns email when present", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserEmailKey, "test@example.com")
		email, ok := GetUserEmail(ctx)
		assert.True(t, ok)
		assert.Equal(t, "test@example.com", email)
	})

	t.Run("returns false when not present", func(t *testing.T) {
		ctx := context.Background()
		email, ok := GetUserEmail(ctx)
		assert.False(t, ok)
		assert.Empty(t, email)
	})

	t.Run("returns false for wrong type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserEmailKey, 123)
		email, ok := GetUserEmail(ctx)
		assert.False(t, ok)
		assert.Empty(t, email)
	})
}

func TestMustGetUserID(t *testing.T) {
	t.Run("returns user ID when present", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIDKey, "user-123")
		userID, err := MustGetUserID(ctx)
		require.NoError(t, err)
		assert.Equal(t, "user-123", userID)
	})

	t.Run("returns error when not present", func(t *testing.T) {
		ctx := context.Background()
		_, err := MustGetUserID(ctx)
		require.Error(t, err)

		var connectErr *connect.Error
		require.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})
}

func TestAuthInterceptor_WrapStreamingClient(t *testing.T) {
	interceptor := NewAuthInterceptor(testSecret)

	called := false
	nextFunc := func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		called = true
		return nil
	}

	wrappedFunc := interceptor.WrapStreamingClient(nextFunc)
	wrappedFunc(context.Background(), connect.Spec{})

	assert.True(t, called, "streaming client should pass through")
}

type testRequest struct {
	*connect.Request[emptypb.Empty]
	procedureOverride string
}

func (t *testRequest) Spec() connect.Spec {
	if t.procedureOverride != "" {
		return connect.Spec{Procedure: t.procedureOverride}
	}
	return t.Request.Spec()
}

func TestAuthInterceptor_WrapUnary(t *testing.T) {
	interceptor := NewAuthInterceptor(testSecret)

	t.Run("public method bypasses auth", func(t *testing.T) {
		called := false
		nextFunc := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			called = true
			return connect.NewResponse(new(emptypb.Empty)), nil
		}

		wrappedFunc := interceptor.WrapUnary(nextFunc)
		realReq := connect.NewRequest(new(emptypb.Empty))
		req := &testRequest{
			Request:           realReq,
			procedureOverride: "/user.v1.UserService/Register",
		}

		_, err := wrappedFunc(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("valid token extracts user info", func(t *testing.T) {
		userID := "user-123"
		email := "test@example.com"
		token := generateTestToken(t, userID, email, time.Now().Add(time.Hour))

		var capturedUserID string
		var capturedEmail string

		nextFunc := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			capturedUserID, _ = GetUserID(ctx)
			capturedEmail, _ = GetUserEmail(ctx)
			return connect.NewResponse(new(emptypb.Empty)), nil
		}

		wrappedFunc := interceptor.WrapUnary(nextFunc)
		realReq := connect.NewRequest(new(emptypb.Empty))
		realReq.Header().Set("Authorization", "Bearer "+token)
		req := &testRequest{
			Request:           realReq,
			procedureOverride: "/user.v1.UserService/GetProfile",
		}

		_, err := wrappedFunc(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, userID, capturedUserID)
		assert.Equal(t, email, capturedEmail)
	})

	t.Run("missing token returns error", func(t *testing.T) {
		nextFunc := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			return connect.NewResponse(new(emptypb.Empty)), nil
		}

		wrappedFunc := interceptor.WrapUnary(nextFunc)
		realReq := connect.NewRequest(new(emptypb.Empty))
		req := &testRequest{
			Request:           realReq,
			procedureOverride: "/user.v1.UserService/GetProfile",
		}

		_, err := wrappedFunc(context.Background(), req)
		require.Error(t, err)
		assert.Equal(t, ErrMissingToken, err)
	})

	t.Run("missing Bearer prefix returns error", func(t *testing.T) {
		nextFunc := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			return connect.NewResponse(new(emptypb.Empty)), nil
		}

		wrappedFunc := interceptor.WrapUnary(nextFunc)
		realReq := connect.NewRequest(new(emptypb.Empty))
		realReq.Header().Set("Authorization", "InvalidToken")
		req := &testRequest{
			Request:           realReq,
			procedureOverride: "/user.v1.UserService/GetProfile",
		}

		_, err := wrappedFunc(context.Background(), req)
		require.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
	})

	t.Run("expired token returns error", func(t *testing.T) {
		token := generateTestToken(t, "user-123", "test@example.com", time.Now().Add(-time.Hour))

		nextFunc := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			return connect.NewResponse(new(emptypb.Empty)), nil
		}

		wrappedFunc := interceptor.WrapUnary(nextFunc)
		realReq := connect.NewRequest(new(emptypb.Empty))
		realReq.Header().Set("Authorization", "Bearer "+token)
		req := &testRequest{
			Request:           realReq,
			procedureOverride: "/user.v1.UserService/GetProfile",
		}

		_, err := wrappedFunc(context.Background(), req)
		require.Error(t, err)
		assert.Equal(t, ErrTokenExpired, err)
	})

	t.Run("invalid signature returns error", func(t *testing.T) {
		claims := &JwtClaims{
			Email: "test@example.com",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user-123",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString([]byte("wrong-secret"))
		require.NoError(t, err)

		nextFunc := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			return connect.NewResponse(new(emptypb.Empty)), nil
		}

		wrappedFunc := interceptor.WrapUnary(nextFunc)
		realReq := connect.NewRequest(new(emptypb.Empty))
		realReq.Header().Set("Authorization", "Bearer "+signedToken)
		req := &testRequest{
			Request:           realReq,
			procedureOverride: "/user.v1.UserService/GetProfile",
		}

		_, err = wrappedFunc(context.Background(), req)
		require.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
	})

	t.Run("wrong signing method returns error", func(t *testing.T) {
		claims := &JwtClaims{
			Email: "test@example.com",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user-123",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		nextFunc := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			return connect.NewResponse(new(emptypb.Empty)), nil
		}

		wrappedFunc := interceptor.WrapUnary(nextFunc)
		realReq := connect.NewRequest(new(emptypb.Empty))
		realReq.Header().Set("Authorization", "Bearer "+signedToken)
		req := &testRequest{
			Request:           realReq,
			procedureOverride: "/user.v1.UserService/GetProfile",
		}

		_, err = wrappedFunc(context.Background(), req)
		require.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
	})

	t.Run("empty subject is handled", func(t *testing.T) {

		claims := &JwtClaims{
			Email: "test@example.com",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(testSecret)
		require.NoError(t, err)

		var capturedUserID string
		nextFunc := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			capturedUserID, _ = GetUserID(ctx)
			return connect.NewResponse(new(emptypb.Empty)), nil
		}

		wrappedFunc := interceptor.WrapUnary(nextFunc)
		realReq := connect.NewRequest(new(emptypb.Empty))
		realReq.Header().Set("Authorization", "Bearer "+signedToken)
		req := &testRequest{
			Request:           realReq,
			procedureOverride: "/user.v1.UserService/GetProfile",
		}

		_, err = wrappedFunc(context.Background(), req)

		if err != nil {
			assert.Equal(t, ErrInvalidClaims, err, "GetSubject() returned an error for empty subject")
		} else {
			assert.Equal(t, "", capturedUserID, "Empty subject was accepted and stored in context")
		}
	})
}

func TestContextKeys(t *testing.T) {
	t.Run("UserIDKey is unique", func(t *testing.T) {
		assert.Equal(t, contextKey("user_id"), UserIDKey)
	})

	t.Run("UserEmailKey is unique", func(t *testing.T) {
		assert.Equal(t, contextKey("user_email"), UserEmailKey)
	})

	t.Run("keys are different", func(t *testing.T) {
		assert.NotEqual(t, UserIDKey, UserEmailKey)
	})
}

func TestErrorTypes(t *testing.T) {
	t.Run("ErrMissingToken is unauthenticated", func(t *testing.T) {
		var connectErr *connect.Error
		require.ErrorAs(t, ErrMissingToken, &connectErr)
		assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})

	t.Run("ErrInvalidToken is unauthenticated", func(t *testing.T) {
		var connectErr *connect.Error
		require.ErrorAs(t, ErrInvalidToken, &connectErr)
		assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})

	t.Run("ErrTokenExpired is unauthenticated", func(t *testing.T) {
		var connectErr *connect.Error
		require.ErrorAs(t, ErrTokenExpired, &connectErr)
		assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})

	t.Run("ErrInvalidClaims is unauthenticated", func(t *testing.T) {
		var connectErr *connect.Error
		require.ErrorAs(t, ErrInvalidClaims, &connectErr)
		assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})
}

func TestMustGetUserEmail(t *testing.T) {
	t.Run("returns email when present", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserEmailKey, "test@example.com")
		email, err := MustGetUserEmail(ctx)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", email)
	})

	t.Run("returns error when not present", func(t *testing.T) {
		ctx := context.Background()
		_, err := MustGetUserEmail(ctx)
		require.Error(t, err)

		var connectErr *connect.Error
		require.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})

	t.Run("returns error for wrong type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserEmailKey, 123)
		_, err := MustGetUserEmail(ctx)
		require.Error(t, err)

		var connectErr *connect.Error
		require.ErrorAs(t, err, &connectErr)
		assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})
}

func TestAuthInterceptor_WrapStreamingHandler(t *testing.T) {
	interceptor := NewAuthInterceptor(testSecret)

	t.Run("public method bypasses auth", func(t *testing.T) {
		called := false
		nextFunc := func(ctx context.Context, conn connect.StreamingHandlerConn) error {
			called = true
			return nil
		}

		wrappedFunc := interceptor.WrapStreamingHandler(nextFunc)
		require.NotNil(t, wrappedFunc)

		mockConn := &mockStreamingHandlerConn{
			procedure: "/user.v1.UserService/Register",
		}

		err := wrappedFunc(context.Background(), mockConn)
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("valid token extracts user info", func(t *testing.T) {
		userID := "user-123"
		email := "test@example.com"
		token := generateTestToken(t, userID, email, time.Now().Add(time.Hour))

		var capturedUserID string
		var capturedEmail string

		nextFunc := func(ctx context.Context, conn connect.StreamingHandlerConn) error {
			capturedUserID, _ = GetUserID(ctx)
			capturedEmail, _ = GetUserEmail(ctx)
			return nil
		}

		wrappedFunc := interceptor.WrapStreamingHandler(nextFunc)
		mockConn := &mockStreamingHandlerConn{
			procedure:      "/user.v1.UserService/GetProfile",
			authHeader:     "Bearer " + token,
			requestHeaders: make(map[string][]string),
		}
		mockConn.requestHeaders["Authorization"] = []string{"Bearer " + token}

		err := wrappedFunc(context.Background(), mockConn)
		require.NoError(t, err)
		assert.Equal(t, userID, capturedUserID)
		assert.Equal(t, email, capturedEmail)
	})

	t.Run("missing token returns error", func(t *testing.T) {
		nextFunc := func(ctx context.Context, conn connect.StreamingHandlerConn) error {
			return nil
		}

		wrappedFunc := interceptor.WrapStreamingHandler(nextFunc)
		mockConn := &mockStreamingHandlerConn{
			procedure:      "/user.v1.UserService/GetProfile",
			requestHeaders: make(map[string][]string),
		}

		err := wrappedFunc(context.Background(), mockConn)
		require.Error(t, err)
		assert.Equal(t, ErrMissingToken, err)
	})

	t.Run("expired token returns error", func(t *testing.T) {
		token := generateTestToken(t, "user-123", "test@example.com", time.Now().Add(-time.Hour))

		nextFunc := func(ctx context.Context, conn connect.StreamingHandlerConn) error {
			return nil
		}

		wrappedFunc := interceptor.WrapStreamingHandler(nextFunc)
		mockConn := &mockStreamingHandlerConn{
			procedure:      "/user.v1.UserService/GetProfile",
			authHeader:     "Bearer " + token,
			requestHeaders: make(map[string][]string),
		}
		mockConn.requestHeaders["Authorization"] = []string{"Bearer " + token}

		err := wrappedFunc(context.Background(), mockConn)
		require.Error(t, err)
		assert.Equal(t, ErrTokenExpired, err)
	})

	t.Run("invalid claims returns error", func(t *testing.T) {
		claims := &JwtClaims{
			Email: "test@example.com",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user-123",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString([]byte("wrong-secret"))
		require.NoError(t, err)

		nextFunc := func(ctx context.Context, conn connect.StreamingHandlerConn) error {
			return nil
		}

		wrappedFunc := interceptor.WrapStreamingHandler(nextFunc)
		mockConn := &mockStreamingHandlerConn{
			procedure:      "/user.v1.UserService/GetProfile",
			authHeader:     "Bearer " + signedToken,
			requestHeaders: make(map[string][]string),
		}
		mockConn.requestHeaders["Authorization"] = []string{"Bearer " + signedToken}

		err = wrappedFunc(context.Background(), mockConn)
		require.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
	})
}

type mockStreamingHandlerConn struct {
	procedure      string
	authHeader     string
	requestHeaders map[string][]string
}

func (m *mockStreamingHandlerConn) Spec() connect.Spec {
	return connect.Spec{Procedure: m.procedure}
}

func (m *mockStreamingHandlerConn) RequestHeader() http.Header {
	if m.requestHeaders == nil {
		m.requestHeaders = make(map[string][]string)
	}
	if m.authHeader != "" {
		m.requestHeaders["Authorization"] = []string{m.authHeader}
	}
	return http.Header(m.requestHeaders)
}

func (m *mockStreamingHandlerConn) Send(msg any) error {
	return nil
}

func (m *mockStreamingHandlerConn) Receive(msg any) error {
	return nil
}

func (m *mockStreamingHandlerConn) Close() error {
	return nil
}

func (m *mockStreamingHandlerConn) Peer() connect.Peer {
	return connect.Peer{}
}

func (m *mockStreamingHandlerConn) ResponseHeader() http.Header {
	return make(http.Header)
}

func (m *mockStreamingHandlerConn) ResponseTrailer() http.Header {
	return make(http.Header)
}
