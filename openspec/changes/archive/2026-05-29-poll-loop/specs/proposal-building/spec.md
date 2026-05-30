## ADDED Requirements

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

## MODIFIED Requirements

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
