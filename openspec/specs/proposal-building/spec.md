### Requirement: Build a Proposal CR from a single alert
The system SHALL convert an Alertmanager `GettableAlert` into a `Proposal` custom resource with deterministic naming, Kubernetes-safe metadata, and a templated request for the analysis agent.

#### Scenario: Alert with namespace label
- **WHEN** the alert has a `namespace` label
- **THEN** the Proposal name is `{alertname}-{namespace}-{fingerprint[:8]}`, `spec.targetNamespaces` is set to `[namespace]`, and the Proposal is created in `openshift-lightspeed`

#### Scenario: Cluster-scoped alert (no namespace)
- **WHEN** the alert has no `namespace` label
- **THEN** the Proposal name is `{alertname}-{fingerprint[:8]}`, `spec.targetNamespaces` is omitted, and the Proposal is created in `openshift-lightspeed`

#### Scenario: Deterministic naming produces idempotent creates
- **WHEN** the same alert is passed to Build twice
- **THEN** both calls produce Proposals with identical names, enabling Kubernetes 409 deduplication

### Requirement: Sanitize alert data for Kubernetes metadata
The system SHALL sanitize alert values to conform to Kubernetes naming and label restrictions.

#### Scenario: Proposal name contains invalid DNS characters
- **WHEN** the alertname or namespace contains characters not allowed in DNS subdomain names
- **THEN** those characters are replaced with hyphens and the result is lowercased

#### Scenario: Proposal name exceeds 253 characters
- **WHEN** the computed name would exceed 253 characters
- **THEN** the alertname component is truncated to fit within the limit while preserving the namespace and fingerprint suffix

#### Scenario: Label value exceeds 63 characters
- **WHEN** an alert field used as a label value exceeds 63 characters
- **THEN** the value is truncated to 63 characters and trimmed of trailing non-alphanumeric characters

#### Scenario: Label value contains invalid characters
- **WHEN** an alert field used as a label value contains characters not allowed in Kubernetes labels
- **THEN** those characters are replaced with hyphens and leading/trailing non-alphanumeric characters are trimmed

### Requirement: Render a structured request from alert data
The system SHALL render the `spec.request` field using an embedded Go template that includes the alert name, severity, namespace, summary, description, and all labels.

#### Scenario: Alert with all annotation fields populated
- **WHEN** the alert has summary and description annotations
- **THEN** both are included in the rendered request

#### Scenario: Alert with missing annotations
- **WHEN** the alert has no summary or description annotations
- **THEN** the corresponding fields are empty in the rendered request and no error is returned

### Requirement: Configure all three workflow steps
The system SHALL set the analysis, execution, and verification steps on the Proposal, each referencing the `default` agent.

#### Scenario: Built Proposal has full workflow
- **WHEN** a Proposal is built from any alert
- **THEN** `spec.analysis`, `spec.execution`, and `spec.verification` all have `agent` set to `"default"`

### Requirement: Create Proposal resources in the cluster
The system SHALL provide a Kubernetes client that creates Proposal CRs using controller-runtime with in-cluster config.

#### Scenario: Successful creation
- **WHEN** CreateProposal is called with a valid Proposal
- **THEN** the Proposal is created in the cluster and no error is returned

#### Scenario: Creation failure
- **WHEN** the Kubernetes API returns an error
- **THEN** CreateProposal returns a wrapped error with context
