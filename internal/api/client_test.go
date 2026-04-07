package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAPIURLAndDashboardURL(t *testing.T) {
	t.Setenv("SEXYVOICE_API_URL", "https://example.com/api/v1/speech")
	t.Setenv("SEXYVOICE_DASHBOARD_URL", "https://example.com")

	if got := GetAPIURL(); got != "https://example.com/api/v1/speech" {
		t.Fatalf("unexpected api url: %s", got)
	}
	if got := GetDashboardURL(); got != "https://example.com" {
		t.Fatalf("unexpected dashboard url: %s", got)
	}
}

func TestTTS_DownloadsAudioFromReturnedURL(t *testing.T) {
	audio := []byte("RIFF....WAVE")

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/speech":
			if r.Header.Get("Authorization") != "Bearer test-key" {
				t.Fatalf("unexpected auth header: %s", r.Header.Get("Authorization"))
			}
			var body TTSRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if body.Model != ModelIDGPro || body.Voice != "poe" || body.Input != "hello" {
				t.Fatalf("unexpected request body: %+v", body)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"url": server.URL + "/audio.wav",
			})
		case "/audio.wav":
			w.Header().Set("Content-Type", "audio/wav")
			_, _ = w.Write(audio)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		APIKey:  "test-key",
		APIURL:  server.URL + "/api/v1/speech",
		Version: "test",
	})
	got, err := client.TTS("hello", &TTSOptions{Speaker: "poe", ModelID: ModelIDGPro})
	if err != nil {
		t.Fatalf("TTS failed: %v", err)
	}
	if string(got) != string(audio) {
		t.Fatalf("unexpected audio payload: %q", got)
	}
}

func TestValidateAPIKey_UsesModelsEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/models" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		APIKey:  "test-key",
		APIURL:  server.URL + "/api/v1/speech",
		Version: "test",
	})
	if err := client.ValidateAPIKey(); err != nil {
		t.Fatalf("ValidateAPIKey failed: %v", err)
	}
}

func TestTTS_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"bad key"}}`))
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		APIKey:  "bad-key",
		APIURL:  server.URL,
		Version: "test",
	})
	_, err := client.TTS("hello", &TTSOptions{Speaker: "poe", ModelID: ModelIDGPro})
	if err == nil || err.Error() != "authentication failed: invalid API key" {
		t.Fatalf("unexpected error: %v", err)
	}
}
