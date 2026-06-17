## 1. Config Package — Skills Parsing

- [x] 1.1 Add `skillsEntry` struct with `yaml` tags and `Skills []agenticv1alpha1.SkillsSource` field to the `Config` struct in `internal/config/config.go`
- [x] 1.2 Add `skills` field to the `configFile` struct using the local `skillsEntry` type for YAML unmarshalling
- [x] 1.3 Implement skills validation in `Load()`: skip entries with empty `image` or empty `paths`, log warnings for skipped entries, convert valid entries to `SkillsSource`
- [x] 1.4 Add unit tests for skills config parsing: valid entries, missing skills key, empty image, empty paths, mix of valid and invalid entries

## 2. Proposal Builder — Accept Tools Config

- [x] 2.1 Change `Build()` signature to accept `skills []agenticv1alpha1.SkillsSource` parameter
- [x] 2.2 Set `spec.tools.skills` on the Proposal when the skills slice is non-empty; leave `spec.tools` at zero value when nil or empty
- [x] 2.3 Update existing `Build()` tests to pass nil skills (preserving current behavior)
- [x] 2.4 Add tests for `Build()` with skills: single skill, multiple skills, empty skills slice

## 3. Adapter Integration

- [x] 3.1 Update the `reconcile()` method in `internal/adapter/adapter.go` to pass `cfg.Skills` from the loaded config to `proposal.Build()`
- [x] 3.2 Update adapter tests to account for the new `Build()` parameter

## 4. Manifest Update

- [x] 4.1 Update `manifests/configmap.yaml` with commented example of skills configuration
