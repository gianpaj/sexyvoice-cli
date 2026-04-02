package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/rimelabs/rime-cli/internal/output/styles"
	"github.com/rimelabs/rime-cli/internal/updater"
)

const rimeCLIGitHub = "https://github.com/rimelabs/rime-cli"

func NewUpgradeCmd() *cobra.Command {
	var targetVersion string

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the Rime CLI to the latest version",
		Long:  "Downloads and installs the latest version of the Rime CLI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(targetVersion)
		},
	}

	cmd.Flags().StringVar(&targetVersion, "version", "", "Version to install (e.g. v1.0.0). Defaults to latest.")

	return cmd
}

func runUpdate(targetVersion string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}

	if isHomebrewInstall(execPath) {
		fmt.Println(styles.Info("Rime was installed via Homebrew."))
		fmt.Println(styles.Dim("To update, run: brew upgrade rime"))
		return nil
	}

	target, err := updatePlatformTarget()
	if err != nil {
		return err
	}

	if targetVersion == "" {
		fmt.Println(styles.Dim("Checking for updates..."))
		latest, err := updater.FetchLatestVersion()
		if err != nil {
			return fmt.Errorf("could not check latest version: %w", err)
		}
		targetVersion = latest
	}

	current := Version
	if current != "" && current != "dev" &&
		strings.TrimPrefix(current, "v") == strings.TrimPrefix(targetVersion, "v") {
		fmt.Println(styles.Success(fmt.Sprintf("Already up to date (%s)", current)))
		return nil
	}

	downloadURL := fmt.Sprintf(
		"%s/releases/download/%s/rime-%s.tar.gz",
		rimeCLIGitHub, targetVersion, target,
	)

	fmt.Printf("%s %s\n", styles.Dim("Downloading"), targetVersion)

	tmpDir, err := os.MkdirTemp("", "rime-update.*")
	if err != nil {
		return fmt.Errorf("could not create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tarball := filepath.Join(tmpDir, "rime.tar.gz")
	if err := downloadToFile(tarball, downloadURL); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	newBin := filepath.Join(tmpDir, "rime")
	if err := extractRimeBinary(tarball, newBin); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	if err := os.Chmod(newBin, 0755); err != nil {
		return fmt.Errorf("could not set permissions: %w", err)
	}

	if runtime.GOOS == "darwin" {
		_ = exec.Command("xattr", "-d", "com.apple.quarantine", newBin).Run()
	}

	if err := os.Rename(newBin, execPath); err != nil {
		// Rename can fail across filesystems; fall back to copy.
		if err := copyFileContents(newBin, execPath); err != nil {
			return fmt.Errorf("could not replace binary: %w", err)
		}
	}

	if current != "" && current != "dev" {
		fmt.Println(styles.Success(fmt.Sprintf("Updated %s → %s", current, targetVersion)))
		// Exec into the new binary to show what changed. syscall.Exec replaces
		// the current process, so the new binary's embedded changelog is used.
		if err := syscall.Exec(execPath, []string{"rime", "changelog", "--since", current}, os.Environ()); err != nil {
			fmt.Println(styles.Dim(fmt.Sprintf("Run `rime changelog --since %s` to see what's new", current)))
		}
	} else {
		fmt.Println(styles.Success(fmt.Sprintf("Updated to %s", targetVersion)))
	}

	return nil
}

func updatePlatformTarget() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			return "darwin-amd64", nil
		case "arm64":
			return "darwin-arm64", nil
		}
	case "linux":
		if runtime.GOARCH == "amd64" {
			return "linux-amd64", nil
		}
	}
	return "", fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
}

func downloadToFile(dest, url string) error {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d downloading from %s", resp.StatusCode, url)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func extractRimeBinary(tarball, dest string) error {
	f, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag == tar.TypeReg && filepath.Base(hdr.Name) == "rime" {
			out, err := os.Create(dest)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(out, tr)
			closeErr := out.Close()
			if copyErr != nil {
				return copyErr
			}
			return closeErr
		}
	}

	return fmt.Errorf("rime binary not found in archive")
}

func copyFileContents(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// Preserve original permissions, fall back to 0755.
	perm := os.FileMode(0755)
	if fi, err := os.Stat(dst); err == nil {
		perm = fi.Mode()
	}

	// Write to a temp file in the same directory so the final rename is
	// atomic on the same filesystem.
	tmp, err := os.CreateTemp(filepath.Dir(dst), "."+filepath.Base(dst)+".tmp")
	if err != nil {
		return err
	}

	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	if err := os.Chmod(tmp.Name(), perm); err != nil {
		os.Remove(tmp.Name())
		return err
	}

	return os.Rename(tmp.Name(), dst)
}
