## Context

The adapter creates Proposals with `spec.tools` empty — no skills images or paths are configured. The `ToolsSpec` and `SkillsSource` types are already defined in the operator API (`agenticv1alpha1`), and the Proposal CRD supports skills at four levels: `spec.tools.skills` as a shared default for all steps, and per-step overrides via `spec.analysis.tools.skills`, `spec.execution.tools.skills`, `spec.verification.tools.skills`. Per-step tools replace the shared default for that step.

The existing `internal/config` package reads `alerts-adapter-config` ConfigMap on each reconcile cycle and exposes timing parameters. The `proposal.Build()` function constructs Proposals from alert data but does not accept any tools configuration.

## Goals / Non-Goals

**Goals:**
- Allow operators to specify shared skills via `tools.skills` in the ConfigMap, applied to `spec.tools.skills` on every Proposal.
- Allow operators to specify per-step skills overrides via `analysis.tools.skills`, `execution.tools.skills`, `verification.tools.skills` in the ConfigMap. Per-step tools map to the corresponding `spec.{step}.tools.skills` on the Proposal and replace the shared default for that step.
- When no tools are configured (default), Proposals are created with no `spec.tools` and no per-step tools — preserving current behavior.
- Invalid skills entries are skipped with warning logs; valid entries are still applied.

**Non-Goals:**
- MCP server or required secrets configuration via ConfigMap — only skills are covered.
- Per-alert skills selection (e.g., different skills based on alert name or severity).
- Skills image validation beyond basic non-empty checks.

## Decisions

### 1. Hierarchical ConfigMap YAML schema mirroring the Proposal CRD

Replace the flat `skills` key with a hierarchical structure that mirrors the CRD:

```yaml
config.yaml: |
  pollInterval: 30s
  initialDelay: 5m
  cooldownWindow: 1h

  # Shared tools for all steps (maps to spec.tools)
  tools:
    skills:
      - image: registry.redhat.io/openshift-lightspeed/agentic-skills:latest
        paths:
          - /skills/prometheus
          - /skills/cluster-diagnostics

  # Per-step overrides (replace shared tools for that step)
  analysis:
    tools:
      skills:
        - image: registry.redhat.io/openshift-lightspeed/analysis-skills:latest
          paths:
            - /skills/diagnostic
  execution:
    tools:
      skills:
        - image: registry.redhat.io/openshift-lightspeed/exec-skills:latest
          paths:
            - /skills/remediation
  verification:
    tools:
      skills:
        - image: registry.redhat.io/openshift-lightspeed/verify-skills:latest
          paths:
            - /skills/validation
```

This directly mirrors the CRD's structure, making the mapping obvious and reducing cognitive overhead for operators who are already familiar with the Proposal CRD.

**Alternative considered:** Keep the flat `skills` key and add separate `analysisSkills`, `executionSkills`, `verificationSkills` keys. Rejected because it invents a new schema shape that doesn't match the CRD, requiring operators to learn two different structures.

### 2. Introduce a `ToolsConfig` struct to hold shared and per-step tools

```go
type ToolsConfig struct {
    Shared       []agenticv1alpha1.SkillsSource
    Analysis     []agenticv1alpha1.SkillsSource
    Execution    []agenticv1alpha1.SkillsSource
    Verification []agenticv1alpha1.SkillsSource
}
```

The `Config` struct replaces the `Skills` field with a single `Tools ToolsConfig` field. This groups all tools-related configuration and makes the `Build()` interface clean — it receives one struct instead of four separate slices.

**Alternative considered:** Four separate fields on `Config` (`SharedSkills`, `AnalysisSkills`, etc.). Rejected because it scatters related config and makes the `Build()` signature unwieldy.

### 3. Change `Build()` signature to accept `ToolsConfig`

Change `Build(alert, skills) → Build(alert, tools)` where `tools` is `ToolsConfig`. The builder sets:
- `spec.tools.skills` from `tools.Shared` (when non-empty)
- `spec.analysis.tools.skills` from `tools.Analysis` (when non-empty)
- `spec.execution.tools.skills` from `tools.Execution` (when non-empty)
- `spec.verification.tools.skills` from `tools.Verification` (when non-empty)

When all slices are empty, the Proposal is created with no tools — preserving current behavior via `omitzero`.

### 4. Use `yaml` struct tags on config-local structs for parsing

Define local structs with `yaml` tags that mirror the hierarchical YAML structure:

```go
type configFile struct {
    PollInterval   Duration       `yaml:"pollInterval"`
    InitialDelay   Duration       `yaml:"initialDelay"`
    CooldownWindow Duration       `yaml:"cooldownWindow"`
    Tools          toolsEntry     `yaml:"tools"`
    Analysis       stepEntry      `yaml:"analysis"`
    Execution      stepEntry      `yaml:"execution"`
    Verification   stepEntry      `yaml:"verification"`
}

type toolsEntry struct {
    Skills []skillsEntry `yaml:"skills"`
}

type stepEntry struct {
    Tools toolsEntry `yaml:"tools"`
}

type skillsEntry struct {
    Image string   `yaml:"image"`
    Paths []string `yaml:"paths"`
}
```

The `agenticv1alpha1.SkillsSource` type uses `json` tags, not `yaml` tags. Using local structs for unmarshalling avoids depending on the YAML library's ability to read `json` tags.

### 5. Validate skills entries during config parsing with a shared helper

A single `parseSkills` method validates and converts `[]skillsEntry` to `[]SkillsSource`. It is called for both shared and per-step skills, with a `context` string parameter for log messages (e.g., `"tools.skills"`, `"analysis.tools.skills"`).

Entries with empty `image` or empty `paths` are skipped with a warning. Valid entries are applied even if some are invalid.

Paths are not validated beyond non-emptiness — the Proposal CRD has CEL validation rules that enforce path format.

## Risks / Trade-offs

- **BREAKING: flat `skills` key is removed** → Operators using the previous `skills` key must migrate to `tools.skills`. Since this feature was just added and not yet released, the migration impact is minimal.
- **Increased config complexity** → The hierarchical structure adds depth but mirrors the CRD exactly, so operators familiar with Proposals will recognize the shape.
- **Per-step tools replace shared tools entirely** → This matches the CRD semantics. An operator who sets `analysis.tools.skills` must include all skills needed for analysis, not just the additions. This is documented in the ConfigMap manifest comments.
