package cli

import (
	"bytes"
	"encoding/base64"
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
	if !strings.Contains(stdout, "│ *      │ work │ -     │") {
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

	stdout, stderr, code := runCLI(t, strings.NewReader("work\n"), "--codex-home", codexHome, "use")
	if code != 0 {
		t.Fatalf("use prompt exit code = %d, stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, "Select account:") || !strings.Contains(stdout, `Switched Codex auth to "work".`) {
		t.Fatalf("use prompt stdout = %q", stdout)
	}
}

func TestExecuteUsePromptDoesNotAcceptNumbers(t *testing.T) {
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
	if code == 0 {
		t.Fatalf("use prompt unexpectedly accepted numeric selection, stdout = %q", stdout)
	}
	if strings.Contains(stdout, "1)") || strings.Contains(stdout, "2)") {
		t.Fatalf("use prompt rendered numeric choices, stdout = %q", stdout)
	}
	if !strings.Contains(stderr, "No account selected") {
		t.Fatalf("use prompt stderr = %q", stderr)
	}
}

func TestExecuteUsePromptShowsEmailAndAcceptsEmail(t *testing.T) {
	codexHome := t.TempDir()
	if err := os.MkdirAll(filepath.Join(codexHome, "accounts"), 0o700); err != nil {
		t.Fatal(err)
	}
	auth := `{"tokens":{"id_token":"` + testJWT(`{"email":"work@example.com"}`) + `"}}`
	if err := os.WriteFile(filepath.Join(codexHome, "accounts", "work.json"), []byte(auth), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runCLI(t, strings.NewReader("work@example.com\n"), "--codex-home", codexHome, "use")
	if code != 0 {
		t.Fatalf("use prompt exit code = %d, stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, "work <work@example.com>") || !strings.Contains(stdout, `Switched Codex auth to "work".`) {
		t.Fatalf("use prompt stdout = %q", stdout)
	}
}

func TestExecuteRenameShowsSavedNameAndEmail(t *testing.T) {
	codexHome := t.TempDir()
	if err := os.MkdirAll(filepath.Join(codexHome, "accounts"), 0o700); err != nil {
		t.Fatal(err)
	}
	auth := `{"tokens":{"id_token":"` + testJWT(`{"email":"work@example.com"}`) + `"}}`
	if err := os.WriteFile(filepath.Join(codexHome, "accounts", "work.json"), []byte(auth), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runCLI(t, nil, "--codex-home", codexHome, "rename", "work", "office")
	if code != 0 {
		t.Fatalf("rename exit code = %d, stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, `Renamed "work" to office (email: work@example.com).`) {
		t.Fatalf("rename stdout = %q", stdout)
	}
	if _, err := os.Stat(filepath.Join(codexHome, "accounts", "office.json")); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteRenamePromptsWithEmail(t *testing.T) {
	codexHome := t.TempDir()
	if err := os.MkdirAll(filepath.Join(codexHome, "accounts"), 0o700); err != nil {
		t.Fatal(err)
	}
	auth := `{"tokens":{"id_token":"` + testJWT(`{"email":"work@example.com"}`) + `"}}`
	if err := os.WriteFile(filepath.Join(codexHome, "accounts", "work.json"), []byte(auth), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runCLI(t, strings.NewReader("work@example.com\noffice\n"), "--codex-home", codexHome, "rename")
	if code != 0 {
		t.Fatalf("rename prompt exit code = %d, stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, "work <work@example.com>") || !strings.Contains(stdout, "New saved name:") {
		t.Fatalf("rename prompt stdout = %q", stdout)
	}
	if _, err := os.Stat(filepath.Join(codexHome, "accounts", "office.json")); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteCanForceColorizedListOutput(t *testing.T) {
	codexHome := t.TempDir()
	if err := os.MkdirAll(filepath.Join(codexHome, "accounts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codexHome, "accounts", "work.json"), []byte(`{"token":"work"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runCLI(t, nil, "--codex-home", codexHome, "--color", "always", "list")
	if code != 0 {
		t.Fatalf("list exit code = %d, stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, "\x1b[") || !strings.Contains(stdout, "Saved Codex accounts") {
		t.Fatalf("colorized list stdout = %q", stdout)
	}
}

func TestExecuteListRendersAccountTable(t *testing.T) {
	codexHome := t.TempDir()
	if err := os.MkdirAll(filepath.Join(codexHome, "accounts"), 0o700); err != nil {
		t.Fatal(err)
	}
	auth := `{"tokens":{"id_token":"` + testJWT(`{"email":"work@example.com"}`) + `"}}`
	if err := os.WriteFile(filepath.Join(codexHome, "accounts", "work.json"), []byte(auth), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runCLI(t, nil, "--codex-home", codexHome, "list")
	if code != 0 {
		t.Fatalf("list exit code = %d, stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, "│ Active │ Name │ Email            │") {
		t.Fatalf("list stdout missing header: %q", stdout)
	}
	if !strings.Contains(stdout, "│        │ work │ work@example.com │") {
		t.Fatalf("list stdout missing row: %q", stdout)
	}
	if !strings.Contains(stdout, "┌") || !strings.Contains(stdout, "┬") || !strings.Contains(stdout, "┘") {
		t.Fatalf("list stdout missing box border: %q", stdout)
	}
}

func TestExecuteListSyncsCurrentFromLiveAuth(t *testing.T) {
	codexHome := t.TempDir()
	accountsDir := filepath.Join(codexHome, "accounts")
	if err := os.MkdirAll(accountsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accountsDir, "personal.json"), []byte(`{"token":"personal"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accountsDir, "work.json"), []byte(`{"token":"work"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codexHome, "current"), []byte("personal\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codexHome, "auth.json"), []byte(`{"token":"work"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runCLI(t, nil, "--codex-home", codexHome, "list")
	if code != 0 {
		t.Fatalf("list exit code = %d, stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, "│        │ personal │ -     │") || !strings.Contains(stdout, "│ *      │ work     │ -     │") {
		t.Fatalf("list stdout = %q", stdout)
	}

	current, err := os.ReadFile(filepath.Join(codexHome, "current"))
	if err != nil {
		t.Fatal(err)
	}
	if string(current) != "work\n" {
		t.Fatalf("current file = %q, want work", current)
	}
}

func TestExecuteUseSuggestsClosestAccountName(t *testing.T) {
	codexHome := t.TempDir()
	if err := os.MkdirAll(filepath.Join(codexHome, "accounts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codexHome, "accounts", "work.json"), []byte(`{"token":"work"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := runCLI(t, nil, "--codex-home", codexHome, "use", "wrk")
	if code == 0 {
		t.Fatal("use with mistyped account succeeded")
	}
	if !strings.Contains(stderr, `No saved Codex account named "wrk" was found. Did you mean "work"?`) {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestExecuteUsePromptsToSaveUnsavedAuthBeforeSwitching(t *testing.T) {
	codexHome := t.TempDir()
	accountsDir := filepath.Join(codexHome, "accounts")
	if err := os.MkdirAll(accountsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accountsDir, "work.json"), []byte(`{"token":"work"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codexHome, "auth.json"), []byte(`{"token":"new-login"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runCLI(t, strings.NewReader("y\npersonal\n"), "--codex-home", codexHome, "use", "work")
	if code != 0 {
		t.Fatalf("use exit code = %d, stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, "Current Codex auth is not saved as an account.") {
		t.Fatalf("use stdout missing save prompt: %q", stdout)
	}
	if !strings.Contains(stdout, `Saved current Codex auth tokens as "personal".`) {
		t.Fatalf("use stdout missing save confirmation: %q", stdout)
	}
	if !strings.Contains(stdout, `Switched Codex auth to "work".`) {
		t.Fatalf("use stdout missing switch confirmation: %q", stdout)
	}

	contents, err := os.ReadFile(filepath.Join(accountsDir, "personal.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != `{"token":"new-login"}` {
		t.Fatalf("personal snapshot = %q", contents)
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
	if strings.TrimSpace(stdout) != "codex-auth version test" {
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

func testJWT(payload string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
	body := base64.RawURLEncoding.EncodeToString([]byte(payload))
	return header + "." + body + "."
}
