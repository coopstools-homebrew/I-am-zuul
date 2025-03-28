package config

import (
	"os"
	"sync"
)

type Config struct {
	GitHubClientID     string
	GitHubClientSecret string
	GitHubCallbackURL  string
	PrivateKey         interface{}
	PublicKey          interface{}
}

func LoadConfig() (*Config, error) {
	privateKey, privateKeyErr := os.ReadFile("private.pem")
	if privateKeyErr != nil {
		return nil, privateKeyErr
	}
	publicKey, publicKeyErr := os.ReadFile("public.pem")
	if publicKeyErr != nil {
		return nil, publicKeyErr
	}

	var once sync.Once
	var config *Config
	var err error

	once.Do(func() {
		config = &Config{
			GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			GitHubCallbackURL:  os.Getenv("GITHUB_CALLBACK_URL"),
			PrivateKey:         privateKey,
			PublicKey:          publicKey,
		}
	})

	return config, err
}
