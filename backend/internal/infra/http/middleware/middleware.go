package middleware

import (
	"fpart/internal/application/ports"
	"sync"

	"github.com/valyala/fasthttp"
)

type Middleware struct {
	tokenService ports.TokenService

	exceptions map[string]bool
	mu         sync.RWMutex
}

func NewMiddleware(tokenService ports.TokenService, exceptions ...string) *Middleware {
	m := &Middleware{
		tokenService: tokenService,

		exceptions: map[string]bool{},
		mu:         sync.RWMutex{},
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, exception := range exceptions {
		m.exceptions[exception] = true
	}

	return m
}

func (m *Middleware) Handler(router func(c *fasthttp.RequestCtx)) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		m.mu.RLock()
		if isExc := m.exceptions[string(ctx.Request.URI().Path())]; isExc {
			router(ctx)
			return
		}
		m.mu.RUnlock()

		token := ctx.Request.Header.Cookie("token")
		if len(token) == 0 || token == nil {
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			return
		}

		uid, err := m.tokenService.Validate(string(token))
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			return
		}
		ctx.Request.Header.Add("User-ID", uid)

		router(ctx)
	}
}
