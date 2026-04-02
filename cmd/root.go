package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/rimelabs/rime-cli/internal/output/styles"
	"github.com/rimelabs/rime-cli/internal/output/ui"
	"github.com/rimelabs/rime-cli/internal/updater"
)

var Quiet bool
var JSONOutput bool
var Version string
var ConfigEnv string
var ConfigFile string

func NewRootCmd(version, changelog string) *cobra.Command {
	Version = version
	ChangelogContent = changelog
	root := &cobra.Command{
		Use:           "rime",
		Short:         "Rime TTS CLI",
		Long:          "Command-line interface for Rime text-to-speech synthesis",
		Version:       version,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Help()
			if term.IsTerminal(int(os.Stdout.Fd())) {
				fmt.Println(ui.PaddedLogo())
			}
			return err
		},
	}

	root.PersistentFlags().BoolVarP(&Quiet, "quiet", "q", false, "Suppress non-essential output")
	root.PersistentFlags().BoolVar(&JSONOutput, "json", false, "Output results as JSON")
	root.PersistentFlags().StringVarP(&ConfigEnv, "env", "e", "", "Environment to use from config")
	root.PersistentFlags().StringVarP(&ConfigFile, "config", "c", "", "Path to config file")

	root.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if Quiet || JSONOutput {
			return
		}
		if isUpgradeCmd(cmd) {
			return
		}
		if Version == "" || Version == "dev" {
			return
		}
		updater.RefreshInBackground()
		latest := updater.CachedLatestVersion()
		if latest != "" && compareVersions(latest, Version) > 0 {
			fmt.Fprintf(os.Stderr, "\n%s\n",
				styles.Dim(fmt.Sprintf("Update available: %s → %s  ·  Run: rime upgrade", Version, latest)),
			)
		}
	}

	root.AddCommand(NewLoginCmd())
	root.AddCommand(NewLogoutCmd())
	root.AddCommand(NewCurlCmd())
	root.AddCommand(NewKeyCmd())
	root.AddCommand(NewTTSCmd())
	root.AddCommand(NewHelloCmd())
	root.AddCommand(NewPlayCmd())
	root.AddCommand(NewUninstallCmd())
	root.AddCommand(NewUpgradeCmd())
	root.AddCommand(NewChangelogCmd())
	root.AddCommand(NewConfigCmd())
	root.AddCommand(NewSpeedtestCmd())
	root.AddCommand(NewUsageCmd())

	return root
}

func isUpgradeCmd(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "upgrade" {
			return true
		}
	}
	return false
}
