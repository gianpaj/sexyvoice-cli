package cmd

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"v0.5.0", "v0.4.0", 1},
		{"v0.4.0", "v0.5.0", -1},
		{"v0.4.0", "v0.4.0", 0},
		{"0.5.0", "v0.4.0", 1},   // with/without leading v
		{"v0.4.1", "v0.4.0", 1},  // patch version
		{"v0.10.0", "v0.9.0", 1}, // double-digit minor
		{"v1.0.0", "v0.9.9", 1},  // major bump
	}

	for _, tt := range tests {
		got := compareVersions(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestParseChangelog(t *testing.T) {
	content := `# Changelog

## [Unreleased]

## [v0.5.0] - 2026-03-25

### Added
- upgrade command
- some other thing

## [v0.4.0] - 2026-03-12

### Changed
- simplified waveform
`
	sections := parseChangelog(content)
	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}

	if sections[0].version != "Unreleased" || sections[0].date != "" {
		t.Errorf("unexpected first section: %+v", sections[0])
	}
	if sections[1].version != "v0.5.0" || sections[1].date != "2026-03-25" {
		t.Errorf("unexpected second section: %+v", sections[1])
	}
	if sections[2].version != "v0.4.0" || sections[2].date != "2026-03-12" {
		t.Errorf("unexpected third section: %+v", sections[2])
	}
}

func TestFilterChangelogSince(t *testing.T) {
	sections := []changelogSection{
		{version: "Unreleased"},
		{version: "v0.5.0", date: "2026-03-25"},
		{version: "v0.4.0", date: "2026-03-12"},
		{version: "v0.3.0", date: "2026-03-04"},
	}

	got := filterChangelogSince(sections, "v0.4.0")
	if len(got) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(got))
	}
	if got[0].version != "Unreleased" {
		t.Errorf("expected Unreleased, got %q", got[0].version)
	}
	if got[1].version != "v0.5.0" {
		t.Errorf("expected v0.5.0, got %q", got[1].version)
	}
}

func TestHasBullets(t *testing.T) {
	withBullets := changelogSection{lines: []string{"", "- foo", "- bar"}}
	withoutBullets := changelogSection{lines: []string{"", "### Added", ""}}

	if !hasBullets(withBullets) {
		t.Error("expected hasBullets to return true")
	}
	if hasBullets(withoutBullets) {
		t.Error("expected hasBullets to return false")
	}
}
