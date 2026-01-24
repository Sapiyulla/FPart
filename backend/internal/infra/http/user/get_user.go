package user

import (
	"fpart/internal/application/user"

	"github.com/valyala/fasthttp"
)

// dto`s
type GetUserResponse struct {
	ID       string `json:"id"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Picture  string `json:"picture"`
}

func (h *UserHandler) GetUserHandler(c *fasthttp.RequestCtx) {
	uid := string(c.Request.URI().QueryArgs().Peek("userId"))
	if uid == "" {
		uid = string(c.Request.Header.Peek("User-ID"))
	}

	u, err := h.service.GetByID(uid)
	if err != nil {
		switch err {
		case user.ErrUserNotFound:
			c.SetStatusCode(fasthttp.StatusNotFound)
		default:
			c.SetStatusCode(fasthttp.StatusInternalServerError)
		}
		return
	}

	userBytes, err := (&GetUserResponse{
		ID:       u.GetID(),
		Email:    u.GetEmail(),
		Fullname: u.GetFullname(),
		Picture:  u.GetPhotoURL(),
	}).MarshalJSON()
	if err != nil {
		c.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	c.SetBody(userBytes)
	c.SetStatusCode(fasthttp.StatusOK)
}
