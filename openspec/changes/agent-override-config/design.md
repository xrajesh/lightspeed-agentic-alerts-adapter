## Context

The adapter builds AgenticRun CRs with all three workflow steps (analysis, execution, verification) hardcoded to the `"default"` agent (`internal/agenticrun/build.go:78-80`). The operator supports custom agents, but the adapter provides no way to select them. Configuration already supports tools/skills overrides via ConfigMap (`/etc/alerts-adapter/config.yaml`), so adding agent configuration follows the same pattern.

## Goals / Non-Goals

**Goals:**
- Allow operators to configure which agent handles AgenticRun steps via the existing ConfigMap
- Support a global default agent and optional per-step overrides (analysis, execution, verification)
- Maintain full backward compatibility — omitting the config results in the current `"default"` behavior

**Non-Goals:**
- Per-alert agent routing (e.g., mapping specific alert names to specific agents) — future work
- Validating that the configured agent name exists in the cluster — the operator handles that
- Environment variable or CLI flag overrides for agent config

## Decisions

### 1. Extend existing ConfigMap YAML with an `agent` section

Add a new top-level `agent` key to `config.yaml` with a `default` field and optional per-step overrides:

```yaml
agent:
  default: "my-agent"
  analysis: "analysis-specialist"
  execution: "execution-specialist"
  verification: "verification-specialist"
```

**Rationale**: Follows the established pattern of the `tools` config. A flat structure (`agent.default`, `agent.analysis`) is simpler than nesting agent names inside the existing `analysis`/`execution`/`verification` step entries, and avoids conflating agent identity with tools configuration.

**Alternative considered**: Putting the agent name inside the existing step entries (e.g., `analysis.agent: "foo"`). Rejected because it mixes two concerns — tool configuration and agent routing — and would complicate the YAML structure for the common case of setting a single global agent.

### 2. New `AgentConfig` struct in `internal/config`

```go
type AgentConfig struct {
    Default      string
    Analysis     string
    Execution    string
    Verification string
}
```

`Build()` receives `AgentConfig` alongside `ToolsConfig`. For each step, it uses the per-step agent if set, then falls back to `AgentConfig.Default`, then falls back to the hardcoded `"default"` constant.

**Rationale**: The three-level fallback (per-step → global → hardcoded default) gives operators granular control while keeping zero-config deployment unchanged. Passing `AgentConfig` as a separate parameter (not embedded in `ToolsConfig`) keeps the two concerns decoupled.

### 3. Validation: empty strings are ignored, no existence checks

Empty agent name strings are treated as "not configured" and fall through to the next fallback level. No validation is performed against the cluster — the operator validates agent references when it processes the AgenticRun.

**Rationale**: Consistent with how the adapter handles tools config (no validation of image references). Keeps the adapter simple and stateless.

## Risks / Trade-offs

- **Typo in agent name silently accepted** → The operator will fail the AgenticRun with a clear error in its status conditions. The adapter logs which agent was assigned to each AgenticRun at creation time, making misconfiguration diagnosable.
- **No per-alert routing** → Operators who need different agents for different alert types must deploy multiple adapter instances with different configs. This is acceptable for the initial implementation and can be extended later without breaking changes.
