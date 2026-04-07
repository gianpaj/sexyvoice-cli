package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gianpaj/sexyvoice-cli/internal/api"
	"github.com/gianpaj/sexyvoice-cli/internal/audio/playback"
	"github.com/gianpaj/sexyvoice-cli/internal/config"
	"github.com/gianpaj/sexyvoice-cli/internal/tts"
)

func NewTTSCmd() *cobra.Command {
	var output string
	var play bool
	var voice string
	var modelID string
	var format string
	var style string
	var seed int
	var hasSeed bool
	var apiURL string

	cmd := &cobra.Command{
		Use:   "tts TEXT",
		Short: "Generate speech audio",
		Long:  "Generate speech with the SexyVoice /api/v1/speech API, optionally save it, and optionally play it locally.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := args[0]

			if !playback.IsPlaybackEnabled() && output == "" {
				return fmt.Errorf("output file required in headless build (use -o FILE or -o - for stdout)")
			}

			shouldPlay := playback.IsPlaybackEnabled() && (play || output == "")

			if voice == "" {
				return fmt.Errorf("--voice is required. You can use --voice poe")
			}
			if modelID == "" {
				return fmt.Errorf("--model-id is required. You can use --model-id %s", api.ModelIDGPro)
			}
			if !api.IsValidModelID(modelID) {
				return fmt.Errorf("invalid model-id: %s (valid options: %s)", modelID, strings.Join(api.ValidModelIDs(), ", "))
			}
			if format != "" && format != "wav" && format != "mp3" {
				return fmt.Errorf("unsupported format: %s (supported: wav, mp3)", format)
			}

			opts := &api.TTSOptions{
				Speaker:     voice,
				ModelID:     modelID,
				AudioFormat: format,
			}
			if style != "" {
				opts.Style = &style
			}
			if hasSeed {
				opts.Seed = &seed
			}
			if err := api.ValidateModelParams(opts); err != nil {
				return err
			}

			if output == "-" {
				resolved, err := config.ResolveConfigWithOptions(config.ResolveOptions{
					EnvName:        ConfigEnv,
					APIURLOverride: apiURL,
					ConfigFile:     ConfigFile,
				})
				if err != nil {
					return err
				}
				client := api.NewClient(api.ClientOptions{
					APIKey:           resolved.APIKey,
					APIURL:           resolved.APIURL,
					AuthHeaderPrefix: resolved.AuthHeaderPrefix,
					Version:          Version,
				})
				audioData, err := client.TTS(text, opts)
				if err != nil {
					return err
				}
				_, err = os.Stdout.Write(audioData)
				return err
			}

			return tts.RunNonInteractive(tts.RunOptions{
				Text:       text,
				TTSOptions: opts,
				Output:     output,
				Play:       shouldPlay,
				Quiet:      Quiet,
				JSON:       JSONOutput,
				Version:    Version,
				BaseURL:    apiURL,
				ConfigEnv:  ConfigEnv,
				ConfigFile: ConfigFile,
			})
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (or - for stdout)")
	if playback.IsPlaybackEnabled() {
		cmd.Flags().BoolVarP(&play, "play", "p", false, "Play audio (default if no output specified)")
	}
	cmd.Flags().StringVarP(&voice, "voice", "v", "", "Voice name to use (required)")
	cmd.Flags().StringVarP(&voice, "speaker", "s", "", "Alias for --voice")
	cmd.Flags().MarkHidden("speaker")
	cmd.Flags().StringVarP(&modelID, "model-id", "m", "", fmt.Sprintf("Model ID (required, one of: %s)", strings.Join(api.ValidModelIDs(), ", ")))
	cmd.Flags().StringVarP(&format, "format", "f", "", "Audio format: wav or mp3")
	cmd.Flags().StringVar(&style, "style", "", "Optional style prompt")
	cmd.Flags().IntVar(&seed, "seed", 0, "Optional deterministic seed")
	cmd.Flags().StringVar(&apiURL, "api-url", "", "API URL (default: $SEXYVOICE_API_URL or https://sexyvoice.ai/api/v1/speech)")
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		hasSeed = cmd.Flags().Changed("seed")
	}

	return cmd
}
