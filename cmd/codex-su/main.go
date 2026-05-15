package main

import (
	"os"

	"github.com/shayyz-code/codex-su/internal/cli"
)

var version = "dev"

func main() {
	app, err := cli.New(version)
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	os.Exit(app.Run(os.Args[1:]))
}
