## 1. Config Package — Hierarchical Tools Parsing

- [x] 1.1 Replace `Skills []agenticv1alpha1.SkillsSource` on `Config` with `Tools ToolsConfig` where `ToolsConfig` has `Shared`, `Analysis`, `Execution`, `Verification` fields (each `[]agenticv1alpha1.SkillsSource`)
- [x] 1.2 Replace flat `skillsEntry` and `skills` field on `configFile` with hierarchical structs: `toolsEntry` (with `Skills []skillsEntry`), `stepEntry` (with `Tools toolsEntry`), and corresponding fields on `configFile` (`Tools`, `Analysis`, `Execution`, `Verification`)
- [x] 1.3 Refactor `parseSkills` to accept a step label for log messages and call it for shared and each per-step skills list in `Load()`
- [x] 1.4 Update unit tests: replace flat skills tests with hierarchical tests covering shared-only, per-step-only, both shared and per-step, mixed valid/invalid entries across levels, and empty/missing config

## 2. Proposal Builder — Accept Tools Config

- [x] 2.1 Change `Build()` signature from `Build(alert, skills)` to `Build(alert, tools)` where `tools` is `config.ToolsConfig`
- [x] 2.2 Set `spec.tools.skills` from `tools.Shared` when non-empty; set `spec.{analysis,execution,verification}.tools.skills` from the corresponding per-step slice when non-empty; leave zero values when empty (omitzero)
- [x] 2.3 Update existing `Build()` tests to pass zero-value `ToolsConfig` (preserving current behavior)
- [x] 2.4 Add tests for `Build()` with shared skills only, per-step skills only, both shared and per-step, and empty config

## 3. Adapter Integration

- [x] 3.1 Update the `reconcile()` method in `internal/adapter/adapter.go` to pass `cfg.Tools` (ToolsConfig) to `proposal.Build()`
- [x] 3.2 Update adapter tests to account for the new `Build()` parameter type

## 4. Manifest Update

- [x] 4.1 Update `manifests/configmap.yaml` with commented examples showing shared tools and per-step tools configuration
