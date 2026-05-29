## ADDED Requirements

### Requirement: Filter for actionable alerts
The system SHALL request only active, non-silenced, non-inhibited alerts from the Alertmanager API so the adapter never processes suppressed alerts.

#### Scenario: Silenced alerts excluded
- **WHEN** an alert is silenced in Alertmanager
- **THEN** the alert is not included in the response

#### Scenario: Inhibited alerts excluded
- **WHEN** an alert is inhibited by another alert in Alertmanager
- **THEN** the alert is not included in the response

#### Scenario: Resolved alerts excluded
- **WHEN** an alert has resolved and is no longer active
- **THEN** the alert is not included in the response
