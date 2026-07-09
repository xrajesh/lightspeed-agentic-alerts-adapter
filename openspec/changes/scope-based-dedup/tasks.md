## 1. Configuration

- [ ] 1.1 Add `VolatileLabels` field to `Config` struct and `volatileLabels` to `configFile` struct in `internal/config/config.go`
- [ ] 1.2 Add default volatile labels constant (`pod`, `instance`, `endpoint`, `uid`) and apply it when field is absent in `LoadFromFile`
- [ ] 1.3 Add tests for `volatileLabels` parsing: absent (defaults), explicit list (replaces defaults), empty list (no stripping)

## 2. Scope hashing

- [ ] 2.1 Add `StableFingerprint(labels map[string]string, volatileLabels []string) string` function in `internal/proposal/build.go` that strips volatile labels, sorts remaining key=value pairs, and returns `SHA256[:8]`
- [ ] 2.2 Add tests: no volatile labels present, volatile labels stripped, two alerts differing only in volatile labels produce same hash, differing in non-volatile labels produce different hash, empty volatile list includes all labels

## 3. Proposal building

- [ ] 3.1 Update `Build` to accept volatile labels and call `StableFingerprint` for the `alert-fingerprint` label value instead of using AlertManager's fingerprint directly
- [ ] 3.2 Update `buildLabels` signature to accept the stable fingerprint
- [ ] 3.3 Update existing `Build` tests to pass volatile labels and verify the `alert-fingerprint` label uses the stable fingerprint

## 4. Adapter integration

- [ ] 4.1 Thread `VolatileLabels` from config through `Adapter` to the `proposal.Build` call
- [ ] 4.2 Update adapter tests to pass volatile labels config

## 5. Cleanup

- [ ] 5.1 Remove `fingerprintPrefix` function from `internal/adapter/adapter.go` if no longer used (dedup now matches on the stable fingerprint written by `Build`)
- [ ] 5.2 Verify `FingerprintLen` constant is still used for Proposal naming; if not, remove it
