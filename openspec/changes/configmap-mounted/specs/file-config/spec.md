## ADDED Requirements

### Requirement: Load configuration from a file at startup
The system SHALL read configuration from a YAML file at the path `/etc/alerts-adapter/config.yaml` once at startup and use the resulting values for the entire lifetime of the process.

#### Scenario: Config file exists with valid YAML
- **WHEN** the file at `/etc/alerts-adapter/config.yaml` exists and contains valid YAML with valid duration values
- **THEN** the system uses the values from the file

#### Scenario: Config file exists with partial YAML
- **WHEN** the file exists and contains valid YAML with only some fields specified
- **THEN** the system uses the specified values and falls back to defaults for missing fields

#### Scenario: Config file exists with invalid duration value
- **WHEN** the file exists and contains an unparseable duration string for one or more fields
- **THEN** the system falls back to the default value for each invalid field and logs an error for each invalid field

#### Scenario: Config file exists with non-positive duration value
- **WHEN** the file exists and contains a zero or negative duration (e.g., `pollInterval: 0s` or `initialDelay: -1m`)
- **THEN** the system falls back to the default value for each non-positive field and logs an error for each non-positive field

#### Scenario: Config file exists with invalid YAML
- **WHEN** the file exists and contains content that is not valid YAML
- **THEN** the system falls back to all default values and logs an error

#### Scenario: Config file does not exist
- **WHEN** the file at `/etc/alerts-adapter/config.yaml` does not exist
- **THEN** the system falls back to all default values and logs an error

### Requirement: Use well-defined default values
The system SHALL use the following default values when the config file is missing, unreadable, or when individual fields are missing or invalid:
- `pollInterval`: 30 seconds
- `initialDelay`: 5 minutes
- `cooldownWindow`: 1 hour
- `allowedReceivers`: empty list (no alerts are processed until explicitly configured)

#### Scenario: All defaults applied
- **WHEN** the config file does not exist or cannot be read
- **THEN** the system uses `pollInterval=30s`, `initialDelay=5m`, `cooldownWindow=1h`, `allowedReceivers=[]`
