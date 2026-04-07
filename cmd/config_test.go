package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gianpaj/sexyvoice-cli/internal/config"
)

func setupConfigHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path, err := config.ConfigFilePath()
	if err != nil {
		t.Fatalf("ConfigFilePath failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	return path
}

func TestConfigAddCmd_UsesSexyVoiceDefaultURL(t *testing.T) {
	setupConfigHome(t)
	writeConfig(t, `api_key = "global"`+"\n")

	cmd := NewConfigAddCmd()
	cmd.SetArgs([]string{"staging", "--key", "test-key"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config add failed: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	env := cfg.Env["staging"]
	if env.APIURL != "https://sexyvoice.ai/api/v1/speech" {
		t.Fatalf("expected SexyVoice default URL, got %q", env.APIURL)
	}
}

func TestConfigInitCmd_WritesSexyVoiceDefaults(t *testing.T) {
	setupConfigHome(t)

	cmd := NewConfigInitCmd()
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.DefaultEnv != "production" {
		t.Fatalf("expected production default env, got %q", cfg.DefaultEnv)
	}
	if cfg.Env["production"].APIURL != "https://sexyvoice.ai/api/v1/speech" {
		t.Fatalf("unexpected production URL: %q", cfg.Env["production"].APIURL)
	}
}

func TestConfigAddCmd_RejectsDefaultName(t *testing.T) {
	setupConfigHome(t)
	cmd := NewConfigAddCmd()
	cmd.SetArgs([]string{"default", "--url", "https://sexyvoice.ai/api/v1/speech", "--key", "k"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for default env name")
	}
}
