package usecase

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fpart/internal/application/ports"
	"fpart/internal/domain/user"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

type Metrics struct {
	StatesDeletedCount uint32 `json:"deleted_states_count"`
	StatesAddedCount   uint32 `json:"added_states_count"`
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

type GoogleLoginUseCase struct {
	oauth2Cfg      *oauth2.Config
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
	googleUseCase := &GoogleLoginUseCase{
		oauth2Cfg:      cfg,
		tokenService:   tokenService,
		userRepository: userRepo,

		logger: logger.With().Str("usecase", "google_login").Logger(),

		stateMap: map[string]time.Time{},
		mu:       sync.Mutex{},

		Metrics: Metrics{},
	}
	go googleUseCase.stateStorageCleaner(ctx)
	if googleUseCase.logger.GetLevel() == zerolog.DebugLevel {
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					time.Sleep(30 * time.Second)
					logger.Debug().
						Uint32("deleted_states_count", googleUseCase.Metrics.StatesDeletedCount).
						Uint32("added_states_count", googleUseCase.Metrics.StatesAddedCount).
						Msg("metrics loaded")
				}
			}
		}(ctx)
	}
	return googleUseCase
}

func (uc *GoogleLoginUseCase) GetRedirectURL() string {
	randState := rand.Text()[:12]
	if uc.logger.GetLevel() == zerolog.DebugLevel {
		uc.logger.Debug().Str("state", randState).Msg("new state generated")
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.stateMap[randState] = time.Now().Add(5 * time.Minute)
	atomic.AddUint32(&uc.Metrics.StatesAddedCount, 1)
	if uc.logger.GetLevel() == zerolog.DebugLevel {
		uc.logger.Debug().Str("state", randState).Msg("state added to states storage")
	}
	return uc.oauth2Cfg.AuthCodeURL(randState, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func (uc *GoogleLoginUseCase) PrepareCallback(ctx context.Context, state, code string) (string, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if uc.logger.GetLevel() == zerolog.DebugLevel {
		uc.logger.Debug().Str("state", state).Msg("state from user")
	}
	if _, ok := uc.stateMap[state]; !ok {
		if uc.logger.GetLevel() == zerolog.DebugLevel {
			uc.logger.Debug().Bool("found", ok).Msg("state not equal with execute state")
		}
		return "", ErrInvalidState
	}
	if uc.logger.GetLevel() == zerolog.DebugLevel {
		uc.logger.Debug().Bool("found", true).Msg("state valid")
	}
	delete(uc.stateMap, state)
	atomic.AddUint32(&uc.Metrics.StatesDeletedCount, 1)
	if uc.logger.GetLevel() == zerolog.DebugLevel {
		uc.logger.Debug().Str("state", state).Msg("state succesfully deleted")
	}

	token, err := uc.oauth2Cfg.Exchange(ctx, code, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	if err != nil {
		uc.logger.Error().Err(err).Msg(ErrInvalidExchange.Error())
		return "", ErrGetUserInfo
	}

	resp, err := uc.oauth2Cfg.Client(ctx, token).Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		uc.logger.Error().Err(err).Msg(ErrGetUserInfo.Error())
		return "", ErrGetUserInfo
	}
	defer resp.Body.Close()

	var JsonUser jsonUser
	if err := json.NewDecoder(resp.Body).Decode(&JsonUser); err != nil {
		uc.logger.Error().Err(err).Msg(ErrJsonDecodeError.Error())
		return "", ErrGetUserInfo
	}

	Token, err := uc.tokenService.Generate(JsonUser.ID)
	if err != nil {
		uc.logger.Error().Err(err).Msg("token generate error")
		return "", ErrGetUserInfo
	}

	if err := uc.userRepository.AddNewUser(user.NewUser(
		JsonUser.ID, JsonUser.Fullname, JsonUser.Email, JsonUser.Picture,
	)); err != nil {
		uc.logger.Error().Err(err).Msg("user save to repository error")
		return "", ErrGetUserInfo
	}

	if uc.logger.GetLevel() == zerolog.DebugLevel {
		uc.logger.Debug().Str("token", Token).Msg("logined to service")
	}

	return Token, nil
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
					atomic.AddUint32(&uc.Metrics.StatesDeletedCount, 1)
					if uc.logger.GetLevel() == zerolog.DebugLevel {
						uc.logger.Debug().Str("state", state).Msg("expired state deleted")
					}
				}
			}
			uc.mu.Unlock()
		}
	}
}
