## ADDED Requirements

### Requirement: Load configuration from a ConfigMap each reconcile cycle
The system SHALL read the `alerts-adapter-config` ConfigMap from the adapter's namespace on each reconcile cycle and apply the configuration values for that cycle.

#### Scenario: ConfigMap exists with valid YAML
- **WHEN** the `alerts-adapter-config` ConfigMap exists and its `config.yaml` key contains valid YAML with valid duration values
- **THEN** the system uses the values from the ConfigMap for that reconcile cycle

#### Scenario: ConfigMap exists with partial YAML
- **WHEN** the `alerts-adapter-config` ConfigMap exists and its `config.yaml` key contains valid YAML with only some fields specified
- **THEN** the system uses the specified values and falls back to defaults for missing fields

#### Scenario: ConfigMap exists with invalid duration value
- **WHEN** the `alerts-adapter-config` ConfigMap exists and its `config.yaml` key contains an unparseable duration string for one or more fields
- **THEN** the system falls back to the default value for each invalid field and logs a warning for each invalid field

#### Scenario: ConfigMap exists with invalid YAML
- **WHEN** the `alerts-adapter-config` ConfigMap exists and its `config.yaml` key contains content that is not valid YAML
- **THEN** the system falls back to all default values and logs a warning

#### Scenario: ConfigMap exists without the config.yaml key
- **WHEN** the `alerts-adapter-config` ConfigMap exists but does not contain a `config.yaml` key
- **THEN** the system falls back to all default values

### Requirement: Operate normally when ConfigMap does not exist
The system SHALL NOT fail or exit when the `alerts-adapter-config` ConfigMap does not exist. It SHALL use default values and continue operating.

#### Scenario: ConfigMap does not exist at startup
- **WHEN** the adapter starts and the `alerts-adapter-config` ConfigMap does not exist
- **THEN** the system starts normally using default values and logs at Info level that no ConfigMap was found

#### Scenario: ConfigMap is deleted while adapter is running
- **WHEN** the `alerts-adapter-config` ConfigMap is deleted while the adapter is running
- **THEN** the system reverts to default values on the next reconcile cycle and logs at Info level

### Requirement: Use well-defined default values
The system SHALL use the following default values when no ConfigMap is present or when individual fields are missing or invalid:
- `pollInterval`: 30 seconds
- `initialDelay`: 5 minutes
- `cooldownWindow`: 1 hour

#### Scenario: All defaults applied
- **WHEN** no ConfigMap exists
- **THEN** the system uses `pollInterval=30s`, `initialDelay=5m`, `cooldownWindow=1h`

### Requirement: Resolve the adapter namespace from the environment
The system SHALL read the `POD_NAMESPACE` environment variable to determine the namespace for the ConfigMap. If the variable is not set, it SHALL fall back to `openshift-lightspeed`.

#### Scenario: POD_NAMESPACE is set
- **WHEN** the `POD_NAMESPACE` environment variable is set to `my-namespace`
- **THEN** the system reads the ConfigMap from the `my-namespace` namespace

#### Scenario: POD_NAMESPACE is not set
- **WHEN** the `POD_NAMESPACE` environment variable is not set
- **THEN** the system reads the ConfigMap from the `openshift-lightspeed` namespace
