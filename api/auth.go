package api

import (
	"errors"
	"fmt"
	"github.com/evilsocket/pwngrid/models"
	"net/http"
	"os"
	"strconv"
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

func CreateTokenFor(unit *models.Unit) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["unit_id"] = unit.ID
	claims["unit_ident"] = unit.Identity
	claims["expires_at"] = time.Now().Add(TokenTTL)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("API_SECRET")))
}

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

func ExtractTokenID(r *http.Request) (uint32, error) {

	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("API_SECRET")), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		uid, err := strconv.ParseUint(fmt.Sprintf("%.0f", claims["user_id"]), 10, 32)
		if err != nil {
			return 0, err
		}
		return uint32(uid), nil
	}
	return 0, nil
}
