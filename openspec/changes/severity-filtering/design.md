## Context

The adapter's reconcile loop processes all firing alerts from AlertManager, applying three dedup filters (initial delay, active AgenticRun, cooldown). There is no filter based on alert severity. Prometheus alerts carry a `severity` label with common values: `critical`, `warning`, `info`, and `none`. Low-severity alerts (`info`, `none`) are informational and should not trigger automated remediation.

## Goals / Non-Goals

**Goals:**
- Skip alerts with severity `none` or `info` before any AgenticRun processing.
- Log skipped alerts at debug level for observability.
- Keep the filter consistent with the existing skip pattern in the reconcile loop.

**Non-Goals:**
- Making the set of skipped severities configurable (can be added later if needed).
- Filtering at the AlertManager API query level (the API doesn't support label-based filtering in query params).

## Decisions

### 1. Filter placement: first check in the reconcile loop

The severity filter will be the first skip check in the reconcile loop, before `skipInitialDelay`. Rationale: severity is a static property of the alert — checking it first avoids unnecessary work in downstream filters (initial delay, active AgenticRun lookup, cooldown). This is the cheapest check.

**Alternative considered:** Filtering in `alertmanager.GetAlerts()` after the API call. Rejected because filtering logic belongs in the adapter (where all other skip decisions live), not in the API client.

### 2. Hardcoded skip set, not configurable

The skipped severities (`none`, `info`) are hardcoded as constants. Rationale: these are well-established Prometheus conventions. A configurable list adds complexity without clear value today. If needed later, it can be extracted to a config parameter.

**Alternative considered:** Environment variable for skipped severities. Rejected as premature — the skip set is unlikely to change across deployments.

### 3. Case-insensitive comparison

Severity labels are compared case-insensitively (lowercase before matching). Rationale: while Prometheus convention is lowercase, AlertManager rules can set arbitrary label values. Defensive normalization prevents silent misconfiguration.

## Risks / Trade-offs

- **[Risk] Alert with missing severity label** → Treated as non-skipped (processed normally). Missing severity is unusual but shouldn't block remediation.
- **[Trade-off] Hardcoded vs configurable** → Simpler code now, but requires a code change to adjust the skip set. Acceptable given stable Prometheus conventions.
