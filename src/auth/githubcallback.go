package auth

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Persisted represents data that will eventually be stored in a database
type Persisted struct {
	AccessToken string
	UserID      string
	Username    string
}

var storage Persisted

// GitHubCallback handles the OAuth callback flow
type GitHubCallback struct {
	client     *http.Client
	privateKey interface{}
}

// NewGitHubCallback creates a new GitHubCallback handler
func NewGitHubCallback(privateKey interface{}) *GitHubCallback {
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
func (gh *GitHubCallback) getUserInfo(accessToken string) (*struct {
	ID    string `json:"id"`
	Login string `json:"login"`
}, error) {
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

	var userInfo struct {
		ID    string `json:"id"`
		Login string `json:"login"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// generateJWT creates a new JWT token with user claims
func (gh *GitHubCallback) generateJWT(userID, username string) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 2).Unix()
	claims["path"] = "/data"

	return token.SignedString(gh.privateKey)
}

func (gh *GitHubCallback) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	accessToken, err := gh.exchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}
	storage.AccessToken = accessToken

	// Get user info
	userInfo, err := gh.getUserInfo(accessToken)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	storage.UserID = userInfo.ID
	storage.Username = userInfo.Login

	// Generate JWT
	tokenString, err := gh.generateJWT(userInfo.ID, userInfo.Login)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set cookie
	cookie := &http.Cookie{
		Name:   "auth_token",
		Value:  "ghsso_" + tokenString,
		Path:   "/",
		Domain: "coopstools.com",
		MaxAge: 7200, // 2 hours in seconds
	}
	http.SetCookie(w, cookie)

	// Redirect
	http.Redirect(w, r, os.Getenv("GITHUB_REDIRECT_URI"), http.StatusMovedPermanently)
}
