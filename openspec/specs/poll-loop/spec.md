## Purpose
Continuously poll AlertManager for firing alerts and create Proposal CRs for new alerts, with stateless deduplication to avoid duplicate or premature Proposals.

## Requirements
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

### Requirement: Skip transient alerts (initial delay)
The system SHALL not create a Proposal for an alert that has been firing for less than 5 minutes, to filter out transient alerts that resolve on their own.

#### Scenario: Alert firing for less than initial delay
- **WHEN** `now - alert.startsAt` is less than 5 minutes
- **THEN** the alert is skipped and logged at Debug level

#### Scenario: Alert firing for longer than initial delay
- **WHEN** `now - alert.startsAt` is 5 minutes or more
- **THEN** the alert passes the initial delay check

### Requirement: Skip alerts with active Proposals
The system SHALL not create a Proposal for an alert that already has an active (non-terminal) Proposal, identified by matching the alert fingerprint label.

#### Scenario: Active proposal exists for alert
- **WHEN** a Proposal with matching fingerprint label exists and its phase is Pending, Analyzing, Proposed, Executing, Verifying, or Escalating
- **THEN** the alert is skipped and logged at Debug level

#### Scenario: No proposal exists for alert
- **WHEN** no Proposal with matching fingerprint label exists
- **THEN** the alert passes the active-proposal check

### Requirement: Skip alerts within cooldown window
The system SHALL not create a Proposal for an alert that has a terminal Proposal (Completed, Failed, Denied, Escalated) within the cooldown window of 1 hour, to prevent flooding for flapping alerts.

#### Scenario: Terminal proposal within cooldown
- **WHEN** a Proposal with matching fingerprint label is in a terminal phase and its terminal condition's `LastTransitionTime` is less than 1 hour ago
- **THEN** the alert is skipped and logged at Debug level

#### Scenario: Terminal proposal outside cooldown
- **WHEN** a Proposal with matching fingerprint label is in a terminal phase and its terminal condition's `LastTransitionTime` is 1 hour or more ago
- **THEN** the alert passes the cooldown check

### Requirement: Shut down gracefully on OS signals
The system SHALL exit cleanly when it receives SIGTERM or SIGINT, completing any in-flight poll cycle before stopping.

#### Scenario: SIGTERM received while idle
- **WHEN** the adapter receives SIGTERM between poll cycles
- **THEN** the adapter exits with status code 0

#### Scenario: SIGINT received during poll
- **WHEN** the adapter receives SIGINT during a poll cycle
- **THEN** the adapter completes or cancels the in-flight cycle and exits with status code 0
