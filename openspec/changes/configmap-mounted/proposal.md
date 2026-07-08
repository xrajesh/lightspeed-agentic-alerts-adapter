## Why

The adapter currently reads the `alerts-adapter-config` ConfigMap via the Kubernetes API on every poll cycle (every 30s by default). This creates unnecessary API server load and couples the adapter to ConfigMap RBAC (`get` verb) at runtime. Since the operator already deploys the adapter, it can mount the ConfigMap as a volume and watch for changes — restarting the pod when the config changes. This simplifies the adapter to a read-once-at-startup model, reduces API calls, and removes the need for dynamic config-reload logic inside the adapter.

## What Changes

- The adapter reads configuration from a local file (mounted ConfigMap volume) at startup instead of calling the Kubernetes API each cycle.
- The `ConfigSource` interface and `ConfigMapSource` implementation are replaced with a file-based config loader.
- The poll loop no longer reloads config on each cycle — it uses the config loaded at startup for its entire lifetime.
- The `configEqual` function and dynamic ticker-reset logic in the `Run` loop are removed.
- The deployment manifest gains a ConfigMap volume mount.
- The ConfigMap RBAC (`Role` + `RoleBinding` for `get` on configmaps) is removed — the adapter no longer needs Kubernetes API access to ConfigMaps.
- **Note:** The operator-side watch + restart logic is out of scope for this repo — it is the operator's responsibility.

## Capabilities

### New Capabilities
- `file-config`: Loading configuration from a file path at startup, replacing the Kubernetes API-based ConfigMap reader.

### Modified Capabilities
- `configmap-config`: Requirements change from "load each reconcile cycle" to "load once at startup from a mounted file". Fallback-to-defaults behavior is preserved but triggered only at startup.
- `poll-loop`: Requirements change to remove per-cycle config reload and dynamic poll-interval adjustment.

## Impact

- **`internal/config/`** — `ConfigMapSource` and its Kubernetes client dependency are removed. A new file-based loader replaces it. The `Config`, `ToolsConfig`, and YAML parsing types are retained.
- **`internal/adapter/`** — `ConfigSource` interface simplified (no `context.Context` needed). `Run` loop simplified: no per-cycle `Load`, no `configEqual`, no ticker reset. `reconcile` receives config as a parameter instead of loading it.
- **`cmd/alerts-adapter/main.go`** — No longer needs `corev1` scheme or `controller-runtime` client for config. Reads a config file path from an environment variable or flag.
- **`manifests/deployment.yaml`** — Adds ConfigMap volume and volumeMount.
- **`manifests/rbac.yaml`** — Removes the `Role`/`RoleBinding` for ConfigMap access.
- **Dependencies** — `k8s.io/api/core/v1` and `sigs.k8s.io/controller-runtime/pkg/client` may become unused if AgenticRuns also stop needing them (AgenticRuns still need the client, so `controller-runtime` stays but `corev1` registration in `main.go` can be removed).
