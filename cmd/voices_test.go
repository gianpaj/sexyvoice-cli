package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gianpaj/sexyvoice-cli/internal/config"
)

func TestVoicesCmd_TableOutput(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := config.SaveAPIKey("test-key"); err != nil {
		t.Fatalf("SaveAPIKey failed: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/voices" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
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

	var out bytes.Buffer
	cmd := NewVoicesCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--api-url", server.URL + "/api/v1/speech"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := out.String()
	for _, fragment := range []string{"NAME", "poe", "gpro", "multiple", "wav", "yes", "voice-1"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q\n%s", fragment, output)
		}
	}
}

func TestVoicesCmd_JSONOutput(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := config.SaveAPIKey("test-key"); err != nil {
		t.Fatalf("SaveAPIKey failed: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":             "voice-2",
					"name":           "tara",
					"language":       "en",
					"model":          "orpheus",
					"formats":        []string{"mp3"},
					"supports_style": false,
				},
			},
		})
	}))
	defer server.Close()

	oldJSONOutput := JSONOutput
	JSONOutput = true
	defer func() { JSONOutput = oldJSONOutput }()

	var out bytes.Buffer
	cmd := NewVoicesCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--api-url", server.URL + "/api/v1/speech"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var payload struct {
		Data []struct {
			ID            string   `json:"id"`
			Name          string   `json:"name"`
			Language      string   `json:"language"`
			Model         string   `json:"model"`
			Formats       []string `json:"formats"`
			SupportsStyle bool     `json:"supports_style"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	if len(payload.Data) != 1 || payload.Data[0].Name != "tara" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}
