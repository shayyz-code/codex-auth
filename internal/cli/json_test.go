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
