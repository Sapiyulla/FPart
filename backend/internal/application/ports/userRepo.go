package ports

import "fpart/internal/domain/user"

type UserRepository interface {
	GetUserByID(string) (*user.User, error)
	AddNewUser(*user.User) error
}
