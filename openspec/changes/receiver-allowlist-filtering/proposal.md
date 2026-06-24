## Why

The adapter currently creates Proposals for all firing alerts that pass severity and timing filters. In production OpenShift clusters, AlertManager routes alerts to different receivers based on severity and team ownership. Operators need to control which alerts trigger automated remediation by selecting only alerts routed to specific receivers (e.g., only "Critical" alerts handled by the on-call team). Without this, the adapter creates Proposals for alerts that don't warrant automated intervention.

## What Changes

- Add an `allowedReceivers` allowlist field to the existing ConfigMap-based configuration (`alerts-adapter-config`). The field holds a list of receiver names; an alert is only processed if at least one of its AlertManager receivers matches an entry in the list.
- Add receiver-based filtering to the adapter's reconcile loop as a new filter step.
- Default the receiver allowlist to `["Critical"]` when the field is absent, ensuring the adapter works out of the box with OpenShift's default AlertManager routing. When explicitly set to an empty list (`[]`), the adapter processes no alerts.

## Capabilities

### New Capabilities
- `receiver-filtering`: Receiver-based allowlist filtering that skips alerts not routed to any allowed receiver

### Modified Capabilities
- `poll-loop`: Add receiver allowlist check as a new filtering step in the reconcile loop

## Impact

- **Code**: `internal/config/` gains an `AllowedReceivers []string` field parsed from the ConfigMap; `internal/adapter/` gains a new filtering function using the config.
- **Deployment**: No new resources — uses the existing `alerts-adapter-config` ConfigMap with an optional `allowedReceivers` key in `config.yaml`.
- **APIs**: No CRD changes. The GettableAlert's `Receivers` field (already available from AlertManager) is now consumed.
- **Dependencies**: None — all required types and libraries are already in use.
