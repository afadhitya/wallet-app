# Wallet CLI — Command Reference

> All commands accept `--json` for structured JSON output. Use `wallet <command> --help` to discover available flags and options.

## Transaction

`wallet add expense <amount> <description>`
`wallet add income <amount> <description>`
`wallet add transfer <amount>`
`wallet list`
`wallet edit <id>`
`wallet rm <id>`
`wallet adjust <account> <amount> <description>`

## Account

`wallet account add <name>`
`wallet account list`
`wallet account edit <id>`
`wallet account archive <id>`

## Category

`wallet category list`
`wallet category add <name>`
`wallet category edit <id>`
`wallet category rm <id>`

## Tag

`wallet tag list`
`wallet tag add <name>`
`wallet tag rm <name>`

## Budget

`wallet budget set <name> <amount>`
`wallet budget list`
`wallet budget check`
`wallet budget edit <id>`
`wallet budget rm <id>`

## Bill

`wallet bill add <name> <amount>`
`wallet bill list`
`wallet bill due`
`wallet bill pay <id>`
`wallet bill skip <id>`
`wallet bill pause <id>`
`wallet bill resume <id>`
`wallet bill edit <id>`
`wallet bill rm <id>`

## Forecast

`wallet forecast`
`wallet forecast bills`

## Report

`wallet report`

## Rate

`wallet rate list`
`wallet rate add <currency> <rate>`
`wallet rate set <currency> <rate>`
`wallet rate rm <currency>`

## Init

`wallet init`

## System

`wallet version` — Show the current wallet binary version
`wallet version --check` — Compare against latest GitHub release
`wallet update` — Download and install the latest version from GitHub
`wallet update --force` — Force reinstall even if already at latest
