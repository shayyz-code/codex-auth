package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/shayyz-code/codex-auth/internal/accounts"
	"github.com/spf13/cobra"
)

type service interface {
	ListAccounts() ([]accounts.AccountInfo, error)
	ListAccountNames() ([]string, error)
	CurrentAccountName() (string, bool, error)
	AuthFileExists() (bool, error)
	CurrentAuthSavedAccount() (string, bool, error)
	SyncCurrentAccount() (string, bool, error)
	AccountEmail(string) (string, error)
	RenameAccount(string, string) (accounts.AccountInfo, error)
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
		accountsService, err := newService(codexHome)
		if err != nil {
			return nil, err
		}
		if _, _, err := accountsService.SyncCurrentAccount(); err != nil {
			return nil, err
		}
		return accountsService, nil
	}

	root.AddCommand(newSaveCommand(serviceForCommand, &jsonOutput, &colorMode))
	root.AddCommand(newUseCommand(serviceForCommand, &jsonOutput, &colorMode))
	root.AddCommand(newRenameCommand(serviceForCommand, &jsonOutput, &colorMode))
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

func newRenameCommand(serviceForCommand func() (service, error), jsonOutput *bool, colorMode *string) *cobra.Command {
	return &cobra.Command{
		Use:   "rename [old-name] [new-name]",
		Short: "Rename a saved Codex account",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || len(args) == 2 {
				return nil
			}
			return fmt.Errorf("accepts either 0 or 2 arg(s), received %d", len(args))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			accountsService, err := serviceForCommand()
			if err != nil {
				return err
			}
			input := bufio.NewReader(cmd.InOrStdin())
			oldRaw := ""
			newRaw := ""
			if len(args) == 2 {
				oldRaw = args[0]
				newRaw = args[1]
			} else {
				if *jsonOutput {
					return errors.New("The [old-name] and [new-name] arguments are required when using --json.")
				}
				oldRaw, err = promptForAccount(input, cmd.OutOrStdout(), accountsService, *colorMode)
				if err != nil {
					return err
				}
				style := newStyle(cmd.OutOrStdout(), *colorMode)
				fmt.Fprint(cmd.OutOrStdout(), style.prompt("New saved name: "))
				line, err := input.ReadString('\n')
				if err != nil && !errors.Is(err, io.EOF) {
					return err
				}
				newRaw = strings.TrimSpace(line)
			}
			oldName, err := accounts.NormalizeAccountName(oldRaw)
			if err != nil {
				return err
			}
			info, err := accountsService.RenameAccount(oldRaw, newRaw)
			if err != nil {
				return addAccountSuggestion(err, oldRaw, accountsService)
			}
			if *jsonOutput {
				return printJSON(cmd.OutOrStdout(), map[string]any{
					"from":  oldName,
					"to":    info.Name,
					"email": nullableString(info.Email),
				})
			}

			style := newStyle(cmd.OutOrStdout(), *colorMode)
			detail := info.Name
			if info.Email != "" {
				detail += fmt.Sprintf(" (email: %s)", info.Email)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", style.successf("Renamed %q to %s.", oldName, detail))
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
			infos, err := accountsService.ListAccounts()
			if err != nil {
				return err
			}
			names := accountNames(infos)
			current, ok, err := accountsService.CurrentAccountName()
			if err != nil {
				return err
			}

			if *jsonOutput {
				res := map[string]any{
					"accounts": names,
					"details":  accountDetails(infos),
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
			printAccountsTable(cmd.OutOrStdout(), infos, current, ok, style)
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
					email, err := accountsService.AccountEmail(name)
					if err != nil {
						return err
					}
					res["email"] = nullableString(email)
				} else {
					res["account"] = nil
					res["email"] = nil
				}
				return printJSON(cmd.OutOrStdout(), res)
			}

			if !ok {
				style := newStyle(cmd.OutOrStdout(), *colorMode)
				fmt.Fprintln(cmd.OutOrStdout(), style.warning("No Codex account is active yet."))
				return nil
			}
			style := newStyle(cmd.OutOrStdout(), *colorMode)
			email, err := accountsService.AccountEmail(name)
			if err != nil {
				return err
			}
			label := formatAccountInfo(accounts.AccountInfo{Name: name, Email: email})
			if style.enabled {
				fmt.Fprintln(cmd.OutOrStdout(), style.account(label, true))
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), label)
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

func accountNames(infos []accounts.AccountInfo) []string {
	names := make([]string, 0, len(infos))
	for _, info := range infos {
		names = append(names, info.Name)
	}
	return names
}

func accountDetails(infos []accounts.AccountInfo) []map[string]any {
	details := make([]map[string]any, 0, len(infos))
	for _, info := range infos {
		details = append(details, map[string]any{
			"name":  info.Name,
			"email": nullableString(info.Email),
		})
	}
	return details
}

func formatAccountInfo(info accounts.AccountInfo) string {
	if info.Email == "" {
		return info.Name
	}
	return fmt.Sprintf("%s <%s>", info.Name, info.Email)
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func printAccountsTable(stdout io.Writer, infos []accounts.AccountInfo, current string, hasCurrent bool, style cliStyle) {
	nameWidth := len("Name")
	emailWidth := len("Email")
	for _, info := range infos {
		if len(info.Name) > nameWidth {
			nameWidth = len(info.Name)
		}
		if len(displayEmail(info.Email)) > emailWidth {
			emailWidth = len(displayEmail(info.Email))
		}
	}

	if style.enabled {
		fmt.Fprintln(stdout, style.title("Saved Codex accounts"))
	}
	border := "+" + strings.Repeat("-", len("Active")+2) + "+" + strings.Repeat("-", nameWidth+2) + "+" + strings.Repeat("-", emailWidth+2) + "+"
	fmt.Fprintln(stdout, border)
	fmt.Fprintf(stdout, "| %-6s | %-*s | %-*s |\n", "Active", nameWidth, "Name", emailWidth, "Email")
	fmt.Fprintln(stdout, border)
	for _, info := range infos {
		active := ""
		isActive := hasCurrent && current == info.Name
		if isActive {
			active = "*"
		}
		fmt.Fprintf(stdout, "| %-6s | %-*s | %-*s |\n", active, nameWidth, info.Name, emailWidth, displayEmail(info.Email))
	}
	fmt.Fprintln(stdout, border)
}

func displayEmail(email string) string {
	if email == "" {
		return "-"
	}
	return email
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
	infos, err := accountsService.ListAccounts()
	if err != nil {
		return "", err
	}
	if len(infos) == 0 {
		return "", accounts.NoAccountsSavedError{}
	}

	current, ok, err := accountsService.CurrentAccountName()
	if err != nil {
		return "", err
	}

	if stdinFile, stdinOK := stdin.(*os.File); stdinOK {
		if stdoutFile, stdoutOK := stdout.(*os.File); stdoutOK && isTerminal(stdinFile) && isTerminal(stdoutFile) {
			if name, err := promptForAccountMenu(stdinFile, stdoutFile, infos, current, ok, colorMode); err == nil {
				return name, nil
			}
		}
	}

	style := newStyle(stdout, colorMode)
	if style.enabled {
		fmt.Fprintln(stdout, style.title("Codex accounts"))
		fmt.Fprintln(stdout, style.hint("Choose by number, saved name, or email."))
	} else {
		fmt.Fprintln(stdout, "Select account:")
	}
	for i, info := range infos {
		label := formatAccountInfo(info)
		if style.enabled {
			status := ""
			if ok && current == info.Name {
				status = " " + style.color("active", "1", "32")
			}
			fmt.Fprintf(stdout, "  %s  %s%s\n", style.color(strconv.Itoa(i+1), "36"), style.account(label, ok && current == info.Name), status)
		} else {
			if ok && current == info.Name {
				label += " (active)"
			}
			fmt.Fprintf(stdout, "  %d) %s\n", i+1, label)
		}
	}
	fmt.Fprint(stdout, style.prompt("Select account: "))

	line, err := promptReader(stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	line = strings.TrimSpace(line)
	index, err := strconv.Atoi(line)
	if err == nil {
		if index < 1 || index > len(infos) {
			return "", errors.New("No account selected. The operation was cancelled.")
		}
		return infos[index-1].Name, nil
	}
	for _, info := range infos {
		if strings.EqualFold(info.Name, line) || strings.EqualFold(info.Email, line) {
			return info.Name, nil
		}
	}
	return "", errors.New("No account selected. The operation was cancelled.")
}

func promptForAccountMenu(stdin *os.File, stdout *os.File, infos []accounts.AccountInfo, current string, hasCurrent bool, colorMode string) (string, error) {
	restore, err := enableRawMode(stdin)
	if err != nil {
		return "", err
	}
	defer restore()

	style := newStyle(stdout, colorMode)
	selected := 0
	for i, info := range infos {
		if hasCurrent && info.Name == current {
			selected = i
			break
		}
	}

	fmt.Fprint(stdout, "\x1b[?25l")
	defer fmt.Fprint(stdout, "\x1b[?25h")

	lines := 0
	render := func() {
		if lines > 0 {
			fmt.Fprintf(stdout, "\x1b[%dA\x1b[J", lines)
		}
		lines = renderAccountMenu(stdout, infos, selected, current, hasCurrent, style)
	}
	render()

	buffer := make([]byte, 8)
	for {
		n, err := stdin.Read(buffer)
		if err != nil {
			return "", err
		}
		key := parseMenuKey(buffer[:n])
		switch key {
		case "up":
			if selected == 0 {
				selected = len(infos) - 1
			} else {
				selected--
			}
			render()
		case "down":
			selected = (selected + 1) % len(infos)
			render()
		case "enter":
			fmt.Fprint(stdout, "\n")
			return infos[selected].Name, nil
		case "cancel":
			fmt.Fprint(stdout, "\n")
			return "", errors.New("No account selected. The operation was cancelled.")
		}
	}
}

func renderAccountMenu(stdout io.Writer, infos []accounts.AccountInfo, selected int, current string, hasCurrent bool, style cliStyle) int {
	lineCount := 0
	title := "Select Codex account"
	if style.enabled {
		title = style.title(title)
	}
	fmt.Fprintln(stdout, title)
	lineCount++
	hint := "Use ↑/↓ or j/k, Enter to confirm, Esc to cancel."
	if style.enabled {
		hint = style.hint(hint)
	}
	fmt.Fprintln(stdout, hint)
	lineCount++

	nameWidth := len("Name")
	emailWidth := len("Email")
	for _, info := range infos {
		if len(info.Name) > nameWidth {
			nameWidth = len(info.Name)
		}
		if len(displayEmail(info.Email)) > emailWidth {
			emailWidth = len(displayEmail(info.Email))
		}
	}
	border := "+" + strings.Repeat("-", 4) + "+" + strings.Repeat("-", len("Active")+2) + "+" + strings.Repeat("-", nameWidth+2) + "+" + strings.Repeat("-", emailWidth+2) + "+"
	fmt.Fprintln(stdout, border)
	fmt.Fprintf(stdout, "| %-2s | %-6s | %-*s | %-*s |\n", "", "Active", nameWidth, "Name", emailWidth, "Email")
	fmt.Fprintln(stdout, border)
	lineCount += 3
	for i, info := range infos {
		cursor := " "
		if i == selected {
			cursor = ">"
		}
		active := ""
		isActive := hasCurrent && current == info.Name
		if isActive {
			active = "*"
		}
		fmt.Fprintf(stdout, "| %-2s | %-6s | %-*s | %-*s |\n", cursor, active, nameWidth, info.Name, emailWidth, displayEmail(info.Email))
		lineCount++
	}
	fmt.Fprintln(stdout, border)
	lineCount++
	return lineCount
}

func parseMenuKey(input []byte) string {
	if len(input) == 0 {
		return ""
	}
	switch input[0] {
	case '\r', '\n':
		return "enter"
	case 3, 27:
		if len(input) >= 3 && input[0] == 27 && input[1] == '[' {
			switch input[2] {
			case 'A':
				return "up"
			case 'B':
				return "down"
			}
		}
		return "cancel"
	case 'k', 'K':
		return "up"
	case 'j', 'J':
		return "down"
	}
	return ""
}

func enableRawMode(stdin *os.File) (func(), error) {
	if runtime.GOOS == "windows" {
		return func() {}, errors.New("raw terminal mode is not available")
	}
	getState := exec.Command("stty", "-g")
	getState.Stdin = stdin
	state, err := getState.Output()
	if err != nil {
		return func() {}, err
	}
	raw := exec.Command("stty", "raw", "-echo")
	raw.Stdin = stdin
	if err := raw.Run(); err != nil {
		return func() {}, err
	}
	return func() {
		restore := exec.Command("stty", strings.TrimSpace(string(state)))
		restore.Stdin = stdin
		_ = restore.Run()
	}, nil
}

func isTerminal(file *os.File) bool {
	stat, err := file.Stat()
	return err == nil && stat.Mode()&os.ModeCharDevice != 0
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
