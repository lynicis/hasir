package auth

import "github.com/golang-jwt/jwt/v5"

type JwtClaims struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
