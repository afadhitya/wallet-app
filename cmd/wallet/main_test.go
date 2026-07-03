package main

import (
	"bytes"
	"testing"

	"github.com/afadhitya/wallet-app/internal/cli"
)

func TestRootCmd(t *testing.T) {
	cmd := cli.NewRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("root command should not error on empty args: %v", err)
	}
}
