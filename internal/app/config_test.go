package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		tomlData string
		wantErr  bool
		wantMain string
	}{
		{
			name:     "valid config",
			tomlData: `main = "/home/user/worktrees"`,
			wantErr:  false,
			wantMain: "/home/user/worktrees",
		},
		{
			name:     "empty config",
			tomlData: "",
			wantErr:  false,
			wantMain: "",
		},
		{
			name:     "invalid toml",
			tomlData: "main = ",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if tt.tomlData != "" {
				cfgPath := filepath.Join(tmpDir, "gowt.toml")
				os.WriteFile(cfgPath, []byte(tt.tomlData), 0o644)
			}

			cfg, err := LoadConfig(tmpDir)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if cfg.Main != tt.wantMain {
				t.Errorf("got main %q, want %q", cfg.Main, tt.wantMain)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{Main: "/home/user/worktrees"}

	err := SaveConfig(tmpDir, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfgPath := filepath.Join(tmpDir, "gowt.toml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	expected := `main = "/home/user/worktrees"
`
	if string(data) != expected {
		t.Errorf("got %q, want %q", string(data), expected)
	}

	cfg, err = LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}
	if cfg.Main != "/home/user/worktrees" {
		t.Errorf("got main %q, want %q", cfg.Main, "/home/user/worktrees")
	}
}

func TestConfigPath(t *testing.T) {
	gitCommonDir := "/home/user/project/.git"
	want := "/home/user/project/.git/gowt.toml"

	got := configPath(gitCommonDir)

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
