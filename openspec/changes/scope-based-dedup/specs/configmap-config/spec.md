## ADDED Requirements

### Requirement: Restructure config YAML with filtering and deduplication sections
The system SHALL support a structured config YAML where `allowedReceivers` is nested under a `filtering` section and `ignoredLabels` is nested under a `deduplication` section. Top-level `allowedReceivers` SHALL continue to be accepted for backward compatibility.

### Requirement: Support ignored labels configuration
The system SHALL support a `deduplication.ignoredLabels` field in the ConfigMap YAML that specifies which alert labels to exclude when computing the stable fingerprint. When the field is absent, the system SHALL use the default list: `[pod, instance, endpoint, uid]`. When the field is explicitly set, the specified list fully replaces the default (no merging).

#### Scenario: ignoredLabels not specified in ConfigMap
- **WHEN** the ConfigMap does not contain a `deduplication.ignoredLabels` field
- **THEN** the system uses the default ignored labels: `pod`, `instance`, `endpoint`, `uid`

#### Scenario: ignoredLabels set to a custom list
- **WHEN** the ConfigMap contains `deduplication.ignoredLabels: [pod, instance, job]`
- **THEN** the system uses exactly `[pod, instance, job]` as the ignored labels (the defaults `endpoint` and `uid` are not included)

#### Scenario: ignoredLabels set to an empty list
- **WHEN** the ConfigMap contains `deduplication.ignoredLabels: []`
- **THEN** no labels are ignored — all alert labels are included in the stable fingerprint hash

#### Scenario: No ConfigMap exists
- **WHEN** no ConfigMap exists
- **THEN** the system uses the default ignored labels: `pod`, `instance`, `endpoint`, `uid`

#### Scenario: allowedReceivers under filtering section
- **WHEN** the ConfigMap contains `filtering.allowedReceivers: [Critical]`
- **THEN** the system uses `[Critical]` as the allowed receivers list
