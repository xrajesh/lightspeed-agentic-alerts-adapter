## Package Structure

```
internal/proposal/
├── build.go          # Build() function: GettableAlert → Proposal
├── build_test.go     # Table-driven tests for Build and helpers
├── client.go         # Thin Kubernetes client for creating Proposals
├── client_test.go    # Client tests using fake controller-runtime client
└── request.tmpl      # Embedded text/template for spec.request
```

## Alert-to-Proposal Mapping

### Deterministic Naming

Proposal names follow the pattern `{alertname}-{namespace}-{fingerprint[:8]}`, or `{alertname}-{fingerprint[:8]}` for cluster-scoped alerts (no namespace label).

- Components are lowercased and sanitized to DNS-1123 subdomain format
- Non-alphanumeric characters are replaced with hyphens
- Names are truncated to 253 characters (Kubernetes maximum for object names)
- The 8-character fingerprint suffix ensures uniqueness across alerts with the same name and namespace

This makes Proposal creation idempotent — the same firing alert always produces the same Proposal name. Kubernetes rejects duplicate creates with 409 AlreadyExists, which callers can treat as success.

### Metadata

**Labels** (for filtering and deduplication):

| Label | Value |
|-------|-------|
| `agentic.openshift.io/source` | `alertmanager` |
| `agentic.openshift.io/alert-fingerprint` | First 8 chars of fingerprint |
| `agentic.openshift.io/alert-name` | Sanitized alert name |
| `agentic.openshift.io/alert-severity` | Sanitized severity |

**Annotations** (debugging context, not indexed):

| Annotation | Value |
|------------|-------|
| `agentic.openshift.io/alert-starts-at` | RFC 3339 timestamp |
| `agentic.openshift.io/alert-summary` | Truncated summary (max 256 chars) |

All label values are sanitized to Kubernetes restrictions: max 63 characters, alphanumeric plus hyphens/underscores/dots, must start and end with alphanumeric.

### Namespace

Proposals are created in `openshift-lightspeed`. The alert's `namespace` label populates `spec.targetNamespaces` when present.

### Spec Fields

| Field | Value |
|-------|-------|
| `request` | Rendered from embedded template (see below) |
| `targetNamespaces` | `[alert.labels["namespace"]]` if present, omitted otherwise |
| `analysis` | `{agent: "default"}` |
| `execution` | `{agent: "default"}` |
| `verification` | `{agent: "default"}` |

All three workflow steps are populated so the operator runs the full remediation cycle: analyze → execute → verify.

### Request Template

Embedded via `//go:embed request.tmpl`. Template data is populated from alert fields:

```
A Kubernetes alert is firing in the cluster.
Investigate the root cause and propose a remediation.

Alert: {{ .AlertName }}
Severity: {{ .Severity }}
Namespace: {{ .Namespace }}
Summary: {{ .Summary }}
Description: {{ .Description }}

Labels:
{{ range $k, $v := .Labels }}  {{ $k }}: {{ $v }}
{{ end }}
```

## Kubernetes Client

A thin wrapper around `controller-runtime/pkg/client`:

- `NewClient()` — creates a typed client with the Proposal scheme registered, using in-cluster config
- `CreateProposal(ctx, *Proposal) error` — creates a single Proposal CR

The client is deliberately minimal. It does not list, update, or delete Proposals — the adapter is create-only. Deduplication logic and polling belong in a future change that wires everything together in `main.go`.

## Dependencies

New direct dependencies:
- `github.com/openshift/lightspeed-agentic-operator/api/v1alpha1` — Proposal types
- `k8s.io/apimachinery` — ObjectMeta, TypeMeta
- `sigs.k8s.io/controller-runtime` — client.New, client.Client, scheme registration

## Decisions

- **Full remediation workflow**: All three steps (analysis, execution, verification) reference the `default` agent. This gives the operator the full analyze → execute → verify pipeline out of the box.
- **Proposals live in `openshift-lightspeed`**: Proposals may reference secrets (via `spec.tools.requiredSecrets`) that the operator mounts into sandbox pods. Kubernetes requires secrets to live in the same namespace as the pod that consumes them, so the Proposal, its sandbox pods, and the referenced secrets must all share a namespace. `openshift-lightspeed` is the operator's home namespace and the natural place for these resources.
- **No runbook URL in labels**: Runbook URLs often exceed the 63-char label value limit. We include them in the request template instead, where they're most useful to the analysis agent.
- **Template over string concatenation**: An embedded template is easier to read, test, and evolve. The prompt is the interface between the adapter and the agent — it deserves to be a first-class artifact.
