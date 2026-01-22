package ports

import (
	"errors"
)

var (
	ErrInvalidToken error = errors.New("invalid token")
	ErrTokenExpired error = errors.New("token expired")
)

// # TokenService
//
// This is a contract for the use of a token service.
type TokenService interface {
	// This method accepts a user ID and returns a token based on it, or an internal service error.
	Generate(user_id string) (token string, err error)
	// This method is used to verify the token's validity. It returns the user ID.
	Validate(token string) (string, error)
}
