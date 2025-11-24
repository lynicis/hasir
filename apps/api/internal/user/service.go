package user

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"apps/api/pkg/config"

	userv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/user/v1"
)

type Service interface {
	Register(ctx context.Context, req *userv1.RegisterRequest) error
	Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.TokenEnvelope, error)
}

type service struct {
	config         *config.Config
	userRepository Repository
}

func NewService(config *config.Config, userRepository Repository) Service {
	return &service{
		config:         config,
		userRepository: userRepository,
	}
}

func (s *service) Register(ctx context.Context, req *userv1.RegisterRequest) error {
	isUserExists, err := s.userRepository.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, ErrNoRows) {
		return err
	}

	if isUserExists != nil {
		return connect.NewError(connect.CodeAlreadyExists, errors.New("user already exists"))
	}

	var hashedPassword []byte
	hashedPassword, err = bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return connect.NewError(connect.CodeInternal, errors.New("something went wrong"))
	}

	user := &UserDTO{
		Id:        uuid.NewString(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now().UTC(),
	}

	err = s.userRepository.CreateUser(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.TokenEnvelope, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid credentials"))
	}

	now := time.Now().UTC()

	accessTokenExpiresAt := now.Add(2 * time.Hour).Unix()
	accessTokenClaims := JwtClaims{
		Claims: jwt.MapClaims{
			"iss": s.config.Server.PublicUrl,
			"sub": user.Id,
			"exp": accessTokenExpiresAt,
			"iat": now.Unix(),
			"aud": s.config.DashboardUrl,
		},
		email:    user.Email,
		username: user.Username,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)

	var signedAccessToken string
	signedAccessToken, err = accessToken.SignedString(s.config.JwtSecret)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to create access token"))
	}

	refreshTokenExpiresAt := now.AddDate(0, 0, 7).Unix()
	refreshTokenClaims := JwtClaims{
		Claims: jwt.MapClaims{
			"iss": s.config.Server.PublicUrl,
			"sub": user.Id,
			"exp": refreshTokenExpiresAt,
			"iat": now.Unix(),
			"aud": s.config.DashboardUrl,
		},
		email:    user.Email,
		username: user.Username,
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	var signedRefreshToken string
	signedRefreshToken, err = refreshToken.SignedString(s.config.JwtSecret)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to create refresh token"))
	}

	expiresAt := time.Now().UTC().AddDate(0, 0, 14)
	if err = s.userRepository.CreateRefreshToken(
		ctx,
		user.Id,
		signedRefreshToken,
		expiresAt,
	); err != nil {
		return nil, err
	}

	return &userv1.TokenEnvelope{
		AccessToken:  signedAccessToken,
		RefreshToken: signedRefreshToken,
	}, nil
}
