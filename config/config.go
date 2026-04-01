package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all values loaded from the .env file.
type Config struct {
	AppID  string
	Secret string

	// OAuth token endpoint (POST)
	Endpoint string

	// Pre-built authorization URL for authorization_code flow.
	// Must include redirect_uri as a query parameter — it is parsed automatically
	// to determine the local callback server port and token exchange redirect_uri.
	AuthURL string

	// Base URL for API calls after obtaining the access token
	APIBaseURL string

	// Optional: space-separated OAuth scopes for client_credentials flow
	Scopes string
}

var requiredKeys = []string{"APP_ID", "SECRET", "ENDPOINT"}

// Load reads the .env file at envPath and returns a validated Config.
// If envPath is the default ".env" and the file doesn't exist, it falls
// through to environment variables already set (useful in CI).
func Load(envPath string) (*Config, error) {
	if err := godotenv.Load(envPath); err != nil {
		if envPath != ".env" {
			return nil, fmt.Errorf("loading env file %q: %w", envPath, err)
		}
		// Default .env missing is acceptable; use existing env vars.
	}

	var missing []string
	for _, key := range requiredKeys {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	appID := os.Getenv("APP_ID")
	authURL := strings.ReplaceAll(os.Getenv("AUTH_URL"), "${APP_ID}", appID)

	return &Config{
		AppID:      appID,
		Secret:     os.Getenv("SECRET"),
		Endpoint:   os.Getenv("ENDPOINT"),
		AuthURL:    authURL,
		APIBaseURL: os.Getenv("API_BASE_URL"),
		Scopes:     os.Getenv("SCOPES"),
	}, nil
}
