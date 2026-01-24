package main

import (
	"context"
	authService "fpart/internal/application/auth"
	userService "fpart/internal/application/user"
	authHandler "fpart/internal/infra/http/auth"
	"fpart/internal/infra/http/middleware"
	userHandler "fpart/internal/infra/http/user"
	userRepo "fpart/internal/infra/repository/user"
	"fpart/internal/infra/secure"
	"fpart/internal/pkg/utils"
	"os"
	"strings"
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

	Mode string

	routes map[string]string = map[string]string{
		"google-login":    "/api/v1/auth/google/login",
		"google-callback": "/api/v1/auth/google/callback",
	}
)

func init() {
	Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	Mode = os.Getenv("MODE")
	if strings.Contains(strings.ToLower(Mode), "prod") {
		Logger = Logger.Level(zerolog.InfoLevel)
	} else {
		Logger = Logger.Level(zerolog.DebugLevel)
	}

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

	userRepository := userRepo.NewUserLStorageRepository()
	tokenService := secure.NewJWTokenService(TokenSecret, 90*24*time.Hour) // need replace on secret.txt source read

	// middleware
	middleware := middleware.NewMiddleware(tokenService,
		routes["google-login"],
		routes["google-callback"],
	)

	// auth module
	authService := authService.NewAuthService(
		ctx,
		OAuth2Cfg, tokenService, userRepository, &Logger,
	)
	authHandler := authHandler.NewAuthHandler(*authService, utils.HandlerOpts{
		RequestTimeout: 5 * time.Second,
	})

	// user module
	userService := userService.NewUserService(&Logger, userRepository)
	userHandler := userHandler.NewUserHandler(userService)

	r := router.New()
	APIv1 := r.Group("/api/v1")
	{
		auth := APIv1.Group("/auth")
		{
			auth.GET("/google/login", authHandler.LoginWithGoogleHandler)
			auth.GET("/google/callback", authHandler.LoginWithGoogleCallback)
		}
		APIv1.GET("/user", userHandler.GetUserHandler)
		user := APIv1.Group("/user")
		{
			user.GET("/{userId}/projects", func(ctx *fasthttp.RequestCtx) {})
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
		Handler:               middleware.Handler(r.Handler),
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
