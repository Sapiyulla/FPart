package secure

import (
	"fpart/internal/application/ports"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTokenService struct {
	secret []byte

	livetime time.Duration
}

type UserClaims struct {
	ID string `json:"id"`
	jwt.StandardClaims
}

func NewJWTokenService(secret string, livetime time.Duration) *JWTokenService {
	return &JWTokenService{
		secret: []byte(secret),

		livetime: livetime,
	}
}

// # Generate
//
// This method generates (and return) a new token based on the passed argument in the form of a user ID.
//
// Returned error(s):
//   - *jwt.Token.SignedString() returned error
func (ts *JWTokenService) Generate(uid string) (token string, err error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
		ID: uid,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(ts.livetime).Unix(),
			Issuer:    "fpart",
		},
	})
	return t.SignedString(ts.secret)
}

// # Validate
//
// This method validates the received value as a token and,
// if successful, returns a unique user identifier.
//
// Validation error is:
//   - fake token
//   - changed token
//
// Returned error(s):
//   - [ports.ErrInvalidToken] (application/ports)
func (ts *JWTokenService) Validate(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&UserClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ports.ErrInvalidToken
			}
			return ts.secret, nil
		},
	)

	if err != nil {
		return "", ports.ErrInvalidToken
	}

	claims, ok := token.Claims.(*UserClaims)

	if !ok || !token.Valid {
		return "", ports.ErrInvalidToken
	}

	return claims.ID, nil
}
