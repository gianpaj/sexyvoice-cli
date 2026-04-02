package cmd

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func makeTarGz(t *testing.T, entries map[string]string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	for name, content := range entries {
		hdr := &tar.Header{
			Name:     name,
			Typeflag: tar.TypeReg,
			Size:     int64(len(content)),
			Mode:     0755,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	return f.Name()
}

func TestExtractRimeBinary(t *testing.T) {
	tests := []struct {
		name      string
		entries   map[string]string
		wantErr   bool
		wantBytes string
	}{
		{
			name:      "binary at archive root",
			entries:   map[string]string{"rime": "binary-content"},
			wantBytes: "binary-content",
		},
		{
			name:      "binary in subdirectory",
			entries:   map[string]string{"bin/rime": "nested-content"},
			wantBytes: "nested-content",
		},
		{
			name:    "binary missing from archive",
			entries: map[string]string{"README.md": "docs"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tarball := makeTarGz(t, tt.entries)
			dest := filepath.Join(t.TempDir(), "rime")

			err := extractRimeBinary(tarball, dest)
			if (err != nil) != tt.wantErr {
				t.Fatalf("extractRimeBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			got, err := os.ReadFile(dest)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.wantBytes {
				t.Errorf("got %q, want %q", string(got), tt.wantBytes)
			}
		})
	}
}
