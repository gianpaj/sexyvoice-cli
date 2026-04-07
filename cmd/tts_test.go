package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gianpaj/sexyvoice-cli/internal/config"
)

func setupTTSServer(t *testing.T, audioURL string, audio []byte, contentType string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/speech":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if body["model"] != "gpro" {
				t.Fatalf("expected model gpro, got %v", body["model"])
			}
			if body["voice"] != "poe" {
				t.Fatalf("expected voice poe, got %v", body["voice"])
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"url": audioURL})
		case "/audio.wav":
			w.Header().Set("Content-Type", contentType)
			_, _ = w.Write(audio)
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestTTSCmd_Stdout(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	audio := []byte("RIFF....WAVE")

	var server *httptest.Server
	server = setupTTSServer(t, "", audio, "audio/wav")
	defer server.Close()

	audioURL := server.URL + "/audio.wav"
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/speech":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"url": audioURL})
		case "/audio.wav":
			w.Header().Set("Content-Type", "audio/wav")
			_, _ = w.Write(audio)
		default:
			http.NotFound(w, r)
		}
	})

	t.Setenv("SEXYVOICE_API_URL", server.URL+"/api/v1/speech")
	if err := config.SaveAPIKey("test-key"); err != nil {
		t.Fatalf("SaveAPIKey failed: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	cmd := NewTTSCmd()
	cmd.SetArgs([]string{"hello", "--voice", "poe", "--model-id", "gpro", "-o", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	if !bytes.Equal(buf.Bytes(), audio) {
		t.Fatalf("unexpected stdout audio: %q", buf.Bytes())
	}
}

func TestTTSCmd_RequiresVoiceAndModel(t *testing.T) {
	cmd := NewTTSCmd()
	if err := cmd.RunE(cmd, []string{"hello"}); err == nil {
		t.Fatal("expected validation error")
	} else if !strings.Contains(err.Error(), "--voice") {
		t.Fatalf("expected voice error, got %v", err)
	}
}
