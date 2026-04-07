package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gianpaj/sexyvoice-cli/internal/api"
	"github.com/gianpaj/sexyvoice-cli/internal/audio/metadata"
)

func TestTTS_Pipeline(t *testing.T) {
	audio := metadata.FixWavHeader([]byte("RIFF....WAVEfmt data"))

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

	client := api.NewClient(api.ClientOptions{
		APIKey:  "test-key",
		APIURL:  server.URL + "/api/v1/speech",
		Version: "test",
	})

	got, err := client.TTS("hello world", &api.TTSOptions{
		Speaker: "poe",
		ModelID: api.ModelIDGPro,
	})
	if err != nil {
		t.Fatalf("TTS failed: %v", err)
	}
	if len(got) == 0 || !bytes.Contains(got, []byte("WAVE")) {
		t.Fatalf("unexpected audio payload: %q", got)
	}
}
