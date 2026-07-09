## ADDED Requirements

### Requirement: Support volatile labels configuration
The system SHALL support a `volatileLabels` field in the ConfigMap YAML that specifies which alert labels to exclude when computing the stable fingerprint. When the field is absent, the system SHALL use the default list: `[pod, instance, endpoint, uid]`. When the field is explicitly set, the specified list fully replaces the default (no merging).

#### Scenario: volatileLabels not specified in ConfigMap
- **WHEN** the ConfigMap does not contain a `volatileLabels` field
- **THEN** the system uses the default volatile labels: `pod`, `instance`, `endpoint`, `uid`

#### Scenario: volatileLabels set to a custom list
- **WHEN** the ConfigMap contains `volatileLabels: [pod, instance, job]`
- **THEN** the system uses exactly `[pod, instance, job]` as the volatile labels (the defaults `endpoint` and `uid` are not included)

#### Scenario: volatileLabels set to an empty list
- **WHEN** the ConfigMap contains `volatileLabels: []`
- **THEN** no labels are stripped — all alert labels are included in the stable fingerprint hash

#### Scenario: No ConfigMap exists
- **WHEN** no ConfigMap exists
- **THEN** the system uses the default volatile labels: `pod`, `instance`, `endpoint`, `uid`
