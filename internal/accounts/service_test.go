package accounts

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSaveAccountSnapshotsAuthAndListSortsAccounts(t *testing.T) {
	paths := NewPaths(t.TempDir())
	service := NewService(paths)

	writeAuth(t, paths, `{"token":"alpha"}`)
	name, err := service.SaveAccount("Beta.json")
	if err != nil {
		t.Fatal(err)
	}
	if name != "Beta" {
		t.Fatalf("name = %q, want Beta", name)
	}

	writeAuth(t, paths, `{"token":"bravo"}`)
	if _, err := service.SaveAccount("alpha"); err != nil {
		t.Fatal(err)
	}

	names, err := service.ListAccountNames()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := join(names), "alpha,Beta"; got != want {
		t.Fatalf("names = %q, want %q", got, want)
	}

	contents, err := os.ReadFile(filepath.Join(paths.AccountsDir, "Beta.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != `{"token":"alpha"}` {
		t.Fatalf("snapshot = %q", contents)
	}
}

func TestUseAccountActivatesSavedAccountAndRecordsCurrent(t *testing.T) {
	paths := NewPaths(t.TempDir())
	service := NewService(paths)

	if err := os.MkdirAll(paths.AccountsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paths.AccountsDir, "work.json"), []byte(`{"token":"work"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	name, err := service.UseAccount("work")
	if err != nil {
		t.Fatal(err)
	}
	if name != "work" {
		t.Fatalf("name = %q, want work", name)
	}

	current, ok, err := service.CurrentAccountName()
	if err != nil {
		t.Fatal(err)
	}
	if !ok || current != "work" {
		t.Fatalf("current = %q, %v; want work, true", current, ok)
	}

	contents, err := os.ReadFile(paths.AuthPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != `{"token":"work"}` {
		t.Fatalf("auth = %q", contents)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Lstat(paths.AuthPath)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Fatal("auth.json is not a symlink")
		}
	}
}

func TestSaveAccountDoesNotDuplicateExistingAuthSnapshot(t *testing.T) {
	paths := NewPaths(t.TempDir())
	service := NewService(paths)

	writeAuth(t, paths, `{"token":"same"}`)
	name, err := service.SaveAccount("work")
	if err != nil {
		t.Fatal(err)
	}
	if name != "work" {
		t.Fatalf("name = %q, want work", name)
	}

	name, err = service.SaveAccount("duplicate")
	if err != nil {
		t.Fatal(err)
	}
	if name != "work" {
		t.Fatalf("duplicate save name = %q, want existing work", name)
	}

	names, err := service.ListAccountNames()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := join(names), "work"; got != want {
		t.Fatalf("names = %q, want %q", got, want)
	}
}

func TestCurrentAuthSavedAccountMatchesByContent(t *testing.T) {
	paths := NewPaths(t.TempDir())
	service := NewService(paths)

	if err := os.MkdirAll(paths.AccountsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paths.AccountsDir, "work.json"), []byte(`{"token":"work"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	writeAuth(t, paths, `{"token":"work"}`)

	name, ok, err := service.CurrentAuthSavedAccount()
	if err != nil {
		t.Fatal(err)
	}
	if !ok || name != "work" {
		t.Fatalf("current auth saved account = %q, %v; want work, true", name, ok)
	}

	writeAuth(t, paths, `{"token":"new"}`)
	name, ok, err = service.CurrentAuthSavedAccount()
	if err != nil {
		t.Fatal(err)
	}
	if ok || name != "" {
		t.Fatalf("current auth saved account = %q, %v; want empty, false", name, ok)
	}
}

func TestCurrentAccountNameInfersSymlinkTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink inference is not used on Windows")
	}

	paths := NewPaths(t.TempDir())
	service := NewService(paths)
	accountPath := filepath.Join(paths.AccountsDir, "personal.json")

	if err := os.MkdirAll(filepath.Dir(accountPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(accountPath, []byte(`{"token":"personal"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(accountPath, paths.AuthPath); err != nil {
		t.Fatal(err)
	}

	current, ok, err := service.CurrentAccountName()
	if err != nil {
		t.Fatal(err)
	}
	if !ok || current != "personal" {
		t.Fatalf("current = %q, %v; want personal, true", current, ok)
	}
}

func TestSaveAccountRejectsInvalidNamesAndMissingAuth(t *testing.T) {
	paths := NewPaths(t.TempDir())
	service := NewService(paths)

	if _, err := service.SaveAccount("../bad"); err == nil {
		t.Fatal("expected invalid account name error")
	}
	if _, err := service.SaveAccount("missing-auth"); err == nil {
		t.Fatal("expected missing auth error")
	}
}

func writeAuth(t *testing.T, paths Paths, contents string) {
	t.Helper()
	if err := os.MkdirAll(paths.CodexDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.AuthPath, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

func join(values []string) string {
	if len(values) == 0 {
		return ""
	}
	result := values[0]
	for _, value := range values[1:] {
		result += "," + value
	}
	return result
}
