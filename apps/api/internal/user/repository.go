package user

import (
	"context"
	"time"
)

type Repository interface {
	CreateUser(ctx context.Context, user *UserDTO) error
	GetUserByEmail(ctx context.Context, email string) (*UserDTO, error)
	GetUserById(ctx context.Context, id string) (*UserDTO, error)
	CreateRefreshToken(ctx context.Context, id, token string, expiresAt time.Time) error
	GetRefreshTokenByTokenId(ctx context.Context, token string) (*RefreshTokensDTO, error)
	DeleteRefreshToken(ctx context.Context, userId, token string) error
	UpdateUserById(ctx context.Context, id string, user *UserDTO) error
	DeleteUser(ctx context.Context, userId string) error
}
