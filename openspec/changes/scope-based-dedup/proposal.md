## Why

The adapter deduplicates alerts using the AlertManager fingerprint, which is a hash that includes volatile labels like `pod`. When a pod is OOM-killed and a new pod is scheduled, the alert re-fires with a different fingerprint (different pod name), bypassing both the active-proposal check and the cooldown window. This causes a flood of duplicate Proposals for the same underlying problem. The adapter needs a stable identity for "same problem" that survives pod restarts.

## What Changes

- Replace the `alert-fingerprint` Proposal label value: instead of AlertManager's fingerprint, compute `SHA256(sorted(labels - volatileLabels))[:8]` from the alert's label set.
- Default volatile labels blocklist: `pod`, `instance`, `endpoint`, `uid`. Configurable via the adapter ConfigMap.
- Proposal naming continues to use AlertManager's fingerprint for uniqueness — each Proposal remains a distinct remediation attempt.
- Dedup logic (`hasActiveProposal`, `inCooldown`) is structurally unchanged — it still matches on `alert-fingerprint`, but the value is now stable across pod restarts.

## Capabilities

### New Capabilities
- `scope-hashing`: Compute a stable fingerprint from alert labels minus configurable volatile labels.

### Modified Capabilities
- `proposal-building`: Compute and write the stable fingerprint as the `alert-fingerprint` label value.
- `configmap-config`: Support a `volatileLabels` configuration field with a default blocklist.

## Impact

- **Code**: `internal/proposal` (fingerprint computation), `internal/config` (new field), `internal/adapter` (pass volatile labels config through).
- **Existing Proposals**: Proposals created before this change will have the old AlertManager fingerprint value. New alerts will compute a different fingerprint and won't match old Proposals — effectively the same as today's behavior. No migration needed.
- **APIs**: No CRD changes. The label key is unchanged; only the value computation changes.
