package auth

import (
	"context"
	"fmt"
	"fpart/internal/application/auth"
	"fpart/internal/pkg/errs"
	"fpart/internal/pkg/utils/options"
	"log/slog"
	"time"

	"github.com/valyala/fasthttp"
)

const defaultRequestTimeout = 5 * time.Second

type AuthHandler struct {
	authService *auth.AuthService

	ctx            context.Context
	requestTimeout time.Duration
}

func NewAuthHandler(ctx context.Context, authService *auth.AuthService, opts ...options.HandlerOpts) *AuthHandler {
	if len(opts) == 0 {
		return &AuthHandler{
			ctx:            ctx,
			authService:    authService,
			requestTimeout: defaultRequestTimeout,
		}
	}
	return &AuthHandler{
		ctx:            ctx,
		authService:    authService,
		requestTimeout: opts[0].RequestTimeout,
	}
}

func (h *AuthHandler) RegistrationHandler(c *fasthttp.RequestCtx) {
	if h.ctx.Err() != nil {
		c.SetStatusCode(fasthttp.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(h.ctx, h.requestTimeout)
	defer cancel()

	var copyedBody []byte
	copy(copyedBody, c.PostBody())
	fmt.Printf("body: %s", copyedBody)

	var req RegRequestDto
	if err := (&req).UnmarshalJSON(copyedBody); err != nil {
		c.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	fmt.Printf("req: %+v", req)
	token, err := h.authService.Registration(ctx, auth.RegCommand{
		Username: req.Username,
		SigninCommand: auth.SigninCommand{
			Email:    req.Email,
			Password: req.Password,
		},
	})
	if err != nil {
		slog.Error("registration error", "error", err.Error())
		switch err.(type) {
		case *errs.InternalError:
			c.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		case *errs.DuplicateError:
			c.SetStatusCode(fasthttp.StatusConflict)
			c.SetBody([]byte(err.Error()))
			return
		default:
			// logging
			c.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}
	}

	c.SetStatusCode(fasthttp.StatusCreated)
	t, err := TokenResp{Token: token}.MarshalJSON()
	if err != nil {
		slog.Error("token marshaling error", "error", err.Error())
		c.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	c.SetBody(t)
}
