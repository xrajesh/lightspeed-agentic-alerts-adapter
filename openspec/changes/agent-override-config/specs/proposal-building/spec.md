## MODIFIED Requirements

### Requirement: Configure all three workflow steps with tools
The system SHALL set the analysis, execution, and verification steps on the Proposal, each referencing a configurable agent name resolved from the agent configuration. The system SHALL support shared tools and per-step tool overrides.

#### Scenario: Built Proposal has full workflow with default agent
- **WHEN** a Proposal is built from any alert and no agent configuration is provided
- **THEN** `spec.analysis`, `spec.execution`, and `spec.verification` all have `agent` set to `"default"`

#### Scenario: Built Proposal with global agent override
- **WHEN** a Proposal is built and the agent configuration has `Default` set to `"my-agent"` with no per-step overrides
- **THEN** `spec.analysis`, `spec.execution`, and `spec.verification` all have `agent` set to `"my-agent"`

#### Scenario: Built Proposal with per-step agent overrides
- **WHEN** a Proposal is built and the agent configuration has per-step overrides for analysis, execution, and verification
- **THEN** each step's `agent` field SHALL be set to the corresponding per-step override value

#### Scenario: Built Proposal with mixed agent configuration
- **WHEN** a Proposal is built with `AgentConfig.Default` set to `"global-agent"`, `AgentConfig.Analysis` set to `"analyzer"`, and `AgentConfig.Execution` empty
- **THEN** `spec.analysis.agent` SHALL be `"analyzer"`, `spec.execution.agent` SHALL be `"global-agent"`, and `spec.verification.agent` SHALL use the fallback chain

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
