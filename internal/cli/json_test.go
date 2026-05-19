package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestExecuteWithJSON(t *testing.T) {
	codexHome := t.TempDir()
	authPath := filepath.Join(codexHome, "auth.json")
	if err := os.WriteFile(authPath, []byte(`{"token":"alpha"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	// Test save --json
	stdout, stderr, code := runCLI(t, nil, "--codex-home", codexHome, "--json", "save", "work")
	if code != 0 {
		t.Fatalf("save --json exit code = %d, stderr = %q", code, stderr)
	}
	var saveRes map[string]string
	if err := json.Unmarshal([]byte(stdout), &saveRes); err != nil {
		t.Fatalf("failed to unmarshal save --json output: %v, stdout: %q", err, stdout)
	}
	if saveRes["account"] != "work" {
		t.Fatalf("save --json account = %q, want %q", saveRes["account"], "work")
	}

	// Test list --json
	stdout, stderr, code = runCLI(t, nil, "--codex-home", codexHome, "--json", "list")
	if code != 0 {
		t.Fatalf("list --json exit code = %d, stderr = %q", code, stderr)
	}
	var listRes map[string]any
	if err := json.Unmarshal([]byte(stdout), &listRes); err != nil {
		t.Fatalf("failed to unmarshal list --json output: %v, stdout: %q", err, stdout)
	}
	expectedAccounts := []any{"work"}
	if !reflect.DeepEqual(listRes["accounts"], expectedAccounts) {
		t.Fatalf("list --json accounts = %v, want %v", listRes["accounts"], expectedAccounts)
	}
	if listRes["current"] != "work" {
		t.Fatalf("list --json current = %v, want %q", listRes["current"], "work")
	}

	// Test use --json
	stdout, stderr, code = runCLI(t, nil, "--codex-home", codexHome, "--json", "use", "work")
	if code != 0 {
		t.Fatalf("use --json exit code = %d, stderr = %q", code, stderr)
	}
	var useRes map[string]string
	if err := json.Unmarshal([]byte(stdout), &useRes); err != nil {
		t.Fatalf("failed to unmarshal use --json output: %v, stdout: %q", err, stdout)
	}
	if useRes["account"] != "work" {
		t.Fatalf("use --json account = %q, want %q", useRes["account"], "work")
	}

	// Test current --json
	stdout, stderr, code = runCLI(t, nil, "--codex-home", codexHome, "--json", "current")
	if code != 0 {
		t.Fatalf("current --json exit code = %d, stderr = %q", code, stderr)
	}
	var currentRes map[string]any
	if err := json.Unmarshal([]byte(stdout), &currentRes); err != nil {
		t.Fatalf("failed to unmarshal current --json output: %v, stdout: %q", err, stdout)
	}
	if currentRes["account"] != "work" {
		t.Fatalf("current --json account = %v, want %q", currentRes["account"], "work")
	}

	// Test list --json with active account
	stdout, stderr, code = runCLI(t, nil, "--codex-home", codexHome, "--json", "list")
	if code != 0 {
		t.Fatalf("list --json exit code = %d, stderr = %q", code, stderr)
	}
	if err := json.Unmarshal([]byte(stdout), &listRes); err != nil {
		t.Fatalf("failed to unmarshal list --json output: %v, stdout: %q", err, stdout)
	}
	if listRes["current"] != "work" {
		t.Fatalf("list --json current = %v, want %q", listRes["current"], "work")
	}
}

func TestUseWithoutNameFailsWithJSON(t *testing.T) {
	codexHome := t.TempDir()
	_, stderr, code := runCLI(t, nil, "--codex-home", codexHome, "--json", "use")
	if code == 0 {
		t.Fatal("use --json without name succeeded")
	}
	if !strings.Contains(stderr, "The [name] argument is required when using --json") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestRenameWithJSONIncludesEmail(t *testing.T) {
	codexHome := t.TempDir()
	accountsDir := filepath.Join(codexHome, "accounts")
	if err := os.MkdirAll(accountsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	auth := `{"email":"work@example.com"}`
	if err := os.WriteFile(filepath.Join(accountsDir, "work.json"), []byte(auth), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runCLI(t, nil, "--codex-home", codexHome, "--json", "rename", "work", "office")
	if code != 0 {
		t.Fatalf("rename --json exit code = %d, stderr = %q", code, stderr)
	}
	var res map[string]any
	if err := json.Unmarshal([]byte(stdout), &res); err != nil {
		t.Fatalf("failed to unmarshal rename --json output: %v, stdout: %q", err, stdout)
	}
	if res["from"] != "work" || res["to"] != "office" || res["email"] != "work@example.com" {
		t.Fatalf("rename --json = %v, want from work to office with email", res)
	}
}
