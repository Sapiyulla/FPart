package auth

import (
	"context"
	"fpart/internal/application/auth"
	"fpart/internal/pkg/utils"
	"time"

	"github.com/valyala/fasthttp"
)

type AuthHandler struct {
	authService *auth.AuthService

	utils.HandlerOpts
}

func NewAuthHandler(authService auth.AuthService, opts ...utils.HandlerOpts) *AuthHandler {
	if len(opts) == 0 {
		return &AuthHandler{
			authService: &authService,
			HandlerOpts: utils.HandlerOpts{
				RequestTimeout: 10 * time.Second,
			},
		}
	}
	return &AuthHandler{
		authService: &authService,
		HandlerOpts: opts[0],
	}
}

func (h *AuthHandler) LoginWithGoogleHandler(c *fasthttp.RequestCtx) {
	c.Redirect(h.authService.GetRedirectURL(), fasthttp.StatusTemporaryRedirect)
}

func (h *AuthHandler) LoginWithGoogleCallback(c *fasthttp.RequestCtx) {
	state := c.Request.URI().QueryArgs().Peek("state")
	code := c.Request.URI().QueryArgs().Peek("code")
	ctx, cancel := context.WithTimeout(context.Background(), h.RequestTimeout)
	defer cancel()
	token, err := h.authService.PrepareCallback(ctx, string(state), string(code))
	if err != nil {
		switch err {
		case auth.ErrGetUserInfo:
			c.SetStatusCode(fasthttp.StatusInternalServerError)
		case auth.ErrUserAlreadyExists:
			c.SetStatusCode(fasthttp.StatusNotFound)
		case auth.ErrInternalService:
			c.SetStatusCode(fasthttp.StatusInternalServerError)
		}
		c.SetBodyString(err.Error())
		return
	}
	cookie := fasthttp.AcquireCookie()
	defer fasthttp.ReleaseCookie(cookie)
	cookie.SetHTTPOnly(true)
	cookie.SetPath("/api")
	cookie.SetKey("token")
	cookie.SetValue(token)
	cookie.SetMaxAge(90 * 24 * 60 * 60)

	c.Response.Header.SetCookie(cookie)
}
