package auth

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
)

var (
	ErrMissingToken  = connect.NewError(connect.CodeUnauthenticated, errors.New("missing authorization token"))
	ErrInvalidToken  = connect.NewError(connect.CodeUnauthenticated, errors.New("invalid authorization token"))
	ErrTokenExpired  = connect.NewError(connect.CodeUnauthenticated, errors.New("token has expired"))
	ErrInvalidClaims = connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token claims"))
)

type AuthInterceptor struct {
	jwtSecret []byte
}

func NewAuthInterceptor(jwtSecret []byte) *AuthInterceptor {
	return &AuthInterceptor{
		jwtSecret: jwtSecret,
	}
}

func (a *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		authHeader := req.Header().Get("Authorization")
		if authHeader == "" {
			return nil, ErrMissingToken
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return nil, ErrInvalidToken
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return a.jwtSecret, nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				return nil, ErrTokenExpired
			}
			return nil, ErrInvalidToken
		}

		if !token.Valid {
			return nil, ErrInvalidToken
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, ErrInvalidClaims
		}

		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			return nil, ErrInvalidClaims
		}

		ctx = context.WithValue(ctx, UserIDKey, userID)

		if email, ok := claims["email"].(string); ok {
			ctx = context.WithValue(ctx, UserEmailKey, email)
		}

		return next(ctx, req)
	}
}

func (a *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (a *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		authHeader := conn.RequestHeader().Get("Authorization")
		if authHeader == "" {
			return ErrMissingToken
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return ErrInvalidToken
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return a.jwtSecret, nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				return ErrTokenExpired
			}
			return ErrInvalidToken
		}

		if !token.Valid {
			return ErrInvalidToken
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return ErrInvalidClaims
		}

		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			return ErrInvalidClaims
		}

		ctx = context.WithValue(ctx, UserIDKey, userID)

		if email, ok := claims["email"].(string); ok {
			ctx = context.WithValue(ctx, UserEmailKey, email)
		}

		return next(ctx, conn)
	}
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

func MustGetUserID(ctx context.Context) (string, error) {
	userID, ok := GetUserID(ctx)
	if !ok {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("user not authenticated"))
	}
	return userID, nil
}
