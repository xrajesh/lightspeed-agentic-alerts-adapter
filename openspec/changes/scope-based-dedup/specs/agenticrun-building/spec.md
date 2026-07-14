## MODIFIED Requirements

### Requirement: Build an AgenticRun CR from a single alert
The system SHALL convert an Alertmanager `GettableAlert` into an `AgenticRun` custom resource with deterministic naming, Kubernetes-safe metadata, and a templated request for the analysis agent. The `agentic.openshift.io/alert-fingerprint` label SHALL use the stable fingerprint computed from the alert's labels minus ignored labels, instead of AlertManager's fingerprint.

#### Scenario: Alert with namespace label
- **WHEN** the alert has a `namespace` label
- **THEN** the AgenticRun name is `{alertname}-{namespace}-{startsAt_hash}`, `spec.targetNamespaces` is set to `[namespace]`, the `agentic.openshift.io/alert-fingerprint` label is set to the stable fingerprint, and the AgenticRun is created in `openshift-lightspeed`

#### Scenario: Cluster-scoped alert (no namespace)
- **WHEN** the alert has no `namespace` label
- **THEN** the AgenticRun name is `{alertname}-{startsAt_hash}`, `spec.targetNamespaces` is omitted, the `agentic.openshift.io/alert-fingerprint` label is set to the stable fingerprint, and the AgenticRun is created in `openshift-lightspeed`

#### Scenario: Deterministic naming produces idempotent creates
- **WHEN** the same alert is passed to Build twice
- **THEN** both calls produce AgenticRuns with identical names, enabling Kubernetes 409 deduplication for the exact same alert instance

#### Scenario: Two alerts for the same problem produce AgenticRuns with the same fingerprint label
- **WHEN** two alerts differ only in ignored labels (e.g., different pod names) but represent the same underlying problem
- **THEN** both AgenticRuns have the same `agentic.openshift.io/alert-fingerprint` label value but different AgenticRun names. Deduplication across these alerts relies on `hasActiveRun` matching on the label, not on 409
