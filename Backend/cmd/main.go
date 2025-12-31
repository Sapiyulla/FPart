package main

import (
	"context"
	"database/sql"
	authService "fpart/internal/application/auth"
	authHandler "fpart/internal/infra/http/auth"
	"fpart/internal/infra/postgres/user"
	"os"

	"github.com/fasthttp/router"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

type mode uint8

const (
	DEBUG mode = 0
	PROD  mode = iota
)

type AppSettings struct {
	Mode   mode
	Config any // *config.Config
}

var (
	SystemLogger *zerolog.Logger
	ModuleLogger *zerolog.Logger
)

var Settings AppSettings

func init() {
	if os.Getenv("MODE") == "prod" {
		Settings.Mode = PROD
	}
}

// @title FastHTTP API
// @version 1.0
// @description API built with fasthttp
// @host localhost:8080
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize all depends
	database_pg, err := sql.Open("postgres", "postgres://postgres:password@database:5432/fpart?sslmode=disable")
	if err != nil {
		println(err.Error())
		return
	}
	defer database_pg.Close()

	userRepo := user.NewUserRepository(database_pg)

	// 2. Connect all depends
	authService := authService.NewAuthService(userRepo)
	authHandler := authHandler.NewAuthHandler(ctx, authService)

	// 3. Server initialize
	r := router.New()
	apiV1 := r.Group("/api/v1")
	{
		authApi := apiV1.Group("/auth")
		{
			authApi.POST("/registration", authHandler.RegistrationHandler)
		}
	}

	srv := fasthttp.Server{
		Handler:               r.Handler,
		NoDefaultServerHeader: true,
	}

	switch Settings.Mode {
	case DEBUG:
		if err := srv.ListenAndServe(":8080"); err != nil {
			panic(err.Error())
		}
	case PROD:
		if err := srv.ListenAndServeTLS(":443", "", ""); err != nil {
			panic(err.Error())
		}
	}

	// 4. Prepare Shutdown & etc.

}
