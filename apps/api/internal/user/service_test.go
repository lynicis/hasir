package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"hasir-api/pkg/auth"
	"hasir-api/pkg/config"

	userv1 "buf.build/gen/go/hasir/hasir/protocolbuffers/go/user/v1"
)

func TestNewService(t *testing.T) {
	s := NewService(nil, nil)
	assert.Implements(t, (*Service)(nil), s)
}

func TestService_Register(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	t.Run("happy path", func(t *testing.T) {
		mockUserRepository := NewMockRepository(mockController)
		mockUserRepository.
			EXPECT().
			GetUserByEmail(gomock.Any(), gomock.Any()).
			Return(nil, ErrNoRows).
			Times(1)
		mockUserRepository.
			EXPECT().
			CreateUser(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		s := NewService(nil, mockUserRepository)
		err := s.Register(t.Context(), &userv1.RegisterRequest{
			Email:    "test@mail.com",
			Username: "test-user",
			Password: "Asdfg12345_",
		})

		assert.NoError(t, err)
	})

	t.Run("error occurred while finding user", func(t *testing.T) {
		mockUserRepository := NewMockRepository(mockController)
		mockUserRepository.
			EXPECT().
			GetUserByEmail(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("something went wrong")).
			Times(1)

		s := NewService(nil, mockUserRepository)
		err := s.Register(t.Context(), &userv1.RegisterRequest{
			Email:    "test@mail.com",
			Username: "test-user",
			Password: "Asdfg12345_",
		})

		assert.Errorf(t, err, "something went wrong")
	})

	t.Run("email already exists", func(t *testing.T) {
		mockUserRepository := NewMockRepository(mockController)
		mockUserRepository.
			EXPECT().
			GetUserByEmail(gomock.Any(), gomock.Any()).
			Return(&UserDTO{
				Id:        uuid.NewString(),
				Username:  "test-user",
				Email:     "test@mail.com",
				Password:  "Asdfg12345_",
				CreatedAt: time.Now().UTC(),
			}, nil).
			Times(1)

		s := NewService(nil, mockUserRepository)
		err := s.Register(t.Context(), &userv1.RegisterRequest{
			Email:    "test@mail.com",
			Username: "test-user",
			Password: "Asdfg12345_",
		})

		assert.Errorf(t, err, "user already exists")
	})

	t.Run("error occured while creating user", func(t *testing.T) {
		mockUserRepository := NewMockRepository(mockController)
		mockUserRepository.
			EXPECT().
			GetUserByEmail(gomock.Any(), gomock.Any()).
			Return(nil, ErrNoRows).
			Times(1)
		mockUserRepository.
			EXPECT().
			CreateUser(gomock.Any(), gomock.Any()).
			Return(errors.New("something went wrong")).
			Times(1)

		s := NewService(nil, mockUserRepository)
		err := s.Register(t.Context(), &userv1.RegisterRequest{
			Email:    "test@mail.com",
			Username: "test-user",
			Password: "Asdfg12345_",
		})

		assert.Errorf(t, err, "something went wrong")
	})
}

func TestService_Login(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	cfg := &config.Config{
		Server: config.ServerConfig{
			PublicUrl: "http://api.test.com",
		},
		DashboardUrl: "http://test.com/dashboard",
		JwtSecret:    []byte("jwt-secret"),
	}

	t.Run("happy path", func(t *testing.T) {
		hashedPwd, err := bcrypt.GenerateFromPassword([]byte("Asdfg12345_"), bcrypt.DefaultCost)
		require.NoError(t, err)

		mockUserRepository := NewMockRepository(mockController)
		mockUserRepository.
			EXPECT().
			GetUserByEmail(gomock.Any(), gomock.Any()).
			Return(&UserDTO{
				Id:        uuid.NewString(),
				Username:  "test-user",
				Email:     "test@mail.com",
				Password:  string(hashedPwd),
				CreatedAt: time.Now().UTC(),
			}, nil).
			Times(1)
		mockUserRepository.
			EXPECT().
			CreateRefreshToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		s := NewService(cfg, mockUserRepository)
		tokens, err := s.Login(t.Context(), &userv1.LoginRequest{
			Email:    "test@mail.com",
			Password: "Asdfg12345_",
		})

		assert.NoError(t, err)
		assert.NotNil(t, tokens.AccessToken)
		assert.NotNil(t, tokens.RefreshToken)
	})

	t.Run("wrong password", func(t *testing.T) {
		hashedPwd, err := bcrypt.GenerateFromPassword([]byte("Asdfg12345_"), bcrypt.DefaultCost)
		require.NoError(t, err)

		mockUserRepository := NewMockRepository(mockController)
		mockUserRepository.
			EXPECT().
			GetUserByEmail(gomock.Any(), gomock.Any()).
			Return(&UserDTO{
				Id:        uuid.NewString(),
				Username:  "test-user",
				Email:     "test@mail.com",
				Password:  string(hashedPwd),
				CreatedAt: time.Now().UTC(),
			}, nil).
			Times(1)

		s := NewService(cfg, mockUserRepository)
		tokens, err := s.Login(t.Context(), &userv1.LoginRequest{
			Email:    "test@mail.com",
			Password: "wrong-password123_",
		})

		assert.Errorf(t, err, "unauthenticated: invalid credentials")
		assert.Nil(t, tokens)
	})

	t.Run("error occurred while creating new refresh token", func(t *testing.T) {
		hashedPwd, err := bcrypt.GenerateFromPassword([]byte("Asdfg12345_"), bcrypt.DefaultCost)
		require.NoError(t, err)

		mockUserRepository := NewMockRepository(mockController)
		mockUserRepository.
			EXPECT().
			GetUserByEmail(gomock.Any(), gomock.Any()).
			Return(&UserDTO{
				Id:        uuid.NewString(),
				Username:  "test-user",
				Email:     "test@mail.com",
				Password:  string(hashedPwd),
				CreatedAt: time.Now().UTC(),
			}, nil).
			Times(1)
		mockUserRepository.
			EXPECT().
			CreateRefreshToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("something went wrong")).
			Times(1)

		s := NewService(cfg, mockUserRepository)
		tokens, err := s.Login(t.Context(), &userv1.LoginRequest{
			Email:    "test@mail.com",
			Password: "Asdfg12345_",
		})

		assert.Errorf(t, err, "something went wrong")
		assert.Nil(t, tokens)
	})
}

func TestService_UpdateUser(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	cfg := &config.Config{
		Server: config.ServerConfig{
			PublicUrl: "http://api.test.com",
		},
		DashboardUrl: "http://test.com/dashboard",
		JwtSecret:    []byte("jwt-secret"),
	}

	userId := uuid.NewString()
	oldPassword := "Asdfg12345_"
	newPassword := "NewPassword123_"
	newUsername := "new-username"
	newEmail := "new@mail.com"
	hashedOldPwd, err := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	t.Run("happy path", func(t *testing.T) {
		mockUserRepository := NewMockRepository(mockController)
		mockUserRepository.
			EXPECT().
			GetUserById(gomock.Any(), userId).
			Return(&UserDTO{
				Id:        userId,
				Username:  "old-username",
				Email:     "old@mail.com",
				Password:  string(hashedOldPwd),
				CreatedAt: time.Now().UTC(),
			}, nil).
			Times(1)
		mockUserRepository.
			EXPECT().
			UpdateUserById(gomock.Any(), userId, gomock.Any()).
			Return(nil).
			Times(1)
		mockUserRepository.
			EXPECT().
			CreateRefreshToken(gomock.Any(), userId, gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		s := NewService(cfg, mockUserRepository)
		ctx := context.WithValue(t.Context(), auth.UserIDKey, userId)
		tokens, err := s.UpdateUser(ctx, &userv1.UpdateUserRequest{
			Username:    &newUsername,
			Email:       &newEmail,
			Password:    oldPassword,
			NewPassword: &newPassword,
		})

		assert.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
	})

	t.Run("user not found", func(t *testing.T) {
		mockUserRepository := NewMockRepository(mockController)
		mockUserRepository.
			EXPECT().
			GetUserById(gomock.Any(), userId).
			Return(nil, ErrNoRows).
			Times(1)

		s := NewService(cfg, mockUserRepository)
		ctx := context.WithValue(t.Context(), auth.UserIDKey, userId)
		tokens, err := s.UpdateUser(ctx, &userv1.UpdateUserRequest{
			Username:    &newUsername,
			Email:       &newEmail,
			Password:    oldPassword,
			NewPassword: &newPassword,
		})

		assert.Error(t, err)
		assert.Nil(t, tokens)
	})

	t.Run("wrong password", func(t *testing.T) {
		mockUserRepository := NewMockRepository(mockController)
		mockUserRepository.
			EXPECT().
			GetUserById(gomock.Any(), userId).
			Return(&UserDTO{
				Id:        userId,
				Username:  "old-username",
				Email:     "old@mail.com",
				Password:  string(hashedOldPwd),
				CreatedAt: time.Now().UTC(),
			}, nil).
			Times(1)

		s := NewService(cfg, mockUserRepository)
		wrongPassword := "wrong-password"
		ctx := context.WithValue(t.Context(), auth.UserIDKey, userId)
		tokens, err := s.UpdateUser(ctx, &userv1.UpdateUserRequest{
			Username:    &newUsername,
			Email:       &newEmail,
			Password:    wrongPassword,
			NewPassword: &newPassword,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
		assert.Nil(t, tokens)
	})

	t.Run("error occurred while updating user", func(t *testing.T) {
		mockUserRepository := NewMockRepository(mockController)
		mockUserRepository.
			EXPECT().
			GetUserById(gomock.Any(), userId).
			Return(&UserDTO{
				Id:        userId,
				Username:  "old-username",
				Email:     "old@mail.com",
				Password:  string(hashedOldPwd),
				CreatedAt: time.Now().UTC(),
			}, nil).
			Times(1)
		mockUserRepository.
			EXPECT().
			UpdateUserById(gomock.Any(), userId, gomock.Any()).
			Return(errors.New("something went wrong")).
			Times(1)

		s := NewService(cfg, mockUserRepository)
		ctx := context.WithValue(t.Context(), auth.UserIDKey, userId)
		tokens, err := s.UpdateUser(ctx, &userv1.UpdateUserRequest{
			Username:    &newUsername,
			Email:       &newEmail,
			Password:    oldPassword,
			NewPassword: &newPassword,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "something went wrong")
		assert.Nil(t, tokens)
	})

	t.Run("error occurred while creating refresh token", func(t *testing.T) {
		mockUserRepository := NewMockRepository(mockController)
		mockUserRepository.
			EXPECT().
			GetUserById(gomock.Any(), userId).
			Return(&UserDTO{
				Id:        userId,
				Username:  "old-username",
				Email:     "old@mail.com",
				Password:  string(hashedOldPwd),
				CreatedAt: time.Now().UTC(),
			}, nil).
			Times(1)
		mockUserRepository.
			EXPECT().
			UpdateUserById(gomock.Any(), userId, gomock.Any()).
			Return(nil).
			Times(1)
		mockUserRepository.
			EXPECT().
			CreateRefreshToken(gomock.Any(), userId, gomock.Any(), gomock.Any()).
			Return(errors.New("something went wrong")).
			Times(1)

		s := NewService(cfg, mockUserRepository)
		ctx := context.WithValue(t.Context(), auth.UserIDKey, userId)
		tokens, err := s.UpdateUser(ctx, &userv1.UpdateUserRequest{
			Username:    &newUsername,
			Email:       &newEmail,
			Password:    oldPassword,
			NewPassword: &newPassword,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "something went wrong")
		assert.Nil(t, tokens)
	})
}

func TestService_RenewTokens(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	cfg := &config.Config{
		Server: config.ServerConfig{
			PublicUrl: "http://api.test.com",
		},
		DashboardUrl: "http://test.com/dashboard",
		JwtSecret:    []byte("jwt-secret"),
	}

	t.Run("happy path", func(t *testing.T) {
		userId := uuid.NewString()
		refreshTokenID := uuid.NewString()
		now := time.Now().UTC()

		claims := auth.JwtClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        refreshTokenID,
				Issuer:    cfg.Server.PublicUrl,
				Subject:   userId,
				ExpiresAt: jwt.NewNumericDate(now.AddDate(0, 0, 7)),
				IssuedAt:  jwt.NewNumericDate(now),
				Audience:  jwt.ClaimStrings{cfg.DashboardUrl},
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(cfg.JwtSecret)
		require.NoError(t, err)

		mockUserRepository := NewMockRepository(mockController)
		mockUserRepository.
			EXPECT().
			GetRefreshTokenByTokenId(gomock.Any(), refreshTokenID).
			Return(&RefreshTokensDTO{
				UserId:    userId,
				Jti:       refreshTokenID,
				ExpiresAt: now.AddDate(0, 0, 14),
				CreatedAt: now,
			}, nil).
			Times(1)
		mockUserRepository.
			EXPECT().
			GetUserById(gomock.Any(), userId).
			Return(&UserDTO{
				Id:        userId,
				Username:  "test-user",
				Email:     "test@mail.com",
				Password:  "hashed-password",
				CreatedAt: now,
			}, nil).
			Times(1)

		s := NewService(cfg, mockUserRepository)
		resp, err := s.RenewTokens(t.Context(), &userv1.RenewTokensRequest{
			RefreshToken: signedToken,
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.AccessToken)
	})

	t.Run("missing token id", func(t *testing.T) {
		userId := uuid.NewString()
		now := time.Now().UTC()

		claims := auth.JwtClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    cfg.Server.PublicUrl,
				Subject:   userId,
				ExpiresAt: jwt.NewNumericDate(now.AddDate(0, 0, 7)),
				IssuedAt:  jwt.NewNumericDate(now),
				Audience:  jwt.ClaimStrings{cfg.DashboardUrl},
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(cfg.JwtSecret)
		require.NoError(t, err)

		mockUserRepository := NewMockRepository(mockController)

		s := NewService(cfg, mockUserRepository)
		resp, err := s.RenewTokens(t.Context(), &userv1.RenewTokensRequest{
			RefreshToken: signedToken,
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid refresh token id")
	})
}
