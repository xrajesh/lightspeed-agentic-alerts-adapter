## 1. Dependencies and Project Setup

- [x] 1.1 Add the Alertmanager client library dependency to `go.mod`
- [x] 1.2 Run `go mod tidy` to resolve transitive dependencies

## 2. Alertmanager Client

- [x] 2.1 Create `internal/alertmanager/` package with a client struct that wraps the Alertmanager client library
- [x] 2.2 Implement in-cluster authentication: read SA bearer token from `/var/run/secrets/kubernetes.io/serviceaccount/token` and configure TLS with `/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt`
- [x] 2.3 Re-read the token file on each request to handle token rotation
- [x] 2.4 Implement a `GetAlerts` function that queries Alertmanager and returns alerts using the library's types
- [x] 2.5 Support configurable Alertmanager URL via `ALERTMANAGER_URL` environment variable with default `https://alertmanager-main.openshift-monitoring.svc:9094`
- [x] 2.6 Return meaningful errors for unreachable Alertmanager, auth failures, and missing token/CA files

## 3. Startup Integration

- [x] 3.1 Wire the Alertmanager client into `cmd/alerts-adapter/main.go` `run()` function
- [x] 3.2 Fetch alerts on startup and log the count and key details using structured logging (`slog`)
- [x] 3.3 Exit with non-zero status code if alert retrieval fails on startup

## 4. Testing

- [x] 4.1 Write unit tests for the Alertmanager client using a test HTTP server to simulate Alertmanager responses
- [x] 4.2 Test error cases: unreachable server, auth failure (401/403), missing token file
- [x] 4.3 Test configuration: custom URL via environment variable, default URL fallback
