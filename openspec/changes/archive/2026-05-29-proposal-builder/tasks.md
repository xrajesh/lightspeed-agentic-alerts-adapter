## Tasks

- [ ] Add `github.com/openshift/lightspeed-agentic-operator/api/v1alpha1` dependency and run `go mod tidy`
- [ ] Create `internal/proposal/request.tmpl` with the embedded request template
- [ ] Create `internal/proposal/build.go` with `Build(*models.GettableAlert) (*Proposal, error)` and helper functions for naming, label sanitization, and annotation mapping
- [ ] Create `internal/proposal/build_test.go` with table-driven tests covering: namespaced alert, cluster-scoped alert, long names, invalid characters, missing annotations, deterministic naming
- [ ] Create `internal/proposal/client.go` with `NewClient()` and `CreateProposal(ctx, *Proposal) error`
- [ ] Create `internal/proposal/client_test.go` with tests using a fake controller-runtime client
