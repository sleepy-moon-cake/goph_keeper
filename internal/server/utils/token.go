package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var ErrParseToken = errors.New("Parse error")

type Claims struct {
	jwt.RegisteredClaims
	UserName string
}

type TokenConfig struct {
	Name      string
	SecretKey string
}

func GenerateToken(userName string, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 12)),
		},
		UserName: userName,
	})

	signedToken, err := token.SignedString([]byte(secretKey))

	if err != nil {
		return "", fmt.Errorf("buildJWTString: %w", err)
	}

	return signedToken, nil
}

func ParseToken(jwtString string, secretKey string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(jwtString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return "", errors.Join(ErrParseToken, err)
	}

	if !token.Valid {
		return "", ErrParseToken
	}

	return claims.UserName, nil
}
