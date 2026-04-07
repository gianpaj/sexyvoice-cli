package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gianpaj/sexyvoice-cli/internal/api"
	"github.com/gianpaj/sexyvoice-cli/internal/config"
)

type CurlOptions struct {
	Text       string
	Voice      string
	ModelID    string
	Format     string
	Oneline    bool
	APIURL     string
	AuthPrefix string

	// Legacy fields retained so older tests still compile.
	Speaker string
	Lang    string
}

func audioFormatToExt(format string) string {
	if format == "audio/mp3" || format == "audio/mpeg" || format == "mp3" {
		return "mp3"
	}
	return "wav"
}

func generateCurlCommand(opts CurlOptions, ttsOpts *api.TTSOptions) (string, error) {
	reqBody := api.TTSRequest{
		Model:          opts.ModelID,
		Input:          opts.Text,
		Voice:          opts.Voice,
		ResponseFormat: opts.Format,
		Style:          ttsOpts.Style,
		Seed:           ttsOpts.Seed,
	}

	var jsonBody []byte
	var err error
	if opts.Oneline {
		jsonBody, err = json.Marshal(reqBody)
	} else {
		jsonBody, err = json.MarshalIndent(reqBody, "", "  ")
	}
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := opts.APIURL
	if apiURL == "" {
		apiURL = api.GetAPIURL()
	}

	authPrefix := opts.AuthPrefix
	if authPrefix == "" {
		authPrefix = "Bearer"
	}

	contentType := api.GetAudioFormat(opts.ModelID)
	if opts.Format == "wav" {
		contentType = "audio/wav"
	} else if opts.Format == "mp3" {
		contentType = "audio/mp3"
	}
	outputFile := "output." + audioFormatToExt(contentType)

	var b strings.Builder
	jsonStr := strings.ReplaceAll(string(jsonBody), "'", "'\\''")

	if opts.Oneline {
		b.WriteString(fmt.Sprintf("curl -X POST '%s' -H 'Accept: application/json' -H \"Authorization: %s $(sexyvoice key)\" -H 'Content-Type: application/json' -o '%s' -f -d '%s'", apiURL, authPrefix, outputFile, jsonStr))
	} else {
		b.WriteString("curl --request POST \\\n")
		b.WriteString(fmt.Sprintf("  --url '%s' \\\n", apiURL))
		b.WriteString("  --header 'Accept: application/json' \\\n")
		b.WriteString(fmt.Sprintf("  --header \"Authorization: %s $(sexyvoice key)\" \\\n", authPrefix))
		b.WriteString("  --header 'Content-Type: application/json' \\\n")
		b.WriteString(fmt.Sprintf("  --output '%s' \\\n", outputFile))
		b.WriteString("  --fail \\\n")
		b.WriteString(fmt.Sprintf("  --data '%s'", jsonStr))
	}

	return b.String(), nil
}

func NewCurlCmd() *cobra.Command {
	var voice string
	var modelID string
	var format string
	var style string
	var seed int
	var hasSeed bool
	var oneline bool
	var apiURL string

	cmd := &cobra.Command{
		Use:   "curl TEXT",
		Short: "Generate curl command for speech request",
		Long: `Generate a curl command for making SexyVoice speech API requests.

Run without arguments to see an example:
  sexyvoice curl

For easy copy-paste (single line):
  sexyvoice curl --oneline

Or provide your own text:
  sexyvoice curl "your text here" --voice poe --model-id gpro`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := "Hello from SexyVoice"
			if len(args) > 0 {
				text = args[0]
				if voice == "" {
					return fmt.Errorf("--voice is required when providing text")
				}
				if modelID == "" {
					return fmt.Errorf("--model-id is required when providing text")
				}
			} else {
				if voice == "" {
					voice = "poe"
				}
				if modelID == "" {
					modelID = api.ModelIDGPro
				}
			}

			if !api.IsValidModelID(modelID) {
				return fmt.Errorf("invalid model-id: %s (valid options: %s)", modelID, strings.Join(api.ValidModelIDs(), ", "))
			}
			if format != "" && format != "wav" && format != "mp3" {
				return fmt.Errorf("unsupported format: %s (supported: wav, mp3)", format)
			}

			resolved, err := config.ResolveConfigWithOptions(config.ResolveOptions{
				EnvName:        ConfigEnv,
				APIURLOverride: apiURL,
				ConfigFile:     ConfigFile,
			})
			if err != nil {
				return err
			}

			ttsOpts := &api.TTSOptions{
				Speaker:     voice,
				ModelID:     modelID,
				AudioFormat: format,
			}
			if style != "" {
				ttsOpts.Style = &style
			}
			if hasSeed {
				ttsOpts.Seed = &seed
			}
			if err := api.ValidateModelParams(ttsOpts); err != nil {
				return err
			}

			curlCmd, err := generateCurlCommand(CurlOptions{
				Text:       text,
				Voice:      voice,
				ModelID:    modelID,
				Format:     format,
				Oneline:    oneline,
				APIURL:     resolved.APIURL,
				AuthPrefix: resolved.AuthHeaderPrefix,
			}, ttsOpts)
			if err != nil {
				return err
			}

			fmt.Println(curlCmd)
			return nil
		},
	}

	cmd.Flags().StringVarP(&voice, "voice", "v", "", "Voice name to use (required)")
	cmd.Flags().StringVarP(&voice, "speaker", "s", "", "Alias for --voice")
	cmd.Flags().MarkHidden("speaker")
	cmd.Flags().StringVarP(&modelID, "model-id", "m", "", fmt.Sprintf("Model ID (required, one of: %s)", strings.Join(api.ValidModelIDs(), ", ")))
	cmd.Flags().StringVarP(&format, "format", "f", "", "Response format: wav or mp3")
	cmd.Flags().StringVar(&style, "style", "", "Optional style prompt")
	cmd.Flags().IntVar(&seed, "seed", 0, "Optional deterministic seed")
	cmd.Flags().BoolVar(&oneline, "oneline", false, "Output as single line (easier to copy-paste)")
	cmd.Flags().StringVar(&apiURL, "api-url", "", "API URL (default: $SEXYVOICE_API_URL or https://sexyvoice.ai/api/v1/speech)")
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		hasSeed = cmd.Flags().Changed("seed")
	}

	return cmd
}
