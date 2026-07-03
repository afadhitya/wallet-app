package main

import (
	"os"

	"github.com/afadhitya/wallet-app/internal/cli"
)

var exit = os.Exit

func main() {
	cmd := cli.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		exit(1)
	}
}
