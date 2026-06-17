## Context

The adapter creates Proposals with `spec.tools` empty — no skills images or paths are configured. The `ToolsSpec` and `SkillsSource` types are already defined in the operator API (`agenticv1alpha1`), and the Proposal CRD supports `spec.tools.skills` as a list of OCI images with mount paths. The existing `internal/config` package reads `alerts-adapter-config` ConfigMap on each reconcile cycle and exposes timing parameters. The `proposal.Build()` function constructs Proposals from alert data but does not accept any tools configuration.

## Goals / Non-Goals

**Goals:**
- Allow operators to specify skills images and paths via the existing `alerts-adapter-config` ConfigMap.
- Configured skills appear on every Proposal the adapter creates, in `spec.tools.skills`.
- When no skills are configured (default), Proposals are created with no `spec.tools` — preserving current behavior.
- Invalid skills entries are skipped with warning logs; valid entries are still applied.

**Non-Goals:**
- Per-step tools overrides (e.g., different skills for analysis vs execution) — all steps share the same tools.
- MCP server or required secrets configuration via ConfigMap — only skills are covered.
- Per-alert skills selection (e.g., different skills based on alert name or severity).
- Skills image validation beyond basic non-empty checks.

## Decisions

### 1. Extend the existing ConfigMap YAML schema with a `skills` key

Add a `skills` key to the `config.yaml` data that mirrors the Proposal CRD's `ToolsSpec.Skills` structure:

```yaml
config.yaml: |
  pollInterval: 30s
  initialDelay: 5m
  cooldownWindow: 1h
  skills:
    - image: registry.redhat.io/openshift-lightspeed/agentic-skills:latest
      paths:
        - /skills/prometheus
        - /skills/cluster-diagnostics
```

Each entry has `image` (OCI pullspec) and `paths` (list of absolute paths to mount). This directly maps to `agenticv1alpha1.SkillsSource` — no translation layer needed.

**Alternative considered:** A separate ConfigMap key (e.g., `tools.yaml`) or a separate ConfigMap entirely. Rejected because the single `config.yaml` key is already established, and adding a second key or ConfigMap increases operator complexity for a simple list of skills.

### 2. Add `Skills` field to the `Config` struct using the operator API type

```go
type Config struct {
    PollInterval   time.Duration
    InitialDelay   time.Duration
    CooldownWindow time.Duration
    Skills         []agenticv1alpha1.SkillsSource
}
```

Using the operator API type directly avoids translation code and ensures the config struct always produces valid `ToolsSpec.Skills` entries. The `internal/config` package already imports `agenticv1alpha1` transitively via the project's dependency.

**Alternative considered:** A separate config-local type (e.g., `SkillConfig{Image string, Paths []string}`) that is later mapped to `SkillsSource`. Rejected because it adds a mapping step with no benefit — the API type is stable and already available.

### 3. Change `Build()` signature to accept skills configuration

Change `Build(alert) → Build(alert, skills)` where `skills` is `[]agenticv1alpha1.SkillsSource`. When the slice is non-empty, `spec.tools.skills` is set on the Proposal. When nil or empty, `spec.tools` is omitted (zero value via `omitzero`), preserving current behavior.

**Alternative considered:** Pass the entire `Config` struct to `Build()`. Rejected because `Build()` has no use for timing parameters — passing only what it needs keeps the interface focused.

### 4. Validate skills entries during config parsing

During ConfigMap parsing, skip entries with empty `image` or empty `paths`. Log a warning for each skipped entry. Valid entries are applied even if some entries are invalid — partial config is better than no config.

Paths are not validated beyond non-emptiness — the Proposal CRD has CEL validation rules that enforce path format (absolute, no `..`, etc.). The adapter does not duplicate that validation.

### 5. Use `yaml` struct tags on a config-local skills struct for parsing

The `agenticv1alpha1.SkillsSource` type uses `json` tags, not `yaml` tags. Define a local `skillsEntry` struct with `yaml` tags for unmarshalling, then convert to `SkillsSource`:

```go
type skillsEntry struct {
    Image string   `yaml:"image"`
    Paths []string `yaml:"paths"`
}
```

This avoids depending on the `yaml` library's ability to read `json` tags, which is not guaranteed across YAML library versions.

## Risks / Trade-offs

- **Skills config increases ConfigMap complexity** → The `skills` key is optional and the adapter works identically without it. The ConfigMap manifest includes a commented example.
- **No per-step tools override** → All steps get the same skills. This is the common case for the alerts adapter; per-step overrides can be added later if needed.
- **CRD validation rejects bad paths at create time** → If an operator configures an invalid path (e.g., relative path), the Proposal creation will fail with a Kubernetes validation error. The adapter logs this as a creation failure and continues — the alert will be retried on the next cycle. This is acceptable because path format errors are operator configuration bugs, not runtime issues.
