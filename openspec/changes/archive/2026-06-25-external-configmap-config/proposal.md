## Why

The adapter's operational parameters (poll interval, initial delay, cooldown window) are currently hard-coded constants. Changing any value requires rebuilding the binary and restarting the Pod. Operators need to tune these parameters at runtime — for example, shortening the cooldown during an incident or adjusting the poll interval — without disrupting the adapter's operation.

## What Changes

- Add a new `internal/config` package that reads a ConfigMap (`alerts-adapter-config` in the adapter's namespace) and exposes current configuration values to the adapter.
- The ConfigMap is polled on each reconcile cycle using the existing controller-runtime client — no informer, no extra goroutine. This fits the adapter's existing poll-based, stateless design.
- The adapter must start and operate normally when the ConfigMap does not exist, falling back to the current default values.
- Configuration changes in the ConfigMap take effect on the next poll cycle without Pod restart.

## Capabilities

### New Capabilities
- `configmap-config`: Read a ConfigMap for configuration on each reconcile cycle using the existing controller-runtime client, expose current values thread-safely, and fall back to defaults when the ConfigMap is absent or contains invalid values.

### Modified Capabilities
- `poll-loop`: The poll loop reads operational parameters (poll interval, initial delay, cooldown window) from the config reader each cycle instead of using hard-coded constants. The poll interval change takes effect on the next tick.

## Impact

- **New package**: `internal/config` — ConfigMap reader, parsing, default values.
- **Modified**: `internal/adapter/adapter.go` — accepts a config source, reads parameters per cycle.
- **Modified**: `cmd/alerts-adapter/main.go` — wires up the config reader.
- **Dependencies**: Uses `k8s.io/client-go/kubernetes` (already an indirect dependency) or the existing controller-runtime client.
- **New manifest**: ConfigMap manifest with documented fields and defaults.
