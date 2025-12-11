package user

import (
	"context"
	"errors"
	"time"

	userv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/user/v1"
	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"hasir-api/pkg/authentication"
	"hasir-api/pkg/config"
)

var ErrInternalServer = connect.NewError(connect.CodeInternal, errors.New("something went wrong"))

type Service interface {
	Register(ctx context.Context, req *userv1.RegisterRequest) error
	Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.TokenEnvelope, error)
	UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.TokenEnvelope, error)
	RenewTokens(ctx context.Context, req *userv1.RenewTokensRequest) (*userv1.RenewTokensResponse, error)
	ForgotPassword(ctx context.Context, req *userv1.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *userv1.ResetPasswordRequest) error
}

type service struct {
	config         *config.Config
	userRepository Repository
	emailService   EmailService
}

type EmailService interface {
	SendForgotPassword(to, resetToken string) error
}

func NewService(config *config.Config, userRepository Repository, emailService EmailService) *service {
	return &service{
		config:         config,
		userRepository: userRepository,
		emailService:   emailService,
	}
}

func (s *service) Register(ctx context.Context, req *userv1.RegisterRequest) error {
	isUserExists, err := s.userRepository.GetUserByEmail(ctx, req.Email)
	if err != nil {
		var connectErr *connect.Error
		if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeNotFound {
			return err
		}
	}

	if isUserExists != nil {
		return connect.NewError(connect.CodeAlreadyExists, errors.New("user already exists"))
	}

	var hashedPassword []byte
	hashedPassword, err = bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return ErrInternalServer
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

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid credentials"))
	}

	tokens, refreshTokenID, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().AddDate(0, 0, 14)
	if err = s.userRepository.CreateRefreshToken(
		ctx,
		user.Id,
		refreshTokenID,
		expiresAt,
	); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *service) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.TokenEnvelope, error) {
	userID, err := authentication.MustGetUserID(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepository.GetUserById(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid credentials"))
	}

	updatedUser := &UserDTO{
		Id:       user.Id,
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Password: user.Password,
	}

	if req.GetNewPassword() != "" {
		var hashedNewPassword []byte
		hashedNewPassword, err = bcrypt.GenerateFromPassword([]byte(req.GetNewPassword()), bcrypt.DefaultCost)
		if err != nil {
			return nil, ErrInternalServer
		}
		updatedUser.Password = string(hashedNewPassword)
	}

	if err = s.userRepository.UpdateUserById(ctx, updatedUser.Id, updatedUser); err != nil {
		return nil, err
	}

	tokens, refreshTokenID, err := s.generateTokens(updatedUser)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().AddDate(0, 0, 14)
	if err = s.userRepository.CreateRefreshToken(
		ctx,
		updatedUser.Id,
		refreshTokenID,
		expiresAt,
	); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *service) RenewTokens(ctx context.Context, req *userv1.RenewTokensRequest) (*userv1.RenewTokensResponse, error) {
	refreshTokenString := req.GetRefreshToken()

	token, err := jwt.ParseWithClaims(refreshTokenString, &authentication.JwtClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return s.config.JwtSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("refresh token has expired"))
		}

		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid refresh token"))
	}

	if !token.Valid {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid refresh token"))
	}

	claims, ok := token.Claims.(*authentication.JwtClaims)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid refresh token claims"))
	}

	userID, err := claims.GetSubject()
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid refresh token subject"))
	}

	if claims.ID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid refresh token id"))
	}

	refreshTokenRecord, err := s.userRepository.GetRefreshTokenByTokenId(ctx, claims.ID)
	if err != nil {
		var connectErr *connect.Error
		if errors.As(err, &connectErr) && connectErr.Code() == connect.CodeNotFound {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid refresh token"))
		}

		return nil, err
	}

	now := time.Now().UTC()
	if now.After(refreshTokenRecord.ExpiresAt) {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("refresh token has expired"))
	}

	if refreshTokenRecord.UserId != userID {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid refresh token"))
	}

	user, err := s.userRepository.GetUserById(ctx, userID)
	if err != nil {
		return nil, err
	}

	tokens, _, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	return &userv1.RenewTokensResponse{
		AccessToken: tokens.AccessToken,
	}, nil
}

func (s *service) generateTokens(user *UserDTO) (*userv1.TokenEnvelope, string, error) {
	now := time.Now().UTC()
	accessTokenExpiresAt := now.Add(2 * time.Hour)
	accessTokenClaims := authentication.JwtClaims{
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    s.config.Server.PublicUrl,
			Subject:   user.Id,
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Audience:  jwt.ClaimStrings{s.config.DashboardUrl},
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)

	signedAccessToken, err := accessToken.SignedString(s.config.JwtSecret)
	if err != nil {
		return nil, "", connect.NewError(connect.CodeInternal, errors.New("failed to create access token"))
	}

	refreshTokenExpiresAt := now.AddDate(0, 0, 7)
	refreshTokenClaims := authentication.JwtClaims{
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    s.config.Server.PublicUrl,
			Subject:   user.Id,
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Audience:  jwt.ClaimStrings{s.config.DashboardUrl},
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	var signedRefreshToken string
	signedRefreshToken, err = refreshToken.SignedString(s.config.JwtSecret)
	if err != nil {
		return nil, "", connect.NewError(connect.CodeInternal, errors.New("failed to create refresh token"))
	}

	return &userv1.TokenEnvelope{
		AccessToken:  signedAccessToken,
		RefreshToken: signedRefreshToken,
	}, refreshTokenClaims.ID, nil
}

func (s *service) ForgotPassword(ctx context.Context, req *userv1.ForgotPasswordRequest) error {
	user, err := s.userRepository.GetUserByEmail(ctx, req.Email)
	if err != nil {
		var connectErr *connect.Error
		if errors.As(err, &connectErr) && connectErr.Code() == connect.CodeNotFound {
			return nil
		}
		return err
	}

	resetToken := uuid.NewString()
	expiresAt := time.Now().UTC().Add(1 * time.Hour)

	if err := s.userRepository.CreatePasswordResetToken(ctx, user.Id, resetToken, expiresAt); err != nil {
		return err
	}

	if err := s.emailService.SendForgotPassword(user.Email, resetToken); err != nil {
		return connect.NewError(connect.CodeInternal, errors.New("failed to send reset email"))
	}

	return nil
}

func (s *service) ResetPassword(ctx context.Context, req *userv1.ResetPasswordRequest) error {
	resetToken, err := s.userRepository.GetPasswordResetToken(ctx, req.Token)
	if err != nil {
		return err
	}

	if resetToken.UsedAt != nil {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("reset token has already been used"))
	}

	now := time.Now().UTC()
	if now.After(resetToken.ExpiresAt) {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("reset token has expired"))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return ErrInternalServer
	}

	user, err := s.userRepository.GetUserById(ctx, resetToken.UserId)
	if err != nil {
		return err
	}

	updatedUser := &UserDTO{
		Id:       user.Id,
		Username: user.Username,
		Email:    user.Email,
		Password: string(hashedPassword),
	}

	if err := s.userRepository.UpdateUserById(ctx, user.Id, updatedUser); err != nil {
		return err
	}

	if err := s.userRepository.MarkPasswordResetTokenAsUsed(ctx, req.Token); err != nil {
		return err
	}

	return nil
}
