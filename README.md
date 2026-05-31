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

Internal polling parameters (constants in `internal/adapter/adapter.go`):

- **Poll interval**: 30s
- **Initial delay**: 5 min (minimum time an alert must fire before a Proposal is created)
- **Cooldown window**: 1 hour (minimum time after a terminal Proposal before retrying)

## Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) — design rationale, requirements, deployment model, and future work
- [openspec/specs/](openspec/specs/) — detailed specs for each subsystem, managed with the [OpenSpec](https://github.com/samber/openspec) framework

## License

[Apache License 2.0](LICENSE)
