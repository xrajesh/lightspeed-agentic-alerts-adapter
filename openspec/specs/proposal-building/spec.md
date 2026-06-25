## Purpose
Translate Alertmanager alerts into Proposal custom resources so the agentic operator can act on firing alerts through its analyze-execute-verify workflow.

## Requirements
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

#### Scenario: Proposal name exceeds 63 characters
- **WHEN** the computed name would exceed 63 characters (the Kubernetes label value limit, since the agentic operator uses the Proposal name as a label value)
- **THEN** the alertname component is truncated to fit within the 63-character limit while preserving the namespace and fingerprint suffix

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

### Requirement: Configure all three workflow steps with tools
The system SHALL set the analysis, execution, and verification steps on the Proposal, each referencing the `default` agent. The system SHALL support shared tools and per-step tool overrides.

#### Scenario: Built Proposal has full workflow
- **WHEN** a Proposal is built from any alert
- **THEN** `spec.analysis`, `spec.execution`, and `spec.verification` all have `agent` set to `"default"`

#### Scenario: Built Proposal with shared skills configured
- **WHEN** a Proposal is built and shared skills configuration is provided with one or more skills entries
- **THEN** `spec.tools.skills` SHALL contain the configured skills entries with their images and paths

#### Scenario: Built Proposal with per-step skills configured
- **WHEN** a Proposal is built and per-step skills are configured for analysis, execution, or verification
- **THEN** the corresponding `spec.{step}.tools.skills` SHALL contain the configured skills entries for that step

#### Scenario: Built Proposal with both shared and per-step skills
- **WHEN** a Proposal is built with both shared skills and per-step skills for a given step
- **THEN** `spec.tools.skills` SHALL contain the shared skills AND `spec.{step}.tools.skills` SHALL contain the per-step skills for steps that have overrides

#### Scenario: Built Proposal with no tools configured
- **WHEN** a Proposal is built and no tools configuration is provided (all slices empty)
- **THEN** `spec.tools` SHALL be omitted from the Proposal (zero value) and no per-step tools SHALL be set

### Requirement: List existing Proposals by source
The system SHALL list Proposal CRs filtered by the `agentic.openshift.io/source=alertmanager` label to support deduplication queries.

#### Scenario: Proposals exist
- **WHEN** ListProposals is called and Proposals with the alertmanager source label exist
- **THEN** the system returns the matching Proposals with their status conditions

#### Scenario: No proposals exist
- **WHEN** ListProposals is called and no Proposals with the alertmanager source label exist
- **THEN** the system returns an empty list and no error

#### Scenario: Kubernetes API error
- **WHEN** the Kubernetes API returns an error during listing
- **THEN** ListProposals returns a wrapped error with context

### Requirement: Create Proposal resources in the cluster
The system SHALL provide a Kubernetes client that creates Proposal CRs using controller-runtime with in-cluster config. The client SHALL return a boolean indicating whether the Proposal was created, and treat 409 AlreadyExists as a non-error.

#### Scenario: Successful creation
- **WHEN** CreateProposal is called with a valid Proposal
- **THEN** the Proposal is created in the cluster, returns true and no error

#### Scenario: Proposal already exists
- **WHEN** the Kubernetes API returns 409 AlreadyExists
- **THEN** CreateProposal logs at Info level and returns false and no error

#### Scenario: Creation failure
- **WHEN** the Kubernetes API returns a non-409 error
- **THEN** CreateProposal returns false and a wrapped error with context
