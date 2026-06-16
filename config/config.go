// Package config loads Asana credentials from the environment and an optional
// YAML config file (~/.config/asana-cli.yaml). Environment variables take
// precedence over file values; the file lets users persist credentials without
// exporting env vars on every shell.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the resolved Asana credentials.
type Config struct {
	AccessToken      string
	DefaultWorkspace string
}

// fileConfig mirrors the on-disk YAML structure of the config file.
type fileConfig struct {
	AccessToken      string `yaml:"access_token"`
	DefaultWorkspace string `yaml:"default_workspace"`
}

// Getenv is the signature of a single-variable environment lookup, allowing
// tests to inject a fake environment.
type Getenv func(string) string

// trimmed returns the trimmed value of name, or "" when unset/blank.
func trimmed(get Getenv, name string) string {
	return strings.TrimSpace(get(name))
}

// firstNonEmpty returns the first argument that is non-empty after trimming.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

// ConfigPath returns the resolved config file path: $ASANA_CONFIG when set,
// else $XDG_CONFIG_HOME/asana-cli.yaml, else ~/.config/asana-cli.yaml. It
// returns "" only when the home directory cannot be determined.
func ConfigPath(get Getenv) string {
	if p := trimmed(get, "ASANA_CONFIG"); p != "" {
		return p
	}
	if dir := trimmed(get, "XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "asana-cli.yaml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "asana-cli.yaml")
}

// loadFile reads and parses the config file at path. A missing file is not an
// error and yields a zero fileConfig.
func loadFile(path string) (fileConfig, error) {
	if path == "" {
		return fileConfig{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fileConfig{}, nil
		}
		return fileConfig{}, fmt.Errorf("read config %s: %w", path, err)
	}
	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		return fileConfig{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	return fc, nil
}

// Load resolves credentials from the config file and process environment.
// Environment variables (ASANA_ACCESS_TOKEN, ASANA_DEFAULT_WORKSPACE) take
// precedence over the corresponding file values.
func Load() (Config, error) {
	fc, err := loadFile(ConfigPath(os.Getenv))
	if err != nil {
		return Config{}, err
	}
	return resolve(os.Getenv, fc)
}

// LoadFrom resolves credentials from the given environment only (no config
// file). It is retained for callers and tests that exercise env-only behavior.
func LoadFrom(get Getenv) (Config, error) {
	return resolve(get, fileConfig{})
}

// resolve merges environment values (highest precedence) over file values.
func resolve(get Getenv, fc fileConfig) (Config, error) {
	token := firstNonEmpty(trimmed(get, "ASANA_ACCESS_TOKEN"), fc.AccessToken)
	if token == "" {
		return Config{}, errors.New("Set ASANA_ACCESS_TOKEN or add access_token to ~/.config/asana-cli.yaml before using Asana tools.")
	}
	return Config{
		AccessToken:      token,
		DefaultWorkspace: firstNonEmpty(trimmed(get, "ASANA_DEFAULT_WORKSPACE"), fc.DefaultWorkspace),
	}, nil
}

// ResolveWorkspace returns the explicit workspace GID when non-empty, else the
// configured default. It errors when neither is available.
func (c Config) ResolveWorkspace(explicit string) (string, error) {
	if ws := strings.TrimSpace(explicit); ws != "" {
		return ws, nil
	}
	if c.DefaultWorkspace != "" {
		return c.DefaultWorkspace, nil
	}
	return "", errors.New("Provide workspace_gid or set ASANA_DEFAULT_WORKSPACE.")
}
