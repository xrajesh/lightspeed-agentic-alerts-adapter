## Context

The adapter has two building blocks implemented: an AlertManager client (`internal/alertmanager`) that fetches firing alerts, and a proposal package (`internal/proposal`) that builds and creates Proposal CRs. Today `cmd/alerts-adapter/main.go` calls `GetAlerts` once, logs the results, and exits.

The poll loop connects these pieces into a continuously running controller: fetch alerts, diff against existing Proposals, create new ones, repeat. The dedup logic is entirely stateless — computed fresh each cycle from two authoritative sources (AlertManager API and Kubernetes API).

## Goals / Non-Goals

**Goals:**
- Continuously poll AlertManager and create Proposals for new firing alerts.
- Stateless deduplication: initial delay, active-proposal detection, cooldown after terminal Proposals.
- Graceful shutdown on SIGTERM/SIGINT.
- Testable with fake clients (no real cluster required for unit tests).

**Non-Goals:**
- Health/readiness probes (`/healthz`, `/readyz`) — deferred.
- Leader election or multi-replica coordination — single replica is sufficient.
- Prometheus metrics — deferred.
- Configurable parameters via CR/ConfigMap — constants for now.

## Decisions

### Package structure: `internal/adapter`

The poll loop lives in a new `internal/adapter` package, separate from both the AlertManager client and the proposal package.

```
internal/adapter/
├── adapter.go        # Adapter struct, Run(), reconcile()
└── adapter_test.go   # Table-driven tests with fakes
```

**Why not in `main.go`**: The reconcile logic has meaningful branching (three dedup checks per alert) that needs unit tests. Keeping it in a package with injected dependencies makes it testable without cluster access.

**Why not controller-runtime Manager**: The adapter is not a Kubernetes controller — it doesn't watch resources or reconcile on events. A `time.Ticker` loop is simpler, has no framework overhead, and matches the "poll and diff" design. If we need leader election or health probes later, we can wrap the adapter in a Manager without changing the core logic.

### Interfaces for testability

```go
type AlertSource interface {
    GetAlerts(ctx context.Context) (models.GettableAlerts, error)
}

type ProposalClient interface {
    ListProposals(ctx context.Context) ([]agenticv1alpha1.Proposal, error)
    CreateProposal(ctx context.Context, p *agenticv1alpha1.Proposal) error
}
```

The adapter depends on these interfaces, not concrete types. Tests inject fakes that return canned alerts and proposals. The real implementations are `alertmanager.Client` and `proposal.Client`.

**Alternative considered**: Using `envtest` (controller-runtime's test harness with a real API server). Rejected because it's slow, requires CRD installation, and adds test-infrastructure complexity. The dedup logic is pure functions over data — fakes are the right tool.

### AlertManager query filtering

Add `?active=true&silenced=false&inhibited=false` parameters to `GetAlerts`. This filters at the API level so the adapter never sees suppressed alerts.

**Why modify the client**: These filters are semantically part of "get actionable alerts" — they belong in the client, not scattered across callers. The current `GetAlerts` fetches everything, which includes resolved and silenced alerts the adapter should never act on.

### 409 handling inside `CreateProposal`

`CreateProposal` will check for `errors.IsAlreadyExists` and log at Info level instead of returning an error. From the caller's perspective, a 409 means the Proposal already exists — that's success, not failure.

**Why in the client, not the adapter**: Every caller of `CreateProposal` should treat 409 the same way. Pushing this into the client avoids duplicating the check.

### Dedup logic: determining terminal time for cooldown

The cooldown check needs "when did this Proposal become terminal?" This requires inspecting the condition that caused the terminal phase:

| Terminal Phase | Condition to check |
|---------------|-------------------|
| Completed | `Verified`, Status=True |
| Failed | First condition with Status=False |
| Denied | `Denied`, Status=True |
| Escalated | `Escalated`, Status=True |

A helper function `terminalTime(proposal) *time.Time` will use `DerivePhase()` to classify, then find the relevant condition's `LastTransitionTime`. Returns nil for non-terminal Proposals.

### Signal handling

`main()` uses `signal.NotifyContext` with `SIGTERM` and `SIGINT`. The context is passed to `adapter.Run()`, which exits cleanly when the context is cancelled. No drain logic needed — the adapter is stateless and the next startup will pick up where it left off.

### Logging

| Event | Level | Fields |
|-------|-------|--------|
| Poll cycle start | Debug | — |
| Alert skipped (initial delay) | Debug | alertname, fingerprint, startsAt, threshold |
| Alert skipped (active proposal) | Debug | alertname, fingerprint, proposal |
| Alert skipped (cooldown) | Debug | alertname, fingerprint, proposal, terminalTime |
| Proposal created | Info | alertname, fingerprint, proposal |
| Proposal already exists (409) | Info | alertname, fingerprint, proposal |
| AlertManager error | Error | error |
| Kubernetes list/create error | Error | error |
| Poll cycle complete | Debug | alertsTotal, skipped, created |

## Risks / Trade-offs

**Risk: Large alert volume → slow poll cycles.** Each cycle does 1 AlertManager GET + 1 Kubernetes LIST + N CREATEs. With hundreds of firing alerts, the create calls dominate. → Mitigation: Acceptable for v1 — the architecture targets single-digit to low-tens of alerts per cycle. If alert storms become a concern, batching and token budgets are in the future-work roadmap.

**Risk: Proposal LIST grows unbounded over time.** The adapter lists all Proposals with `source=alertmanager`, including ancient terminal ones. → Mitigation: Kubernetes label-selector queries are efficient (server-side filtering). If the list grows very large, a future iteration can add a `createdAfter` or field selector, or rely on operator-side garbage collection.

**Risk: Clock skew between adapter pod and AlertManager.** The initial-delay check compares `time.Now()` against the alert's `startsAt` from AlertManager. Clock skew could cause premature or delayed Proposal creation. → Mitigation: Both run in the same cluster with NTP. A few seconds of skew is absorbed by the 5-minute initial delay.
