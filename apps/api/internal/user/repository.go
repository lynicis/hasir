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
	CreateApiKey(ctx context.Context, userId, apiKey string) error
	GetApiKeys(ctx context.Context, userId string, page, pageSize int) (*[]ApiKeyDTO, error)
	GetApiKeysCount(ctx context.Context, userId string) (int, error)
	RevokeApiKey(ctx context.Context, userId, keyId string) error
	CreateSshKey(ctx context.Context, userId, publicKey string) error
	GetSshKeys(ctx context.Context, userId string, page, pageSize int) (*[]SshKeyDTO, error)
	GetSshKeysCount(ctx context.Context, userId string) (int, error)
	RevokeSshKey(ctx context.Context, userId, keyId string) error
}
