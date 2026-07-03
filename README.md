# Wallet App

A CLI-first wallet application built in Go.

## Prerequisites

- Go 1.25+
- [sqlc](https://sqlc.dev) (for code generation, install with `brew install sqlc` or `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

## Quality Commands

```sh
make build           # Build the wallet binary to bin/wallet
make run             # Build and run the wallet binary
make install         # Install the wallet binary to $GOPATH/bin
make test            # Run unit tests
make test-cover      # Run tests with coverage profile and HTML report
make coverage-check  # Run tests and enforce 100% coverage
make lint            # Run golangci-lint across all packages
make fmt             # Format Go source files
make tidy            # Clean up go.mod and go.sum
make verify-deps     # Verify module dependencies are unchanged
make sqlc-gen        # Generate Go code from SQL queries
make clean           # Remove generated artifacts
```
