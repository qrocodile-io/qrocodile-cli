package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// withTempConfigDir points os.UserConfigDir() (via HOME/AppData/XDG_CONFIG_HOME, depending on
// OS) at a fresh temp directory, so tests never touch a real config file.
func withTempConfigDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("AppData", dir)
	case "darwin":
		t.Setenv("HOME", dir)
	default:
		t.Setenv("XDG_CONFIG_HOME", dir)
		t.Setenv("HOME", dir)
	}
}

func TestLoad_missingFileReturnsEmptyConfig(t *testing.T) {
	withTempConfigDir(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.APIKey() != "" {
		t.Errorf("APIKey() = %q, want empty", cfg.APIKey())
	}
}

func TestSaveAndLoad_roundTrip(t *testing.T) {
	withTempConfigDir(t)

	err := Save(&Config{Credential: &Credential{Type: CredentialTypeAPIKey, Value: "qk_live_abc"}})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := cfg.APIKey(); got != "qk_live_abc" {
		t.Errorf("APIKey() = %q, want qk_live_abc", got)
	}
}

func TestSave_permissionsAre0600(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits don't apply on Windows")
	}
	withTempConfigDir(t)

	if err := Save(&Config{Credential: &Credential{Type: CredentialTypeAPIKey, Value: "qk_live_abc"}}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	path, err := Path()
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("permissions = %o, want 600", perm)
	}
}

func TestAPIKey_nilCredentialIsEmpty(t *testing.T) {
	cfg := &Config{}
	if got := cfg.APIKey(); got != "" {
		t.Errorf("APIKey() = %q, want empty", got)
	}
}

func TestAPIKey_nilConfigIsEmpty(t *testing.T) {
	var cfg *Config
	if got := cfg.APIKey(); got != "" {
		t.Errorf("APIKey() = %q, want empty", got)
	}
}

func TestPath_underQrocodileDir(t *testing.T) {
	withTempConfigDir(t)
	path, err := Path()
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	if filepath.Base(filepath.Dir(path)) != "qrocodile" {
		t.Errorf("Path() = %q, want it under a qrocodile/ directory", path)
	}
}
