## Context

The adapter polls AlertManager and creates Proposal CRs for firing alerts. It already has ConfigMap-based runtime configuration (`alerts-adapter-config` in `openshift-lightspeed`) that supports `pollInterval`, `initialDelay`, `cooldownWindow`, and `tools` (skills). The reconcile loop applies filters in order: severity → initial delay → active proposal → cooldown. AlertManager's `GettableAlert` type includes a `Receivers` field (`[]*models.Receiver`, each with a `Name *string`) that identifies which AlertManager receivers matched the alert via routing, but this field is not currently consumed.

## Goals / Non-Goals

**Goals:**
- Add an `allowedReceivers` allowlist to the existing ConfigMap config that controls which alerts get Proposals.
- Filter alerts in the reconcile loop based on receiver matching — an alert passes if any of its receivers appears in the allowlist.
- Default to `["Critical"]` when the config field is absent, so the adapter works out of the box with OpenShift's default AlertManager routing.
- Log the active receiver allowlist at Info level so operators can verify their configuration.

**Non-Goals:**
- Blocklist/denylist approach — only allowlist semantics.
- Regex or glob matching on receiver names — exact case-insensitive match only.
- Watching the ConfigMap for changes via an informer — the existing pattern re-reads the ConfigMap on every poll cycle, which is sufficient.
- Wildcard "all receivers" mode — operators who want all alerts can list all their receiver names.

## Decisions

### Extend the existing ConfigMap config with an `allowedReceivers` field

Add an `allowedReceivers` YAML key to `configFile` in `internal/config/config.go` that maps to `[]string`. The parsed result is stored as `AllowedReceivers []string` on `config.Config`. When the field is absent, default to `["Critical"]`. When explicitly set to an empty list (`[]`), honour it — the adapter processes no alerts, giving operators an explicit way to disable Proposal creation.

**Rationale**: Follows the established pattern — the ConfigMap is already read on every poll cycle, validated, and falls back to defaults. The name `allowedReceivers` makes the allowlist semantics self-evident in both Go and YAML. No new configuration mechanism needed.

### Case-insensitive receiver matching

Receiver names from the allowlist and from the alert's `Receivers` field are compared case-insensitively. The allowlist is normalized to lowercase at parse time and alert receiver names are lowered at comparison time.

**Rationale**: AlertManager receiver names are user-configured strings with no case guarantees. Case-insensitive matching avoids surprising mismatches (e.g., `Critical` vs `critical`).

### Filter placement: first in the reconcile loop, before severity

The receiver filter runs as the first check in the reconcile loop, before severity filtering. This is the cheapest filter (string set lookup) and will skip the most alerts in typical deployments, so placing it first avoids unnecessary work.

**Rationale**: In a typical OpenShift cluster, most alerts are routed to non-critical receivers. Filtering them out first avoids computing severity, timing, and dedup checks for alerts that will never produce Proposals.

### Default allowlist: `["Critical"]`

When no `allowedReceivers` key is present in the ConfigMap (or it's empty), the default is `["Critical"]`. This matches the OpenShift default AlertManager configuration where critical alerts are routed to a "Critical" receiver.

**Rationale**: Secure-by-default — without explicit configuration, only critical alerts trigger automated remediation. Operators can widen the scope by adding more receivers.

### Log the active receiver allowlist at Info level

The adapter logs the effective `allowedReceivers` list at Info level at startup and whenever the config is reloaded. This gives operators immediate visibility into which receivers are enabled.

**Rationale**: Receiver misconfiguration (e.g., typo in name, wrong default) would silently suppress all Proposals. Info-level logging makes this easy to diagnose without enabling Debug.

## Risks / Trade-offs

- **Receiver name mismatch**: If the OpenShift AlertManager routing is customized and doesn't use a "Critical" receiver, the default allowlist will match nothing and no Proposals will be created. → Mitigation: Log the full allowlist at Info level at startup and log skipped alerts with their receiver names at Debug level.
- **Empty receivers on alert**: Some alerts may have an empty `Receivers` field (unlikely from AlertManager but possible). → Mitigation: An alert with no receivers matches no allowlist entry and is skipped, which is the safe behavior.
