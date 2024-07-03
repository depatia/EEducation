package jwt

import (
	"AuthService/internal/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrBadJWT     = errors.New("incorrect JWT")
	ErrJWTExpired = errors.New("JWT is expired")
)

type JwtWrapper struct {
	SecretKey       string
	Issuer          string
	ExpirationHours int64
}

type jwtClaims struct {
	jwt.RegisteredClaims
	Uid         int64
	Email       string
	Permissions int64
}

func (w *JwtWrapper) NewToken(user models.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["permissions"] = user.PermissionLevel
	claims["exp"] = time.Now().Local().Add(time.Hour * time.Duration(w.ExpirationHours)).Unix()
	claims["issuer"] = w.Issuer
	tokenString, err := token.SignedString([]byte(w.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (w *JwtWrapper) ValidateToken(signedToken string) (*jwtClaims, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&jwtClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(w.SecretKey), nil
		},
	)

	if err != nil {
		return nil, ErrBadJWT
	}

	claims, ok := token.Claims.(*jwtClaims)

	if !ok {
		return nil, errors.New("Couldn't parse claims")
	}

	return claims, nil
}
