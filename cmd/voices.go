package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/gianpaj/sexyvoice-cli/internal/api"
	"github.com/gianpaj/sexyvoice-cli/internal/config"
)

func NewVoicesCmd() *cobra.Command {
	var apiURL string

	cmd := &cobra.Command{
		Use:   "voices",
		Short: "List available voices",
		Long:  "List the voices available from the SexyVoice /api/v1/voices endpoint.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			voices, err := client.ListVoices()
			if err != nil {
				return err
			}

			sort.SliceStable(voices, func(i, j int) bool {
				if voices[i].Model != voices[j].Model {
					return voices[i].Model < voices[j].Model
				}
				return voices[i].Name < voices[j].Name
			})

			if JSONOutput {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"data": voices,
				})
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAME\tMODEL\tLANGUAGE\tFORMATS\tSTYLE\tID")
			for _, voice := range voices {
				style := "no"
				if voice.SupportsStyle {
					style = "yes"
				}
				_, _ = fmt.Fprintf(
					w,
					"%s\t%s\t%s\t%s\t%s\t%s\n",
					voice.Name,
					voice.Model,
					voice.Language,
					strings.Join(voice.Formats, ","),
					style,
					voice.ID,
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&apiURL, "api-url", "", "API URL (default: $SEXYVOICE_API_URL or https://sexyvoice.ai/api/v1/speech)")
	return cmd
}
