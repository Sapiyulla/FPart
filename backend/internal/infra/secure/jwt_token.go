package secure

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTokenService struct {
	secret []byte
}

var (
	ErrInvalidToken error = errors.New("invalid token")
)

type UserClaims struct {
	ID string `json:"id"`
	jwt.Claims
}

func NewJWTokenService(secret string) *JWTokenService {
	return &JWTokenService{
		secret: []byte(secret),
	}
}

func (ts *JWTokenService) Generate(uid string) (token string, err error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
		ID: uid,
		Claims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(90 * (24 * time.Hour)).Unix(),
			Issuer:    "fpart",
			NotBefore: time.Now().Add(20 * time.Second).Unix(),
		},
	})
	return t.SignedString(ts.secret)
}

func (ts *JWTokenService) Validate(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&UserClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return ts.secret, nil
		},
	)

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
