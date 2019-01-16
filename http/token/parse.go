package token

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

func Parse(publicKey []byte, rawToken string) (*Token, error) {
	jwtToken, err := jwt.Parse(rawToken, func(t *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if err != nil {
		return nil, errors.Errorf("couldn't parse JWT token ('%s'), err: %s", rawToken, err)
	}
	return (*Token)(jwtToken), nil
}
