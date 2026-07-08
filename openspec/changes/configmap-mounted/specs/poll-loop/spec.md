## MODIFIED Requirements

### Requirement: Poll AlertManager on a fixed interval
The system SHALL use the configuration provided at startup for the poll interval and all filtering/deduplication parameters. The poll interval is fixed for the lifetime of the process. The filter order SHALL be: receiver allowlist → severity → initial delay → active AgenticRun → cooldown.

#### Scenario: Normal poll cycle
- **WHEN** the poll interval elapses
- **THEN** the system fetches alerts from AlertManager, lists existing AgenticRuns, applies receiver filtering then dedup rules, and creates AgenticRuns for qualifying alerts

#### Scenario: Configuration used from startup
- **WHEN** a reconcile cycle begins
- **THEN** the system uses the config values loaded at startup for that cycle's initial delay check, cooldown window check, and receiver filtering

#### Scenario: AlertManager unreachable during poll
- **WHEN** the AlertManager API returns an error during a poll cycle
- **THEN** the system logs the error and skips the cycle; the next poll retries

#### Scenario: Kubernetes API unreachable during poll
- **WHEN** the Kubernetes API returns an error during AgenticRun listing or creation
- **THEN** the system logs the error and skips the cycle; the next poll retries

## REMOVED Requirements

### Requirement: Poll interval changes between cycles
**Reason**: The adapter no longer reloads config each cycle. The poll interval is fixed at startup and changes require a pod restart (triggered by the operator).
**Migration**: Config changes are applied by the operator restarting the adapter pod when the ConfigMap changes.
