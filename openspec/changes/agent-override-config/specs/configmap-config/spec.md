## MODIFIED Requirements

### Requirement: Load configuration from a ConfigMap each reconcile cycle
The system SHALL read the `alerts-adapter-config` ConfigMap from the adapter's namespace on each reconcile cycle and apply the configuration values for that cycle, including the optional agent configuration.

#### Scenario: ConfigMap exists with valid YAML
- **WHEN** the `alerts-adapter-config` ConfigMap exists and its `config.yaml` key contains valid YAML with valid duration values
- **THEN** the system uses the values from the ConfigMap for that reconcile cycle

#### Scenario: ConfigMap exists with partial YAML
- **WHEN** the `alerts-adapter-config` ConfigMap exists and its `config.yaml` key contains valid YAML with only some fields specified
- **THEN** the system uses the specified values and falls back to defaults for missing fields

#### Scenario: ConfigMap exists with agent section
- **WHEN** the `alerts-adapter-config` ConfigMap exists and its `config.yaml` key contains an `agent` section with `default` and/or per-step overrides
- **THEN** the system populates `Config.Agent` with the specified values for that reconcile cycle

#### Scenario: ConfigMap exists without agent section
- **WHEN** the `alerts-adapter-config` ConfigMap exists and its `config.yaml` key does not contain an `agent` section
- **THEN** the system uses an empty `AgentConfig` (all fields empty, resulting in hardcoded `"default"` agent for all steps)

#### Scenario: ConfigMap exists with invalid duration value
- **WHEN** the `alerts-adapter-config` ConfigMap exists and its `config.yaml` key contains an unparseable duration string for one or more fields
- **THEN** the system falls back to the default value for each invalid field and logs a warning for each invalid field

#### Scenario: ConfigMap exists with non-positive duration value
- **WHEN** the `alerts-adapter-config` ConfigMap exists and its `config.yaml` key contains a zero or negative duration (e.g., `pollInterval: 0s` or `initialDelay: -1m`)
- **THEN** the system falls back to the default value for each non-positive field and logs a warning for each non-positive field

#### Scenario: ConfigMap exists with invalid YAML
- **WHEN** the `alerts-adapter-config` ConfigMap exists and its `config.yaml` key contains content that is not valid YAML
- **THEN** the system falls back to all default values and logs a warning

#### Scenario: ConfigMap exists without the config.yaml key
- **WHEN** the `alerts-adapter-config` ConfigMap exists but does not contain a `config.yaml` key
- **THEN** the system falls back to all default values
