## Purpose
Filter alerts based on a configurable allowlist of AlertManager receivers, ensuring only alerts routed to approved receivers are processed for AgenticRun creation.

## Requirements
### Requirement: Skip alerts not routed to an allowed receiver
The system SHALL skip any alert whose AlertManager receivers do not include at least one entry from the configured `allowedReceivers` list. Comparison SHALL be case-insensitive.

#### Scenario: Alert has a matching receiver
- **WHEN** an alert has receivers `["Critical", "slack-oncall"]` and the allowlist is `["Critical"]`
- **THEN** the alert passes the receiver filter

#### Scenario: Alert has no matching receiver
- **WHEN** an alert has receivers `["Default", "slack-info"]` and the allowlist is `["Critical"]`
- **THEN** the alert is skipped and logged at Debug level with its receiver names

#### Scenario: Alert has no receivers
- **WHEN** an alert has an empty receivers list
- **THEN** the alert is skipped and logged at Debug level

#### Scenario: Case-insensitive matching
- **WHEN** an alert has receivers `["critical"]` and the allowlist is `["Critical"]`
- **THEN** the alert passes the receiver filter

#### Scenario: Allowlist is empty
- **WHEN** the allowlist is empty (explicitly set to `[]`)
- **THEN** all alerts are skipped

### Requirement: Configure allowed receivers via ConfigMap
The system SHALL read the `allowedReceivers` field from the `alerts-adapter-config` ConfigMap's `config.yaml` data key as a YAML list of receiver name strings.

#### Scenario: Receivers configured in ConfigMap
- **WHEN** the ConfigMap contains `allowedReceivers: ["Critical", "PagerDuty"]`
- **THEN** the system uses `["Critical", "PagerDuty"]` as the allowlist

#### Scenario: Receivers field absent from ConfigMap
- **WHEN** the ConfigMap exists but does not contain an `allowedReceivers` key
- **THEN** the system uses the default empty allowlist and no alerts are processed

#### Scenario: Receivers field is empty list
- **WHEN** the ConfigMap contains `allowedReceivers: []`
- **THEN** the system uses an empty allowlist and no alerts are processed

#### Scenario: ConfigMap not found
- **WHEN** the `alerts-adapter-config` ConfigMap does not exist
- **THEN** the system uses the default empty allowlist and no alerts are processed

### Requirement: Log the active receiver allowlist
The system SHALL log the effective `allowedReceivers` list at Info level at startup and when the configuration is reloaded.

#### Scenario: Startup with default receivers
- **WHEN** the adapter starts and no ConfigMap overrides the receivers
- **THEN** the system logs the default empty allowlist at Info level

#### Scenario: Startup with custom receivers
- **WHEN** the adapter starts and the ConfigMap sets `allowedReceivers: ["Critical", "PagerDuty"]`
- **THEN** the system logs `["Critical", "PagerDuty"]` at Info level
