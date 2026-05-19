package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func NewRootCommand(version string, newService serviceFactory) *cobra.Command {
	var codexHome string
	var jsonOutput bool

	root := &cobra.Command{
		Use:           "codex-auth",
		Short:         "Manage named Codex auth snapshots",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.PersistentFlags().StringVar(&codexHome, "codex-home", "", "Codex config directory; defaults to CODEX_HOME or ~/.codex")
	root.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	serviceForCommand := func() (service, error) {
		return newService(codexHome)
	}

	root.AddCommand(newSaveCommand(serviceForCommand, &jsonOutput))
	root.AddCommand(newUseCommand(serviceForCommand, &jsonOutput))
	root.AddCommand(newListCommand(serviceForCommand, &jsonOutput))
	root.AddCommand(newCurrentCommand(serviceForCommand, &jsonOutput))
	return root
}

func newSaveCommand(serviceForCommand func() (service, error), jsonOutput *bool) *cobra.Command {
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

			fmt.Fprintf(cmd.OutOrStdout(), "Saved current Codex auth tokens as %q.\n", savedName)
			return nil
		},
	}
}

func newUseCommand(serviceForCommand func() (service, error), jsonOutput *bool) *cobra.Command {
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
				picked, err := promptForAccount(input, cmd.OutOrStdout(), accountsService)
				if err != nil {
					return err
				}
				accountName = picked
			}

			if err := ensureSavedAccountExists(accountName, accountsService); err != nil {
				return err
			}
			if !*jsonOutput {
				if err := promptToSaveUnsavedAuth(input, cmd.OutOrStdout(), accountsService); err != nil {
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

			fmt.Fprintf(cmd.OutOrStdout(), "Switched Codex auth to %q.\n", activated)
			return nil
		},
	}
}

func newListCommand(serviceForCommand func() (service, error), jsonOutput *bool) *cobra.Command {
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
				fmt.Fprintln(cmd.OutOrStdout(), "No saved Codex accounts yet. Run `codex-auth save <name>`.")
				return nil
			}

			for _, name := range names {
				mark := " "
				if ok && current == name {
					mark = "*"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", mark, name)
			}
			return nil
		},
	}
}

func newCurrentCommand(serviceForCommand func() (service, error), jsonOutput *bool) *cobra.Command {
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
				fmt.Fprintln(cmd.OutOrStdout(), "No Codex account is active yet.")
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), name)
			return nil
		},
	}
}

func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
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

func promptToSaveUnsavedAuth(stdin io.Reader, stdout io.Writer, accountsService service) error {
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
	fmt.Fprint(stdout, "Current Codex auth is not saved as an account. Save it before switching? [y/N]: ")
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	if answer != "y" && answer != "yes" {
		fmt.Fprintln(stdout, "Continuing without saving current Codex auth.")
		return nil
	}

	fmt.Fprint(stdout, "Account name: ")
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
	fmt.Fprintf(stdout, "Saved current Codex auth tokens as %q.\n", savedName)
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

func promptForAccount(stdin io.Reader, stdout io.Writer, accountsService service) (string, error) {
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

	fmt.Fprintln(stdout, "Select account:")
	for i, name := range names {
		label := name
		if ok && current == name {
			label += " (active)"
		}
		fmt.Fprintf(stdout, "  %d) %s\n", i+1, label)
	}
	fmt.Fprint(stdout, "Enter number: ")

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
