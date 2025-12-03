package auth

import (
	"context"
	"errors"
	"strings"

	"buf.build/gen/go/hasir/hasir/connectrpc/go/organization/v1/organizationv1connect"
	"buf.build/gen/go/hasir/hasir/connectrpc/go/user/v1/userv1connect"
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
	jwtSecret     []byte
	publicMethods map[string]bool
}

func NewAuthInterceptor(jwtSecret []byte) *AuthInterceptor {
	return &AuthInterceptor{
		jwtSecret: jwtSecret,
		publicMethods: map[string]bool{
			userv1connect.UserServiceRegisterProcedure:                          true,
			userv1connect.UserServiceLoginProcedure:                             true,
			userv1connect.UserServiceRenewTokensProcedure:                       true,
			organizationv1connect.OrganizationServiceIsInvitationValidProcedure: true,
		},
	}
}

func (a *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if a.publicMethods[req.Spec().Procedure] {
			return next(ctx, req)
		}

		authHeader := req.Header().Get("Authorization")
		if authHeader == "" {
			return nil, ErrMissingToken
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return nil, ErrInvalidToken
		}

		token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (any, error) {
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

		claims, ok := token.Claims.(*JwtClaims)
		if !ok {
			return nil, ErrInvalidClaims
		}

		userID, err := claims.GetSubject()
		if err != nil {
			return nil, ErrInvalidClaims
		}

		ctx = context.WithValue(ctx, UserIDKey, userID)

		email := claims.Email
		ctx = context.WithValue(ctx, UserEmailKey, email)

		return next(ctx, req)
	}
}

func (a *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (a *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if a.publicMethods[conn.Spec().Procedure] {
			return next(ctx, conn)
		}

		authHeader := conn.RequestHeader().Get("Authorization")
		if authHeader == "" {
			return ErrMissingToken
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return ErrInvalidToken
		}

		token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
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

		claims, ok := token.Claims.(*JwtClaims)
		if !ok {
			return ErrInvalidClaims
		}

		userID, err := claims.GetSubject()
		if err != nil {
			return ErrInvalidClaims
		}

		ctx = context.WithValue(ctx, UserIDKey, userID)

		email := claims.Email
		ctx = context.WithValue(ctx, UserEmailKey, email)

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
