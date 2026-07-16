## Why

The adapter deduplicates alerts using the AlertManager fingerprint, which is a hash that includes labels like `pod` that change across restarts. When a pod is OOM-killed and a new pod is scheduled, the alert re-fires with a different fingerprint (different pod name), bypassing both the active-run check and the cooldown window. This causes a flood of duplicate AgenticRuns for the same underlying problem. The adapter needs a stable identity for "same problem" that survives pod restarts.

## What Changes

- Replace the `alert-fingerprint` AgenticRun label value: instead of AlertManager's fingerprint, compute `FNV-64a(sorted(labels - ignoredLabels))[:8]` from the alert's label set.
- Default ignored labels: `pod`, `instance`, `endpoint`, `uid`. Configurable via `deduplication.ignoredLabels` in the adapter ConfigMap.
- AgenticRun naming continues to use a `startsAt` hash for uniqueness — each AgenticRun remains a distinct remediation attempt.
- Dedup logic (`hasActiveRun`, `inCooldown`) is structurally unchanged — it still matches on `alert-fingerprint`, but the value is now stable across pod restarts.

## Capabilities

### New Capabilities
- `scope-hashing`: Compute a stable fingerprint from alert labels minus configurable ignored labels.

### Modified Capabilities
- `agenticrun-building`: Compute and write the stable fingerprint as the `alert-fingerprint` label value.
- `configmap-config`: Restructure config YAML with `filtering` and `deduplication` sections; support `deduplication.ignoredLabels` field with a default list.

## Impact

- **Code**: `internal/agenticrun` (fingerprint computation), `internal/config` (new field), `internal/adapter` (pass ignored labels config through).
- **Existing AgenticRuns**: AgenticRuns created before this change will not match the new scope hash. Breaking change — no migration provided.
- **APIs**: No CRD changes. The label key is unchanged; only the value computation changes.
