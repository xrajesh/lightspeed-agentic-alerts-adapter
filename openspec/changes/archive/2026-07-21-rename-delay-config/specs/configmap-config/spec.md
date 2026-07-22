## MODIFIED Requirements

### Requirement: Use well-defined default values
The system SHALL use the following default values when no ConfigMap is present or when individual fields are missing:
- `pollInterval`: 30 seconds
- `preRunDelay`: 0 seconds
- `postRunDelay`: 1 hour
- `allowedReceivers`: empty list (no alerts are processed until explicitly configured)

#### Scenario: All defaults applied
- **WHEN** no ConfigMap exists
- **THEN** the system uses `pollInterval=30s`, `preRunDelay=0s`, `postRunDelay=1h`, `allowedReceivers=[]`

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
- **THEN** the system logs an error and returns a config load error, causing the reconcile loop to fail

#### Scenario: ConfigMap exists with explicit zero preRunDelay
- **WHEN** the `alerts-adapter-config` ConfigMap contains `preRunDelay: 0s`
- **THEN** the system uses `preRunDelay=0s` (same as default, no delay)

#### Scenario: ConfigMap exists with explicit zero postRunDelay
- **WHEN** the `alerts-adapter-config` ConfigMap contains `postRunDelay: 0s`
- **THEN** the system uses `postRunDelay=0s` (overrides the 1h default, no delay)

#### Scenario: ConfigMap exists with negative preRunDelay or postRunDelay
- **WHEN** the `alerts-adapter-config` ConfigMap contains a negative duration for `preRunDelay` or `postRunDelay`
- **THEN** the system clamps the value to `0s` (no error logged)

#### Scenario: ConfigMap exists with invalid YAML
- **WHEN** the `alerts-adapter-config` ConfigMap exists and its `config.yaml` key contains content that is not valid YAML
- **THEN** the system logs an error and returns a config load error, causing the reconcile loop to fail

#### Scenario: ConfigMap exists without the config.yaml key
- **WHEN** the `alerts-adapter-config` ConfigMap exists but does not contain a `config.yaml` key
- **THEN** the system falls back to all default values

## RENAMED Requirements

### Requirement: ConfigMap key initialDelay
- **FROM:** `initialDelay`
- **TO:** `preRunDelay`

### Requirement: ConfigMap key cooldownWindow
- **FROM:** `cooldownWindow`
- **TO:** `postRunDelay`

## REMOVED Requirements

### Requirement: Non-positive duration fallback to nonzero default
**Reason**: Zero is now a valid value meaning "no delay". When a field is explicitly set to 0s or a negative value, it is clamped to 0s, overriding the default. When a field is absent, the default is used (`preRunDelay`: 0s, `postRunDelay`: 1h).
**Migration**: No action needed. The `pollInterval` field retains its existing validation (must be positive, falls back to 30s default).
