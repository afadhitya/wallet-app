package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/afadhitya/wallet-app/internal/cli"
)

func TestMainSuccess(t *testing.T) {
	exit = func(code int) {
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	}
	defer func() { exit = os.Exit }()

	main()
}

func TestMainError(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"wallet", "nonexistent"}

	var exitCode int
	exit = func(code int) { exitCode = code }
	defer func() { exit = os.Exit }()

	main()

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

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
