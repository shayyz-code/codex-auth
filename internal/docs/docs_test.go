package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestREADMEDocumentsReleaseSecrets(t *testing.T) {
	readmePath := filepath.Join("..", "..", "README.md")
	contents, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read %s: %v", readmePath, err)
	}
	readme := string(contents)

	assertContains(t, readme, "## Release Setup")
	assertContains(t, readme, "NPM_TOKEN")
	assertContains(t, readme, "HOMEBREW_TAP_TOKEN")
	assertContains(t, readme, "shayyz-code/homebrew-tap")
	assertContains(t, readme, "v*.*.*")
}

func assertContains(t *testing.T, contents string, expected string) {
	t.Helper()

	if !strings.Contains(contents, expected) {
		t.Fatalf("README missing %q", expected)
	}
}
