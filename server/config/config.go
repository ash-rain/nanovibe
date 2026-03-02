package config

import (
	"os"
	"path/filepath"
	"strings"
)

// Config holds all application configuration derived from environment variables.
type Config struct {
	Port     string
	Host     string
	AppEnv   string
	DataDir  string

	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURI  string
}

var global *Config

// Load parses environment variables and returns the Config singleton.
func Load() *Config {
	if global != nil {
		return global
	}

	dataDir := getEnv("DATA_DIR", "~/.vibecodepc/data")
	if strings.HasPrefix(dataDir, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			dataDir = filepath.Join(home, dataDir[2:])
		}
	}

	global = &Config{
		Port:    getEnv("PORT", "3000"),
		Host:    getEnv("HOST", "0.0.0.0"),
		AppEnv:  getEnv("APP_ENV", "development"),
		DataDir: dataDir,

		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURI:  getEnv("GITHUB_REDIRECT_URI", "http://localhost:3000/auth/github/callback"),
	}
	return global
}

// Get returns the global config, loading it if not yet loaded.
func Get() *Config {
	return Load()
}

// IsDevelopment returns true when APP_ENV is "development".
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
