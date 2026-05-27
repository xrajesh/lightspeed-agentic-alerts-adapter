## Why

The lightspeed-agentic-alerts-adapter needs to retrieve active alerts from an OpenShift cluster so that downstream consumers (e.g., an agentic LLM pipeline) can reason about cluster health. This is the foundational data-ingestion capability — without it, the adapter has nothing to work with.

## What Changes

- Add the ability to query alerts from the Alertmanager instance running in an OpenShift cluster.
- Use the alert types provided by the Alertmanager client library directly, avoiding premature abstraction.
- Wire alert retrieval into the adapter's startup so it can fetch and log the current set of firing alerts.

## Capabilities

### New Capabilities
- `alert-retrieval`: Query active/firing alerts from an OpenShift cluster's Alertmanager and return them using the Alertmanager client library types.

### Modified Capabilities
<!-- None — this is a greenfield project with no existing specs. -->

## Impact

- **Code**: New `internal/alertmanager/` package for alert retrieval using the Alertmanager client library. Changes to `cmd/alerts-adapter/main.go` to invoke alert retrieval from `run()`.
- **Infrastructure**: The adapter must run with sufficient permissions to access the cluster's Alertmanager API.
- **Dependencies**: To be determined during design — the choice of client library and authentication approach will be made then.
