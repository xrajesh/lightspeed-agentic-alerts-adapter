## 1. Config package

- [ ] 1.1 Create `internal/config` package with `Config` struct (pollInterval, initialDelay, cooldownWindow as `time.Duration`) and default constants
- [ ] 1.2 Implement YAML parsing with a custom `Duration` type that wraps `time.Duration` and supports `yaml.Unmarshal`
- [ ] 1.3 Implement `ConfigMapConfigSource` that reads the `alerts-adapter-config` ConfigMap via controller-runtime `client.Get`, parses the `config.yaml` key, and falls back to defaults on NotFound, missing key, invalid YAML, or invalid duration values
- [ ] 1.4 Add tests for YAML parsing: valid full config, partial config, invalid durations, invalid YAML, empty data, missing key
- [ ] 1.5 Add tests for `ConfigMapConfigSource.Load`: ConfigMap exists, ConfigMap not found, ConfigMap without config.yaml key

## 2. Adapter integration

- [ ] 2.1 Define `ConfigSource` interface in `internal/adapter` with `Load(ctx context.Context) config.Config`
- [ ] 2.2 Add `ConfigSource` field to `Adapter` struct and accept it in `New()`
- [ ] 2.3 Update `reconcile()` to call `ConfigSource.Load()` at the start and use the returned values for initialDelay and cooldownWindow
- [ ] 2.4 Update `Run()` to detect pollInterval changes, reset the ticker, and log the change with previous and new values
- [ ] 2.5 Update existing adapter tests to supply a mock/stub `ConfigSource`

## 3. Wiring and manifests

- [ ] 3.1 Update `cmd/alerts-adapter/main.go` to create a shared controller-runtime client, pass it to both `ConfigMapConfigSource` and proposal client, and wire `ConfigSource` into the adapter
- [ ] 3.2 Add `POD_NAMESPACE` environment variable to the deployment manifest via the downward API
- [ ] 3.3 Add a sample ConfigMap manifest (`manifests/configmap.yaml`) with documented fields and defaults
