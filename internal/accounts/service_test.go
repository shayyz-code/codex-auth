package accounts

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
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

	info, err := os.Lstat(paths.AuthPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("auth.json is a symlink")
	}
}

func TestAccountEmailExtractsJWTEmail(t *testing.T) {
	paths := NewPaths(t.TempDir())
	service := NewService(paths)

	if err := os.MkdirAll(paths.AccountsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	auth := `{"tokens":{"id_token":"` + testJWT(`{"email":"work@example.com"}`) + `"}}`
	if err := os.WriteFile(filepath.Join(paths.AccountsDir, "work.json"), []byte(auth), 0o600); err != nil {
		t.Fatal(err)
	}

	email, err := service.AccountEmail("work")
	if err != nil {
		t.Fatal(err)
	}
	if email != "work@example.com" {
		t.Fatalf("email = %q, want work@example.com", email)
	}
}

func TestRenameAccountRenamesSnapshotAndCurrentMarker(t *testing.T) {
	paths := NewPaths(t.TempDir())
	service := NewService(paths)

	if err := os.MkdirAll(paths.AccountsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	auth := `{"tokens":{"id_token":"` + testJWT(`{"email":"work@example.com"}`) + `"}}`
	if err := os.WriteFile(filepath.Join(paths.AccountsDir, "work.json"), []byte(auth), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.CurrentNamePath, []byte("work\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	info, err := service.RenameAccount("work", "office")
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "office" || info.Email != "work@example.com" {
		t.Fatalf("info = %+v, want office with email", info)
	}
	if _, err := os.Stat(filepath.Join(paths.AccountsDir, "work.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("old account exists after rename: %v", err)
	}
	if _, err := os.Stat(filepath.Join(paths.AccountsDir, "office.json")); err != nil {
		t.Fatal(err)
	}
	current, ok, err := service.CurrentAccountName()
	if err != nil {
		t.Fatal(err)
	}
	if !ok || current != "office" {
		t.Fatalf("current = %q, %v; want office, true", current, ok)
	}
}

func TestRenameAccountRejectsExistingDestination(t *testing.T) {
	paths := NewPaths(t.TempDir())
	service := NewService(paths)

	if err := os.MkdirAll(paths.AccountsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paths.AccountsDir, "work.json"), []byte(`{"token":"work"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paths.AccountsDir, "office.json"), []byte(`{"token":"office"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := service.RenameAccount("work", "office"); err == nil {
		t.Fatal("expected destination exists error")
	}
}

func TestSaveAccountHonorsRequestedNameEvenWhenAuthMatchesExistingSnapshot(t *testing.T) {
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
	if name != "duplicate" {
		t.Fatalf("duplicate save name = %q, want duplicate", name)
	}

	names, err := service.ListAccountNames()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := join(names), "duplicate,work"; got != want {
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

func TestCurrentAuthSavedAccountDoesNotTrustAuthSymlinkTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink compatibility is not used on Windows")
	}

	paths := NewPaths(t.TempDir())
	service := NewService(paths)
	accountPath := filepath.Join(paths.AccountsDir, "work.json")

	if err := os.MkdirAll(paths.AccountsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(accountPath, []byte(`{"token":"new-login-through-link"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(accountPath, paths.AuthPath); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes(accountPath, future, future); err != nil {
		t.Fatal(err)
	}

	name, ok, err := service.CurrentAuthSavedAccount()
	if err != nil {
		t.Fatal(err)
	}
	if ok || name != "" {
		t.Fatalf("current auth saved account = %q, %v; want empty, false", name, ok)
	}
}

func TestSyncCurrentAccountDetachesModifiedAuthSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink compatibility is not used on Windows")
	}

	paths := NewPaths(t.TempDir())
	service := NewService(paths)
	accountPath := filepath.Join(paths.AccountsDir, "work.json")

	if err := os.MkdirAll(paths.AccountsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(accountPath, []byte(`{"token":"new-login-through-link"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.CurrentNamePath, []byte("work\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(accountPath, paths.AuthPath); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes(accountPath, future, future); err != nil {
		t.Fatal(err)
	}

	name, ok, err := service.SyncCurrentAccount()
	if err != nil {
		t.Fatal(err)
	}
	if ok || name != "" {
		t.Fatalf("synced current = %q, %v; want empty, false", name, ok)
	}
	info, err := os.Lstat(paths.AuthPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("auth.json is still a symlink")
	}
	if _, err := os.Stat(accountPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("account path exists after quarantine: %v", err)
	}
	matches, err := filepath.Glob(accountPath + ".overwritten-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("quarantined accounts = %v, want one", matches)
	}
	contents, err := os.ReadFile(paths.AuthPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != `{"token":"new-login-through-link"}` {
		t.Fatalf("auth = %q", contents)
	}
}

func TestSyncCurrentAccountMatchesLiveAuthToSavedAccount(t *testing.T) {
	paths := NewPaths(t.TempDir())
	service := NewService(paths)

	if err := os.MkdirAll(paths.AccountsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paths.AccountsDir, "personal.json"), []byte(`{"token":"personal"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paths.AccountsDir, "work.json"), []byte(`{"token":"work"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.CurrentNamePath, []byte("personal\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	writeAuth(t, paths, `{"token":"work"}`)

	name, ok, err := service.SyncCurrentAccount()
	if err != nil {
		t.Fatal(err)
	}
	if !ok || name != "work" {
		t.Fatalf("synced current = %q, %v; want work, true", name, ok)
	}

	current, ok, err := service.CurrentAccountName()
	if err != nil {
		t.Fatal(err)
	}
	if !ok || current != "work" {
		t.Fatalf("current = %q, %v; want work, true", current, ok)
	}
}

func TestSyncCurrentAccountClearsStaleCurrentWhenLiveAuthIsUnsaved(t *testing.T) {
	paths := NewPaths(t.TempDir())
	service := NewService(paths)

	if err := os.MkdirAll(paths.AccountsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paths.AccountsDir, "work.json"), []byte(`{"token":"work"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.CurrentNamePath, []byte("work\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	writeAuth(t, paths, `{"token":"unsaved"}`)

	name, ok, err := service.SyncCurrentAccount()
	if err != nil {
		t.Fatal(err)
	}
	if ok || name != "" {
		t.Fatalf("synced current = %q, %v; want empty, false", name, ok)
	}

	current, ok, err := service.CurrentAccountName()
	if err != nil {
		t.Fatal(err)
	}
	if ok || current != "" {
		t.Fatalf("current = %q, %v; want empty, false", current, ok)
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

func testJWT(payload string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
	body := base64.RawURLEncoding.EncodeToString([]byte(payload))
	return header + "." + body + "."
}
