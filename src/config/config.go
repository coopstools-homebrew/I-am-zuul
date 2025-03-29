package config

import (
	"fmt"
	"os"
	"sync"
)

type Config struct {
	GitHubClientID     string
	GitHubClientSecret string
	GitHubCallbackURL  string
	PrivateKey         string
	PublicKey          string
}

func LoadConfig() (*Config, error) {
	privateKey, err := loadKeyFile("PRIVATE_KEY")
	if err != nil {
		return nil, err
	}

	publicKey, err := loadKeyFile("PUBLIC_KEY")
	if err != nil {
		return nil, err
	}

	var once sync.Once
	var config *Config

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

func loadKeyFile(keyName string) (string, error) {
	keyPath := os.Getenv(keyName)
	if keyPath == "" {
		return "", fmt.Errorf("%s environment variable not set", keyName)
	}

	key, err := os.ReadFile(keyPath)
	if err != nil {
		return "", err
	}

	return string(key), nil
}
