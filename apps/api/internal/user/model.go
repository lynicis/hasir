package user

import "time"

type UserDTO struct {
	Id        string     `json:"id" db:"id"`
	Username  string     `json:"username" db:"username"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"password" db:"password"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type RefreshTokensDTO struct {
	UserId    string    `json:"id" db:"id"`
	Jti       string    `json:"jti" db:"jti"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ApiKeyDTO struct {
	Id        string     `json:"id" db:"id"`
	UserId    string     `json:"user_id" db:"user_id"`
	Name      string     `json:"name" db:"name"`
	Key       string     `json:"key" db:"key"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type SshKeyDTO struct {
	Id        string     `json:"id" db:"id"`
	UserId    string     `json:"user_id" db:"user_id"`
	Name      string     `json:"name" db:"name"`
	PublicKey string     `json:"public_key" db:"public_key"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
