package main

import (
	"os"

	"github.com/shayyz-code/codex-su/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(cli.Execute(version, os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
