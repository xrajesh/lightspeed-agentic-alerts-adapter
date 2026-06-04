## ADDED Requirements

### Requirement: Skip alerts with low severity
The adapter SHALL skip alerts whose `severity` label is `none` or `info` (case-insensitive) during reconciliation. Skipped alerts MUST NOT result in Proposal CR creation. The severity check SHALL be performed before all other skip checks (initial delay, active proposal, cooldown).

#### Scenario: Alert with severity none is skipped
- **WHEN** an alert has severity label `none`
- **THEN** the adapter skips the alert and does not create a Proposal

#### Scenario: Alert with severity info is skipped
- **WHEN** an alert has severity label `info`
- **THEN** the adapter skips the alert and does not create a Proposal

#### Scenario: Alert with severity warning is processed
- **WHEN** an alert has severity label `warning`
- **THEN** the adapter processes the alert through remaining filters and may create a Proposal

#### Scenario: Alert with severity critical is processed
- **WHEN** an alert has severity label `critical`
- **THEN** the adapter processes the alert through remaining filters and may create a Proposal

#### Scenario: Case-insensitive severity matching
- **WHEN** an alert has severity label `Info` or `NONE` (mixed case)
- **THEN** the adapter skips the alert

#### Scenario: Alert with missing severity label is processed
- **WHEN** an alert has no `severity` label
- **THEN** the adapter processes the alert through remaining filters (does not skip)

### Requirement: Log skipped alerts
The adapter SHALL log each severity-skipped alert at debug level, including the alert name, fingerprint, and severity value.

#### Scenario: Debug log for skipped alert
- **WHEN** an alert is skipped due to low severity
- **THEN** the adapter logs a debug message with alertname, fingerprint, and severity
