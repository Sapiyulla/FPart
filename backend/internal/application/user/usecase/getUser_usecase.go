package usecase

import (
	"fpart/internal/application/ports"
	"fpart/internal/domain/user"
	"fpart/internal/pkg/errs"

	"github.com/rs/zerolog"
)

type UserGetUseCase struct {
	repo ports.UserRepository

	logger zerolog.Logger
}

func NewUserGetUseCase(
	logger *zerolog.Logger,
	repo ports.UserRepository,
) *UserGetUseCase {
	return &UserGetUseCase{
		repo:   repo,
		logger: logger.With().Str("usecase", "user_get").Logger(),
	}
}

func (u *UserGetUseCase) GetByID(id string) (*user.User, error) {
	user, err := u.repo.GetUserByID(id)
	if err != nil {
		u.logger.Error().
			Err(err).
			Str("op", "get_by_id").
			Msg("repository error")
		switch err {
		case ports.ErrUserNotFound:
			return nil, err
		default:
			return nil, &errs.InternalError{}
		}
	}
	u.logger.Debug().
		Str("op", "get_by_id").
		Msg("success operation")
	return user, nil
}
