package tts

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gianpaj/sexyvoice-cli/internal/api"
	"github.com/gianpaj/sexyvoice-cli/internal/config"
)

func TestRunNonInteractive_SaveToFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	audio := []byte("RIFF....WAVE")

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/speech":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"url": server.URL + "/audio.wav"})
		case "/audio.wav":
			w.Header().Set("Content-Type", "audio/wav")
			_, _ = w.Write(audio)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("SEXYVOICE_API_URL", server.URL+"/api/v1/speech")
	if err := config.SaveAPIKey("test-key"); err != nil {
		t.Fatalf("SaveAPIKey failed: %v", err)
	}

	outputFile := filepath.Join(t.TempDir(), "output.wav")
	err := RunNonInteractive(RunOptions{
		Text:       "hello",
		TTSOptions: &api.TTSOptions{Speaker: "poe", ModelID: api.ModelIDGPro},
		Output:     outputFile,
		Quiet:      true,
		Version:    "test",
	})
	if err != nil {
		t.Fatalf("RunNonInteractive failed: %v", err)
	}
}

func TestRunNonInteractive_MissingAPIKey(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	err := RunNonInteractive(RunOptions{
		Text:       "hello",
		TTSOptions: &api.TTSOptions{Speaker: "poe", ModelID: api.ModelIDGPro},
		Output:     filepath.Join(t.TempDir(), "output.wav"),
		Quiet:      true,
		Version:    "test",
	})
	if err == nil {
		t.Fatal("expected missing API key error")
	}
	if bytes.Contains([]byte(err.Error()), []byte("RIME")) {
		t.Fatalf("unexpected old branding in error: %v", err)
	}
}
