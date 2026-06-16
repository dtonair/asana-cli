package config

import (
	"os"
	"path/filepath"
	"testing"
)

// envFunc builds a Getenv from a map.
func envFunc(m map[string]string) Getenv {
	return func(k string) string { return m[k] }
}

func TestLoadFrom(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		wantErr   bool
		wantToken string
		wantWS    string
	}{
		{
			name:      "token and workspace present",
			env:       map[string]string{"ASANA_ACCESS_TOKEN": "tok", "ASANA_DEFAULT_WORKSPACE": "ws1"},
			wantToken: "tok",
			wantWS:    "ws1",
		},
		{
			name:      "token only",
			env:       map[string]string{"ASANA_ACCESS_TOKEN": "tok"},
			wantToken: "tok",
		},
		{
			name:      "token trimmed",
			env:       map[string]string{"ASANA_ACCESS_TOKEN": "  tok  ", "ASANA_DEFAULT_WORKSPACE": "  ws  "},
			wantToken: "tok",
			wantWS:    "ws",
		},
		{
			name:    "missing token",
			env:     map[string]string{},
			wantErr: true,
		},
		{
			name:    "blank token",
			env:     map[string]string{"ASANA_ACCESS_TOKEN": "   "},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadFrom(envFunc(tt.env))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.AccessToken != tt.wantToken {
				t.Errorf("token = %q, want %q", cfg.AccessToken, tt.wantToken)
			}
			if cfg.DefaultWorkspace != tt.wantWS {
				t.Errorf("workspace = %q, want %q", cfg.DefaultWorkspace, tt.wantWS)
			}
		})
	}
}

func TestMissingTokenMessage(t *testing.T) {
	_, err := LoadFrom(envFunc(nil))
	if err == nil || err.Error() != "Set ASANA_ACCESS_TOKEN or add access_token to ~/.config/asana-cli.yaml before using Asana tools." {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigPath(t *testing.T) {
	if got := ConfigPath(envFunc(map[string]string{"ASANA_CONFIG": "/tmp/x.yaml"})); got != "/tmp/x.yaml" {
		t.Errorf("ASANA_CONFIG override: got %q", got)
	}
	got := ConfigPath(envFunc(map[string]string{"XDG_CONFIG_HOME": "/cfg"}))
	if want := filepath.Join("/cfg", "asana-cli.yaml"); got != want {
		t.Errorf("XDG_CONFIG_HOME: got %q want %q", got, want)
	}
}

// writeConfig writes a config file into a temp dir and returns its path.
func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "asana-cli.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestLoadReadsFile(t *testing.T) {
	path := writeConfig(t, "access_token: file-tok\ndefault_workspace: file-ws\n")
	t.Setenv("ASANA_CONFIG", path)
	t.Setenv("ASANA_ACCESS_TOKEN", "")
	t.Setenv("ASANA_DEFAULT_WORKSPACE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AccessToken != "file-tok" || cfg.DefaultWorkspace != "file-ws" {
		t.Fatalf("got %+v", cfg)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	path := writeConfig(t, "access_token: file-tok\ndefault_workspace: file-ws\n")
	t.Setenv("ASANA_CONFIG", path)
	t.Setenv("ASANA_ACCESS_TOKEN", "env-tok")
	t.Setenv("ASANA_DEFAULT_WORKSPACE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AccessToken != "env-tok" {
		t.Errorf("env should win for token, got %q", cfg.AccessToken)
	}
	if cfg.DefaultWorkspace != "file-ws" {
		t.Errorf("file workspace should be used, got %q", cfg.DefaultWorkspace)
	}
}

func TestLoadMissingFileIsNotError(t *testing.T) {
	t.Setenv("ASANA_CONFIG", filepath.Join(t.TempDir(), "absent.yaml"))
	t.Setenv("ASANA_ACCESS_TOKEN", "env-tok")
	t.Setenv("ASANA_DEFAULT_WORKSPACE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AccessToken != "env-tok" {
		t.Fatalf("got %+v", cfg)
	}
}

func TestLoadInvalidFileErrors(t *testing.T) {
	path := writeConfig(t, "access_token: [unterminated\n")
	t.Setenv("ASANA_CONFIG", path)
	t.Setenv("ASANA_ACCESS_TOKEN", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestResolveWorkspace(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		explicit string
		want     string
		wantErr  bool
	}{
		{name: "explicit wins", cfg: Config{DefaultWorkspace: "def"}, explicit: "exp", want: "exp"},
		{name: "explicit trimmed", cfg: Config{DefaultWorkspace: "def"}, explicit: "  exp  ", want: "exp"},
		{name: "falls back to default", cfg: Config{DefaultWorkspace: "def"}, explicit: "", want: "def"},
		{name: "blank explicit falls back", cfg: Config{DefaultWorkspace: "def"}, explicit: "   ", want: "def"},
		{name: "neither present", cfg: Config{}, explicit: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cfg.ResolveWorkspace(tt.explicit)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
