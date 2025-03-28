package auth

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func NewMiddleware(publicKey interface{}) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get JWT token from cookie
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				http.Error(w, "Unauthorized - No token found", http.StatusUnauthorized)
				return
			}

			// Remove "ghsso_" prefix
			tokenString := cookie.Value[6:] // Skip "ghsso_" prefix

			// Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return publicKey, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
				return
			}

			// Verify path claim matches current path
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Unauthorized - Invalid claims", http.StatusUnauthorized)
				return
			}

			if claims["path"] != r.URL.Path {
				http.Error(w, "Unauthorized - Invalid path", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}
