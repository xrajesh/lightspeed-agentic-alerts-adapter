## Why

The adapter currently creates Proposals with no `spec.tools` configuration — all three workflow steps use the `default` agent with no skills images or paths. Operators deploying custom skills (e.g., ACS skills, cluster-update skills) must manually patch every Proposal after creation or modify the adapter source code. The existing ConfigMap-based external configuration covers only timing parameters (`pollInterval`, `initialDelay`, `cooldownWindow`). Extending it to include skills and their paths allows operators to configure which skills images and paths are injected into every Proposal the adapter creates, without code changes or Pod restarts.

## What Changes

- Extend the `alerts-adapter-config` ConfigMap schema to accept a `skills` section that defines OCI images and their mount paths. These map directly to the `ToolsSpec.Skills` field on the Proposal CRD.
- Modify `proposal.Build()` to accept a tools configuration and set `spec.tools.skills` on the generated Proposal when skills are configured.
- Update the adapter's reconcile loop to pass the loaded tools config through to the proposal builder.
- Update the ConfigMap manifest with example skills configuration.

## Capabilities

### New Capabilities
- `tools-config`: Read skills images and paths from the external ConfigMap and expose them as part of the runtime configuration. Parse and validate skills entries, falling back to no tools (current behavior) when absent or invalid.

### Modified Capabilities
- `proposal-building`: The proposal builder accepts an optional tools configuration and sets `spec.tools.skills` on the Proposal when skills are configured. The existing three-step workflow with the `default` agent remains unchanged.

## Impact

- **Modified**: `internal/config/config.go` — extend `Config` struct and `configFile` struct with skills fields; parse skills from ConfigMap YAML.
- **Modified**: `internal/proposal/build.go` — accept tools config parameter; set `spec.tools.skills` on Proposal.
- **Modified**: `internal/adapter/adapter.go` — pass tools config from loaded config to `proposal.Build()`.
- **Modified**: `manifests/configmap.yaml` — add example skills configuration.
- **Tests**: Update existing tests and add new tests for skills config parsing and proposal building with tools.
- **Dependencies**: No new dependencies. Uses existing `agenticv1alpha1.ToolsSpec`, `SkillsSource` types from the operator API.
