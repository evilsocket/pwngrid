package api

import (
	"net/http"
)

var (
	AllowedOrigin  = "*"
	AllowedHeaders = "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization"
	AllowedMethods = "POST, GET, OPTIONS, PUT, DELETE"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Frame-Options", "DENY")
		w.Header().Add("X-Content-Type-Options", "nosniff")
		w.Header().Add("X-XSS-Protection", "1; mode=block")
		w.Header().Add("Referrer-Policy", "same-origin")
		w.Header().Set("Access-Control-Allow-Origin", AllowedOrigin)
		w.Header().Add("Access-Control-Allow-Headers", AllowedHeaders)
		w.Header().Add("Access-Control-Allow-Methods", AllowedMethods)

		next.ServeHTTP(w, r)
	})
}

func CORSOptionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}