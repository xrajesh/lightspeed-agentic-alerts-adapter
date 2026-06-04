## 1. Core Implementation

- [x] 1.1 Add `skipSeverity` function to `internal/adapter/adapter.go` that returns true for alerts with severity `none` or `info` (case-insensitive), false otherwise (including missing label)
- [x] 1.2 Integrate severity check as the first skip in the `reconcile` loop, before `skipInitialDelay`, with debug logging (alertname, fingerprint, severity)

## 2. Tests

- [x] 2.1 Add unit tests for `skipSeverity` covering: none, info, warning, critical, mixed-case, empty string, and missing severity label
- [x] 2.2 Add reconcile-level test verifying that alerts with severity `none`/`info` do not result in Proposal creation
