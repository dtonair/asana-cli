package config

import "testing"

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
	if err == nil || err.Error() != "Set ASANA_ACCESS_TOKEN before using Asana tools." {
		t.Fatalf("unexpected error: %v", err)
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
