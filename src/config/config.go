package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

type Config struct {
	Port string

	LoremIpsumAccessToken string
	LoremIpsumRepo        string
	LoremIpsumBranch      string
	LoremIpsumPath        string

	GitHubClientID     string
	GitHubClientSecret string
	GitHubCallbackURL  string
	PrivateKey         string
	PublicKey          string

	DatabaseURL      string
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string

	AllowedOrigins []string
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
			LoremIpsumAccessToken: os.Getenv("LOREM_IPSUM_ACCESS_TOKEN"),
			LoremIpsumRepo:        os.Getenv("LOREM_IPSUM_REPO"),
			LoremIpsumBranch:      os.Getenv("LOREM_IPSUM_BRANCH"),
			LoremIpsumPath:        os.Getenv("LOREM_IPSUM_PATH"),
			GitHubClientID:        os.Getenv("GITHUB_CLIENT_ID"),
			GitHubClientSecret:    os.Getenv("GITHUB_CLIENT_SECRET"),
			GitHubCallbackURL:     os.Getenv("GITHUB_CALLBACK_URL"),
			PrivateKey:            privateKey,
			PublicKey:             publicKey,
			AllowedOrigins:        strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
			Port:                  os.Getenv("PORT"),
			DatabaseURL:           os.Getenv("DATABASE_URL"),
			DatabaseName:          os.Getenv("DATABASE_NAME"),
			DatabaseUser:          os.Getenv("DATABASE_USER"),
			DatabasePassword:      os.Getenv("DATABASE_PASSWORD"),
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

	log.Printf("%s not found; loading from file %s", keyName, keyFileName)
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
