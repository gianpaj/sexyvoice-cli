package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gianpaj/sexyvoice-cli/internal/config"
)

func TestResolveModelIDForVoice_UsesExplicitModel(t *testing.T) {
	got, err := resolveModelIDForVoice("poe", "gpro", "")
	if err != nil {
		t.Fatalf("resolveModelIDForVoice returned error: %v", err)
	}
	if got != "gpro" {
		t.Fatalf("expected gpro, got %s", got)
	}
}

func TestResolveModelIDForVoice_InferByVoiceID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := config.SaveAPIKey("test-key"); err != nil {
		t.Fatalf("SaveAPIKey failed: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/voices" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":             "e55e4fc7-c140-46eb-82da-4d8c97f03d95",
					"name":           "tara",
					"language":       "en-US",
					"model":          "orpheus",
					"formats":        []string{"mp3"},
					"supports_style": false,
				},
			},
		})
	}))
	defer server.Close()

	got, err := resolveModelIDForVoice("e55e4fc7-c140-46eb-82da-4d8c97f03d95", "", server.URL+"/api/v1/speech")
	if err != nil {
		t.Fatalf("resolveModelIDForVoice returned error: %v", err)
	}
	if got != "orpheus" {
		t.Fatalf("expected orpheus, got %s", got)
	}
}

func TestResolveModelIDForVoice_InferByUniqueVoiceName(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := config.SaveAPIKey("test-key"); err != nil {
		t.Fatalf("SaveAPIKey failed: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":             "voice-1",
					"name":           "poe",
					"language":       "multiple",
					"model":          "gpro",
					"formats":        []string{"wav"},
					"supports_style": true,
				},
			},
		})
	}))
	defer server.Close()

	got, err := resolveModelIDForVoice("poe", "", server.URL+"/api/v1/speech")
	if err != nil {
		t.Fatalf("resolveModelIDForVoice returned error: %v", err)
	}
	if got != "gpro" {
		t.Fatalf("expected gpro, got %s", got)
	}
}
