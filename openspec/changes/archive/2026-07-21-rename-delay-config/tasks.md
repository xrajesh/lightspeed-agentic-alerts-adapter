## 1. Config package

- [x] 1.1 Rename constants: `DefaultInitialDelay` to `DefaultPreRunDelay` (value 0), `DefaultCooldownWindow` to `DefaultPostRunDelay` (value 1h)
- [x] 1.2 Rename `Config` struct fields: `InitialDelay` to `PreRunDelay`, `CooldownWindow` to `PostRunDelay`
- [x] 1.3 Rename `configFile` YAML tags: `initialDelay` to `preRunDelay`, `cooldownWindow` to `postRunDelay`
- [x] 1.4 Update `LoadFromFile` validation: use `isSet` to distinguish "not set" (keep default) from "explicit 0" (override to 0); clamp negative values to 0 silently; return an error on unparseable duration syntax
- [x] 1.5 Update config tests: rename references, update expected defaults to 0, test zero/negative clamping without error log, test that invalid syntax returns an error

## 2. Adapter package

- [x] 2.1 Rename field references: `cfg.InitialDelay` to `cfg.PreRunDelay`, `cfg.CooldownWindow` to `cfg.PostRunDelay`
- [x] 2.2 Rename helper functions: `skipInitialDelay` to `tooEarly`, `inCooldown` to `tooRecent`
- [x] 2.3 Add short-circuit: skip `tooEarly` check when `PreRunDelay` is 0, skip `tooRecent` check when `PostRunDelay` is 0
- [x] 2.4 Update log messages to use new field names
- [x] 2.5 Update adapter tests: rename references, update `defaultTestConfig()` defaults to 0, add test cases for 0-value short-circuits

## 3. Docs and manifests

- [x] 3.1 Update `manifests/configmap.yaml`: rename keys, comment them out, show `preRunDelay: 5m` and `postRunDelay: 1h` as examples
- [x] 3.2 Update `README.md`: rename config table entries, update defaults, update example ConfigMap
- [x] 3.3 Update `ARCHITECTURE.md`: rename references in flow pseudocode, requirements tables, and config reference
- [x] 3.4 Update `AGENTS.md`: rename references in the architecture summary
