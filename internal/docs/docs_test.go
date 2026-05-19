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
	assertContains(t, readme, "@shayyz-code/codex-auth")
	assertContains(t, readme, "HOMEBREW_TAP_TOKEN")
	assertContains(t, readme, "shayyz-code/homebrew-tap")
	assertContains(t, readme, "v*.*.*")
	assertContains(t, readme, "make version VERSION=0.2.0")
	assertNotContains(t, readme, "pick interactively by number")
}

func TestREADMEDocumentsBadgesAndInstallation(t *testing.T) {
	readme := readREADME(t)

	assertContains(t, readme, "actions/workflows/ci.yml")
	assertContains(t, readme, "img.shields.io/github/actions/workflow/status/shayyz-code/codex-auth/ci.yml")
	assertContains(t, readme, "img.shields.io/github/v/release/shayyz-code/codex-auth")
	assertContains(t, readme, "img.shields.io/npm/v/%40shayyz-code%2Fcodex-auth")
	assertContains(t, readme, "img.shields.io/github/license/shayyz-code/codex-auth")
	assertContains(t, readme, "## Installation")
	assertContains(t, readme, "brew tap shayyz-code/tap")
	assertContains(t, readme, "brew install codex-auth")
	assertContains(t, readme, "npm install -g @shayyz-code/codex-auth")
	assertContains(t, readme, "https://github.com/shayyz-code/codex-auth/releases/latest")
	assertContains(t, readme, "go install github.com/shayyz-code/codex-auth/cmd/codex-auth@latest")
}

func TestContributingDocumentsChecksAndReleasePreparation(t *testing.T) {
	contributing := readProjectFile(t, "CONTRIBUTING.md")

	assertContains(t, contributing, "make check")
	assertContains(t, contributing, "git diff --check")
	assertContains(t, contributing, "make version VERSION=0.2.0")
	assertContains(t, contributing, "Update `README.md` and `CHANGELOG.md`")
	assertContains(t, contributing, "Use `--codex-home <path>` or `CODEX_HOME`")
}

func readREADME(t *testing.T) string {
	t.Helper()

	return readProjectFile(t, "README.md")
}

func readProjectFile(t *testing.T, name string) string {
	t.Helper()

	filePath := filepath.Join("..", "..", name)
	contents, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read %s: %v", filePath, err)
	}
	return string(contents)
}

func assertContains(t *testing.T, contents string, expected string) {
	t.Helper()

	if !strings.Contains(contents, expected) {
		t.Fatalf("README missing %q", expected)
	}
}

func assertNotContains(t *testing.T, contents string, unexpected string) {
	t.Helper()

	if strings.Contains(contents, unexpected) {
		t.Fatalf("README contains unexpected %q", unexpected)
	}
}
