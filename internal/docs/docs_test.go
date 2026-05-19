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
	assertContains(t, readme, "make version VERSION=0.2.1")
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

func TestREADMEDocumentsBuildRequirementsAndFeatures(t *testing.T) {
	readme := readREADME(t)

	assertContains(t, readme, "These are only needed when building from source or running repository checks:")
	assertContains(t, readme, "No Go or Node.js installation is required when installing a prebuilt release")
	assertContains(t, readme, "## Features")
	assertContains(t, readme, "Switch accounts by exact name, detected email, or an arrow-key terminal picker.")
	assertContains(t, readme, "Suggest the closest saved account when a name is mistyped.")
	assertContains(t, readme, "Ask whether to save an unsaved live Codex login before switching away from it.")
}

func TestREADMEDocumentsScreenshotsAndNotices(t *testing.T) {
	readme := readREADME(t)

	assertContains(t, readme, "## Screenshots")
	assertContains(t, readme, "Accounts list table:")
	assertContains(t, readme, "See current account and interactive use command for account switching:")
	assertContains(t, readme, "github.com/user-attachments/assets/")
	assertContains(t, readme, "## Notice")
	assertContains(t, readme, "Treat files under `~/.codex` as sensitive credentials.")
	assertContains(t, readme, "Scoop, Winget, Arch AUR, Docker")
}

func TestContributingDocumentsChecksAndReleasePreparation(t *testing.T) {
	contributing := readProjectFile(t, "CONTRIBUTING.md")

	assertContains(t, contributing, "make check")
	assertContains(t, contributing, "git diff --check")
	assertContains(t, contributing, "make version VERSION=0.2.1")
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
