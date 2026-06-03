# Lightspeed Agentic Alerts Adapter

A Go component that polls OpenShift AlertManager for firing alerts and creates `Proposal` CRs (`agentic.openshift.io/v1alpha1`) to trigger automated remediation via the Lightspeed Agentic operator. Stateless, single-replica, create-only design — no internal state, diffs AlertManager vs Kubernetes API each cycle.

## Commands

```sh
make build          # → ./bin/alerts-adapter
make test           # go test ./...
make lint           # golangci-lint run ./...
make fmt            # go fmt ./...
make vet            # go vet ./...
make coverage       # HTML coverage report → coverage.html
make container-build # podman build

# Run a single test
go test -run TestFunctionName ./internal/adapter/

# Run a single subtest
go test -run TestFunctionName/subtest_name ./internal/adapter/
```

## Architecture

Three internal packages, each behind an interface, wired together in `cmd/alerts-adapter/main.go`:

- **`internal/alertmanager`** — AlertManager API client. Reads bearer token on every call (handles rotation). TLS via in-cluster CA. Implements `adapter.AlertSource`.
- **`internal/proposal`** — Two concerns: `build.go` translates an alert into a `Proposal` CR (deterministic name from fingerprint, embedded Go template `request.tmpl` for the request field); `client.go` wraps controller-runtime to create/list Proposals. Implements `adapter.ProposalClient`.
- **`internal/adapter`** — Poll loop (`Run` → `reconcile` on ticker). Stateless deduplication: skips alerts below `initialDelay` (5m), with an active (non-terminal) Proposal, or within `cooldownWindow` (1h) of a terminal Proposal. Matching is by `alert-fingerprint` label (first 8 chars).

The Proposal CRD types come from `github.com/openshift/lightspeed-agentic-operator/api`.

## Key design decisions

- Polls (not webhooks) for resilience — restart immediately sees all firing alerts.
- Proposals always created in `openshift-lightspeed` namespace.
- 409 AlreadyExists on create is expected and handled as a no-op (returns `false, nil`).
- Fingerprint prefix (8 chars) is used for both Proposal naming and dedup matching (`proposal.FingerprintLen`).
- Terminal phases: Completed, Failed, Denied, Escalated.

## Conventions

- Commit messages: [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `test:`, `docs:`, `refactor:`).
- Structured logging with `log/slog` (JSON handler), passed explicitly — no globals except `slog.SetDefault` in main.
- Interfaces defined in the consumer package (`adapter`), not the provider.
- Tests use table-driven style.
