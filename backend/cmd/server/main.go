package main

import (
	"context"
	authService "fpart/internal/application/auth"
	authHandler "fpart/internal/infra/http/auth"
	userRepo "fpart/internal/infra/repository/user"
	"fpart/internal/infra/secure"
	"fpart/internal/pkg/utils"
	"os"
	"sync"
	"time"

	"github.com/fasthttp/router"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	Logger    zerolog.Logger
	OAuth2Cfg *oauth2.Config

	TokenSecret string
)

func init() {
	Logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.DebugLevel)

	TokenSecret = os.Getenv("TOKEN_SECRET")

	OAuth2Cfg = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URI"),
		Scopes: []string{
			"openid",
			"email",
			"profile",
		},
		Endpoint: google.Endpoint,
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userRepo := userRepo.NewUserLStorageRepository()
	tokenService := secure.NewJWTokenService(TokenSecret) // need replace on secret.txt source read

	// auth module
	authService := authService.NewAuthService(
		ctx,
		OAuth2Cfg, tokenService, userRepo, &Logger,
	)
	authHandler := authHandler.NewAuthHandler(*authService, utils.HandlerOpts{
		RequestTimeout: 5 * time.Second,
	})

	r := router.New()
	APIv1 := r.Group("/api/v1")
	{
		auth := APIv1.Group("/auth")
		{
			auth.GET("/google/login", authHandler.LoginWithGoogleHandler)
			auth.GET("/google/callback", authHandler.LoginWithGoogleCallback)
		}
	}

	if Logger.GetLevel() == zerolog.DebugLevel {
		for method, paths := range r.List() {
			for _, path := range paths {
				Logger.Debug().Str("method", method).Msg(path)
			}
		}
	}

	srv := fasthttp.Server{
		Handler:               r.Handler,
		NoDefaultServerHeader: true,
	}

	wg := sync.WaitGroup{}
	wg.Go(func() {
		Logger.Info().Str("addr", "0.0.0.0:8000").Str("server-type", "http/rest-api").Msg("server starting")
		if err := srv.ListenAndServe(":8000"); err != nil {
			Logger.Error().Err(err).Msg("server starting error")
		}
	})

	wg.Wait()
	go srv.Shutdown()
	time.Sleep(1 * time.Second)
	Logger.Info().Msg("server shutting down")
}
