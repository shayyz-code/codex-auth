package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteRunsAccountWorkflowWithCodexHomeFlag(t *testing.T) {
	codexHome := t.TempDir()
	authPath := filepath.Join(codexHome, "auth.json")
	if err := os.WriteFile(authPath, []byte(`{"token":"alpha"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runCLI(t, nil, "--codex-home", codexHome, "save", "work")
	if code != 0 {
		t.Fatalf("save exit code = %d, stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, `Saved current Codex auth tokens as "work".`) {
		t.Fatalf("save stdout = %q", stdout)
	}

	stdout, stderr, code = runCLI(t, nil, "--codex-home", codexHome, "list")
	if code != 0 {
		t.Fatalf("list exit code = %d, stderr = %q", code, stderr)
	}
	if strings.TrimSpace(stdout) != "work" {
		t.Fatalf("list stdout = %q", stdout)
	}

	stdout, stderr, code = runCLI(t, nil, "--codex-home", codexHome, "use", "work")
	if code != 0 {
		t.Fatalf("use exit code = %d, stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, `Switched Codex auth to "work".`) {
		t.Fatalf("use stdout = %q", stdout)
	}

	stdout, stderr, code = runCLI(t, nil, "--codex-home", codexHome, "current")
	if code != 0 {
		t.Fatalf("current exit code = %d, stderr = %q", code, stderr)
	}
	if strings.TrimSpace(stdout) != "work" {
		t.Fatalf("current stdout = %q", stdout)
	}
}

func TestExecuteUsePromptsForAccount(t *testing.T) {
	codexHome := t.TempDir()
	if err := os.MkdirAll(filepath.Join(codexHome, "accounts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codexHome, "accounts", "personal.json"), []byte(`{"token":"personal"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codexHome, "accounts", "work.json"), []byte(`{"token":"work"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runCLI(t, strings.NewReader("2\n"), "--codex-home", codexHome, "use")
	if code != 0 {
		t.Fatalf("use prompt exit code = %d, stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, "Select account:") || !strings.Contains(stdout, `Switched Codex auth to "work".`) {
		t.Fatalf("use prompt stdout = %q", stdout)
	}
}

func TestExecuteReportsCommandFailures(t *testing.T) {
	_, stderr, code := runCLI(t, nil, "save")
	if code == 0 {
		t.Fatal("save without name succeeded")
	}
	if !strings.Contains(stderr, "accepts 1 arg(s), received 0") {
		t.Fatalf("stderr = %q", stderr)
	}

	_, stderr, code = runCLI(t, nil, "missing")
	if code == 0 {
		t.Fatal("unknown command succeeded")
	}
	if !strings.Contains(stderr, `unknown command "missing"`) {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestExecuteVersion(t *testing.T) {
	stdout, stderr, code := runCLI(t, nil, "--version")
	if code != 0 {
		t.Fatalf("version exit code = %d, stderr = %q", code, stderr)
	}
	if strings.TrimSpace(stdout) != "codex-su version test" {
		t.Fatalf("version stdout = %q", stdout)
	}
}

func runCLI(t *testing.T, stdin *strings.Reader, args ...string) (string, string, int) {
	t.Helper()
	var input bytes.Buffer
	if stdin != nil {
		if _, err := input.ReadFrom(stdin); err != nil {
			t.Fatal(err)
		}
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Execute("test", args, &input, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}
