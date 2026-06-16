// Package config loads Asana credentials from the environment, mirroring the
// env-variable subset of the original Pi extension's configuration.
package config

import (
	"errors"
	"os"
	"strings"
)

// Config holds the resolved Asana credentials.
type Config struct {
	AccessToken      string
	DefaultWorkspace string
}

// Getenv is the signature of a single-variable environment lookup, allowing
// tests to inject a fake environment.
type Getenv func(string) string

// trimmed returns the trimmed value of name, or "" when unset/blank.
func trimmed(get Getenv, name string) string {
	return strings.TrimSpace(get(name))
}

// Load reads ASANA_ACCESS_TOKEN (required) and ASANA_DEFAULT_WORKSPACE
// (optional) using the process environment.
func Load() (Config, error) {
	return LoadFrom(os.Getenv)
}

// LoadFrom is Load with an injectable environment lookup.
func LoadFrom(get Getenv) (Config, error) {
	token := trimmed(get, "ASANA_ACCESS_TOKEN")
	if token == "" {
		return Config{}, errors.New("Set ASANA_ACCESS_TOKEN before using Asana tools.")
	}
	return Config{
		AccessToken:      token,
		DefaultWorkspace: trimmed(get, "ASANA_DEFAULT_WORKSPACE"),
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
