## ADDED Requirements

### Requirement: Parse skills configuration from ConfigMap
The system SHALL parse an optional `skills` key from the `config.yaml` data in the `alerts-adapter-config` ConfigMap. Each skills entry specifies an OCI image and a list of mount paths.

#### Scenario: ConfigMap contains valid skills entries
- **WHEN** the ConfigMap `config.yaml` contains a `skills` list with entries that have non-empty `image` and non-empty `paths`
- **THEN** the loaded `Config` SHALL contain the corresponding `Skills` entries with their images and paths

#### Scenario: ConfigMap has no skills key
- **WHEN** the ConfigMap `config.yaml` does not contain a `skills` key
- **THEN** the loaded `Config` SHALL have an empty `Skills` slice and no error is logged

#### Scenario: ConfigMap does not exist
- **WHEN** the ConfigMap is not found in the cluster
- **THEN** the loaded `Config` SHALL use defaults for all timing parameters and have an empty `Skills` slice

#### Scenario: Skills entry with empty image
- **WHEN** a skills entry has an empty `image` field
- **THEN** that entry SHALL be skipped and a warning SHALL be logged, and remaining valid entries SHALL still be applied

#### Scenario: Skills entry with empty paths
- **WHEN** a skills entry has a non-empty `image` but an empty `paths` list
- **THEN** that entry SHALL be skipped and a warning SHALL be logged, and remaining valid entries SHALL still be applied

#### Scenario: Mix of valid and invalid skills entries
- **WHEN** the skills list contains both valid entries (non-empty image and paths) and invalid entries (empty image or empty paths)
- **THEN** only the valid entries SHALL be included in the loaded `Config`, and warnings SHALL be logged for each invalid entry
