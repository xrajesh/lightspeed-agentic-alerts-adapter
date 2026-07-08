## Purpose
Parse and validate tools/skills configuration from the alerts-adapter-config ConfigMap so the adapter can populate shared and per-step skills on AgenticRun custom resources.

## Requirements
### Requirement: Parse shared skills configuration from ConfigMap
The system SHALL parse an optional `tools.skills` key from the `config.yaml` data in the `alerts-adapter-config` ConfigMap. Each skills entry specifies an OCI image and a list of mount paths. These shared skills map to `spec.tools.skills` on the AgenticRun.

#### Scenario: ConfigMap contains valid shared skills entries
- **WHEN** the ConfigMap `config.yaml` contains a `tools.skills` list with entries that have non-empty `image` and non-empty `paths`
- **THEN** the loaded `Config` SHALL contain the corresponding shared skills entries with their images and paths

#### Scenario: ConfigMap has no tools key
- **WHEN** the ConfigMap `config.yaml` does not contain a `tools` key
- **THEN** the loaded `Config` SHALL have empty shared skills and no error is logged

#### Scenario: ConfigMap does not exist
- **WHEN** the ConfigMap is not found in the cluster
- **THEN** the loaded `Config` SHALL use defaults for all timing parameters and have empty tools config

### Requirement: Parse per-step skills configuration from ConfigMap
The system SHALL parse optional `analysis.tools.skills`, `execution.tools.skills`, and `verification.tools.skills` keys from the `config.yaml` data. Per-step skills map to `spec.{analysis,execution,verification}.tools.skills` on the AgenticRun and replace the shared default for that step.

#### Scenario: ConfigMap contains per-step skills for analysis
- **WHEN** the ConfigMap `config.yaml` contains `analysis.tools.skills` with valid entries
- **THEN** the loaded `Config` SHALL contain the corresponding analysis skills entries

#### Scenario: ConfigMap contains per-step skills for execution and verification
- **WHEN** the ConfigMap `config.yaml` contains `execution.tools.skills` and `verification.tools.skills` with valid entries
- **THEN** the loaded `Config` SHALL contain the corresponding execution and verification skills entries

#### Scenario: Only shared tools configured, no per-step overrides
- **WHEN** the ConfigMap `config.yaml` contains `tools.skills` but no `analysis.tools.skills`, `execution.tools.skills`, or `verification.tools.skills`
- **THEN** the loaded `Config` SHALL have shared skills populated and all per-step skills empty

#### Scenario: Only per-step tools configured, no shared tools
- **WHEN** the ConfigMap `config.yaml` contains per-step skills but no `tools.skills`
- **THEN** the loaded `Config` SHALL have empty shared skills and per-step skills populated

### Requirement: Validate skills entries
The system SHALL validate each skills entry in both shared and per-step contexts.

#### Scenario: Skills entry with empty image
- **WHEN** a skills entry has an empty `image` field
- **THEN** that entry SHALL be skipped and a warning SHALL be logged indicating the context (shared or per-step), and remaining valid entries SHALL still be applied

#### Scenario: Skills entry with empty paths
- **WHEN** a skills entry has a non-empty `image` but an empty `paths` list
- **THEN** that entry SHALL be skipped and a warning SHALL be logged indicating the context (shared or per-step), and remaining valid entries SHALL still be applied

#### Scenario: Mix of valid and invalid skills entries
- **WHEN** a skills list contains both valid entries (non-empty image and paths) and invalid entries (empty image or empty paths)
- **THEN** only the valid entries SHALL be included in the loaded `Config`, and warnings SHALL be logged for each invalid entry
