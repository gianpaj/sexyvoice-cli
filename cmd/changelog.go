package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rimelabs/rime-cli/internal/output/styles"
)

// ChangelogContent is set by main via NewRootCmd.
var ChangelogContent string

func NewChangelogCmd() *cobra.Command {
	var since string

	cmd := &cobra.Command{
		Use:   "changelog",
		Short: "Show the Rime CLI changelog",
		Long:  "Displays the changelog, or only entries newer than a given version with --since.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChangelog(since)
		},
	}

	cmd.Flags().StringVar(&since, "since", "", "Only show versions newer than this version (e.g. v0.4.0)")

	return cmd
}

func runChangelog(since string) error {
	if Quiet || JSONOutput {
		return nil
	}

	sections := parseChangelog(ChangelogContent)
	if since != "" {
		sections = filterChangelogSince(sections, since)
	}

	printChangelog(sections)
	return nil
}

type changelogSection struct {
	version string // "Unreleased", "v0.4.0", etc.
	date    string // "2026-03-12" or "" for Unreleased
	lines   []string
}

func parseChangelog(content string) []changelogSection {
	var sections []changelogSection
	var current *changelogSection

	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "## [") {
			if current != nil {
				sections = append(sections, *current)
			}
			version, date := parseVersionHeader(line)
			current = &changelogSection{version: version, date: date}
		} else if current != nil {
			current.lines = append(current.lines, line)
		}
	}

	if current != nil {
		sections = append(sections, *current)
	}

	return sections
}

func parseVersionHeader(line string) (version, date string) {
	// "## [Unreleased]" or "## [v0.4.0] - 2026-03-12"
	inner := strings.TrimPrefix(line, "## [")
	parts := strings.SplitN(inner, "]", 2)
	if len(parts) == 0 {
		return "", ""
	}
	version = parts[0]
	if len(parts) == 2 {
		rest := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(parts[1]), "-"))
		date = rest
	}
	return version, date
}

func filterChangelogSince(sections []changelogSection, since string) []changelogSection {
	var result []changelogSection
	for _, s := range sections {
		if s.version == "Unreleased" || compareVersions(s.version, since) > 0 {
			result = append(result, s)
		}
	}
	return result
}

// compareVersions compares two semver strings (with or without leading "v").
// Returns -1, 0, or 1.
func compareVersions(a, b string) int {
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")

	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	max := len(aParts)
	if len(bParts) > max {
		max = len(bParts)
	}

	for i := 0; i < max; i++ {
		var aNum, bNum int
		if i < len(aParts) {
			aNum, _ = strconv.Atoi(aParts[i])
		}
		if i < len(bParts) {
			bNum, _ = strconv.Atoi(bParts[i])
		}
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
	}
	return 0
}

func hasBullets(s changelogSection) bool {
	for _, line := range s.lines {
		if strings.HasPrefix(line, "- ") {
			return true
		}
	}
	return false
}

func highlightCode(s string) string {
	parts := strings.Split(s, "`")
	for i := 1; i < len(parts); i += 2 {
		parts[i] = styles.Info(parts[i])
	}
	return strings.Join(parts, "")
}

func printChangelog(sections []changelogSection) {
	for _, s := range sections {
		if s.version == "Unreleased" && !hasBullets(s) {
			continue
		}
		if s.date != "" {
			fmt.Printf("%s  %s\n", s.version, styles.Dim(s.date))
		} else {
			fmt.Println(s.version)
		}

		for _, line := range s.lines {
			line = strings.TrimRight(line, " ")
			switch {
			case line == "":
				// skip blank lines within sections
			case strings.HasPrefix(line, "### "):
				fmt.Println("  " + strings.ToUpper(strings.TrimPrefix(line, "### ")))
			case strings.HasPrefix(line, "- "):
				fmt.Printf("  %s %s\n", styles.Dim("·"), highlightCode(strings.TrimPrefix(line, "- ")))
			}
		}
		fmt.Println()
	}
}
