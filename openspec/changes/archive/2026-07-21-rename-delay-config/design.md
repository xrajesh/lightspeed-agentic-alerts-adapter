## Context

The adapter uses two duration-based config fields to gate AgenticRun creation: `initialDelay` (wait before acting on a new alert) and `cooldownWindow` (wait before re-acting after a terminal AgenticRun). Both names are vague. This change renames them to `preRunDelay` and `postRunDelay`, defaults `preRunDelay` to 0s (act immediately on new alerts) and `postRunDelay` to 1h (avoid repeated analysis), and allows explicit `0s` to disable either delay.

## Goals / Non-Goals

**Goals:**
- Rename `initialDelay` to `preRunDelay` and `cooldownWindow` to `postRunDelay` across Go code, config YAML, docs, and manifests
- Default `preRunDelay` to 0s and `postRunDelay` to 1h when not set in the ConfigMap
- Allow explicit `0s` to disable either delay (distinguish "not set" from "set to 0")
- Clamp negative values to 0s silently (no error log)

**Non-Goals:**
- Backward compatibility with old key names (clean break)
- Changing the semantics of what these delays do (only names and defaults change)
- Changing pollInterval defaults or behavior

## Decisions

**Rename Go identifiers to match new YAML keys**

Rename `Config.InitialDelay` to `Config.PreRunDelay`, `DefaultInitialDelay` to `DefaultPreRunDelay`, `skipInitialDelay()` to `tooEarly()`, and similarly for the cooldown side: `Config.CooldownWindow` to `Config.PostRunDelay`, `DefaultCooldownWindow` to `DefaultPostRunDelay`. The helper `inCooldown()` becomes `tooRecent()`. These names describe what the check answers rather than the mechanism.

Alternative considered: keeping the old Go identifiers and only changing the YAML keys. Rejected because split naming creates confusion between code and config.

**Default `preRunDelay` to 0s, `postRunDelay` to 1h**

`DefaultPreRunDelay = 0` and `DefaultPostRunDelay = 1 * time.Hour`. The pre-run delay defaults to 0 so the adapter acts immediately on new alerts. The post-run delay defaults to 1h to avoid repeated analysis of an alert that has already been investigated. Both can be explicitly set to `0s` in the ConfigMap to disable the delay.

**Distinguish "not set" from "explicit 0" using `Duration.isSet`**

The `Duration` type's `isSet` flag tracks whether a key was present in YAML. When `isSet` is true and the value is zero or negative, the config is set to 0 (overriding the default). When `isSet` is false, the default is kept. Negative values are clamped to 0 silently.

**Fail hard on invalid syntax**

If a duration string cannot be parsed (e.g., `preRunDelay: "abc"`), `LoadFromFile` returns an error. The reconcile loop propagates this error, logging it and skipping the cycle. This is a change from the current behavior, which silently falls back to defaults. Rationale: a typo in config should be visible immediately, not masked by a silent fallback that changes operational behavior.

**No backward compatibility shim**

Old ConfigMaps using `initialDelay` or `cooldownWindow` will be silently ignored (treated as unrecognized keys by strict YAML parsing). This is acceptable because the adapter is pre-GA and the ConfigMap is deployed alongside it.

## Risks / Trade-offs

**Existing deployments break silently** - A deployment with `initialDelay: 10m` or `cooldownWindow: 2h` in its ConfigMap will lose those settings after upgrade. `preRunDelay` will default to 0s (act immediately) and `postRunDelay` will default to 1h. Mitigation: release notes must call this out.
