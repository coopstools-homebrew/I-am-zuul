package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/coopstools-homebrew/I-am-zuul/src/auth"
	"github.com/coopstools-homebrew/I-am-zuul/src/config"
	"github.com/coopstools-homebrew/I-am-zuul/src/github"
	"github.com/coopstools-homebrew/I-am-zuul/src/persistence"
	"github.com/coopstools-homebrew/I-am-zuul/src/utils"
)

type DummyData struct {
	OrgName   string `json:"org_name"`
	OrgID     int32  `json:"org_id"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

type UserTable interface {
	GetUserByID(id int32) (*persistence.UserInfo, error)
}

func getDummyData(userTable UserTable) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(utils.UserIDKey).(int32)
		user, err := userTable.GetUserByID(userID)
		if err != nil {
			log.Printf("Failed to get user: %v", err)
			http.Error(w, "Failed to get user", http.StatusInternalServerError)
			return
		}
		// Create dummy data
		data := DummyData{
			OrgName:   user.LoginName,
			OrgID:     user.ID,
			AvatarURL: user.AvatarURL,
			Email:     user.Email,
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
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

	userTable := persistence.NewUserTable(db)

	authMiddleware := auth.NewMiddleware(config.PublicKey)
	corsMiddleware := auth.NewCORSMiddleware(config.AllowedOrigins...)

	dummyDataRetriever := corsMiddleware(authMiddleware(getDummyData(userTable)))
	githubCallback := auth.NewGitHubCallback(config.PrivateKey, userTable)
	appender := github.NewLoremIpsumAppender(config.LoremIpsumAccessToken)

	http.HandleFunc("GET /generate-jwt", githubCallback.HandleGenerateJWT)
	http.HandleFunc("GET /callback", githubCallback.HandleGitHubCallback)
	http.HandleFunc("/data", dummyDataRetriever)
	http.HandleFunc("/lorem-ipsum", func(w http.ResponseWriter, r *http.Request) {
		err := appender.AppendLoremIpsum(config.LoremIpsumRepo, config.LoremIpsumBranch, config.LoremIpsumPath)
		if err != nil {
			log.Printf("Failed to append lorem ipsum: %v", err)
		}
	})
	log.Println("Server starting on :" + config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
