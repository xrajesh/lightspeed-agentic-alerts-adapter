## 1. Config loader

- [x] 1.1 Replace `ConfigMapSource` with `LoadFromFile(path string, logger *slog.Logger) Config` function in `internal/config/config.go` — reads file, parses YAML, falls back to defaults on any error and logs it
- [x] 1.2 Remove `ConfigMapSource`, `NewConfigMapSource`, and all Kubernetes client imports (`corev1`, `apierrors`, `types`, `client`) from `internal/config/config.go`
- [x] 1.3 Add the `DefaultConfigPath` constant (`/etc/alerts-adapter/config.yaml`) to `internal/config/config.go`
- [x] 1.4 Update `internal/config/config_test.go` — replace ConfigMap-based tests with file-based tests covering: valid file, partial YAML, invalid duration, non-positive duration, invalid YAML, missing file

## 2. Adapter simplification

- [x] 2.1 Remove `ConfigSource` interface from `internal/adapter/adapter.go`
- [x] 2.2 Change `Adapter` struct to hold `config.Config` value instead of `ConfigSource` — update `New` to accept `config.Config`
- [x] 2.3 Simplify `Run` loop — remove per-cycle `config.Load`, `configEqual`, and dynamic ticker reset
- [x] 2.4 Simplify `reconcile` — remove the `config.Load` call, use the stored config directly
- [x] 2.5 Remove the `configEqual` function
- [x] 2.6 Update `internal/adapter/adapter_test.go` — remove `ConfigSource` mock, pass `config.Config` directly to `New`

## 3. Main wiring

- [x] 3.1 Update `cmd/alerts-adapter/main.go` — call `config.LoadFromFile` at startup, pass the resulting `Config` to `adapter.New`
- [x] 3.2 Remove `corev1` scheme registration from `newClient` in `main.go`
- [x] 3.3 Remove the `corev1` import from `main.go`

## 4. Manifests

- [x] 4.1 Add ConfigMap volume and volumeMount to `manifests/deployment.yaml` — mount `alerts-adapter-config` at `/etc/alerts-adapter/`
- [x] 4.2 Remove the ConfigMap `Role` and `RoleBinding` from `manifests/rbac.yaml`

## 5. Verification

- [x] 5.1 Run `make test` — all tests pass
- [x] 5.2 Run `make lint` — no lint errors
- [x] 5.3 Run `make vet` — no vet errors
