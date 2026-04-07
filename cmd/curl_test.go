package cmd

import (
	"strings"
	"testing"

	"github.com/gianpaj/sexyvoice-cli/internal/api"
)

func TestGenerateCurlCommand_Basic(t *testing.T) {
	cmd, err := generateCurlCommand(CurlOptions{
		Text:    "hello world",
		Voice:   "poe",
		ModelID: api.ModelIDGPro,
		APIURL:  "https://sexyvoice.ai/api/v1/speech",
	}, &api.TTSOptions{})
	if err != nil {
		t.Fatalf("generateCurlCommand failed: %v", err)
	}

	for _, fragment := range []string{
		`"voice": "poe"`,
		`"model": "gpro"`,
		"$(sexyvoice key)",
		"--output 'output.wav'",
	} {
		if !strings.Contains(cmd, fragment) {
			t.Fatalf("expected curl command to contain %q\n%s", fragment, cmd)
		}
	}
}

func TestGenerateCurlCommand_WithStyleAndSeed(t *testing.T) {
	style := "happy"
	seed := 42

	cmd, err := generateCurlCommand(CurlOptions{
		Text:    "hello",
		Voice:   "tara",
		ModelID: api.ModelIDGrok,
		Format:  "mp3",
		Oneline: true,
		APIURL:  "https://sexyvoice.ai/api/v1/speech",
	}, &api.TTSOptions{Style: &style, Seed: &seed})
	if err != nil {
		t.Fatalf("generateCurlCommand failed: %v", err)
	}

	for _, fragment := range []string{`"style":"happy"`, `"seed":42`, `-o 'output.mp3'`} {
		if !strings.Contains(cmd, fragment) {
			t.Fatalf("expected curl command to contain %q\n%s", fragment, cmd)
		}
	}
}

func TestCurlCmd_RequiresVoiceAndModelForText(t *testing.T) {
	cmd := NewCurlCmd()
	cmd.SetArgs([]string{"hello"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected validation error")
	}
}
