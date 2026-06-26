## REMOVED Requirements

### Requirement: Load configuration from a ConfigMap each reconcile cycle
**Reason**: Replaced by file-based config loading at startup. The ConfigMap is now mounted as a volume by the operator, and the adapter reads the resulting file once at startup instead of querying the Kubernetes API each cycle.
**Migration**: Configuration is still provided via the same `alerts-adapter-config` ConfigMap, but it is now consumed as a mounted file rather than via API calls. The operator watches the ConfigMap and restarts the adapter on changes.

### Requirement: Operate normally when ConfigMap does not exist
**Reason**: Replaced by equivalent behavior in the `file-config` capability. The adapter falls back to defaults when the mounted config file does not exist.
**Migration**: No action needed — the fallback-to-defaults behavior is preserved in the file-based loader.

### Requirement: Use well-defined default values
**Reason**: Moved to the `file-config` capability with identical default values.
**Migration**: No action needed — same defaults apply.

### Requirement: Resolve the adapter namespace from the environment
**Reason**: The adapter no longer reads ConfigMaps via the Kubernetes API, so the namespace is irrelevant. The config file path is hardcoded.
**Migration**: The `POD_NAMESPACE` environment variable is no longer used for config loading. It may still be used for other purposes (e.g., Proposal creation namespace).
