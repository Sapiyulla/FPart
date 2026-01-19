package ports

import "fpart/internal/domain/user"

type UserRepository interface {
	AddNewUser(*user.User) error
}
