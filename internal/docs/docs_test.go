package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestREADMEDocumentsReleaseSecrets(t *testing.T) {
	readme := readREADME(t)

	assertContains(t, readme, "## Release Setup")
	assertContains(t, readme, "NPM_TOKEN")
	assertContains(t, readme, "HOMEBREW_TAP_TOKEN")
	assertContains(t, readme, "shayyz-code/homebrew-tap")
	assertContains(t, readme, "v*.*.*")
}

func TestREADMEDocumentsBadgesAndInstallation(t *testing.T) {
	readme := readREADME(t)

	assertContains(t, readme, "actions/workflows/ci.yml/badge.svg")
	assertContains(t, readme, "img.shields.io/github/v/release/shayyz-code/codex-su")
	assertContains(t, readme, "img.shields.io/npm/v/codex-su")
	assertContains(t, readme, "img.shields.io/github/license/shayyz-code/codex-su")
	assertContains(t, readme, "## Installation")
	assertContains(t, readme, "brew tap shayyz-code/tap")
	assertContains(t, readme, "brew install codex-su")
	assertContains(t, readme, "npm install -g codex-su")
	assertContains(t, readme, "https://github.com/shayyz-code/codex-su/releases/latest")
	assertContains(t, readme, "go install github.com/shayyz-code/codex-su/cmd/codex-su@latest")
}

func readREADME(t *testing.T) string {
	t.Helper()

	readmePath := filepath.Join("..", "..", "README.md")
	contents, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read %s: %v", readmePath, err)
	}
	return string(contents)
}

func assertContains(t *testing.T, contents string, expected string) {
	t.Helper()

	if !strings.Contains(contents, expected) {
		t.Fatalf("README missing %q", expected)
	}
}
