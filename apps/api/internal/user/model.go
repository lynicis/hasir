package user

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserDTO struct {
	Id        string     `json:"id" db:"id"`
	Username  string     `json:"username" db:"username"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"password" db:"password"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at,omitempty"`
}

type UserUpdateHistoryDTO struct {
	UserId    string    `json:"userId" db:"userId"`
	Fields    []string  `json:"fields" db:"fields"`
	UpdatedAt time.Time `json:"updatedAt" db:"updatedAt"`
}

type RefreshTokensDTO struct {
	UserId    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type JwtClaims struct {
	jwt.Claims
	email    string
	username string
}
