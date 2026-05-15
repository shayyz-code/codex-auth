package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/shayyz-code/codex-su/internal/accounts"
)

type App struct {
	Name    string
	Version string
	Service *accounts.Service
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

func New(version string) (*App, error) {
	paths, err := accounts.DefaultPaths()
	if err != nil {
		return nil, err
	}
	return &App{
		Name:    "codex-su",
		Version: version,
		Service: accounts.NewService(paths),
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}, nil
}

func (a *App) Run(args []string) int {
	if len(args) == 0 {
		a.printUsage()
		return 1
	}

	switch args[0] {
	case "-h", "--help", "help":
		a.printUsage()
		return 0
	case "-v", "--version", "version":
		fmt.Fprintf(a.Stdout, "%s %s\n", a.Name, a.Version)
		return 0
	case "save":
		return a.runSave(args[1:])
	case "use":
		return a.runUse(args[1:])
	case "list":
		return a.runList(args[1:])
	case "current":
		return a.runCurrent(args[1:])
	default:
		fmt.Fprintf(a.Stderr, "Unknown command: %s\n\n", args[0])
		a.printUsage()
		return 1
	}
}

func (a *App) runSave(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(a.Stderr, "Usage: codex-su save <name>")
		return 1
	}
	savedName, err := a.Service.SaveAccount(args[0])
	if err != nil {
		return a.handleError(err)
	}
	fmt.Fprintf(a.Stdout, "Saved current Codex auth tokens as %q.\n", savedName)
	return 0
}

func (a *App) runUse(args []string) int {
	if len(args) > 1 {
		fmt.Fprintln(a.Stderr, "Usage: codex-su use [name]")
		return 1
	}

	accountName := ""
	if len(args) == 1 {
		accountName = args[0]
	} else {
		picked, err := a.promptForAccount()
		if err != nil {
			return a.handleError(err)
		}
		accountName = picked
	}

	activated, err := a.Service.UseAccount(accountName)
	if err != nil {
		return a.handleError(err)
	}
	fmt.Fprintf(a.Stdout, "Switched Codex auth to %q.\n", activated)
	return 0
}

func (a *App) runList(args []string) int {
	if len(args) != 0 {
		fmt.Fprintln(a.Stderr, "Usage: codex-su list")
		return 1
	}

	names, err := a.Service.ListAccountNames()
	if err != nil {
		return a.handleError(err)
	}
	current, ok, err := a.Service.CurrentAccountName()
	if err != nil {
		return a.handleError(err)
	}

	if len(names) == 0 {
		fmt.Fprintln(a.Stdout, "No saved Codex accounts yet. Run `codex-su save <name>`.")
		return 0
	}

	for _, name := range names {
		mark := " "
		if ok && current == name {
			mark = "*"
		}
		fmt.Fprintf(a.Stdout, "%s %s\n", mark, name)
	}
	return 0
}

func (a *App) runCurrent(args []string) int {
	if len(args) != 0 {
		fmt.Fprintln(a.Stderr, "Usage: codex-su current")
		return 1
	}
	name, ok, err := a.Service.CurrentAccountName()
	if err != nil {
		return a.handleError(err)
	}
	if !ok {
		fmt.Fprintln(a.Stdout, "No Codex account is active yet.")
		return 0
	}
	fmt.Fprintln(a.Stdout, name)
	return 0
}

func (a *App) promptForAccount() (string, error) {
	names, err := a.Service.ListAccountNames()
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		return "", accounts.NoAccountsSavedError{}
	}

	current, ok, err := a.Service.CurrentAccountName()
	if err != nil {
		return "", err
	}

	fmt.Fprintln(a.Stdout, "Select account:")
	for i, name := range names {
		label := name
		if ok && current == name {
			label += " (active)"
		}
		fmt.Fprintf(a.Stdout, "  %d) %s\n", i+1, label)
	}
	fmt.Fprint(a.Stdout, "Enter number: ")

	line, err := bufio.NewReader(a.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	line = strings.TrimSpace(line)
	index, err := strconv.Atoi(line)
	if err != nil || index < 1 || index > len(names) {
		return "", errors.New("No account selected. The operation was cancelled.")
	}
	return names[index-1], nil
}

func (a *App) handleError(err error) int {
	fmt.Fprintln(a.Stderr, err.Error())
	return 1
}

func (a *App) printUsage() {
	fmt.Fprintln(a.Stdout, `codex-su manages named Codex auth snapshots.

Usage:
  codex-su save <name>     Save the current ~/.codex/auth.json as an account
  codex-su use [name]      Switch ~/.codex/auth.json to an account
  codex-su list            List saved accounts
  codex-su current         Show the active account name
  codex-su --version       Show version

Set CODEX_HOME to use a nonstandard Codex config directory.`)
}
