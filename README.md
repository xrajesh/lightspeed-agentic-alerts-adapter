# Lightspeed Agentic Alerts Adapter

A component that bridges OpenShift cluster alerts into the [Lightspeed Agentic](https://github.com/openshift/lightspeed-agentic-operator) system. It polls the in-cluster AlertManager API for firing alerts and creates `Proposal` custom resources (`agentic.openshift.io/v1alpha1`) to trigger automated analysis and remediation workflows.

## Quick start

### Prerequisites

- Go 1.26+
- Access to an OpenShift cluster (for deployment)
- [golangci-lint](https://golangci-lint.run/) (for linting)

### Build

```sh
make build        # outputs ./bin/alerts-adapter
```

### Test

```sh
make test         # run all tests
make coverage     # generate HTML coverage report
make lint         # run golangci-lint
```

### Run locally

```sh
make run
```

### Container

```sh
make container-build
```

### Deploy to OpenShift

```sh
kubectl apply -f manifests/
```

The adapter runs as a single-replica Deployment in the `openshift-lightspeed` namespace, using in-cluster authentication.

## Configuration

| Environment variable | Default | Description |
|---|---|---|
| `ALERTMANAGER_URL` | `https://alertmanager-main.openshift-monitoring.svc:9094` | AlertManager API endpoint |

### ConfigMap

Runtime-tunable parameters are read from the `alerts-adapter-config` ConfigMap in the `openshift-lightspeed` namespace (key: `config.yaml`). Changes are picked up on the next poll cycle — no restart required. If the ConfigMap is missing or malformed, defaults are used.

| Field | Default | Description |
|---|---|---|
| `pollInterval` | `30s` | How often to poll AlertManager |
| `initialDelay` | `5m` | Minimum time an alert must fire before a Proposal is created |
| `cooldownWindow` | `1h` | Minimum time after a terminal Proposal before retrying the same alert |
| `allowedReceivers` | `[]` | Receiver allowlist — only alerts routed to at least one of these receivers are processed (case-insensitive). Empty by default; no proposals are created until receivers are explicitly configured |

#### Tools / Skills

Skills (OCI images with runbook paths) can be configured at a shared level or per Proposal step (`analysis`, `execution`, `verification`). Per-step skills override shared skills for that step.

| Field | Description |
|---|---|
| `tools.skills` | Shared skills applied to all steps |
| `analysis.tools.skills` | Skills for the analysis step only |
| `execution.tools.skills` | Skills for the execution step only |
| `verification.tools.skills` | Skills for the verification step only |

Each skills entry requires `image` (OCI image reference) and `paths` (list of paths within the image).

#### Example ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alerts-adapter-config
  namespace: openshift-lightspeed
data:
  config.yaml: |
    pollInterval: "45s"
    initialDelay: "10m"
    cooldownWindow: "2h"
    allowedReceivers:
      - critical
      - warning
    tools:
      skills:
        - image: quay.io/example/shared-runbooks:latest
          paths:
            - /runbooks/common
    analysis:
      tools:
        skills:
          - image: quay.io/example/analysis-runbooks:latest
            paths:
              - /runbooks/analysis
    execution:
      tools:
        skills:
          - image: quay.io/example/exec-runbooks:latest
            paths:
              - /runbooks/exec
```

## Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) — design rationale, requirements, deployment model, and future work
- [docs/receivers.md](docs/receivers.md) — what AlertManager receivers are and how the adapter uses them for filtering
- [openspec/specs/](openspec/specs/) — detailed specs for each subsystem, managed with the [OpenSpec](https://github.com/Fission-AI/OpenSpec) framework

## License

[Apache License 2.0](LICENSE)
