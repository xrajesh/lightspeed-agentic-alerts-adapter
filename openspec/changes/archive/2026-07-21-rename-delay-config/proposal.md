## Why

The names `initialDelay` and `cooldownWindow` are vague and inconsistent. `preRunDelay` and `postRunDelay` clearly communicate their relationship to the AgenticRun lifecycle: `preRunDelay` prevents an AgenticRun from being created until the alert has been firing for a minimum time, filtering out transient alerts that resolve on their own. `postRunDelay` prevents a new AgenticRun from being created when the alert is still firing after a previous run has finished, avoiding repeated analysis of an alert that has already been investigated. The `preRunDelay` default changes from 5m to 0s so the adapter acts immediately on new alerts. The `postRunDelay` default stays at 1h to avoid repeated analysis of an alert that has already been investigated. Both can be explicitly set to `0s` to disable the delay.

## What Changes

- **BREAKING**: Rename ConfigMap key `initialDelay` to `preRunDelay`
- **BREAKING**: Rename ConfigMap key `cooldownWindow` to `postRunDelay`
- Change default for `preRunDelay` from 5m to 0s (act immediately on new alerts)
- Keep default for `postRunDelay` at 1h (avoid repeated analysis of an alert that has already been investigated)
- Allow explicit `0s` to disable either delay (distinguishes "not set" from "set to 0")
- Negative values are clamped to 0s silently

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `configmap-config`: Rename `initialDelay`/`cooldownWindow` keys to `preRunDelay`/`postRunDelay`, default `preRunDelay` to 0s and `postRunDelay` to 1h, allow explicit 0s to disable delays
- `poll-loop`: Update references to use the renamed config fields; adapt dedup logic to handle 0s delays (skip the check entirely when delay is 0)

## Impact

- **Config**: Go struct fields, constants, YAML keys, and validation logic in `internal/config/`
- **Adapter**: References in `internal/adapter/` (field access, helper functions, log messages)
- **Docs**: README.md, ARCHITECTURE.md, AGENTS.md, manifests/configmap.yaml
- **Tests**: All tests referencing the old names or old defaults
- **Breaking**: Any existing ConfigMap using `initialDelay` or `cooldownWindow` will need to be updated
