## MODIFIED Requirements

### Requirement: Poll AlertManager on a fixed interval
The system SHALL read operational parameters (`pollInterval`, `initialDelay`, `cooldownWindow`) from the `ConfigSource` at the start of each reconcile cycle instead of using hard-coded constants. When the loaded `pollInterval` differs from the current ticker interval, the system SHALL reset the ticker to the new interval.

#### Scenario: Configuration loaded each cycle
- **WHEN** a reconcile cycle begins
- **THEN** the system calls `ConfigSource.Load()` and uses the returned values for that cycle's initial delay check and cooldown window check

#### Scenario: Poll interval changes between cycles
- **WHEN** the `pollInterval` value from `ConfigSource.Load()` differs from the current ticker interval
- **THEN** the system resets the ticker to the new interval and logs the change
