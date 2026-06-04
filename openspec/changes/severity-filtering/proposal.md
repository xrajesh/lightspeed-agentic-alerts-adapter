## Why

The adapter currently processes all firing alerts regardless of severity. Alerts with severity `none` or `info` are typically informational and do not warrant automated remediation via Proposal CRs. Processing them creates unnecessary Proposals that clutter the system and waste agentic operator resources on non-actionable alerts.

## What Changes

- Add severity-based filtering to the reconcile loop so alerts with severity `none` or `info` are skipped before Proposal creation.
- Log skipped alerts at debug level for observability.

## Capabilities

### New Capabilities
- `severity-filtering`: Filter out alerts based on their severity label during reconciliation. Alerts with severity `none` or `info` are skipped.

### Modified Capabilities

## Impact

- `internal/adapter/adapter.go` — new severity check in the reconcile loop alongside existing dedup filters.
- No API changes, no new dependencies, no breaking changes.
- Existing alerts with severity `warning`, `critical`, or any other value continue to be processed as before.
