package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/shayyz-code/codex-auth/internal/accounts"
	"github.com/spf13/cobra"
)

type service interface {
	ListAccountNames() ([]string, error)
	CurrentAccountName() (string, bool, error)
	AuthFileExists() (bool, error)
	CurrentAuthSavedAccount() (string, bool, error)
	SaveAccount(string) (string, error)
	UseAccount(string) (string, error)
}

type serviceFactory func(codexHome string) (service, error)

func Execute(version string, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	cmd := NewRootCommand(version, defaultServiceFactory)
	cmd.SetArgs(args)
	cmd.SetIn(stdin)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	if err := cmd.Execute(); err != nil {
		style := newStyle(stderr, "auto")
		fmt.Fprintln(stderr, style.error(err.Error()))
		return 1
	}
	return 0
}

func NewRootCommand(version string, newService serviceFactory) *cobra.Command {
	var codexHome string
	var jsonOutput bool
	var colorMode string

	root := &cobra.Command{
		Use:           "codex-auth",
		Short:         "Manage named Codex auth snapshots",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.PersistentFlags().StringVar(&codexHome, "codex-home", "", "Codex config directory; defaults to CODEX_HOME or ~/.codex")
	root.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	root.PersistentFlags().StringVar(&colorMode, "color", "auto", "Color output: auto, always, or never")

	serviceForCommand := func() (service, error) {
		return newService(codexHome)
	}

	root.AddCommand(newSaveCommand(serviceForCommand, &jsonOutput, &colorMode))
	root.AddCommand(newUseCommand(serviceForCommand, &jsonOutput, &colorMode))
	root.AddCommand(newListCommand(serviceForCommand, &jsonOutput, &colorMode))
	root.AddCommand(newCurrentCommand(serviceForCommand, &jsonOutput, &colorMode))
	return root
}

func newSaveCommand(serviceForCommand func() (service, error), jsonOutput *bool, colorMode *string) *cobra.Command {
	return &cobra.Command{
		Use:   "save <name>",
		Short: "Save the current Codex auth file as a named account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			accountsService, err := serviceForCommand()
			if err != nil {
				return err
			}
			savedName, err := accountsService.SaveAccount(args[0])
			if err != nil {
				return err
			}

			if *jsonOutput {
				return printJSON(cmd.OutOrStdout(), map[string]string{"account": savedName})
			}

			style := newStyle(cmd.OutOrStdout(), *colorMode)
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", style.successf("Saved current Codex auth tokens as %q.", savedName))
			return nil
		},
	}
}

func newUseCommand(serviceForCommand func() (service, error), jsonOutput *bool, colorMode *string) *cobra.Command {
	return &cobra.Command{
		Use:   "use [name]",
		Short: "Switch Codex auth to a saved account",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			accountsService, err := serviceForCommand()
			if err != nil {
				return err
			}

			accountName := ""
			input := bufio.NewReader(cmd.InOrStdin())
			if len(args) == 1 {
				accountName = args[0]
			} else {
				if *jsonOutput {
					return errors.New("The [name] argument is required when using --json.")
				}
				picked, err := promptForAccount(input, cmd.OutOrStdout(), accountsService, *colorMode)
				if err != nil {
					return err
				}
				accountName = picked
			}

			if err := ensureSavedAccountExists(accountName, accountsService); err != nil {
				return err
			}
			if !*jsonOutput {
				if err := promptToSaveUnsavedAuth(input, cmd.OutOrStdout(), accountsService, *colorMode); err != nil {
					return err
				}
			}

			activated, err := accountsService.UseAccount(accountName)
			if err != nil {
				return addAccountSuggestion(err, accountName, accountsService)
			}

			if *jsonOutput {
				return printJSON(cmd.OutOrStdout(), map[string]string{"account": activated})
			}

			style := newStyle(cmd.OutOrStdout(), *colorMode)
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", style.successf("Switched Codex auth to %q.", activated))
			return nil
		},
	}
}

func newListCommand(serviceForCommand func() (service, error), jsonOutput *bool, colorMode *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List saved Codex accounts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			accountsService, err := serviceForCommand()
			if err != nil {
				return err
			}
			names, err := accountsService.ListAccountNames()
			if err != nil {
				return err
			}
			current, ok, err := accountsService.CurrentAccountName()
			if err != nil {
				return err
			}

			if *jsonOutput {
				res := map[string]any{
					"accounts": names,
				}
				if ok {
					res["current"] = current
				} else {
					res["current"] = nil
				}
				return printJSON(cmd.OutOrStdout(), res)
			}

			if len(names) == 0 {
				style := newStyle(cmd.OutOrStdout(), *colorMode)
				fmt.Fprintln(cmd.OutOrStdout(), style.warning("No saved Codex accounts yet. Run `codex-auth save <name>`."))
				return nil
			}

			style := newStyle(cmd.OutOrStdout(), *colorMode)
			if style.enabled {
				fmt.Fprintln(cmd.OutOrStdout(), style.title("Saved Codex accounts"))
			}
			for _, name := range names {
				mark := " "
				if ok && current == name {
					mark = "*"
				}
				if style.enabled {
					fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", style.marker(mark, mark == "*"), style.account(name, ok && current == name))
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", mark, name)
				}
			}
			return nil
		},
	}
}

func newCurrentCommand(serviceForCommand func() (service, error), jsonOutput *bool, colorMode *string) *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the active Codex account",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			accountsService, err := serviceForCommand()
			if err != nil {
				return err
			}
			name, ok, err := accountsService.CurrentAccountName()
			if err != nil {
				return err
			}

			if *jsonOutput {
				res := map[string]any{}
				if ok {
					res["account"] = name
				} else {
					res["account"] = nil
				}
				return printJSON(cmd.OutOrStdout(), res)
			}

			if !ok {
				style := newStyle(cmd.OutOrStdout(), *colorMode)
				fmt.Fprintln(cmd.OutOrStdout(), style.warning("No Codex account is active yet."))
				return nil
			}
			style := newStyle(cmd.OutOrStdout(), *colorMode)
			if style.enabled {
				fmt.Fprintln(cmd.OutOrStdout(), style.account(name, true))
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), name)
			}
			return nil
		},
	}
}

func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

type cliStyle struct {
	enabled bool
}

func newStyle(w io.Writer, colorMode string) cliStyle {
	switch strings.ToLower(strings.TrimSpace(colorMode)) {
	case "always":
		return cliStyle{enabled: true}
	case "never":
		return cliStyle{}
	}
	file, ok := w.(*os.File)
	if !ok {
		return cliStyle{}
	}
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return cliStyle{}
	}
	stat, err := file.Stat()
	if err != nil || stat.Mode()&os.ModeCharDevice == 0 {
		return cliStyle{}
	}
	return cliStyle{enabled: true}
}

func (s cliStyle) title(message string) string {
	return s.color(message, "1", "36")
}

func (s cliStyle) account(name string, active bool) string {
	if active {
		return s.color(name, "1", "32")
	}
	return s.color(name, "37")
}

func (s cliStyle) marker(mark string, active bool) string {
	if active {
		return s.color(mark, "1", "32")
	}
	return s.color(mark, "90")
}

func (s cliStyle) successf(format string, args ...any) string {
	return s.success(fmt.Sprintf(format, args...))
}

func (s cliStyle) success(message string) string {
	if !s.enabled {
		return message
	}
	return s.color("[OK]", "1", "32") + " " + message
}

func (s cliStyle) warning(message string) string {
	if !s.enabled {
		return message
	}
	return s.color("[!]", "1", "33") + " " + message
}

func (s cliStyle) error(message string) string {
	if !s.enabled {
		return message
	}
	return s.color("[ERR]", "1", "31") + " " + message
}

func (s cliStyle) prompt(message string) string {
	return s.color(message, "1", "36")
}

func (s cliStyle) hint(message string) string {
	return s.color(message, "90")
}

func (s cliStyle) color(message string, codes ...string) string {
	if !s.enabled {
		return message
	}
	return "\x1b[" + strings.Join(codes, ";") + "m" + message + "\x1b[0m"
}

func ensureSavedAccountExists(rawName string, accountsService service) error {
	name, err := accounts.NormalizeAccountName(rawName)
	if err != nil {
		return err
	}
	names, err := accountsService.ListAccountNames()
	if err != nil {
		return err
	}
	for _, candidate := range names {
		if candidate == name {
			return nil
		}
	}
	return addAccountSuggestion(accounts.AccountNotFoundError{Name: name}, name, accountsService)
}

func promptToSaveUnsavedAuth(stdin io.Reader, stdout io.Writer, accountsService service, colorMode string) error {
	hasAuth, err := accountsService.AuthFileExists()
	if err != nil {
		return err
	}
	if !hasAuth {
		return nil
	}
	if _, ok, err := accountsService.CurrentAuthSavedAccount(); err != nil || ok {
		return err
	}

	reader := promptReader(stdin)
	style := newStyle(stdout, colorMode)
	fmt.Fprintf(stdout, "%s ", style.warning("Current Codex auth is not saved as an account."))
	fmt.Fprint(stdout, style.prompt("Save it before switching? [y/N]: "))
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	if answer != "y" && answer != "yes" {
		fmt.Fprintln(stdout, style.hint("Continuing without saving current Codex auth."))
		return nil
	}

	fmt.Fprint(stdout, style.prompt("Account name: "))
	line, err = reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	name := strings.TrimSpace(line)
	if name == "" {
		return errors.New("No account name entered. The operation was cancelled.")
	}
	savedName, err := accountsService.SaveAccount(name)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "%s\n", style.successf("Saved current Codex auth tokens as %q.", savedName))
	return nil
}

func addAccountSuggestion(err error, requested string, accountsService service) error {
	var notFound accounts.AccountNotFoundError
	if !errors.As(err, &notFound) {
		return err
	}

	names, listErr := accountsService.ListAccountNames()
	if listErr != nil {
		return err
	}
	suggestion := closestAccountName(requested, names)
	if suggestion == "" {
		return err
	}
	return fmt.Errorf("%w Did you mean %q?", err, suggestion)
}

func closestAccountName(rawName string, names []string) string {
	name, err := accounts.NormalizeAccountName(rawName)
	if err != nil {
		name = strings.TrimSpace(rawName)
	}
	name = strings.ToLower(name)
	if name == "" || len(names) == 0 {
		return ""
	}

	bestName := ""
	bestDistance := 0
	for _, candidate := range names {
		distance := levenshteinDistance(name, strings.ToLower(candidate))
		if bestName == "" || distance < bestDistance || distance == bestDistance && strings.ToLower(candidate) < strings.ToLower(bestName) {
			bestName = candidate
			bestDistance = distance
		}
	}

	limit := len(name) / 2
	if limit < 2 {
		limit = 2
	}
	if bestDistance > limit {
		return ""
	}
	return bestName
}

func levenshteinDistance(a string, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)
	for j := range previous {
		previous[j] = j
	}
	for i := 1; i <= len(a); i++ {
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			current[j] = minInt(current[j-1]+1, previous[j]+1, previous[j-1]+cost)
		}
		previous, current = current, previous
	}
	return previous[len(b)]
}

func minInt(values ...int) int {
	minimum := values[0]
	for _, value := range values[1:] {
		if value < minimum {
			minimum = value
		}
	}
	return minimum
}

func promptForAccount(stdin io.Reader, stdout io.Writer, accountsService service, colorMode string) (string, error) {
	names, err := accountsService.ListAccountNames()
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		return "", accounts.NoAccountsSavedError{}
	}

	current, ok, err := accountsService.CurrentAccountName()
	if err != nil {
		return "", err
	}

	style := newStyle(stdout, colorMode)
	if style.enabled {
		fmt.Fprintln(stdout, style.title("Select account"))
	} else {
		fmt.Fprintln(stdout, "Select account:")
	}
	for i, name := range names {
		label := name
		if ok && current == name {
			label += " (active)"
		}
		if style.enabled {
			fmt.Fprintf(stdout, "  %s) %s\n", style.color(strconv.Itoa(i+1), "36"), style.account(label, ok && current == name))
		} else {
			fmt.Fprintf(stdout, "  %d) %s\n", i+1, label)
		}
	}
	fmt.Fprint(stdout, style.prompt("Enter number: "))

	line, err := promptReader(stdin).ReadString('\n')
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

func promptReader(stdin io.Reader) *bufio.Reader {
	if reader, ok := stdin.(*bufio.Reader); ok {
		return reader
	}
	return bufio.NewReader(stdin)
}

func defaultServiceFactory(codexHome string) (service, error) {
	var paths accounts.Paths
	var err error
	if codexHome == "" {
		paths, err = accounts.DefaultPaths()
		if err != nil {
			return nil, err
		}
	} else {
		paths = accounts.NewPaths(codexHome)
	}
	return accounts.NewService(paths), nil
}
