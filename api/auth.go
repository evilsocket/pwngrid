package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	TokenTTL        = time.Minute * 30
	ErrTokenClaims  = errors.New("can't extract claims from jwt token")
	ErrTokenInvalid = errors.New("jwt token not valid")
	ErrTokenExpired = errors.New("jwt token expired")
)

func ValidateToken(r *http.Request) (jwt.MapClaims, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("API_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); !ok {
		return nil, ErrTokenClaims
	} else if !token.Valid {
		return nil, ErrTokenInvalid
	} else if claims["expires_at"].(time.Time).Before(time.Now()) {
		return nil, ErrTokenExpired
	} else {
		return claims, nil
	}
}

func ExtractToken(r *http.Request) string {
	keys := r.URL.Query()
	token := keys.Get("token")
	if token != "" {
		return token
	}
	bearerToken := r.Header.Get("Authorization")
	if parts := strings.Split(bearerToken, " "); len(parts) == 2 {
		return parts[1]
	}
	return ""
}
