package config

import (
	"os"
	"path/filepath"
)

// Config holds the application configuration.
type Config struct {
	SubscriptionKey string // EDINET API key (from env only, never persisted)
	ConfigDir       string // Application config directory
	CacheDir        string // Application cache directory
	NoCache         bool   // Bypass local cache
	Debug           bool   // Enable debug output
	Format          string // Output format: "json" or "table"
}

// Load reads configuration from environment variables and system defaults.
func Load() (*Config, error) {
	cfg := &Config{
		SubscriptionKey: os.Getenv("EDINET_API_KEY"),
		Format:          "json",
	}

	if dir := os.Getenv("EDINET_CONFIG_DIR"); dir != "" {
		cfg.ConfigDir = dir
	} else {
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		cfg.ConfigDir = filepath.Join(userConfigDir, "edinet-cli")
	}

	if dir := os.Getenv("EDINET_CACHE_DIR"); dir != "" {
		cfg.CacheDir = dir
	} else {
		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			return nil, err
		}
		cfg.CacheDir = filepath.Join(userCacheDir, "edinet-cli")
	}

	return cfg, nil
}
