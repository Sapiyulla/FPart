package usecase

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fpart/internal/application/ports"
	"fpart/internal/domain/user"
	"fpart/internal/pkg/errs"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

type Metrics struct {
	StatesDeletedCount uint32 `json:"deleted_states_count"`
	StatesAddedCount   uint32 `json:"added_states_count"`

	LoginedUsers          uint32 `json:"succesfully_login_users"`
	LoginedErrorProcesses uint32 `json:"error_login_processes"`
}

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

var oauth2Cfg *oauth2.Config

type GoogleLoginUseCase struct {
	tokenService   ports.TokenService
	userRepository ports.UserRepository

	logger zerolog.Logger

	stateMap map[string]time.Time
	mu       sync.Mutex

	Metrics Metrics
}

func NewGoogleLoginUseCase(
	ctx context.Context,
	cfg *oauth2.Config,
	tokenService ports.TokenService,
	userRepo ports.UserRepository,
	logger zerolog.Logger,
) *GoogleLoginUseCase {
	oauth2Cfg = cfg
	googleUseCase := &GoogleLoginUseCase{
		tokenService:   tokenService,
		userRepository: userRepo,

		logger: logger.With().Str("usecase", "google_login").Logger(),

		stateMap: map[string]time.Time{},
		mu:       sync.Mutex{},

		Metrics: Metrics{},
	}
	go googleUseCase.stateStorageCleaner(ctx)
	return googleUseCase
}

// # RedirectURL
//
// This function returns only the path to the callback endpoint.
// It is required to be called before processing user data for user authorization.
func (uc *GoogleLoginUseCase) GetRedirectURL() string {
	randState := rand.Text()[:12]
	uc.logger.Debug().Msg("new oauth state generated")

	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.stateMap[randState] = time.Now().Add(3 * time.Minute)
	atomic.AddUint32(&uc.Metrics.StatesAddedCount, 1)

	return oauth2Cfg.AuthCodeURL(randState, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// # Callback (Google)
//
// This function processes user data after the user has verified their Google account.
// It receives state and code as input, which are used to verify and retrieve user information using the Google API.
//
// Returned error(s):
//   - [ErrGetUserInfo]
//   - [user.ErrUserAlreadyExists] (domain/user)
//   - [errs.InternalError] (pkg/errs)
func (uc *GoogleLoginUseCase) PrepareCallback(ctx context.Context, state, code string) (string, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if ttl, ok := uc.stateMap[state]; !ok {
		uc.logger.Debug().Str("op", "prepare_callback").Msg(ErrInvalidState.Error())
		atomic.AddUint32(&uc.Metrics.LoginedErrorProcesses, 1)
		return "", ErrGetUserInfo
	} else if ttl.Before(time.Now()) {
		uc.logger.Debug().Str("op", "prepare_callback").Msg("state expired")
		atomic.AddUint32(&uc.Metrics.LoginedErrorProcesses, 1)
		return "", ErrGetUserInfo
	}

	delete(uc.stateMap, state)
	atomic.AddUint32(&uc.Metrics.StatesDeletedCount, 1)

	googleToken, err := oauth2Cfg.Exchange(ctx, code, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	if err != nil {
		uc.logger.Error().
			Err(err).
			Str("op", "prepare_callback").
			Msg(ErrInvalidExchange.Error())
		atomic.AddUint32(&uc.Metrics.LoginedErrorProcesses, 1)
		return "", &errs.InternalError{}
	}

	resp, err := oauth2Cfg.Client(ctx, googleToken).Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		uc.logger.Error().
			Err(err).
			Str("op", "prepare_callback").
			Msg(ErrGetUserInfo.Error())
		atomic.AddUint32(&uc.Metrics.LoginedErrorProcesses, 1)
		return "", &errs.InternalError{}
	}
	defer resp.Body.Close()

	var JsonUser jsonUser
	if err := json.NewDecoder(resp.Body).Decode(&JsonUser); err != nil {
		uc.logger.Error().
			Err(err).
			Str("op", "prepare_callback").
			Msg(ErrJsonDecodeError.Error())
		atomic.AddUint32(&uc.Metrics.LoginedErrorProcesses, 1)
		return "", &errs.InternalError{}
	}

	token, err := uc.tokenService.Generate(JsonUser.ID)
	if err != nil {
		uc.logger.Error().
			Err(err).
			Str("op", "prepare_callback").
			Msg("token generate error")
		atomic.AddUint32(&uc.Metrics.LoginedErrorProcesses, 1)
		return "", ErrGetUserInfo
	}

	if err := uc.userRepository.AddNewUser(user.NewUser(
		JsonUser.ID, JsonUser.Fullname, JsonUser.Email, JsonUser.Picture,
	)); err != nil {
		uc.logger.Error().
			Err(err).
			Str("op", "prepare_callback").
			Msg("user save error")
		atomic.AddUint32(&uc.Metrics.LoginedErrorProcesses, 1)
		return "", err
	}

	uc.logger.Debug().Msg("logined to service")
	atomic.AddUint32(&uc.Metrics.LoginedUsers, 1)
	return token, nil
}

func (uc *GoogleLoginUseCase) stateStorageCleaner(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(30 * time.Second)
			uc.mu.Lock()
			for state, ttl := range uc.stateMap {
				if ctx.Err() != nil {
					return
				}
				if ttl.Before(time.Now()) {
					delete(uc.stateMap, state)
					uc.logger.Trace().
						Str("component", "state_cleaner").
						Msg("expired oauth state removed")
					atomic.AddUint32(&uc.Metrics.StatesDeletedCount, 1)
				}
			}
			uc.mu.Unlock()
		}
	}
}
