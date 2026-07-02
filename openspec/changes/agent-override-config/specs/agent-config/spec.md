## ADDED Requirements

### Requirement: Parse agent configuration from ConfigMap
The system SHALL parse an optional `agent` section from the `config.yaml` data in the `alerts-adapter-config` ConfigMap. The section supports a `default` field for the global agent name and optional per-step overrides (`analysis`, `execution`, `verification`).

#### Scenario: ConfigMap contains a global default agent
- **WHEN** the ConfigMap `config.yaml` contains `agent.default` set to `"my-agent"`
- **THEN** the loaded `Config` SHALL have `AgentConfig.Default` set to `"my-agent"`

#### Scenario: ConfigMap contains per-step agent overrides
- **WHEN** the ConfigMap `config.yaml` contains `agent.analysis` set to `"analyzer"`, `agent.execution` set to `"executor"`, and `agent.verification` set to `"verifier"`
- **THEN** the loaded `Config` SHALL have the corresponding `AgentConfig` fields set to those values

#### Scenario: ConfigMap contains only a global agent with no per-step overrides
- **WHEN** the ConfigMap `config.yaml` contains `agent.default` set to `"my-agent"` but no `agent.analysis`, `agent.execution`, or `agent.verification`
- **THEN** the loaded `Config` SHALL have `AgentConfig.Default` set to `"my-agent"` and all per-step fields empty

#### Scenario: ConfigMap has no agent section
- **WHEN** the ConfigMap `config.yaml` does not contain an `agent` key
- **THEN** the loaded `Config` SHALL have an empty `AgentConfig` (all fields empty strings) and no error is logged

#### Scenario: ConfigMap does not exist
- **WHEN** the ConfigMap is not found in the cluster
- **THEN** the loaded `Config` SHALL use defaults for all parameters and have an empty `AgentConfig`

### Requirement: Resolve the effective agent for each Proposal step
The system SHALL resolve the agent name for each workflow step using a three-level fallback: per-step override → global default → hardcoded `"default"`.

#### Scenario: Per-step agent set for analysis
- **WHEN** `AgentConfig.Analysis` is `"analyzer"` and `AgentConfig.Default` is `"my-agent"`
- **THEN** the analysis step SHALL use `"analyzer"`

#### Scenario: Per-step agent not set, global default set
- **WHEN** `AgentConfig.Analysis` is empty and `AgentConfig.Default` is `"my-agent"`
- **THEN** the analysis step SHALL use `"my-agent"`

#### Scenario: No agent config set at all
- **WHEN** both `AgentConfig.Analysis` and `AgentConfig.Default` are empty
- **THEN** the analysis step SHALL use `"default"`

#### Scenario: Mixed per-step and global configuration
- **WHEN** `AgentConfig.Default` is `"global-agent"`, `AgentConfig.Analysis` is `"analyzer"`, `AgentConfig.Execution` is empty, and `AgentConfig.Verification` is `"verifier"`
- **THEN** analysis SHALL use `"analyzer"`, execution SHALL use `"global-agent"`, and verification SHALL use `"verifier"`
