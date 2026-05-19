package workflows

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCIWorkflowRunsRequiredChecks(t *testing.T) {
	workflow := readWorkflow(t, "ci.yml")

	assertContains(t, workflow, "on:", "CI workflow trigger")
	assertContains(t, workflow, "pull_request:", "pull request trigger")
	assertContains(t, workflow, "- main", "main branch push trigger")
	assertContains(t, workflow, "go-version-file: go.mod", "Go version pin")
	assertContains(t, workflow, "run: go test ./...", "Go test step")
	assertContains(t, workflow, "run: go build -o bin/codex-auth ./cmd/codex-auth", "CLI build step")
}

func TestReleaseWorkflowBuildsAuditableTagArtifacts(t *testing.T) {
	workflow := readWorkflow(t, "release.yml")

	assertContains(t, workflow, `- "v*.*.*"`, "semver tag trigger")
	assertContains(t, workflow, "workflow_dispatch:", "manual release dry-run trigger")
	assertContains(t, workflow, "permissions:", "workflow permissions")
	assertContains(t, workflow, "contents: write", "release publishing permission")
	assertContains(t, workflow, "id-token: write", "npm provenance permission")
	assertContains(t, workflow, "run: go test ./...", "pre-release test step")
	assertContains(t, workflow, "GOOS: ${{ matrix.goos }}", "cross-platform GOOS matrix")
	assertContains(t, workflow, "GOARCH: ${{ matrix.goarch }}", "cross-platform GOARCH matrix")
	assertContains(t, workflow, "go build -trimpath", "reproducible release build")
	assertContains(t, workflow, "main.version=${GITHUB_REF_NAME}", "version injection")
	assertContains(t, workflow, "shasum -a 256", "checksum generation")
	assertContains(t, workflow, "sigstore/cosign-installer@v4.1.0", "Cosign installer")
	assertContains(t, workflow, "cosign sign-blob", "release binary signing")
	assertContains(t, workflow, "--bundle \"dist/${{ matrix.artifact }}.sigstore.json\"", "Sigstore bundle output")
	assertContains(t, workflow, "actions/upload-artifact@v4", "artifact upload")
	assertContains(t, workflow, "if: startsWith(github.ref, 'refs/tags/')", "tag-only release publishing")
	assertContains(t, workflow, "softprops/action-gh-release@v2", "GitHub Release publishing")
	assertContains(t, workflow, "needs: binaries", "npm package validation waits for binaries")
	assertContains(t, workflow, "actions/setup-node@v4", "Node setup for npm package validation")
	assertContains(t, workflow, "actions/download-artifact@v4", "release binary download")
	assertContains(t, workflow, "merge-multiple: true", "downloaded binaries are staged together")
	assertContains(t, workflow, "npm run stage:npm-binaries", "npm binary staging")
	assertContains(t, workflow, "npm pack --dry-run", "npm package dry run")
	assertContainsCount(t, workflow, "if: startsWith(github.ref, 'refs/tags/')", 6, "tag-only signing and publishing gates")
	assertContains(t, workflow, "npm publish \"$package_dir\" --access public --provenance", "platform npm package publishing")
	assertContains(t, workflow, "run: npm publish --access public --provenance", "root npm package publishing")
	assertContains(t, workflow, "NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}", "npm publish token")
	assertContains(t, workflow, "name: Homebrew tap release", "Homebrew release job")
	assertContains(t, workflow, "repository: shayyz-code/homebrew-tap", "Homebrew tap repository")
	assertContains(t, workflow, "token: ${{ secrets.HOMEBREW_TAP_TOKEN }}", "Homebrew tap token")
	assertContains(t, workflow, "npm run generate:homebrew-formula", "Homebrew formula generation")
	assertContains(t, workflow, "Formula/codex-auth.rb", "Homebrew formula path")
	assertContains(t, workflow, "git push", "Homebrew tap push")
}

func readWorkflow(t *testing.T, name string) string {
	t.Helper()

	workflowPath := filepath.Join("..", "..", ".github", "workflows", name)
	contents, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read %s: %v", workflowPath, err)
	}
	return string(contents)
}

func assertContains(t *testing.T, contents string, expected string, context string) {
	t.Helper()

	if !strings.Contains(contents, expected) {
		t.Fatalf("%s missing %q", context, expected)
	}
}

func assertContainsCount(t *testing.T, contents string, expected string, count int, context string) {
	t.Helper()

	if actual := strings.Count(contents, expected); actual != count {
		t.Fatalf("%s has %d occurrences of %q, want %d", context, actual, expected, count)
	}
}
