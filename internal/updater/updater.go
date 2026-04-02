package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rimelabs/rime-cli/internal/config"
)

const (
	rimeCLIGitHubAPI = "https://api.github.com/repos/rimelabs/rime-cli"
	cacheFileName    = "update_check.json"
	cacheTTL         = 24 * time.Hour
	fetchTimeout     = 5 * time.Second
)

type cacheEntry struct {
	LatestVersion string    `json:"latest_version"`
	CheckedAt     time.Time `json:"checked_at"`
}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// FetchLatestVersion fetches the latest release version from GitHub.
func FetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: fetchTimeout}
	req, err := http.NewRequest("GET", rimeCLIGitHubAPI+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	if release.TagName == "" {
		return "", fmt.Errorf("could not determine latest version")
	}

	return release.TagName, nil
}

func cacheFilePath() (string, error) {
	dir, err := config.CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, cacheFileName), nil
}

func readCache() (*cacheEntry, error) {
	path, err := cacheFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func writeCache(entry cacheEntry) error {
	path, err := cacheFilePath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// CachedLatestVersion returns the cached latest version, or "" if the cache is
// missing, stale, or unreadable. Never returns an error — update checks are best-effort.
func CachedLatestVersion() string {
	entry, err := readCache()
	if err != nil || entry == nil {
		return ""
	}
	if time.Since(entry.CheckedAt) > cacheTTL {
		return ""
	}
	return entry.LatestVersion
}

// RefreshInBackground fires a goroutine that fetches the latest version from
// GitHub and writes it to the cache. Returns immediately. The goroutine may not
// complete before the process exits — this is intentional; the result will be
// available on the next invocation.
func RefreshInBackground() {
	go func() {
		entry, err := readCache()
		if err == nil && entry != nil && time.Since(entry.CheckedAt) < cacheTTL {
			return // cache is fresh
		}

		version, err := FetchLatestVersion()
		if err != nil {
			return // network unavailable or API error — fail silently
		}

		_ = writeCache(cacheEntry{
			LatestVersion: version,
			CheckedAt:     time.Now().UTC(),
		})
	}()
}
