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
	assertContains(t, workflow, "run: go build -o bin/codex-su ./cmd/codex-su", "CLI build step")
}

func TestReleaseWorkflowBuildsAuditableTagArtifacts(t *testing.T) {
	workflow := readWorkflow(t, "release.yml")

	assertContains(t, workflow, `- "v*.*.*"`, "semver tag trigger")
	assertContains(t, workflow, "workflow_dispatch:", "manual release dry-run trigger")
	assertContains(t, workflow, "permissions:", "workflow permissions")
	assertContains(t, workflow, "contents: write", "release publishing permission")
	assertContains(t, workflow, "run: go test ./...", "pre-release test step")
	assertContains(t, workflow, "GOOS: ${{ matrix.goos }}", "cross-platform GOOS matrix")
	assertContains(t, workflow, "GOARCH: ${{ matrix.goarch }}", "cross-platform GOARCH matrix")
	assertContains(t, workflow, "go build -trimpath", "reproducible release build")
	assertContains(t, workflow, "main.version=${GITHUB_REF_NAME}", "version injection")
	assertContains(t, workflow, "shasum -a 256", "checksum generation")
	assertContains(t, workflow, "actions/upload-artifact@v4", "artifact upload")
	assertContains(t, workflow, "if: startsWith(github.ref, 'refs/tags/')", "tag-only release publishing")
	assertContains(t, workflow, "softprops/action-gh-release@v2", "GitHub Release publishing")
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
