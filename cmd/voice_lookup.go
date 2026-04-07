package cmd

import (
	"fmt"

	"github.com/gianpaj/sexyvoice-cli/internal/api"
	"github.com/gianpaj/sexyvoice-cli/internal/config"
)

func resolveModelIDForVoice(voice, modelID, apiURL string) (string, error) {
	if voice == "" {
		return "", fmt.Errorf("--voice is required. You can use --voice poe")
	}
	if modelID != "" {
		if !api.IsValidModelID(modelID) {
			return "", fmt.Errorf("invalid model-id: %s (valid options: %s)", modelID, joinValidModelIDs())
		}
		return modelID, nil
	}

	resolved, err := config.ResolveConfigWithOptions(config.ResolveOptions{
		EnvName:        ConfigEnv,
		APIURLOverride: apiURL,
		ConfigFile:     ConfigFile,
	})
	if err != nil {
		return "", err
	}

	client := api.NewClient(api.ClientOptions{
		APIKey:           resolved.APIKey,
		APIURL:           resolved.APIURL,
		AuthHeaderPrefix: resolved.AuthHeaderPrefix,
		Version:          Version,
	})

	voices, err := client.ListVoices()
	if err != nil {
		return "", fmt.Errorf("--model-id is required unless the CLI can infer it from /api/v1/voices: %w", err)
	}

	var matches []api.Voice
	for _, item := range voices {
		if item.ID == voice || item.Name == voice {
			matches = append(matches, item)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("--model-id is required when the voice cannot be resolved from /api/v1/voices")
	case 1:
		if !api.IsValidModelID(matches[0].Model) {
			return "", fmt.Errorf("voice %q resolved to unsupported model %q", voice, matches[0].Model)
		}
		return matches[0].Model, nil
	default:
		return "", fmt.Errorf("--model-id is required because voice %q matched multiple entries", voice)
	}
}

func joinValidModelIDs() string {
	return "gpro, grok, orpheus"
}
