package user

import (
	"fpart/internal/application/ports"
	"fpart/internal/application/user/usecase"
	"fpart/internal/domain/user"

	"github.com/rs/zerolog"
)

type UserService struct {
	*usecase.UserGetUseCase
}

var (
	ErrUserNotFound error = user.ErrUserNotFound
)

func NewUserService(
	logger *zerolog.Logger,
	repo ports.UserRepository,
) *UserService {
	userServiceLogger := logger.With().Str("service", "user").Logger()
	return &UserService{
		UserGetUseCase: usecase.NewUserGetUseCase(&userServiceLogger, repo),
	}
}
