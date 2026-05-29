## Why

The adapter can retrieve alerts from Alertmanager but has no way to act on them. The agentic operator expects Proposal CRs as its input — without a translation layer, alerts remain log lines instead of actionable workflows.

## What Changes

- Add a `Build` function that converts an Alertmanager `GettableAlert` into a `Proposal` CR, mapping alert metadata into labels, annotations, target namespaces, and a templated request prompt.
- Add a thin Kubernetes client wrapper that creates Proposal resources using controller-runtime with in-cluster config.
- Use an embedded Go template for the request field so the prompt can evolve independently of the code.

## Capabilities

### New Capabilities
- `proposal-building`: Translate Alertmanager alerts into Proposal CRs with deterministic naming, Kubernetes-safe metadata, and a structured request for the analysis agent.

### Modified Capabilities
<!-- None -->

## Impact

- **Code**: New `internal/proposal/` package with `build.go`, `client.go`, `request.tmpl`, and corresponding tests.
- **Dependencies**: `github.com/openshift/lightspeed-agentic-operator/api/v1alpha1`, `k8s.io/apimachinery`, `sigs.k8s.io/controller-runtime`.
- **Infrastructure**: The adapter's ServiceAccount will need RBAC to create Proposals (not part of this change).
