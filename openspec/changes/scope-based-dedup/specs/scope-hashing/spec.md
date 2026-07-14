## ADDED Requirements

### Requirement: Compute a stable fingerprint from alert labels
The system SHALL compute a stable fingerprint by removing a configurable set of ignored labels from the alert's label set, sorting the remaining key=value pairs lexicographically, and hashing the result with FNV-64a (Fowler–Noll–Vo 1a, 64-bit variant) truncated to 8 hex characters.

#### Scenario: Alert with no ignored labels
- **WHEN** an alert has labels `{alertname=HighCPU, namespace=myns, container=app}` and the ignored labels list is `[pod, instance, endpoint, uid]`
- **THEN** the stable fingerprint is `FNV64a("alertname=HighCPU\ncontainer=app\nnamespace=myns")[:8]`

#### Scenario: Alert with ignored labels present
- **WHEN** an alert has labels `{alertname=KubePodCrashLooping, namespace=myns, pod=app-abc123, container=app}` and the ignored labels list is `[pod, instance, endpoint, uid]`
- **THEN** the `pod` label is removed and the stable fingerprint is `FNV64a("alertname=KubePodCrashLooping\ncontainer=app\nnamespace=myns")[:8]`

#### Scenario: Two alerts differing only in ignored labels produce the same fingerprint
- **WHEN** alert A has labels `{alertname=X, namespace=ns, pod=pod-aaa}` and alert B has labels `{alertname=X, namespace=ns, pod=pod-bbb}` and `pod` is in the ignored labels list
- **THEN** both alerts produce the same stable fingerprint

#### Scenario: Two alerts differing in non-ignored labels produce different fingerprints
- **WHEN** alert A has labels `{alertname=X, namespace=ns, container=foo}` and alert B has labels `{alertname=X, namespace=ns, container=bar}`
- **THEN** the alerts produce different stable fingerprints

#### Scenario: Alert with nil fingerprint
- **WHEN** an alert has a nil AlertManager fingerprint
- **THEN** the system returns an error (existing behavior preserved)

#### Scenario: Empty ignored labels list
- **WHEN** the ignored labels list is empty
- **THEN** all alert labels are included in the hash (no labels stripped)
