## Why

The adapter currently creates Proposals with no `spec.tools` configuration — all three workflow steps use the `default` agent with no skills images or paths. Operators deploying custom skills (e.g., ACS skills, cluster-update skills) must manually patch every Proposal after creation or modify the adapter source code. The existing ConfigMap-based external configuration covers only timing parameters (`pollInterval`, `initialDelay`, `cooldownWindow`). Extending it to include skills configuration allows operators to configure which skills images and paths are injected into Proposals the adapter creates, without code changes or Pod restarts.

The Proposal CRD supports skills at four levels: `spec.tools` (shared default for all steps) and per-step overrides via `spec.analysis.tools`, `spec.execution.tools`, `spec.verification.tools`. Per-step tools replace the shared default for that step. The ConfigMap schema mirrors this hierarchy so operators can assign different skills to different workflow steps (e.g., diagnostic skills for analysis, remediation skills for execution).

## What Changes

- **BREAKING**: Replace the flat `skills` key in the ConfigMap schema with a hierarchical `tools.skills` key for shared skills and optional `analysis.tools.skills`, `execution.tools.skills`, `verification.tools.skills` for per-step overrides, mirroring the Proposal CRD's `ToolsSpec` structure.
- Restructure `Config` struct to hold shared tools and per-step tool overrides.
- Modify `proposal.Build()` to accept a tools configuration struct and set `spec.tools`, `spec.analysis.tools`, `spec.execution.tools`, `spec.verification.tools` on the generated Proposal as appropriate.
- Update the adapter's reconcile loop to pass the loaded tools config through to the proposal builder.
- Update the ConfigMap manifest with example configuration showing both shared and per-step skills.

## Capabilities

### New Capabilities
- `tools-config`: Read skills configuration from the external ConfigMap supporting shared tools (`tools.skills`) and per-step overrides (`analysis.tools.skills`, `execution.tools.skills`, `verification.tools.skills`). Parse and validate skills entries, falling back to no tools (current behavior) when absent or invalid.

### Modified Capabilities
- `proposal-building`: The proposal builder accepts a tools configuration and sets `spec.tools.skills` for shared skills, and `spec.{analysis,execution,verification}.tools.skills` for per-step overrides. Per-step tools are only set when explicitly configured. The existing three-step workflow with the `default` agent remains unchanged.

## Impact

- **Modified**: `internal/config/config.go` — restructure `Config` struct to hold shared and per-step tools config; replace flat `skills` parsing with hierarchical `tools`/`analysis.tools`/`execution.tools`/`verification.tools` parsing.
- **Modified**: `internal/proposal/build.go` — accept tools config struct; set shared and per-step skills on the Proposal.
- **Modified**: `internal/adapter/adapter.go` — pass tools config from loaded config to `proposal.Build()`.
- **Modified**: `manifests/configmap.yaml` — update example to show shared and per-step skills configuration.
- **Tests**: Update existing tests and add new tests for hierarchical skills config parsing and proposal building with per-step tools.
- **Dependencies**: No new dependencies. Uses existing `agenticv1alpha1.ToolsSpec`, `SkillsSource` types from the operator API.
