package secure

import (
	"fpart/internal/application/ports"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToken_ValidCases(t *testing.T) {
	secret := "secret-key"

	testCases := []struct {
		desc        string
		id          string
		livetime    time.Duration
		expectedErr bool
		err         error
	}{
		{
			desc:     "valid token",
			id:       "valid-id",
			livetime: 10 * time.Second,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// service init
			TokenService := NewJWTokenService(secret, tC.livetime)

			// generate token with use id
			token, err := TokenService.Generate(tC.id)
			assert.Equal(t, false, err != nil)

			// validate token
			uid, err := TokenService.Validate(token)
			assert.Equal(t, false, err != nil)
			assert.Equal(t, tC.err, err)
			assert.Equal(t, tC.id, uid)
		})
	}
}

func TestToken_ChangeUseCase(t *testing.T) {
	secret := "secret123"
	token_service := NewJWTokenService(secret, 10*time.Second)

	uid := "user_123"

	token, err := token_service.Generate(uid)
	assert.Equal(t, false, err != nil)

	token = func() string {
		tokenBytes := []byte(token)
		tokenBytes[3] = byte('g')

		return string(tokenBytes)
	}()

	tokenUserID, err := token_service.Validate(token)
	assert.EqualError(t, err, ports.ErrInvalidToken.Error())
	assert.EqualValues(t, "", tokenUserID)
}

func TestToken_ExpiredUseCase(t *testing.T) {
	secret := "secret456"
	token_service := NewJWTokenService(secret, 1*time.Second)

	uid := "user_7463"

	token, err := token_service.Generate(uid)
	assert.Equal(t, false, err != nil)

	time.Sleep(2 * time.Second)

	tokenUserID, err := token_service.Validate(token)
	assert.EqualError(t, err, ports.ErrInvalidToken.Error())
	assert.EqualValues(t, "", tokenUserID)
}
