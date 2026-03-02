package auth

import "github.com/golang-jwt/jwt/v5"

//go:generate mockgen -source=./auth.go -destination=../../cmd/api/mock/auth/Mock_Mailer.go -package=mock_auth Authenticator
type Authenticator interface {
	GenerateToken(claims jwt.Claims) (string, error)
	ValidateToken(string) (*jwt.Token, error)
}
