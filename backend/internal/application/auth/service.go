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
