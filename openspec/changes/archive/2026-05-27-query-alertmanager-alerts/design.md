## Context

The lightspeed-agentic-alerts-adapter is a greenfield Go service that will run inside an OpenShift cluster. Its first job is to retrieve active alerts from the cluster's Alertmanager so downstream consumers can reason about cluster health. Today, the project contains only a skeleton `main.go` with structured logging — no alert retrieval exists yet.

OpenShift clusters expose Alertmanager behind an internal service (`alertmanager-main` in the `openshift-monitoring` namespace) accessible over HTTPS. Pods authenticate using their ServiceAccount bearer token and must trust the serving certificate issued by the OpenShift service CA (mounted at `service-ca.crt`).

## Goals / Non-Goals

**Goals:**
- Retrieve currently firing alerts from the cluster's Alertmanager using an existing client library.
- Use the alert types provided by the Alertmanager client library directly.
- Log retrieved alerts on startup as a proof-of-life for the capability.
- Handle authentication and TLS for in-cluster communication.
- Define least-privilege RBAC so the adapter's ServiceAccount has only the permissions it needs.

**Non-Goals:**
- Exposing an HTTP API from the adapter (future work).
- Filtering, deduplication, or enrichment of alerts (future work).
- Running outside the cluster (out-of-cluster kubeconfig support).
- Watching or streaming alerts — this is a point-in-time query.
- Custom domain alert types — use the types from the Alertmanager library; introduce our own if decoupling becomes necessary later.

## Decisions

### 1. Use an existing Alertmanager Go client library

Use the official Alertmanager client library (from the `prometheus/alertmanager` module or its generated API client) rather than hand-rolling HTTP calls.

**Rationale:** An existing client handles endpoint paths, response parsing, and API versioning. It reduces boilerplate and aligns with how other Go projects in the Prometheus ecosystem interact with Alertmanager.

**Alternatives considered:**
- *Raw `net/http` with manual JSON parsing* — more code to maintain, easy to get wrong as the API evolves, and duplicates work the client library already does.

### 2. Use Alertmanager library types directly

Use the alert types provided by the Alertmanager client dependency rather than defining custom domain types.

**Rationale:** Adding a separate domain model introduces mapping boilerplate with no benefit at this stage. The Alertmanager types already represent alerts accurately. Custom domain types can be introduced later if we need to decouple from the upstream API model.

**Alternatives considered:**
- *Separate `internal/model/` package* — premature abstraction; adds code and maintenance burden without a current consumer that needs a different representation.

### 3. Authenticate with the in-cluster ServiceAccount token and OpenShift service CA

Read the bearer token from `/var/run/secrets/kubernetes.io/serviceaccount/token` and send it in the `Authorization` header. Trust the OpenShift service-serving CA from `/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt` for TLS verification.

**Rationale:** This is the standard in-cluster authentication mechanism for OpenShift. The `service-ca.crt` is the correct trust anchor for services whose serving certificates are issued by the OpenShift service CA operator, which includes the monitoring stack.

**Alternatives considered:**
- *Using `ca.crt`* — that's the Kubernetes API server CA, not the service-serving CA used by Alertmanager in OpenShift.
- *OAuth proxy sidecar* — adds deployment complexity for no benefit when the SA token is already available.

### 4. Least-privilege RBAC

Define a dedicated ServiceAccount, ClusterRole, and ClusterRoleBinding scoped to the minimum permissions required:
- The ClusterRole grants only `GET` on the Alertmanager API (specifically the `monitoring.coreos.com` resources or the `/api/v2/alerts` route, depending on how access is gated).
- The ClusterRoleBinding binds only the adapter's ServiceAccount in its namespace.
- No `list`, `watch`, `create`, `update`, or `delete` permissions beyond what alert retrieval requires.

**Rationale:** Following the principle of least privilege minimizes blast radius if the adapter's token is compromised. The adapter only reads alerts — it should have no write access to anything.

### 5. Package layout

- `internal/alertmanager/` — Client wrapper around the Alertmanager client library, handling auth, TLS, and configuration.

**Rationale:** `internal/` prevents external import. A single package is sufficient since we're using the library's own types.

### 6. Configuration via environment variables

The Alertmanager URL will be configurable via an environment variable (e.g., `ALERTMANAGER_URL`) with a sensible default pointing to the in-cluster service (`https://alertmanager-main.openshift-monitoring.svc:9094`).

**Rationale:** Environment variables are the simplest configuration mechanism for a Kubernetes workload and align with 12-factor practices. No config file parsing needed at this stage.

## Risks / Trade-offs

- **[RBAC misconfiguration]** → If the ClusterRole is too restrictive, requests will return 403. Mitigation: document the exact RBAC manifests and test them in a real cluster.
- **[Alertmanager URL may vary]** → Different OpenShift versions or custom monitoring stacks may place Alertmanager at a different address or port. Mitigation: make the URL configurable via environment variable.
- **[Client library compatibility]** → The Alertmanager client library version must be compatible with the Alertmanager version shipped in the target OpenShift release. Mitigation: pin the dependency and test against the target cluster.
- **[Large alert volume]** → A cluster with many firing alerts could return a large response. Mitigation: acceptable for now; pagination or filtering can be added later.
- **[Token expiry]** → Projected SA tokens rotate. Mitigation: re-read the token file on each request rather than caching it at startup.
- **[Coupling to Alertmanager types]** → Using library types directly couples consumers to the upstream model. Mitigation: acceptable trade-off for now; introduce a domain layer when a concrete need arises.
