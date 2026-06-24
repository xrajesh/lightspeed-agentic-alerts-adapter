## MODIFIED Requirements

### Requirement: Poll AlertManager on a fixed interval
The system SHALL poll AlertManager every 30 seconds for firing alerts and process each alert against the filtering and deduplication rules. The filter order SHALL be: receiver allowlist → severity → initial delay → active proposal → cooldown.

#### Scenario: Normal poll cycle
- **WHEN** the poll interval elapses
- **THEN** the system fetches alerts from AlertManager, lists existing Proposals, applies receiver filtering then dedup rules, and creates Proposals for qualifying alerts

#### Scenario: AlertManager unreachable during poll
- **WHEN** the AlertManager API returns an error during a poll cycle
- **THEN** the system logs the error and skips the cycle; the next poll retries

#### Scenario: Kubernetes API unreachable during poll
- **WHEN** the Kubernetes API returns an error during proposal listing or creation
- **THEN** the system logs the error and skips the cycle; the next poll retries
