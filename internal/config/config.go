// Package config reads and writes the CLI's local config file: the stored API key.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Credential is a typed record rather than a bare string so a future credential kind (e.g. an
// OAuth token pair) is a second variant of this shape, not a breaking migration for everyone
// already storing an API key.
type Credential struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// CredentialTypeAPIKey is the only Credential.Type this version writes or understands.
const CredentialTypeAPIKey = "api_key"

// Config is the on-disk shape of the config file.
type Config struct {
	Credential *Credential `json:"credential,omitempty"`
}

// APIKey returns the stored API key, or "" if none is stored (or the stored credential isn't
// an API key — not possible yet, but Load never rejects a config file over it).
func (c *Config) APIKey() string {
	if c == nil || c.Credential == nil || c.Credential.Type != CredentialTypeAPIKey {
		return ""
	}
	return c.Credential.Value
}

// Path returns the config file's location: the OS-appropriate config directory, same tier as
// `gh`'s ~/.config/gh/hosts.yml or `aws`'s ~/.aws/credentials — a plain file, not the OS
// keychain.
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locating config directory: %w", err)
	}
	return filepath.Join(dir, "qrocodile", "config.json"), nil
}

// Load reads the config file. A missing file is not an error — it means nothing is stored yet.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &cfg, nil
}

// Save writes the config file with 0600 permissions, creating its directory if needed.
func Save(cfg *Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}
