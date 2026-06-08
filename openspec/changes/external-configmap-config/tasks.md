## 1. Config package

- [x] 1.1 Create `internal/config` package with `Config` struct (pollInterval, initialDelay, cooldownWindow as `time.Duration`) and default constants
- [x] 1.2 Implement YAML parsing with a custom `Duration` type that wraps `time.Duration` and supports `yaml.Unmarshal`
- [x] 1.3 Implement `ConfigMapConfigSource` that reads the `alerts-adapter-config` ConfigMap via controller-runtime `client.Get`, parses the `config.yaml` key, and falls back to defaults on NotFound, missing key, invalid YAML, or invalid duration values
- [x] 1.4 Add tests for YAML parsing: valid full config, partial config, invalid durations, invalid YAML, empty data, missing key
- [x] 1.5 Add tests for `ConfigMapConfigSource.Load`: ConfigMap exists, ConfigMap not found, ConfigMap without config.yaml key

## 2. Adapter integration

- [x] 2.1 Define `ConfigSource` interface in `internal/adapter` with `Load(ctx context.Context) config.Config`
- [x] 2.2 Add `ConfigSource` field to `Adapter` struct and accept it in `New()`
- [x] 2.3 Update `reconcile()` to call `ConfigSource.Load()` at the start and use the returned values for initialDelay and cooldownWindow
- [x] 2.4 Update `Run()` to detect pollInterval changes and reset the ticker
- [x] 2.5 Update existing adapter tests to supply a mock/stub `ConfigSource`

## 3. Wiring and manifests

- [x] 3.1 Update `cmd/alerts-adapter/main.go` to create a shared controller-runtime client, pass it to both `ConfigMapConfigSource` and proposal client, and wire `ConfigSource` into the adapter
- [x] 3.2 Add `POD_NAMESPACE` environment variable to the deployment manifest via the downward API
- [x] 3.3 Add a sample ConfigMap manifest (`manifests/configmap.yaml`) with documented fields and defaults
