## 1. Configuration Parsing

- [x] 1.1 Add `AgentConfig` struct to `internal/config/config.go` with `Default`, `Analysis`, `Execution`, `Verification` string fields
- [x] 1.2 Add `Agent AgentConfig` field to the `Config` struct
- [x] 1.3 Add `agentEntry` to `configFile` struct with YAML tag `agent` and fields `Default`, `Analysis`, `Execution`, `Verification`
- [x] 1.4 Parse the agent section in `LoadFromFile` and populate `cfg.Agent` from the deserialized `configFile`
- [x] 1.5 Add tests for agent config parsing: global only, per-step only, mixed, missing section, empty strings

## 2. Proposal Building

- [x] 2.1 Add `config.AgentConfig` parameter to `Build()` function signature
- [x] 2.2 Add `resolveAgent` helper that implements the three-level fallback: per-step → global default → hardcoded `"default"`
- [x] 2.3 Replace hardcoded `defaultAgent` assignment with `resolveAgent` calls for each step
- [x] 2.4 Update all existing callers of `Build()` in `internal/adapter/adapter.go` to pass `cfg.Agent`
- [x] 2.5 Add tests for `Build()` with agent overrides: no config (backward compat), global override, per-step overrides, mixed config

## 3. Test Fixture Updates

- [x] 3.1 Update existing `Build()` tests to pass the new `AgentConfig` parameter (zero value for backward compat)

## 4. Manifest and Documentation

- [x] 4.1 Update `manifests/configmap.yaml` with a commented-out `agent` section showing the available fields
