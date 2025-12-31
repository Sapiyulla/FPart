package auth

import (
	"context"
	"fpart/internal/domain/user"
)

type UserRepository interface {
	AddUser(ctx context.Context, username, email, password string) error
	FindUserByEmail(ctx context.Context, email string) (*user.User, error)
}
