## MODIFIED Requirements

### Requirement: Configure all three workflow steps
The system SHALL set the analysis, execution, and verification steps on the Proposal, each referencing the `default` agent. When skills are provided, the system SHALL set `spec.tools.skills` on the Proposal with the configured skills images and paths.

#### Scenario: Built Proposal has full workflow
- **WHEN** a Proposal is built from any alert
- **THEN** `spec.analysis`, `spec.execution`, and `spec.verification` all have `agent` set to `"default"`

#### Scenario: Built Proposal with skills configured
- **WHEN** a Proposal is built and skills configuration is provided with one or more skills entries
- **THEN** `spec.tools.skills` SHALL contain the configured skills entries with their images and paths

#### Scenario: Built Proposal with no skills configured
- **WHEN** a Proposal is built and no skills configuration is provided (nil or empty)
- **THEN** `spec.tools` SHALL be omitted from the Proposal (zero value)
