## MODIFIED Requirements

### Requirement: Build a Proposal CR from a single alert
The system SHALL convert an Alertmanager `GettableAlert` into a `Proposal` custom resource with deterministic naming, Kubernetes-safe metadata, and a templated request for the analysis agent. The `agentic.openshift.io/alert-fingerprint` label SHALL use the stable fingerprint computed from the alert's labels minus volatile labels, instead of AlertManager's fingerprint.

#### Scenario: Alert with namespace label
- **WHEN** the alert has a `namespace` label
- **THEN** the Proposal name is `{alertname}-{namespace}-{alertmanager_fingerprint[:8]}`, `spec.targetNamespaces` is set to `[namespace]`, the `agentic.openshift.io/alert-fingerprint` label is set to the stable fingerprint, and the Proposal is created in `openshift-lightspeed`

#### Scenario: Cluster-scoped alert (no namespace)
- **WHEN** the alert has no `namespace` label
- **THEN** the Proposal name is `{alertname}-{alertmanager_fingerprint[:8]}`, `spec.targetNamespaces` is omitted, the `agentic.openshift.io/alert-fingerprint` label is set to the stable fingerprint, and the Proposal is created in `openshift-lightspeed`

#### Scenario: Deterministic naming produces idempotent creates
- **WHEN** the same alert is passed to Build twice
- **THEN** both calls produce Proposals with identical names, enabling Kubernetes 409 deduplication

#### Scenario: Two alerts for the same problem produce Proposals with the same fingerprint label
- **WHEN** two alerts differ only in volatile labels (e.g., different pod names) but represent the same underlying problem
- **THEN** both Proposals have the same `agentic.openshift.io/alert-fingerprint` label value but different Proposal names
