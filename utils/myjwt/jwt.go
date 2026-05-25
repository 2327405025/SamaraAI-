package myjwt

import (
	"SamaraAI/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(id int64, username string) (string, error) {
	cfg := config.Get()
	claims := Claims{
		ID:       id,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.Jwt.ExpireDuration) * time.Hour)),
			Issuer:    cfg.Jwt.Issuer,
			Subject:   cfg.Jwt.Subject,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Jwt.Key))
}

func ParseToken(token string) (string, bool) {
	claims := new(Claims)
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.Get().Jwt.Key), nil
	})
	if !t.Valid || err != nil || claims == nil {
		return "", false
	}
	return claims.Username, true
}
