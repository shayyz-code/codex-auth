package accounts

import "fmt"

type AuthFileMissingError struct {
	Path string
}

func (e AuthFileMissingError) Error() string {
	return fmt.Sprintf("No Codex auth file found at %s. Log into Codex first so ~/.codex/auth.json exists.", e.Path)
}

type AccountNotFoundError struct {
	Name string
}

func (e AccountNotFoundError) Error() string {
	return fmt.Sprintf("No saved Codex account named %q was found.", e.Name)
}

type NoAccountsSavedError struct{}

func (NoAccountsSavedError) Error() string {
	return `No saved Codex accounts yet. Run "codex-auth save <name>" first.`
}

type InvalidAccountNameError struct{}

func (InvalidAccountNameError) Error() string {
	return "Account names must include at least one non-space character and may contain letters, numbers, dashes, underscores, and dots."
}
