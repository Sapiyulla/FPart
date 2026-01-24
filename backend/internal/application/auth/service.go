package auth

import (
	"context"
	"fpart/internal/application/auth/usecase"
	"fpart/internal/application/ports"

	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

type AuthService struct {
	*usecase.GoogleLoginUseCase
}

var (
	ErrGetUserInfo       error = usecase.ErrGetUserInfo
	ErrInternalService   error = usecase.ErrInternalService
	ErrUserAlreadyExists error = usecase.ErrUserAlreadyExists
)

func NewAuthService(
	ctx context.Context,
	oauthCfg *oauth2.Config,
	tokenService ports.TokenService,
	userRepo ports.UserRepository,
	logger *zerolog.Logger,
) *AuthService {
	return &AuthService{
		usecase.NewGoogleLoginUseCase(
			ctx,
			oauthCfg,
			tokenService,
			userRepo,
			*logger,
		),
	}
}
