package auth

import (
	"net/http"
)

func NewCORSMiddleware(allowedOrigins ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					break
				}
			}

			// Handle preflight OPTIONS request
			if r.Method == "OPTIONS" {
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}
