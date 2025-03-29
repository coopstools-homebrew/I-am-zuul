package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/coopstools-homebrew/I-am-zuul/src/auth"
	"github.com/coopstools-homebrew/I-am-zuul/src/config"
)

type DummyData struct {
	OrgName   string `json:"org_name"`
	OrgID     string `json:"org_id"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

func getDummyData(w http.ResponseWriter, r *http.Request) {
	// Create dummy data
	data := DummyData{
		OrgName:   "CoopsTools",
		OrgID:     "12345",
		AvatarURL: "https://avatars.githubusercontent.com/u/28509719?v=4",
		Email:     "coopstools@gmail.com",
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
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

func main() {
	if os.Getenv("GITHUB_CLIENT_ID") == "" || os.Getenv("GITHUB_CLIENT_SECRET") == "" {
		log.Fatal("GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET must be set")
	}

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	githubCallback := auth.NewGitHubCallback(config.PrivateKey)
	authMiddleware := auth.NewMiddleware(config.PublicKey)
	corsMiddleware := NewCORSMiddleware("https://gh.coopstools.com", "https://a28d-104-129-206-200.ngrok-free.app")

	http.HandleFunc("GET /generate-jwt", githubCallback.HandleGenerateJWT)
	http.HandleFunc("GET /callback", githubCallback.HandleGitHubCallback)
	http.HandleFunc("/data", corsMiddleware(authMiddleware(getDummyData)))

	log.Println("Server starting on :" + config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
