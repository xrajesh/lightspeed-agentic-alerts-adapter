## Context

The adapter currently uses a `ConfigMapSource` that calls the Kubernetes API (`client.Get`) to read the `alerts-adapter-config` ConfigMap on every reconcile cycle. This works but generates one API call per poll interval (every 30s by default). The `Run` loop includes logic to detect config changes between cycles, reset the ticker if `pollInterval` changed, and log reloads.

The operator that deploys the adapter already manages the Deployment resource. Kubernetes natively supports mounting ConfigMaps as volumes and the operator can watch the ConfigMap to trigger a rolling restart when it changes. This moves config-change detection from the adapter to the platform.

## Goals / Non-Goals

**Goals:**
- Replace the Kubernetes API-based config reader with a file-based reader that loads config once at startup.
- Simplify the poll loop by removing per-cycle config reload, `configEqual`, and dynamic ticker reset.
- Mount the ConfigMap as a volume in the Deployment manifest.
- Remove ConfigMap RBAC (the adapter no longer queries the API for ConfigMaps).
- Preserve all existing fallback-to-defaults behavior (missing file, invalid YAML, non-positive durations).

**Non-Goals:**
- Implementing the operator-side ConfigMap watch and pod restart — that belongs in the `lightspeed-agentic-operator` repo.
- Changing the ConfigMap YAML schema or config field semantics.
- Adding a file-watcher or hot-reload capability to the adapter.

## Decisions

### 1. Load-once file-based config replaces `ConfigMapSource` and `ConfigSource` interface

Replace `ConfigMapSource` with a `LoadFromFile(path string, logger *slog.Logger) Config` function. It reads the file at the given path, parses YAML, applies defaults — same logic as today minus the Kubernetes API call. If reading or parsing the file fails for any reason (missing file, invalid YAML, non-positive durations), the function falls back to default values and logs an error. The adapter always starts. `adapter.New` accepts `config.Config` directly instead of a `ConfigSource` interface. The `Run` loop and `reconcile` use this fixed config.

**Rationale:** The `ConfigSource` interface and per-cycle `Load` call only existed to support dynamic config reload. With load-once semantics, a plain function called in `main.go` is simpler — it eliminates the interface, `configEqual`, the ticker-reset branch, and the per-cycle `Load` call in `reconcile`.

**Alternative considered:** Keep the `ConfigSource` interface with a `FileSource` struct that caches the result. Rejected — unnecessary abstraction when the adapter is restarted on config change.

### 2. Hardcoded config file path with volume mount

The adapter reads config from `/etc/alerts-adapter/config.yaml`. The ConfigMap is mounted at `/etc/alerts-adapter/` in the Deployment manifest. No environment variable is needed — the operator programmatically creates the Deployment and controls the mount path.

**Rationale:** The path is a contract between the operator and the adapter, both maintained by the same team. An env var would add indirection with no benefit since the operator always sets both the mount and the path.

### 3. Remove ConfigMap RBAC

The `Role` and `RoleBinding` for ConfigMap `get` access are deleted from `manifests/rbac.yaml`.

**Rationale:** The adapter no longer makes API calls to read ConfigMaps. Least-privilege principle — don't request permissions you don't use.

### 4. Remove `corev1` scheme registration from `main.go`

With no ConfigMap API calls, `corev1.AddToScheme` is no longer needed in `main.go`.

**Rationale:** The `corev1` scheme was only registered to support the `client.Get` call for the ConfigMap. AgenticRun CRs use only the `agenticv1alpha1` scheme.

## Risks / Trade-offs

- **Config changes require pod restart** → Mitigated by the operator watching the ConfigMap and triggering a rolling restart. Until the operator implements this, config changes require a manual pod restart.
- **Startup failure on missing config file** → Mitigated by treating a missing file as "use defaults" and logging an error, same as today's missing-ConfigMap behavior. The adapter always starts.
- **Loss of dynamic poll-interval tuning** → Accepted trade-off. The previous per-cycle reload allowed changing `pollInterval` without a restart. With the new model, changes take effect after the operator-triggered restart (seconds, not cycles). This is acceptable for a 30s default interval.
