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
	OrgName string `json:"org_name"`
	OrgID   string `json:"org_id"`
}

func getDummyData(w http.ResponseWriter, r *http.Request) {
	// Create dummy data
	data := DummyData{
		OrgName: "CoopsTools",
		OrgID:   "12345",
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
		log.Fatal(err)
	}

	githubCallback := auth.NewGitHubCallback(config.PrivateKey)
	authMiddleware := auth.NewMiddleware(config.PublicKey)
	http.HandleFunc("/callback", githubCallback.HandleGitHubCallback)
	http.HandleFunc("/data", authMiddleware(getDummyData))
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
