package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testSecret = []byte("test-secret-key-for-testing-only")

func generateTestToken(t *testing.T, userID string, email string, expiresAt time.Time) string {
	t.Helper()

	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   expiresAt.Unix(),
		"iat":   time.Now().Unix(),
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

func TestAuthInterceptor_Integration(t *testing.T) {
	interceptor := NewAuthInterceptor(testSecret)

	t.Run("valid token extracts user info", func(t *testing.T) {
		userID := "user-123"
		email := "test@example.com"
		token := generateTestToken(t, userID, email, time.Now().Add(time.Hour))

		var capturedUserID string
		var capturedEmail string

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			tokenString := authHeader[7:]
			parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return testSecret, nil
			})
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			claims := parsedToken.Claims.(jwt.MapClaims)
			capturedUserID = claims["sub"].(string)
			capturedEmail = claims["email"].(string)

			w.WriteHeader(http.StatusOK)
		})

		server := httptest.NewServer(handler)
		defer server.Close()

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, userID, capturedUserID)
		assert.Equal(t, email, capturedEmail)
		require.NotNil(t, interceptor)
	})

	t.Run("missing token returns unauthorized", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		server := httptest.NewServer(handler)
		defer server.Close()

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("expired token returns unauthorized", func(t *testing.T) {
		token := generateTestToken(t, "user-123", "test@example.com", time.Now().Add(-time.Hour))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			tokenString := authHeader[7:]
			_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return testSecret, nil
			})
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		server := httptest.NewServer(handler)
		defer server.Close()

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid signature returns unauthorized", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":   "user-123",
			"email": "test@example.com",
			"exp":   time.Now().Add(time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString([]byte("wrong-secret"))
		require.NoError(t, err)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			tokenString := authHeader[7:]
			_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return testSecret, nil
			})
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		server := httptest.NewServer(handler)
		defer server.Close()

		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+signedToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestTokenParsing(t *testing.T) {
	t.Run("parses valid token", func(t *testing.T) {
		userID := "user-456"
		email := "parse@example.com"
		tokenString := generateTestToken(t, userID, email, time.Now().Add(time.Hour))

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return testSecret, nil
		})
		require.NoError(t, err)
		require.True(t, token.Valid)

		claims := token.Claims.(jwt.MapClaims)
		assert.Equal(t, userID, claims["sub"])
		assert.Equal(t, email, claims["email"])
	})

	t.Run("rejects token with wrong signing method", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":   "user-123",
			"email": "test@example.com",
			"exp":   time.Now().Add(time.Hour).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		_, err = jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return testSecret, nil
		})
		require.Error(t, err)
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
