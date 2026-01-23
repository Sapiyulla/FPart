package ports

import (
	"context"
	"fpart/internal/domain/user"
)

type OAuth2Provider interface {
	GetUserInfoByCode(ctx context.Context, code string) (*user.User, error)
}
