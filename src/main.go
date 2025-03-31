package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/coopstools-homebrew/I-am-zuul/src/auth"
	"github.com/coopstools-homebrew/I-am-zuul/src/config"
	"github.com/coopstools-homebrew/I-am-zuul/src/persistence"
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

func main() {
	if os.Getenv("GITHUB_CLIENT_ID") == "" || os.Getenv("GITHUB_CLIENT_SECRET") == "" {
		log.Fatal("GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET must be set")
	}

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	err = persistence.Migrate(db)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	githubCallback := auth.NewGitHubCallback(config.PrivateKey)
	authMiddleware := auth.NewMiddleware(config.PublicKey)
	corsMiddleware := auth.NewCORSMiddleware(config.AllowedOrigins...)

	http.HandleFunc("GET /generate-jwt", githubCallback.HandleGenerateJWT)
	http.HandleFunc("GET /callback", githubCallback.HandleGitHubCallback)
	http.HandleFunc("/data", corsMiddleware(authMiddleware(getDummyData)))

	log.Println("Server starting on :" + config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
