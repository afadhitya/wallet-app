## Why

Users can already track accounts, transactions, and planned payments, but they cannot see how scheduled income and bills will affect future balances. A forecasting command helps users anticipate cash flow needs and avoid surprises by projecting balances from known planned payments.

## What Changes

- Add balance forecasting based on active, unpaused planned payments.
- Add a `wallet forecast` command with configurable forecast horizon, optional account filtering, and JSON output.
- Add a `wallet forecast bills` command to show upcoming bill impact over a forecast horizon.
- Include projected income, projected expenses, net movement, ending balances, and bill/category breakdowns where applicable.
- Warn when a projected balance becomes negative without treating it as a command failure.

## Capabilities

### New Capabilities
- `forecasting`: Project future account balances and bill impact from planned payments.

### Modified Capabilities

## Impact

- Adds CLI commands under the wallet application command layer.
- Adds service logic for monthly balance projection and bill forecasting.
- Adds sqlc queries for planned payment aggregation and account balance lookup.
- Reuses existing account, category, currency formatting, and planned payment data models.
