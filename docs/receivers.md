# Receivers and Alert Filtering

## What is a receiver?

In Prometheus Alertmanager, a **receiver** is a named destination for alert notifications. Each receiver defines where and how alerts are delivered — for example, to PagerDuty, Slack, email, or a webhook endpoint. Receivers are the final stage of AlertManager's routing pipeline: after alerts are grouped, silenced, and inhibited, they are dispatched to one or more receivers based on the routing tree.

A single alert can match multiple receivers when the routing configuration uses `continue: true` on a route, causing the alert to fall through to additional routes.

## Receivers in OpenShift

OpenShift ships a default AlertManager configuration in the `openshift-monitoring` namespace. The platform defines built-in receivers that route alerts based on severity and other labels. Common receivers include:

| Receiver | Typical purpose |
|---|---|
| `Critical` | High-severity alerts requiring immediate action |
| `Default` | Catch-all for alerts that don't match a more specific route |
| `Watchdog` | Heartbeat alerts to verify the monitoring pipeline is alive |
| `null` | Sink for alerts that should be silenced |

Cluster administrators can customize AlertManager routing by creating `AlertmanagerConfig` CRs in user namespaces or by editing the global AlertManager configuration via the `alertmanager-main` secret in `openshift-monitoring`. This allows adding custom receivers (e.g., a `PagerDuty` or `slack-oncall` receiver) and routing specific alerts to them.

## How the alerts adapter uses receivers

The alerts adapter uses receivers as a **scoping mechanism** to control which alerts trigger automated remediation. Rather than processing every firing alert, the adapter checks whether an alert is routed to at least one receiver in a configurable allowlist. Only matching alerts produce `Proposal` CRs.

### Filtering logic

When AlertManager returns its list of firing alerts (`GET /api/v2/alerts`), each alert includes the receivers it was dispatched to. The adapter applies a receiver filter as the **first check** in its reconcile loop, before severity filtering, initial delay, or deduplication checks:

1. Read the `allowedReceivers` list from configuration.
2. For each alert, iterate over its receivers.
3. If any receiver name matches an entry in the allowlist (case-insensitive), the alert passes.
4. If no receiver matches, the alert is skipped.

This means the routing decisions already made by AlertManager — based on labels, severity, namespace, and the routing tree — determine which alerts the adapter acts on. Administrators don't need to duplicate filtering rules in the adapter; they configure AlertManager routing to send the right alerts to a designated receiver, and the adapter picks up only those.

### Default behavior

When no `allowedReceivers` field is configured (or the ConfigMap is absent), the adapter defaults to an empty list. This means no alerts produce Proposals until receivers are explicitly configured in the ConfigMap.

### Configuration

The allowlist is set via the `alerts-adapter-config` ConfigMap in the `openshift-lightspeed` namespace:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alerts-adapter-config
  namespace: openshift-lightspeed
data:
  config.yaml: |
    allowedReceivers:
      - critical
      - pagerduty
```

Key behaviors:

- **Field absent** — uses the default empty list (no alerts are processed).
- **Explicit empty list (`[]`)** — disables all Proposal creation; no alerts are processed.
- **Case-insensitive** — `"Critical"`, `"critical"`, and `"CRITICAL"` all match the same receiver.
- **Multiple receivers** — an alert passes if *any* of its receivers appears in the allowlist.
- **Hot-reloadable** — changes to the ConfigMap take effect on the next poll cycle without restarting the adapter.
