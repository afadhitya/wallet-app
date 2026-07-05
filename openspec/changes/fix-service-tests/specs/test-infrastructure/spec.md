## ADDED Requirements

### Requirement: Service tests configure rate config
The service test helper `setupService()` SHALL configure an in-memory rate configuration so that tests exercising `AddExpense`, `AddIncome`, and any operations that call `GetBaseCurrency()` do not depend on the filesystem for rate data.

#### Scenario: Service test runs without filesystem rate config
- **WHEN** a service test calls `setupService()` to create a service instance
- **THEN** the service instance has an in-memory rate configuration with base currency IDR and no foreign exchange rates
- **AND** the rate configuration is reset to the default filesystem-based loader after the test completes
