package jwt

import (
	clienterrors "APIGateway/internal/client_errors"
	"APIGateway/internal/config"
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type jwtClaims struct {
	jwt.RegisteredClaims
	Uid         int64
	Email       string
	Permissions int64
}

var cfg, _ = config.LoadConfig()

func ValidateToken(signedToken string) (*jwtClaims, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&jwtClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		},
	)

	if err != nil {
		return nil, clienterrors.ErrBadJWT
	}

	claims, ok := token.Claims.(*jwtClaims)

	if !ok {
		return nil, errors.New("Couldn't parse claims")
	}

	return claims, nil
}
