package usecase

import (
	"context"
	"fpart/internal/domain/user"
	userStorage "fpart/internal/infra/repository/user"
	"fpart/internal/infra/secure"
	"fpart/internal/pkg/errs"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

type mockOAuth2Provider struct {
	User *user.User
	Err  error
}

func (m *mockOAuth2Provider) GetUserInfoByCode(ctx context.Context, code string) (*user.User, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return m.User, m.Err
}

func TestGoogleLogin_PrepareCallback_UserDuplicate(t *testing.T) {
	secretForToken := "secret123"

	User := user.NewUser("user-id", "Alex Miller", "example@gmail.com", "url://rand.domain.png")
	TokenLivetime := time.Duration(time.Minute)

	mockProvider := mockOAuth2Provider{
		User: User,
		Err:  nil,
	}

	usecase := NewGoogleLoginUseCase(
		context.Background(),
		&oauth2.Config{},
		secure.NewJWTokenService(secretForToken, TokenLivetime),
		userStorage.NewUserLStorageRepository(),
		zerolog.Logger{},
	)

	usecase.oauth2Provider = &mockProvider

	state := "state123"
	usecase.stateMap[state] = time.Now().Add(time.Minute)

	_, err := usecase.PrepareCallback(context.Background(), state, "someCode")
	assert.Nil(t, err)
	assert.Equal(t, usecase.Metrics.LoginedUsers, uint32(1))
	assert.Equal(t, usecase.Metrics.StatesDeletedCount, uint32(1))
	_, ok := usecase.stateMap[state]
	assert.Equal(t, false, ok)

	state = "state456"
	usecase.stateMap[state] = time.Now().Add(time.Minute)

	_, err = usecase.PrepareCallback(context.Background(), state, "someCode")
	assert.EqualError(t, err, user.ErrUserAlreadyExists.Error())
	assert.Equal(t, usecase.Metrics.LoginedErrorProcesses, uint32(1))
	assert.Equal(t, usecase.Metrics.StatesDeletedCount, uint32(2))
	_, ok = usecase.stateMap[state]
	assert.Equal(t, false, ok)
}

func TestGoogleLogin_PrepareCallback_StateDeleted(t *testing.T) {
	secret := "secret123"

	mockProvider := mockOAuth2Provider{
		User: user.NewUser("randID", "Full Name", "example@domain.org", "url://cat.picture.jpeg"),
		Err:  nil,
	}

	usecase := NewGoogleLoginUseCase(
		context.Background(),
		&oauth2.Config{},
		secure.NewJWTokenService(secret, time.Minute),
		userStorage.NewUserLStorageRepository(),
		zerolog.Logger{},
	)
	usecase.oauth2Provider = &mockProvider
	usecase.cleanInterval = time.Duration(4 * time.Second)

	state := "state456"
	usecase.stateMap[state] = time.Now().Add(2 * time.Second)

	time.Sleep(6 * time.Second) // state cleaner work every 30 seconds
	_, ok := usecase.stateMap[state]
	assert.Equal(t, false, ok)

	_, err := usecase.PrepareCallback(context.Background(), state, "someCode")
	assert.EqualError(t, err, ErrGetUserInfo.Error())
	assert.Equal(t, usecase.Metrics.LoginedErrorProcesses, uint32(1))
}

func TestGoogleLogin_PrepareCallback_ValidCase(t *testing.T) {
	secret := "secret123"

	mockProvider := mockOAuth2Provider{
		User: user.NewUser("user-id", "Bil Gates", "ex.mail@gmail.com", "photo://j1.googleusercontent.com/validcase?size=1024x1440"),
		Err:  nil,
	}

	tokenService := secure.NewJWTokenService(secret, time.Minute)

	usecase := NewGoogleLoginUseCase(
		context.Background(),
		&oauth2.Config{},
		tokenService,
		userStorage.NewUserLStorageRepository(),
		zerolog.Logger{},
	)
	usecase.oauth2Provider = &mockProvider

	state := "rand-state"
	usecase.stateMap[state] = time.Now().Add(time.Minute)

	token, err := usecase.PrepareCallback(context.Background(), state, "code")
	assert.Nil(t, err)
	assert.Equal(t, usecase.Metrics.LoginedUsers, uint32(1))
	assert.Equal(t, usecase.Metrics.StatesDeletedCount, uint32(1))

	uid, err := tokenService.Validate(token)
	assert.Nil(t, err)
	assert.Equal(t, mockProvider.User.GetID(), uid)
}

func TestGoogleLogin_PrepareCallback_ContextExceeded(t *testing.T) {
	secret := "secret123"

	mockProvider := mockOAuth2Provider{
		User: user.NewUser("user-id", "Bil Gates", "ex.mail@gmail.com", "photo://j1.googleusercontent.com/validcase?size=1024x1440"),
		Err:  nil,
	}

	context, cancel := context.WithCancel(context.Background())
	defer cancel()

	cancel()
	usecase := NewGoogleLoginUseCase(
		context,
		&oauth2.Config{},
		secure.NewJWTokenService(secret, time.Minute),
		userStorage.NewUserLStorageRepository(),
		zerolog.Logger{},
	)
	usecase.oauth2Provider = &mockProvider

	state := "state-rand"
	usecase.stateMap[state] = time.Now().Add(time.Minute)

	_, err := usecase.PrepareCallback(context, state, "code")
	assert.EqualError(t, err, ErrGetUserInfo.Error())
	assert.Equal(t, usecase.Metrics.StatesDeletedCount, uint32(1))
	assert.Equal(t, usecase.Metrics.LoginedErrorProcesses, uint32(1))
	assert.NotNil(t, context.Err())
}

func TestGoogleLogin_PrepareCallback_OAuthProviderError(t *testing.T) {
	secret := "secret123"

	mockProvider := mockOAuth2Provider{
		User: nil,
		Err:  &errs.InternalError{},
	}

	usecase := NewGoogleLoginUseCase(
		context.Background(),
		&oauth2.Config{},
		secure.NewJWTokenService(secret, time.Minute),
		userStorage.NewUserLStorageRepository(),
		zerolog.Logger{},
	)
	usecase.oauth2Provider = &mockProvider

	state := "rand-state"
	usecase.stateMap[state] = time.Now().Add(5 * time.Second)

	_, err := usecase.PrepareCallback(context.Background(), state, "code")
	assert.EqualError(t, err, ErrGetUserInfo.Error())
	assert.Equal(t, usecase.Metrics.LoginedErrorProcesses, uint32(1))
	assert.Equal(t, usecase.Metrics.StatesDeletedCount, uint32(1))
}

func TestGoogleLogin_PrepareCallback_StateExpiredCase(t *testing.T) {
	secret := "secret123"

	mockProvider := mockOAuth2Provider{
		User: user.NewUser("user-id", "Bil Gates", "ex.mail@gmail.com", "photo://j1.googleusercontent.com/validcase?size=1024x1440"),
		Err:  nil,
	}

	usecase := NewGoogleLoginUseCase(
		context.Background(),
		&oauth2.Config{},
		secure.NewJWTokenService(secret, time.Minute),
		userStorage.NewUserLStorageRepository(),
		zerolog.Logger{},
	)
	usecase.oauth2Provider = &mockProvider

	usecase.cleanInterval = time.Duration(3 * time.Second)

	state := "state-123"
	usecase.stateMap[state] = time.Now().Add(4 * time.Second)

	time.Sleep(5 * time.Second)
	assert.Equal(t, usecase.Metrics.StatesDeletedCount, uint32(0))
	expired := usecase.stateMap[state].Before(time.Now())
	assert.True(t, expired)

	_, err := usecase.PrepareCallback(context.Background(), state, "code")
	assert.EqualError(t, err, ErrGetUserInfo.Error())
	assert.Equal(t, usecase.Metrics.StatesDeletedCount, uint32(0))
	assert.Equal(t, usecase.Metrics.LoginedErrorProcesses, uint32(1))
}
