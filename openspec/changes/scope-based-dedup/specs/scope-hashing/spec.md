## ADDED Requirements

### Requirement: Compute a stable fingerprint from alert labels
The system SHALL compute a stable fingerprint by removing a configurable set of volatile labels from the alert's label set, sorting the remaining key=value pairs lexicographically, and hashing the result with SHA256 truncated to 8 hex characters.

#### Scenario: Alert with no volatile labels
- **WHEN** an alert has labels `{alertname=HighCPU, namespace=myns, container=app}` and the volatile labels list is `[pod, instance, endpoint, uid]`
- **THEN** the stable fingerprint is `SHA256("alertname=HighCPU\ncontainer=app\nnamespace=myns")[:8]`

#### Scenario: Alert with volatile labels present
- **WHEN** an alert has labels `{alertname=KubePodCrashLooping, namespace=myns, pod=app-abc123, container=app}` and the volatile labels list is `[pod, instance, endpoint, uid]`
- **THEN** the `pod` label is removed and the stable fingerprint is `SHA256("alertname=KubePodCrashLooping\ncontainer=app\nnamespace=myns")[:8]`

#### Scenario: Two alerts differing only in volatile labels produce the same fingerprint
- **WHEN** alert A has labels `{alertname=X, namespace=ns, pod=pod-aaa}` and alert B has labels `{alertname=X, namespace=ns, pod=pod-bbb}` and `pod` is in the volatile labels list
- **THEN** both alerts produce the same stable fingerprint

#### Scenario: Two alerts differing in non-volatile labels produce different fingerprints
- **WHEN** alert A has labels `{alertname=X, namespace=ns, container=foo}` and alert B has labels `{alertname=X, namespace=ns, container=bar}`
- **THEN** the alerts produce different stable fingerprints

#### Scenario: Alert with nil fingerprint
- **WHEN** an alert has a nil AlertManager fingerprint
- **THEN** the system returns an error (existing behavior preserved)

#### Scenario: Empty volatile labels list
- **WHEN** the volatile labels list is empty
- **THEN** all alert labels are included in the hash (no labels stripped)
