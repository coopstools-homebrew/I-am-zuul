package auth

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/coopstools-homebrew/I-am-zuul/src/utils"
	"github.com/golang-jwt/jwt/v5"
)

func NewMiddleware(publicKeyString string) func(http.HandlerFunc) http.HandlerFunc {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyString))
	if err != nil {
		log.Fatalf("Failed to parse public key: %v", err)
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Request received: %s %s", r.Method, r.URL.Path)
			// Get JWT token from cookie
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				log.Printf("No token found: %v", err)
				http.Error(w, "Bad Request - No token found", http.StatusBadRequest)
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
				log.Printf("Invalid token: %v", err)
				http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
				return
			}

			// Verify path claim matches current path
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				log.Printf("Invalid claims: %v", err)
				http.Error(w, "Unauthorized - Invalid claims", http.StatusUnauthorized)
				return
			}

			if claims["path"] != r.URL.Path {
				log.Printf("Invalid path: %v", claims["path"])
				http.Error(w, "Unauthorized - Invalid path", http.StatusUnauthorized)
				return
			}

			// pull user_id out of jwt claims and add it to request context
			userID := claims["user_id"].(int32)
			r = r.WithContext(context.WithValue(r.Context(), utils.UserIDKey, userID))

			next.ServeHTTP(w, r)
		}
	}
}
