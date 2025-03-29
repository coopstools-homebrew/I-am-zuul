package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"crypto/rsa"

	"github.com/golang-jwt/jwt/v5"
)

// Persisted represents data that will eventually be stored in a database
type Persisted struct {
	AccessToken string
	UserID      int32
	Username    string
}

var storage Persisted

// GitHubCallback handles the OAuth callback flow
type GitHubCallback struct {
	client     *http.Client
	privateKey *rsa.PrivateKey
}

type UserInfo struct {
	ID        int32  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

// NewGitHubCallback creates a new GitHubCallback handler
func NewGitHubCallback(privateKeyString string) *GitHubCallback {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyString))
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
		os.Exit(1)
	}
	return &GitHubCallback{
		client:     &http.Client{},
		privateKey: privateKey,
	}
}

// exchangeCodeForToken exchanges the OAuth code for an access token
func (gh *GitHubCallback) exchangeCodeForToken(code string) (string, error) {
	tokenURL := "https://github.com/login/oauth/access_token"
	req, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("client_id", os.Getenv("GITHUB_CLIENT_ID"))
	q.Add("client_secret", os.Getenv("GITHUB_CLIENT_SECRET"))
	q.Add("code", code)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/json")

	resp, err := gh.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

// getUserInfo fetches the GitHub user information using the access token
func (gh *GitHubCallback) getUserInfo(accessToken string) (*UserInfo, error) {
	userReq, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	userReq.Header.Set("Authorization", "Bearer "+accessToken)
	userReq.Header.Set("Accept", "application/json")

	userResp, err := gh.client.Do(userReq)
	if err != nil {
		return nil, err
	}
	defer userResp.Body.Close()

	var userInfo UserInfo
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// generateJWT creates a new JWT token with user claims
func (gh *GitHubCallback) generateJWT(userID int32, username, path string) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 2).Unix()
	claims["path"] = path

	return token.SignedString(gh.privateKey)
}

func (gh *GitHubCallback) HandleGenerateJWT(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		log.Printf("Failed to parse user_id: %v", err)
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}
	username := r.URL.Query().Get("username")

	token, err := gh.generateJWT(int32(userID), username, "/nowhere")
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(token))
}

func (gh *GitHubCallback) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Printf("Code not found in submitted form")
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	accessToken, err := gh.exchangeCodeForToken(code)
	if err != nil {
		log.Printf("Failed to get access token: %v", err)
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}
	storage.AccessToken = accessToken

	// Get user info
	userInfo, err := gh.getUserInfo(accessToken)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	storage.UserID = userInfo.ID
	storage.Username = userInfo.Login
	log.Printf("User onboarded: %s", userInfo.Login)

	// Generate JWT
	tokenString, err := gh.generateJWT(userInfo.ID, userInfo.Login, "/data")
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	redirectURI := os.Getenv("GITHUB_REDIRECT_URI")

	// Set cookie
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "ghsso_" + tokenString,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Domain:   redirectURI[8:],
		Path:     "/",
		MaxAge:   7200, // 2 hours in seconds
	}
	http.SetCookie(w, cookie)

	// Redirect
	http.Redirect(w, r, redirectURI, http.StatusFound)
}
