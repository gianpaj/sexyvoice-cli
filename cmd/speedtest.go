package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/rimelabs/rime-cli/internal/api"
	"github.com/rimelabs/rime-cli/internal/config"
	"github.com/rimelabs/rime-cli/internal/output/styles"
)

type SpeedtestResult struct {
	Environment string        `json:"environment"`
	APIURL      string        `json:"api_url"`
	TTFB        time.Duration `json:"ttfb_ns"`
	TTFBMs      float64       `json:"ttfb_ms"`
	TTFBMinMs   *float64      `json:"ttfb_min_ms,omitempty"`
	TTFBMaxMs   *float64      `json:"ttfb_max_ms,omitempty"`
	Error       string        `json:"error,omitempty"`
}

func NewSpeedtestCmd() *cobra.Command {
	var modelID string
	var speaker string
	var lang string
	var text string
	var extraURLs []string
	var envFilter []string
	var runs int
	var timeout time.Duration
	var autoYes bool

	cmd := &cobra.Command{
		Use:   "speedtest",
		Short: "Measure TTFB for all configured endpoints",
		Long:  "Performs a TTS request against all configured endpoints and reports the time to first byte (TTFB) for each",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg *config.Config
			var err error

			if ConfigFile != "" {
				cfg, err = config.LoadConfigFromPath(ConfigFile)
			} else {
				cfg, err = config.LoadConfig()
			}
			if err != nil {
				return err
			}

			// Build the list of (name, env) entries to test.
			type envEntry struct {
				name string
				env  *config.Environment
				err  error
			}

			var entries []envEntry

			// Add config-based environments.
			// Include them when: --env is explicitly set, OR no --url flags were given
			// (i.e. --url alone means "only test these URLs").
			includeConfigEnvs := len(envFilter) > 0 || len(extraURLs) == 0
			if includeConfigEnvs {
				var names []string
				if len(envFilter) > 0 {
					names = envFilter
				} else if cfg != nil {
					names = cfg.ListEnvironments()
				}
				for _, name := range names {
					env, resolveErr := cfg.ResolveEnvironment(name)
					entries = append(entries, envEntry{name: name, env: env, err: resolveErr})
				}
			}

			// Add synthetic environments from --url flags.
			if len(extraURLs) > 0 {
				// Resolve default env to inherit API key (works even if cfg is nil).
				defaultEnv, _ := cfg.ResolveEnvironment("default")
				if defaultEnv.APIKey != nil && !Quiet {
					fmt.Fprintf(os.Stderr, "Note: sending API credentials to %d custom URL(s): %s\n",
						len(extraURLs), strings.Join(extraURLs, ", "))
				}
				for _, rawURL := range extraURLs {
					synth := &config.Environment{APIURL: rawURL}
					synth.APIKey = defaultEnv.APIKey
					synth.AuthHeaderPrefix = defaultEnv.AuthHeaderPrefix
					entries = append(entries, envEntry{name: rawURL, env: synth})
				}
			}

			if len(entries) == 0 {
				if cfg == nil {
					return fmt.Errorf("no config file found. Run 'rime config init' to create one")
				}
				return fmt.Errorf("no environments configured")
			}

			if text == "" {
				text = "running speedtest."
			}
			opts := &api.TTSOptions{
				Speaker: speaker,
				ModelID: modelID,
				Lang:    lang,
			}

			results := make([]SpeedtestResult, 0, len(entries))

			if runs < 1 {
				return fmt.Errorf("--runs must be at least 1")
			}

			ttfbHeader := "TTFB"
			if runs > 1 {
				ttfbHeader = fmt.Sprintf("TTFB (%d runs)", runs)
			}
			if !JSONOutput && !Quiet {
				fmt.Printf("%-15s %-50s %s\n", "ENV", "URL", ttfbHeader)
				fmt.Println(strings.Repeat("-", 80))
			}

			for _, entry := range entries {
				if entry.err != nil {
					result := SpeedtestResult{
						Environment: entry.name,
						Error:       entry.err.Error(),
					}
					results = append(results, result)
					if !JSONOutput && !Quiet {
						fmt.Printf("%-15s %-50s %s\n", entry.name, "(error)", styles.Error(entry.err.Error()))
					}
					continue
				}

				env := entry.env
				client := api.NewClient(api.ClientOptions{
					APIKey:           env.GetAPIKey(),
					APIURL:           env.APIURL,
					AuthHeaderPrefix: getAuthPrefix(env),
					Version:          Version,
					Timeout:          timeout,
				})

				var ttfbs []time.Duration
				var lastErr error
				for i := 0; i < runs; i++ {
					streamResult, err := client.TTSStream(text, opts)
					if err != nil {
						lastErr = err
						continue
					}
					streamResult.Body.Close()
					ttfbs = append(ttfbs, streamResult.TTFB)
				}

				if len(ttfbs) == 0 {
					result := SpeedtestResult{
						Environment: entry.name,
						APIURL:      env.APIURL,
						Error:       lastErr.Error(),
					}
					results = append(results, result)
					if !JSONOutput && !Quiet {
						fmt.Printf("%-15s %-50s %s\n", entry.name, truncateURL(env.APIURL, 50), styles.Error(lastErr.Error()))
					}
					continue
				}

				mean, minTTFB, maxTTFB := computeStats(ttfbs)
				result := SpeedtestResult{
					Environment: entry.name,
					APIURL:      env.APIURL,
					TTFB:        mean,
					TTFBMs:      float64(mean.Microseconds()) / 1000.0,
				}
				if runs > 1 {
					minMs := float64(minTTFB.Microseconds()) / 1000.0
					maxMs := float64(maxTTFB.Microseconds()) / 1000.0
					result.TTFBMinMs = &minMs
					result.TTFBMaxMs = &maxMs
				}
				results = append(results, result)

				if !JSONOutput && !Quiet {
					var ttfbStr string
					if runs > 1 {
						ttfbStr = fmt.Sprintf("mean=%-10s min=%-10s max=%s",
							formatTTFB(mean), formatTTFB(minTTFB), formatTTFB(maxTTFB))
					} else {
						ttfbStr = formatTTFB(mean)
					}
					fmt.Printf("%-15s %-50s %s\n", entry.name, truncateURL(env.APIURL, 50), styles.Success(ttfbStr))
				}
			}

			if JSONOutput {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(results)
			}

			fastest := findFastest(results)

			if !Quiet && fastest != nil {
				fmt.Printf("\n%s %s (%s)\n", styles.Success("Fastest:"), fastest.Environment, formatTTFB(fastest.TTFB))
			}

			if fastest != nil && !JSONOutput {
				currentDefault := "default"
				if cfg != nil && cfg.DefaultEnv != "" {
					currentDefault = cfg.DefaultEnv
				}

				switchDefault := func(name string) {
					if err := config.SaveDefaultEnv(name); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: could not switch default: %v\n", err)
					} else if !Quiet {
						fmt.Printf("%s\n", styles.Successf("Switched default to %q", name))
					}
				}

				if fastest.Environment != currentDefault && (autoYes || term.IsTerminal(int(os.Stdin.Fd()))) {
					if autoYes {
						switchDefault(fastest.Environment)
					} else if !Quiet {
						fmt.Printf("Switch default to %q? [y/N]: ", fastest.Environment)
						line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
						if strings.ToLower(strings.TrimSpace(line)) == "y" {
							switchDefault(fastest.Environment)
						}
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&modelID, "model", "m", "arcana", "Model ID for the test request")
	cmd.Flags().StringVarP(&speaker, "speaker", "s", "astra", "Speaker for the test request")
	cmd.Flags().StringVarP(&lang, "lang", "l", "eng", "Language code for the test request")
	cmd.Flags().StringVar(&text, "text", "", `Text to synthesize (default "running speedtest.")`)
	cmd.Flags().StringArrayVar(&extraURLs, "url", nil, "Additional URL to test (repeatable)")
	cmd.Flags().StringArrayVar(&envFilter, "env", nil, "Only test these named environments from config (repeatable)")
	cmd.Flags().IntVar(&runs, "runs", 1, "Number of requests per endpoint (reports mean/min/max when >1)")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "Per-request timeout (0 disables timeout)")
	cmd.Flags().BoolVarP(&autoYes, "yes", "y", false, "Automatically switch default to the fastest environment")

	return cmd
}

func getAuthPrefix(env *config.Environment) string {
	if env.AuthHeaderPrefix != nil {
		return *env.AuthHeaderPrefix
	}
	return "Bearer"
}

func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}

func formatTTFB(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Microseconds()))
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000.0)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func computeStats(ttfbs []time.Duration) (mean, min, max time.Duration) {
	min = ttfbs[0]
	max = ttfbs[0]
	var total time.Duration
	for _, t := range ttfbs {
		total += t
		if t < min {
			min = t
		}
		if t > max {
			max = t
		}
	}
	mean = total / time.Duration(len(ttfbs))
	return
}

func findFastest(results []SpeedtestResult) *SpeedtestResult {
	var fastest *SpeedtestResult
	for i := range results {
		if results[i].Error != "" || results[i].TTFB == 0 {
			continue
		}
		if fastest == nil || results[i].TTFB < fastest.TTFB {
			fastest = &results[i]
		}
	}
	return fastest
}
