package handlers

import (
	"context"
	"fpart/internal/application/auth"
	"fpart/internal/pkg/utils"

	"github.com/valyala/fasthttp"
)

type AuthHandler struct {
	authService *auth.AuthService

	utils.HandlerOpts
}

func NewAuthHandler(authService auth.AuthService, opts ...utils.HandlerOpts) *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) LoginWithGoogleHandler(c *fasthttp.RequestCtx) {
	c.Redirect(h.authService.GetRedirectURL(), fasthttp.StatusTemporaryRedirect)
}

func (h *AuthHandler) LoginWithGoogleCallback(c *fasthttp.RequestCtx) {
	state := c.Request.URI().QueryArgs().Peek("state")
	code := c.Request.URI().QueryArgs().Peek("code")
	ctx, cancel := context.WithTimeout(context.Background(), h.RequestTimeout)
	defer cancel()
	user, err := h.authService.PrepareCallback(ctx, string(state), string(code))
	if err != nil {
		c.SetStatusCode(fasthttp.StatusInternalServerError)
		c.SetBodyString(err.Error())
		return
	}
	cookie := fasthttp.AcquireCookie()
	cookie.SetHTTPOnly(true)
	cookie.SetKey("token")
	cookie.SetValue()
	cookie.SetMaxAge(90 * 24 * 60 * 60)
	defer fasthttp.ReleaseCookie(cookie)
}
