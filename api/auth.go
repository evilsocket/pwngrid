package api

import (
	"errors"
	"fmt"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/pwngrid/models"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	ErrTokenClaims       = errors.New("can't extract claims from jwt token")
	ErrTokenInvalid      = errors.New("jwt token not valid")
	ErrTokenExpired      = errors.New("jwt token expired")
	ErrTokenIncomplete   = errors.New("jwt token is missing required fields")
	ErrTokenUnauthorized = errors.New("jwt token authorized field is false (?!)")
)

func validateToken(header string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(header, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("API_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrTokenClaims
	} else if !token.Valid {
		return nil, ErrTokenInvalid
	}

	required := []string{
		"expires_at",
		"authorized",
		"unit_id",
		"unit_ident",
	}
	for _, req := range required {
		if _, found := claims[req]; !found {
			return nil, ErrTokenIncomplete
		}
	}

	log.Debug("%+v", claims)

	if expiresAt, err := time.Parse(time.RFC3339, claims["expires_at"].(string)); err != nil{
		return nil, ErrTokenExpired
	} else if expiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	} else if claims["authorized"].(bool) != true {
		return nil, ErrTokenUnauthorized
	}
	return claims, err
}

func Authenticate(w http.ResponseWriter, r *http.Request) *models.Unit{
	client := clientIP(r)
	tokenHeader := reqToken(r)
	if tokenHeader == "" {
		log.Debug("unauthenticated request from %s", client)
		ERROR(w, http.StatusUnauthorized, ErrEmpty)
		return nil
	}

	claims, err := validateToken(tokenHeader)
	if err != nil {
		log.Warning("token error for %s: %v", client, err)
		ERROR(w, http.StatusUnauthorized, ErrEmpty)
		return nil
	}

	unit := models.FindUnit(nil, claims["user_id"].(uint32))
	if unit == nil {
		log.Warning("client %s authenticated with unknown claims '%v'", client, claims)
		ERROR(w, http.StatusUnauthorized, ErrEmpty)
		return nil
	}

	return unit
}
