## Why

The adapter hardcodes `"default"` as the agent for all three Proposal workflow steps (analysis, execution, verification). Operators deploying custom agents cannot direct alerts to them — every Proposal always uses the default agent regardless of the alert type or operational context.

## What Changes

- Add an optional `agent` configuration section to the ConfigMap that lets operators set a global agent name and per-step overrides (analysis, execution, verification)
- Update `Build()` to accept an agent configuration and use it instead of the hardcoded `"default"` when set
- Fall back to `"default"` when no agent override is configured (backward compatible, no **BREAKING** changes)

## Capabilities

### New Capabilities
- `agent-config`: ConfigMap-based agent name configuration with global and per-step granularity

### Modified Capabilities
- `proposal-building`: Proposal steps use configurable agent names instead of hardcoded `"default"`
- `configmap-config`: ConfigMap parsing includes the new `agent` section

## Impact

- **Code**: `internal/config/config.go` (new struct + parsing), `internal/proposal/build.go` (accept agent config), `internal/adapter/adapter.go` (pass config through)
- **Manifests**: `manifests/configmap.yaml` (document new optional section)
- **APIs**: No CRD changes — uses existing `ProposalStep.Agent` field
- **Dependencies**: None — no new imports
