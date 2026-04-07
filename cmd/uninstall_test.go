package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTildifyPath(t *testing.T) {
	home := "/home/user"

	tests := []struct {
		path string
		want string
	}{
		{"/home/user/.zshrc", "~/.zshrc"},
		{"/home/user/.config/fish/config.fish", "~/.config/fish/config.fish"},
		{"/etc/bashrc", "/etc/bashrc"},
		{"/home/user", "~"},
	}

	for _, tt := range tests {
		got := tildifyPath(home, tt.path)
		if got != tt.want {
			t.Errorf("tildifyPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestFileContains(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.zshrc")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	f.WriteString("export FOO=bar\n. \"/home/user/.sexyvoice/env.sh\"\nexport BAZ=qux\n")

	if !fileContains(f.Name(), `. "/home/user/.sexyvoice/env.sh"`) {
		t.Error("expected fileContains to return true for present line")
	}
	if fileContains(f.Name(), `. "/home/user/.sexyvoice/env.sh" `+" extra") {
		t.Error("expected fileContains to return false for absent line")
	}
	if fileContains("/nonexistent/file", "anything") {
		t.Error("expected fileContains to return false for nonexistent file")
	}
}

func TestRemoveLineFromFile(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		substr string
		want   string
	}{
		{
			name:   "removes source line",
			input:  "export FOO=bar\n. \"/home/user/.sexyvoice/env.sh\"\nexport BAZ=qux\n",
			substr: `. "/home/user/.sexyvoice/env.sh"`,
			want:   "export FOO=bar\nexport BAZ=qux\n",
		},
		{
			name:   "removes source line and preceding sexyvoice comment",
			input:  "export FOO=bar\n# sexyvoice\n. \"/home/user/.sexyvoice/env.sh\"\nexport BAZ=qux\n",
			substr: `. "/home/user/.sexyvoice/env.sh"`,
			want:   "export FOO=bar\nexport BAZ=qux\n",
		},
		{
			name:   "removes source line, sexyvoice comment, and preceding blank line",
			input:  "export FOO=bar\n\n# sexyvoice\n. \"/home/user/.sexyvoice/env.sh\"\nexport BAZ=qux\n",
			substr: `. "/home/user/.sexyvoice/env.sh"`,
			want:   "export FOO=bar\nexport BAZ=qux\n",
		},
		{
			name:   "no change when substr not present",
			input:  "export FOO=bar\nexport BAZ=qux\n",
			substr: `. "/home/user/.sexyvoice/env.sh"`,
			want:   "export FOO=bar\nexport BAZ=qux\n",
		},
		{
			name:   "does not remove unrelated # sexyvoice comment",
			input:  "# sexyvoice\nexport FOO=bar\n. \"/home/user/.sexyvoice/env.sh\"\n",
			substr: `. "/home/user/.sexyvoice/env.sh"`,
			want:   "# sexyvoice\nexport FOO=bar\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), ".zshrc")
			if err := os.WriteFile(path, []byte(tt.input), 0644); err != nil {
				t.Fatal(err)
			}

			if err := removeLineFromFile(path, tt.substr); err != nil {
				t.Fatalf("removeLineFromFile returned error: %v", err)
			}

			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			if string(got) != tt.want {
				t.Errorf("got:\n%q\nwant:\n%q", string(got), tt.want)
			}
		})
	}
}
