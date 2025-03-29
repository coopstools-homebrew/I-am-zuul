package config

import (
	"fmt"
	"log"
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
	privateKey, err := loadKeyOrFile("PRIVATE_KEY", "PRIVATE_KEY_FILE")
	if err != nil {
		return nil, err
	}

	publicKey, err := loadKeyOrFile("PUBLIC_KEY", "PUBLIC_KEY_FILE")
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

func loadKeyOrFile(keyName, keyFileName string) (string, error) {
	key := os.Getenv(keyName)
	if key != "" {
		log.Printf("Using %s from environment variable", keyName)
		return key, nil
	}

	log.Printf("%s not found; trying %s from file", keyName, keyFileName)
	keyPath := os.Getenv(keyFileName)
	if keyPath == "" {
		return "", fmt.Errorf("%s environment variable not set", keyFileName)
	}

	bytekey, err := os.ReadFile(keyPath)
	if err != nil {
		return "", err
	}

	return string(bytekey), nil
}
