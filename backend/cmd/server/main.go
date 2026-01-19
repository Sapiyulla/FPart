package main

import (
	"context"
	"os"

	"github.com/fasthttp/router"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

var (
	Logger    zerolog.Logger
	OAuth2Cfg *oauth2.Config
)

func init() {
	Logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.DebugLevel)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := router.New()
	APIv1 := r.Group("/api/v1")
}
