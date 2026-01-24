package ports

import "fpart/internal/domain/user"

var (
	ErrUserAlreadyExists error = user.ErrUserAlreadyExists
	ErrUserNotFound      error = user.ErrUserNotFound
)

type UserRepository interface {
	GetUserByID(string) (*user.User, error)
	AddNewUser(*user.User) error
}
