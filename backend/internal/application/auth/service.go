package auth

import (
	"fpart/internal/application/auth/usecase"
	"fpart/internal/application/ports"

	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

type AuthService struct {
	*usecase.GoogleLoginUseCase
}

func NewAuthService(
	oauthCfg *oauth2.Config,
	tokenService ports.TokenService,
	userRepo ports.UserRepository,
	logger *zerolog.Logger,
) *AuthService {
	return &AuthService{
		usecase.NewGoogleLoginUseCase(
			oauthCfg,
			tokenService,
			userRepo,
			func(l *zerolog.Logger) *zerolog.Logger {
				Logger := l.With().Str("service", "auth").Logger()
				return &Logger
			}(logger),
		),
	}
}
