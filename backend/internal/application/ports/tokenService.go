package ports

import "fpart/internal/domain/user"

type TokenService interface {
	Generate(*user.User) (token string, err error)
	Validate(token string) bool
}
