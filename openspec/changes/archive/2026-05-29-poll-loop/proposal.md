## Why

The adapter can fetch alerts and build Proposals, but the two are not connected. Today `main()` fetches once, logs, and exits. There is no loop, no deduplication, and no Proposal creation. Without the poll loop the adapter is a one-shot diagnostic tool, not a running controller.

## What Changes

- Add an `internal/adapter` package that wires AlertManager retrieval and Proposal creation into a continuous poll loop with stateless deduplication.
- Extend `internal/proposal.Client` with `ListProposals` (filtered by `source=alertmanager` label) and 409-safe creation (log and swallow `AlreadyExists`).
- Rewrite `cmd/alerts-adapter/main.go` to run the adapter loop under OS signal handling (`SIGTERM`/`SIGINT`), replacing the current fetch-and-exit behavior.
- Filter alerts at the AlertManager API level (`active=true`, `silenced=false`, `inhibited=false`).

## Capabilities

### New Capabilities
- `poll-loop`: Continuously poll AlertManager for firing alerts and create Proposal CRs, with stateless deduplication based on initial delay, active-proposal detection, and cooldown windows.

### Modified Capabilities
- `alert-retrieval`: Add query parameters to filter for active, non-silenced, non-inhibited alerts.
- `proposal-building`: Add `ListProposals` for dedup queries and handle 409 AlreadyExists as success.

## Impact

- **Code**: New `internal/adapter/` package (`adapter.go`, `adapter_test.go`). Modified `internal/proposal/client.go` (add List, 409 handling). Modified `cmd/alerts-adapter/main.go` (signal handling, adapter wiring).
- **Dependencies**: None new — already using `controller-runtime/client` and `alertmanager/api/v2`.
- **Infrastructure**: No changes — RBAC for Proposal list/get is already in the ClusterRole defined in the architecture.
