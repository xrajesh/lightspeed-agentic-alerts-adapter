## Context

The adapter polls AlertManager for firing alerts and creates Proposal CRs. It deduplicates using AlertManager's fingerprint — a hash that includes labels like `pod` that change across restarts. When a pod is OOM-killed and rescheduled, the alert re-fires with a different fingerprint (new pod name), causing duplicate Proposals for the same underlying problem.

The current dedup flow: `reconcile` → `hasActiveProposal(fingerprint)` → `inCooldown(fingerprint)`. The fingerprint is also stored as the `agentic.openshift.io/alert-fingerprint` Proposal label for matching.

## Goals / Non-Goals

**Goals:**
- Prevent duplicate Proposals when the same alert re-fires with a different fingerprint due to ignored label changes (e.g., pod restarts).
- Make the set of ignored labels configurable so operators can extend it for clusters with custom alert labels.
- Maintain the existing stateless dedup architecture — no internal state beyond what's in the Kubernetes API.

**Non-Goals:**
- Changing the Proposal CRD or operator behavior.
- Changing the Proposal naming scheme (it still uses AlertManager's fingerprint for uniqueness).
- Implementing escalating cooldowns or rate limiting.
- Changing how AlertManager fingerprints work.

## Decisions

### Decision 1: Blocklist approach for ignored labels

**Choice**: Strip a configurable set of ignored labels from the alert's label set before hashing, rather than using an allowlist of stable labels.

**Alternatives considered**:
- *Allowlist*: Only include known-stable labels (`alertname`, `namespace`, `container`, etc.). Rejected because its failure mode — merging distinct problems into one scope — is worse than the blocklist's failure mode of creating an extra Proposal.
- *Match on `alertname + namespace` only*: Too coarse. Two different containers crash-looping in the same namespace are distinct problems.

**Rationale**: Unknown labels are included by default (safe default). The blocklist is small and stable. Failure mode is creating a duplicate Proposal, which is the same as current behavior.

### Decision 2: Replace fingerprint label value, don't add a new label

**Choice**: Reuse the existing `agentic.openshift.io/alert-fingerprint` label but change its value from AlertManager's fingerprint to the stable scope hash.

**Alternatives considered**:
- *Add a new `alert-scope` label alongside `alert-fingerprint`*: Adds complexity with no benefit — nobody queries by AlertManager's fingerprint.

**Rationale**: The label's purpose is dedup matching. Changing the value computation is simpler than adding a new label and migrating dedup logic to use it. AlertManager's fingerprint is still embedded in the Proposal name for uniqueness and traceability.

### Decision 3: FNV-64a hash truncated to 8 hex characters

**Choice**: `FNV-64a(sorted key=value pairs)[:8]` — 32 bits of a 64-bit FNV-64a hash, consistent with Prometheus/Alertmanager's hash family.

**Rationale**: FNV-64a is the same hash family used by Prometheus and AlertManager for label fingerprinting. It is fast, non-cryptographic, and available in the Go standard library (`hash/fnv`). Truncating to 8 hex characters matches the existing `FingerprintLen = 8` convention. Collision probability is negligible for the number of distinct alert scopes in a single cluster.

### Decision 4: Default ignored labels

**Choice**: Default ignored labels list is `pod`, `instance`, `endpoint`, `uid`.

**Rationale**:
- `pod` — changes on reschedule (ReplicaSet creates new pod name)
- `instance` — host:port, changes on restart
- `endpoint` — similar to instance
- `uid` — Kubernetes object UID, changes on recreation

These are the labels most commonly present on OpenShift alerts that change across restarts of the same workload.

### Decision 5: Configuration via ConfigMap with structured sections

**Choice**: Add a `deduplication.ignoredLabels` field to the existing `alerts-adapter-config` ConfigMap YAML, and restructure the existing `allowedReceivers` field under a new `filtering` section.

The new config YAML structure:

```yaml
pollInterval: 30s
initialDelay: 5m
cooldownWindow: 1h
filtering:
  allowedReceivers:
    - Critical
deduplication:
  ignoredLabels:
    - pod
    - instance
    - endpoint
    - uid
```

**Rationale**: Grouping related configuration under `filtering` and `deduplication` sections makes the config self-documenting and avoids a flat namespace as more options are added. The default applies when the field is absent, consistent with how `pollInterval`, `initialDelay`, etc. work.

When `deduplication.ignoredLabels` is explicitly set in the ConfigMap, it fully replaces the default list (not merged). This keeps behavior predictable — operators see exactly which labels are ignored.

## Risks / Trade-offs

**[Risk] Rollout transition**: Proposals created before this change have the old AlertManager fingerprint value. New alerts will compute a different value for the same label, so they won't match old Proposals during the transition.
→ **Mitigation**: The 409 AlreadyExists guard on Proposal creation (via deterministic naming with AlertManager's fingerprint) still catches exact duplicates. The transition window is at most one cooldown period (1h default). This is acceptable.

**[Risk] Unknown ignored label**: If an alert has an ignored label not in the list, the scope hash will change when that label changes, creating a duplicate Proposal.
→ **Mitigation**: This is the same failure mode as today's behavior. The configurable blocklist lets operators fix it without code changes. The default list covers the most common cases.

**[Risk] Overly broad blocklist**: If an operator adds a label to the ignored list that actually distinguishes different problems, the adapter will suppress legitimate Proposals.
→ **Mitigation**: Document the default list and the semantics clearly. The default is conservative (only 4 labels).
