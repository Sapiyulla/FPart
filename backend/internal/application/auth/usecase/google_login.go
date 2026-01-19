package usecase

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fpart/internal/application/ports"
	"fpart/internal/domain/user"
	"sync"

	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

type jsonUser struct {
	ID       string `json:"id"`
	Fullname string `json:"name"`
	Email    string `json:"email"`
	Picture  string `json:"picture"`
}

var (
	ErrInvalidState    error = errors.New("invalid state: not equal")
	ErrInvalidExchange error = errors.New("invalid exchange operation")
	ErrGetUserInfo     error = errors.New("get user info error")
	ErrJsonDecodeError error = errors.New("json decoding error")
)

type GoogleLoginUseCase struct {
	oauth2Cfg      *oauth2.Config
	tokenService   ports.TokenService
	userRepository ports.UserRepository

	logger *zerolog.Logger

	stateMap map[string]struct{}
	mu       sync.Mutex
}

func NewGoogleLoginUseCase(
	cfg *oauth2.Config,
	tokenService ports.TokenService,
	userRepo ports.UserRepository,
	logger *zerolog.Logger,
) *GoogleLoginUseCase {
	return &GoogleLoginUseCase{
		oauth2Cfg:      cfg,
		tokenService:   tokenService,
		userRepository: userRepo,

		logger: logger,

		stateMap: map[string]struct{}{},
		mu:       sync.Mutex{},
	}
}

func (uc *GoogleLoginUseCase) GetRedirectURL() string {
	state := rand.Text()[:12]
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.stateMap[state] = struct{}{}
	return uc.oauth2Cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func (uc *GoogleLoginUseCase) PrepareCallback(ctx context.Context, state, code string) (*user.User, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if _, ok := uc.stateMap[state]; !ok {
		return nil, ErrInvalidState
	}
	delete(uc.stateMap, state)

	token, err := uc.oauth2Cfg.Exchange(ctx, code, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	if err != nil {
		uc.logger.Error().Err(err).Msg(ErrInvalidExchange.Error())
		return nil, ErrGetUserInfo
	}

	resp, err := uc.oauth2Cfg.Client(ctx, token).Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		uc.logger.Error().Err(err).Msg(ErrGetUserInfo.Error())
		return nil, ErrGetUserInfo
	}
	defer resp.Body.Close()

	var JsonUser jsonUser
	if err := json.NewDecoder(resp.Body).Decode(&JsonUser); err != nil {
		uc.logger.Error().Err(err).Msg(ErrJsonDecodeError.Error())
		return nil, ErrGetUserInfo
	}

	return user.NewUser(JsonUser.ID, JsonUser.Fullname, JsonUser.Email, JsonUser.Picture), nil
}
