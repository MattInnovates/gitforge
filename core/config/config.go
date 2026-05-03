package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultWebPort     = 3000
	defaultGitHTTPPort = 8080
	defaultGitSSHPort  = 2222
)

// Config stores shared runtime settings loaded from environment variables.
type Config struct {
	WebPort             int
	GitHTTPPort         int
	GitSSHPort          int
	ReposPath           string
	DatabaseDriver      string
	DatabaseDSN         string
	AuthTokenSecret     string
	AuthTokenTTLMinutes int
}

// LoadFromEnv returns configuration using env vars with sane defaults.
func LoadFromEnv() (Config, error) {
	cfg := Config{
		WebPort:             defaultWebPort,
		GitHTTPPort:         defaultGitHTTPPort,
		GitSSHPort:          defaultGitSSHPort,
		ReposPath:           getEnvOrDefault("GITFORGE_REPOS_PATH", "./repos"),
		DatabaseDriver:      getEnvOrDefault("GITFORGE_DATABASE_DRIVER", "sqlite"),
		DatabaseDSN:         getEnvOrDefault("GITFORGE_DATABASE_DSN", "gitforge.db"),
		AuthTokenSecret:     getEnvOrDefault("GITFORGE_AUTH_TOKEN_SECRET", "change-me-gitforge-auth-secret"),
		AuthTokenTTLMinutes: 120,
	}

	var err error
	cfg.WebPort, err = getEnvIntOrDefault("GITFORGE_WEB_PORT", cfg.WebPort)
	if err != nil {
		return Config{}, err
	}

	cfg.GitHTTPPort, err = getEnvIntOrDefault("GITFORGE_GIT_HTTP_PORT", cfg.GitHTTPPort)
	if err != nil {
		return Config{}, err
	}

	cfg.GitSSHPort, err = getEnvIntOrDefault("GITFORGE_GIT_SSH_PORT", cfg.GitSSHPort)
	if err != nil {
		return Config{}, err
	}

	cfg.AuthTokenTTLMinutes, err = getEnvIntOrDefault("GITFORGE_AUTH_TOKEN_TTL_MINUTES", cfg.AuthTokenTTLMinutes)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func getEnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvIntOrDefault(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid integer for %s: %q", key, raw)
	}

	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", key)
	}

	return value, nil
}
