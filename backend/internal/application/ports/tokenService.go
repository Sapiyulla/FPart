package ports

import "fpart/internal/infra/secure"

type TokenService interface {
	Generate(user_id string) (token string, err error)
	Validate(token string) (*secure.UserClaims, error)
}
