## Context

The adapter uses three hard-coded constants (`pollInterval`, `initialDelay`, `cooldownWindow`) in `internal/adapter/adapter.go`. These are set once in `New()` and never change. The adapter already has a controller-runtime `client.Client` in the proposal package, connected via in-cluster config. The adapter runs as a single replica in a known namespace.

## Goals / Non-Goals

**Goals:**
- Allow operators to tune `pollInterval`, `initialDelay`, and `cooldownWindow` at runtime via a ConfigMap.
- Configuration changes take effect on the next poll cycle — no Pod restart required.
- The adapter starts and runs normally when the ConfigMap does not exist, using the current defaults.
- Invalid or unparseable values fall back to defaults with a warning log.

**Non-Goals:**
- Informer or watch-based notification — polling the ConfigMap each cycle is sufficient.
- Configuring AlertManager URL, token path, or CA path via ConfigMap (these are set at startup).
- ConfigMap creation or management by the adapter itself.
- Validation webhooks or CRD-based configuration.

## Decisions

### 1. Poll the ConfigMap each reconcile cycle, not via informer

The adapter already polls AlertManager and the Kubernetes API every 30 seconds. Adding one `client.Get()` call for a ConfigMap adds negligible overhead. An informer would require a typed clientset, factory lifecycle, and cache sync — machinery disproportionate to watching a single object. The poll approach keeps the adapter stateless and simple.

**Alternative considered:** SharedInformer with event handler. Rejected because it adds concurrency (separate goroutine, thread-safe config swap), factory lifecycle management, and error surface — all for a single ConfigMap that the adapter only needs to read once per cycle.

### 2. New `internal/config` package with a `ConfigSource` interface

Define a `ConfigSource` interface in the `adapter` package (consumer-defines-interface convention) and implement it in `internal/config`. The interface returns a `Config` struct with the three duration fields. The adapter calls it at the start of each `reconcile()`.

**Interface:**
```go
// ConfigSource provides runtime configuration for each poll cycle.
type ConfigSource interface {
    Load(ctx context.Context) Config
}
```

`Load` never returns an error — it falls back to defaults on any failure. NotFound is logged at Info level (the ConfigMap is optional); parse and validation errors are logged at Warning level. This keeps the reconcile loop clean: no error branch for config loading.

### 3. ConfigMap name and namespace

- **Name:** `alerts-adapter-config` (hard-coded constant).
- **Namespace:** The adapter's own namespace, read from the downward API via the `POD_NAMESPACE` environment variable, falling back to `openshift-lightspeed`.

### 4. ConfigMap structure — YAML in a single data key

The ConfigMap uses a single `config.yaml` key containing a YAML document:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alerts-adapter-config
data:
  config.yaml: |
    pollInterval: 45s
    initialDelay: 10m
    cooldownWindow: 30m
```

The YAML is unmarshalled into a struct with `yaml.Unmarshal`. Duration fields use Go duration string format (`30s`, `5m`, `1h`). A custom `Duration` type wraps `time.Duration` with YAML unmarshalling support.

| Field | Type | Default | Example |
|-------|------|---------|---------|
| `pollInterval` | duration string | `30s` | `45s`, `1m` |
| `initialDelay` | duration string | `5m` | `10m`, `2m30s` |
| `cooldownWindow` | duration string | `1h` | `30m`, `2h` |

All fields are optional. Missing fields use the default value. Invalid duration strings fall back to the default for that field with a warning log.

### 5. Reuse the existing controller-runtime client

The proposal package already creates a controller-runtime `client.Client` with in-cluster config. Rather than creating a second client or a separate clientset, pass the same `rest.Config` to the config reader (or let `main.go` create the client once and share it). The config reader only needs `client.Get()` for a single ConfigMap — the controller-runtime client supports core types out of the box.

### 6. Poll interval changes take effect on the next tick

When `pollInterval` changes, the adapter resets the ticker. The new interval applies starting from the reset — there is no attempt to adjust the current in-flight tick.

## Risks / Trade-offs

- **One extra API call per cycle** → Negligible for a single-replica adapter polling one ConfigMap every 30s. No mitigation needed.
- **Config changes are not instant** → Changes take up to one poll interval to apply. Acceptable given the adapter's polling nature.
- **Missing ConfigMap on first cycle** → `client.Get` returns NotFound, adapter uses defaults and logs at Info level. No retry needed — next cycle re-reads.
- **Partial ConfigMap** → Only the fields present in the YAML are applied; missing fields use defaults. An operator can set just `cooldownWindow` without specifying the others.
