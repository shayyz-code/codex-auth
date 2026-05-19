package accounts

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var accountNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

type Paths struct {
	CodexDir        string
	AccountsDir     string
	AuthPath        string
	CurrentNamePath string
}

func DefaultPaths() (Paths, error) {
	codexDir := os.Getenv("CODEX_HOME")
	if codexDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Paths{}, err
		}
		codexDir = filepath.Join(home, ".codex")
	}

	return NewPaths(codexDir), nil
}

func NewPaths(codexDir string) Paths {
	return Paths{
		CodexDir:        codexDir,
		AccountsDir:     filepath.Join(codexDir, "accounts"),
		AuthPath:        filepath.Join(codexDir, "auth.json"),
		CurrentNamePath: filepath.Join(codexDir, "current"),
	}
}

type Service struct {
	paths Paths
}

func NewService(paths Paths) *Service {
	return &Service{paths: paths}
}

func (s *Service) ListAccountNames() ([]string, error) {
	entries, err := os.ReadDir(s.paths.AccountsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
	}

	sort.Slice(names, func(i, j int) bool {
		return strings.ToLower(names[i]) < strings.ToLower(names[j])
	})

	return names, nil
}

func (s *Service) CurrentAccountName() (string, bool, error) {
	currentName, err := os.ReadFile(s.paths.CurrentNamePath)
	if err == nil {
		trimmed := strings.TrimSpace(string(currentName))
		if trimmed != "" {
			return trimmed, true, nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", false, err
	}

	stat, err := os.Lstat(s.paths.AuthPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	if stat.Mode()&os.ModeSymlink == 0 {
		return "", false, nil
	}

	rawTarget, err := os.Readlink(s.paths.AuthPath)
	if err != nil {
		return "", false, err
	}
	resolvedTarget := rawTarget
	if !filepath.IsAbs(resolvedTarget) {
		resolvedTarget = filepath.Join(filepath.Dir(s.paths.AuthPath), rawTarget)
	}
	resolvedTarget, err = filepath.Abs(resolvedTarget)
	if err != nil {
		return "", false, err
	}
	accountsRoot, err := filepath.Abs(s.paths.AccountsDir)
	if err != nil {
		return "", false, err
	}
	relative, err := filepath.Rel(accountsRoot, resolvedTarget)
	if err != nil || strings.HasPrefix(relative, "..") || filepath.IsAbs(relative) {
		return "", false, nil
	}

	base := filepath.Base(resolvedTarget)
	return strings.TrimSuffix(base, filepath.Ext(base)), true, nil
}

func (s *Service) SaveAccount(rawName string) (string, error) {
	name, err := NormalizeAccountName(rawName)
	if err != nil {
		return "", err
	}
	authContents, err := os.ReadFile(s.paths.AuthPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", AuthFileMissingError{Path: s.paths.AuthPath}
		}
		return "", err
	}
	if err := os.MkdirAll(s.paths.AccountsDir, 0o700); err != nil {
		return "", err
	}
	if err := writeFileAtomic(s.accountFilePath(name), authContents, 0o600); err != nil {
		return "", err
	}
	return name, nil
}

func (s *Service) UseAccount(rawName string) (string, error) {
	name, err := NormalizeAccountName(rawName)
	if err != nil {
		return "", err
	}

	source := s.accountFilePath(name)
	if _, err := os.Stat(source); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", AccountNotFoundError{Name: name}
		}
		return "", err
	}

	if err := os.MkdirAll(s.paths.AccountsDir, 0o700); err != nil {
		return "", err
	}
	if err := os.MkdirAll(s.paths.CodexDir, 0o700); err != nil {
		return "", err
	}

	if err := copyFileAtomic(source, s.paths.AuthPath, 0o600); err != nil {
		return "", err
	}

	if err := writeFileAtomic(s.paths.CurrentNamePath, []byte(name+"\n"), 0o600); err != nil {
		return "", err
	}
	return name, nil
}

func (s *Service) AuthFileExists() (bool, error) {
	if _, err := os.Stat(s.paths.AuthPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Service) CurrentAuthSavedAccount() (string, bool, error) {
	authContents, err := os.ReadFile(s.paths.AuthPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	skipName := ""
	symlinkName, _, ok, targetChanged, err := s.authSymlinkAccountInfo()
	if err != nil {
		return "", false, err
	}
	if ok && targetChanged {
		skipName = symlinkName
	}
	return s.accountNameForAuthContents(authContents, skipName)
}

func (s *Service) SyncCurrentAccount() (string, bool, error) {
	authContents, authErr := os.ReadFile(s.paths.AuthPath)
	if authErr != nil && !errors.Is(authErr, os.ErrNotExist) {
		return "", false, authErr
	}
	name, ok, err := s.CurrentAuthSavedAccount()
	if err != nil {
		return "", false, err
	}
	if !ok {
		if authErr == nil {
			if _, symlinkTarget, symlinkOK, targetChanged, err := s.authSymlinkAccountInfo(); err != nil {
				return "", false, err
			} else if symlinkOK && targetChanged {
				if err := quarantineOverwrittenAccount(symlinkTarget); err != nil {
					return "", false, err
				}
				if err := writeFileAtomic(s.paths.AuthPath, authContents, 0o600); err != nil {
					return "", false, err
				}
			}
		}
		if err := os.Remove(s.paths.CurrentNamePath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", false, err
		}
		return "", false, nil
	}
	if err := writeFileAtomic(s.paths.CurrentNamePath, []byte(name+"\n"), 0o600); err != nil {
		return "", false, err
	}
	return name, true, nil
}

func NormalizeAccountName(rawName string) (string, error) {
	trimmed := strings.TrimSpace(rawName)
	if trimmed == "" {
		return "", InvalidAccountNameError{}
	}
	withoutExtension := trimmed
	if strings.EqualFold(filepath.Ext(trimmed), ".json") {
		withoutExtension = strings.TrimSuffix(trimmed, filepath.Ext(trimmed))
	}
	if !accountNamePattern.MatchString(withoutExtension) {
		return "", InvalidAccountNameError{}
	}
	return withoutExtension, nil
}

func (s *Service) accountFilePath(name string) string {
	return filepath.Join(s.paths.AccountsDir, name+".json")
}

func (s *Service) authSymlinkAccountInfo() (string, string, bool, bool, error) {
	stat, err := os.Lstat(s.paths.AuthPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", false, false, nil
		}
		return "", "", false, false, err
	}
	if stat.Mode()&os.ModeSymlink == 0 {
		return "", "", false, false, nil
	}

	rawTarget, err := os.Readlink(s.paths.AuthPath)
	if err != nil {
		return "", "", false, false, err
	}
	resolvedTarget := rawTarget
	if !filepath.IsAbs(resolvedTarget) {
		resolvedTarget = filepath.Join(filepath.Dir(s.paths.AuthPath), rawTarget)
	}
	resolvedTarget, err = filepath.Abs(resolvedTarget)
	if err != nil {
		return "", "", false, false, err
	}
	accountsRoot, err := filepath.Abs(s.paths.AccountsDir)
	if err != nil {
		return "", "", false, false, err
	}
	relative, err := filepath.Rel(accountsRoot, resolvedTarget)
	if err != nil || strings.HasPrefix(relative, "..") || filepath.IsAbs(relative) {
		return "", "", false, false, nil
	}
	targetStat, err := os.Stat(resolvedTarget)
	if err != nil {
		return "", "", false, false, err
	}

	base := filepath.Base(resolvedTarget)
	return strings.TrimSuffix(base, filepath.Ext(base)), resolvedTarget, true, targetStat.ModTime().After(stat.ModTime()), nil
}

func quarantineOverwrittenAccount(path string) error {
	destination := path + ".overwritten-" + strings.ReplaceAll(timeNowUTC(), ":", "")
	if err := os.Rename(path, destination); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func timeNowUTC() string {
	return time.Now().UTC().Format("20060102T150405Z")
}

func (s *Service) accountNameForAuthContents(authContents []byte, skipName string) (string, bool, error) {
	names, err := s.ListAccountNames()
	if err != nil {
		return "", false, err
	}
	for _, candidate := range names {
		if skipName != "" && candidate == skipName {
			continue
		}
		contents, err := os.ReadFile(s.accountFilePath(candidate))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", false, err
		}
		if bytes.Equal(contents, authContents) {
			return candidate, true, nil
		}
	}
	return "", false, nil
}

func copyFileAtomic(sourcePath string, destinationPath string, mode os.FileMode) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	return writeAtomic(destinationPath, mode, func(destination *os.File) error {
		_, err := io.Copy(destination, source)
		return err
	})
}

func writeFileAtomic(destinationPath string, data []byte, mode os.FileMode) error {
	return writeAtomic(destinationPath, mode, func(destination *os.File) error {
		_, err := destination.Write(data)
		return err
	})
}

func writeAtomic(destinationPath string, mode os.FileMode, write func(*os.File) error) error {
	destinationDir := filepath.Dir(destinationPath)
	if err := os.MkdirAll(destinationDir, 0o700); err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(destinationDir, "."+filepath.Base(destinationPath)+".tmp-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tempPath)
		}
	}()

	if err := tempFile.Chmod(mode); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := write(tempFile); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, destinationPath); err != nil {
		return err
	}
	cleanup = false
	return nil
}
