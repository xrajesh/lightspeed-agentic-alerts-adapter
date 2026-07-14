## 1. Configuration

- [ ] 1.1 Restructure `configFile` struct: add `Filtering` and `Deduplication` wrapper structs in `internal/config/config.go`; nest `AllowedReceivers` under `Filtering` and add `IgnoredLabels` under `Deduplication`
- [ ] 1.2 Add `IgnoredLabels` field to `Config` struct in `internal/config/config.go`
- [ ] 1.3 Add default ignored labels constant (`pod`, `instance`, `endpoint`, `uid`) and apply it when field is absent in `LoadFromFile`
- [ ] 1.4 Support backward compatibility for top-level `allowedReceivers` in config YAML
- [ ] 1.5 Add tests for `deduplication.ignoredLabels` parsing: absent (defaults), explicit list (replaces defaults), empty list (no stripping)
- [ ] 1.6 Add tests for `filtering.allowedReceivers` parsing and backward compatibility with top-level `allowedReceivers`

## 2. Scope hashing

- [ ] 2.1 Add `StableFingerprint(labels map[string]string, ignoredLabels []string) string` function in `internal/proposal/build.go` that strips ignored labels, sorts remaining key=value pairs, and returns `FNV-64a[:8]`
- [ ] 2.2 Add tests: no ignored labels present, ignored labels stripped, two alerts differing only in ignored labels produce same hash, differing in non-ignored labels produce different hash, empty ignored list includes all labels

## 3. Proposal building

- [ ] 3.1 Update `Build` to accept ignored labels and call `StableFingerprint` for the `alert-fingerprint` label value instead of using AlertManager's fingerprint directly
- [ ] 3.2 Update `buildLabels` signature to accept the stable fingerprint
- [ ] 3.3 Update existing `Build` tests to pass ignored labels and verify the `alert-fingerprint` label uses the stable fingerprint

## 4. Adapter integration

- [ ] 4.1 Thread `IgnoredLabels` from config through `Adapter` to the `proposal.Build` call
- [ ] 4.2 Update adapter tests to pass ignored labels config

## 5. Cleanup

- [ ] 5.1 Remove `fingerprintPrefix` function from `internal/adapter/adapter.go` if no longer used (dedup now matches on the stable fingerprint written by `Build`)
- [ ] 5.2 Verify `FingerprintLen` constant is still used for Proposal naming; if not, remove it
