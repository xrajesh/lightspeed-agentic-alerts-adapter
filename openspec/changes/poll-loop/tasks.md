## 1. AlertManager Client: Add Query Filters

- [x] 1.1 Add `active=true`, `silenced=false`, `inhibited=false` query parameters to `GetAlerts` in `internal/alertmanager/client.go`
- [x] 1.2 Update `internal/alertmanager/client_test.go` to verify filtered parameters are sent

## 2. Proposal Client: Add List and 409 Handling

- [x] 2.1 Add `ListProposals(ctx) ([]Proposal, error)` to `internal/proposal/client.go`, filtered by label `agentic.openshift.io/source=alertmanager`
- [x] 2.2 Handle `errors.IsAlreadyExists` in `CreateProposal`: log at Info and return nil instead of an error
- [x] 2.3 Add tests for `ListProposals` and 409 handling in `internal/proposal/client_test.go` using a fake controller-runtime client

## 3. Adapter Package: Poll Loop and Dedup Logic

- [x] 3.1 Define `AlertSource` and `ProposalClient` interfaces in `internal/adapter/adapter.go`
- [x] 3.2 Implement `Adapter` struct with `Run(ctx) error` (ticker loop, exits on context cancellation) and `reconcile(ctx)` (fetch alerts, list proposals, dedup, create)
- [x] 3.3 Implement dedup helpers: `skipInitialDelay(alert, now, threshold)`, `hasActiveProposal(alert, proposals)`, `inCooldown(alert, proposals, window)`, `terminalTime(proposal) *time.Time`
- [x] 3.4 Write table-driven tests in `internal/adapter/adapter_test.go` covering: new alert creates proposal, transient alert skipped (initial delay), active proposal skipped, terminal proposal within cooldown skipped, terminal proposal past cooldown creates new proposal, AlertManager error skips cycle, Kubernetes error skips cycle

## 4. Main: Signal Handling and Wiring

- [x] 4.1 Rewrite `cmd/alerts-adapter/main.go`: signal.NotifyContext for SIGTERM/SIGINT, construct real clients, wire into `adapter.Run(ctx)`
- [x] 4.2 Update `cmd/alerts-adapter/main_test.go` for the new `run()` signature
