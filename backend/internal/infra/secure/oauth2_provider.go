package secure

import (
	"context"
	"encoding/json"
	"fpart/internal/domain/user"
	"fpart/internal/pkg/errs"

	"golang.org/x/oauth2"
)

type OAuth2Provider struct {
	cfg *oauth2.Config
}

type jsonUser struct {
	ID       string `json:"id"`
	Fullname string `json:"name"`
	Email    string `json:"email"`
	Picture  string `json:"picture"`
}

func NewOAuth2Provider(cfg *oauth2.Config) *OAuth2Provider {
	return &OAuth2Provider{
		cfg: cfg,
	}
}

func (p *OAuth2Provider) GetUserInfoByCode(ctx context.Context, code string) (*user.User, error) {
	token, err := p.cfg.Exchange(ctx, code, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	if err != nil {
		return nil, &errs.InternalError{}
	}

	resp, err := p.cfg.Client(ctx, token).Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, &errs.InternalError{}
	}
	defer resp.Body.Close()

	var userJson jsonUser
	if err := json.NewDecoder(resp.Body).Decode(&userJson); err != nil {
		return nil, &errs.InternalError{}
	}

	return user.NewUser(userJson.ID, userJson.Fullname, userJson.Email, userJson.Picture), nil
}
