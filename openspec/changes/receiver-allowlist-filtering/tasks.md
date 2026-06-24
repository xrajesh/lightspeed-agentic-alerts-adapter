## 1. Configuration

- [x] 1.1 Add `AllowedReceivers []string` field to `config.Config` and `DefaultAllowedReceivers` constant `["Critical"]`
- [x] 1.2 Add `allowedReceivers` YAML key to `configFile` struct and parse it in `ConfigMapSource.Load` — default to `["Critical"]` when absent, honour explicit empty list
- [x] 1.3 Normalize allowlist entries to lowercase at parse time
- [x] 1.4 Add tests for `allowedReceivers` parsing: present, absent, empty list, ConfigMap missing

## 2. Receiver Filtering

- [x] 2.1 Add `skipReceiver` function in `internal/adapter/adapter.go` that checks if any alert receiver matches the allowlist (case-insensitive)
- [x] 2.2 Insert `skipReceiver` call as the first filter in `reconcile`, before `skipSeverity`
- [x] 2.3 Log skipped alerts at Debug level with receiver names
- [x] 2.4 Add tests for `skipReceiver`: matching receiver, no match, empty receivers on alert, empty allowlist, case-insensitive match

## 3. Logging

- [x] 3.1 Log the effective `allowedReceivers` list at Info level at adapter startup and on config reload
