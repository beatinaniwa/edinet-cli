package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_SubscriptionKeyFromEnv(t *testing.T) {
	t.Setenv("EDINET_API_KEY", "test-key-123")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.SubscriptionKey != "test-key-123" {
		t.Errorf("SubscriptionKey = %q, want %q", cfg.SubscriptionKey, "test-key-123")
	}
}

func TestLoad_SubscriptionKeyEmpty(t *testing.T) {
	t.Setenv("EDINET_API_KEY", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.SubscriptionKey != "" {
		t.Errorf("SubscriptionKey = %q, want empty", cfg.SubscriptionKey)
	}
}

func TestLoad_ConfigDirFromEnv(t *testing.T) {
	t.Setenv("EDINET_CONFIG_DIR", "/tmp/test-config")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ConfigDir != "/tmp/test-config" {
		t.Errorf("ConfigDir = %q, want %q", cfg.ConfigDir, "/tmp/test-config")
	}
}

func TestLoad_CacheDirFromEnv(t *testing.T) {
	t.Setenv("EDINET_CACHE_DIR", "/tmp/test-cache")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CacheDir != "/tmp/test-cache" {
		t.Errorf("CacheDir = %q, want %q", cfg.CacheDir, "/tmp/test-cache")
	}
}

func TestLoad_DefaultConfigDir(t *testing.T) {
	t.Setenv("EDINET_CONFIG_DIR", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	userConfigDir, _ := os.UserConfigDir()
	expected := filepath.Join(userConfigDir, "edinet-cli")
	if cfg.ConfigDir != expected {
		t.Errorf("ConfigDir = %q, want %q", cfg.ConfigDir, expected)
	}
}

func TestLoad_DefaultCacheDir(t *testing.T) {
	t.Setenv("EDINET_CACHE_DIR", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	userCacheDir, _ := os.UserCacheDir()
	expected := filepath.Join(userCacheDir, "edinet-cli")
	if cfg.CacheDir != expected {
		t.Errorf("CacheDir = %q, want %q", cfg.CacheDir, expected)
	}
}
