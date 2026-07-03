package main

import (
	"os"

	"github.com/afadhitya/wallet-app/internal/cli"
)

func main() {
	cmd := cli.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
