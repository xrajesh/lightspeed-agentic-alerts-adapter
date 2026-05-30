// Package alertmanager provides a client for querying alerts from an OpenShift cluster's Alertmanager instance.
package alertmanager

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	amclient "github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/alert"
	"github.com/prometheus/alertmanager/api/v2/models"
)

const (
	defaultURL       = "https://alertmanager-main.openshift-monitoring.svc:9094"
	defaultTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultCAPath    = "/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt"
)

// Config holds the settings for connecting to Alertmanager.
// Zero values are replaced with in-cluster defaults.
type Config struct {
	URL       string
	TokenPath string
	CAPath    string
}

func (c *Config) setDefaults() {
	if c.URL == "" {
		c.URL = defaultURL
	}
	if c.TokenPath == "" {
		c.TokenPath = defaultTokenPath
	}
	if c.CAPath == "" {
		c.CAPath = defaultCAPath
	}
}

// Client queries alerts from an Alertmanager instance.
type Client struct {
	api       *amclient.AlertmanagerAPI
	tokenPath string
}

// New creates a Client configured for the given Alertmanager endpoint.
// It reads the CA certificate at construction time and returns an error if the file is missing or invalid.
func New(cfg Config) (*Client, error) {
	cfg.setDefaults()

	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("alertmanager: parsing url: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("alertmanager: invalid url %q: scheme and host are required", cfg.URL)
	}

	caCert, err := os.ReadFile(cfg.CAPath)
	if err != nil {
		return nil, fmt.Errorf("alertmanager: reading ca certificate from %s: %w", cfg.CAPath, err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("alertmanager: no valid certificates found in %s", cfg.CAPath)
	}

	transport := httptransport.New(u.Host, amclient.DefaultBasePath, []string{u.Scheme})
	defaultTransport := http.DefaultTransport.(*http.Transport).Clone()
	defaultTransport.TLSClientConfig = &tls.Config{
		RootCAs:    caPool,
		MinVersion: tls.VersionTLS12,
	}
	transport.Transport = defaultTransport

	return &Client{
		api:       amclient.New(transport, strfmt.Default),
		tokenPath: cfg.TokenPath,
	}, nil
}

// GetAlerts retrieves the current set of alerts from Alertmanager.
// The ServiceAccount token is re-read on each call to handle token rotation.
func (c *Client) GetAlerts(ctx context.Context) (models.GettableAlerts, error) {
	token, err := os.ReadFile(c.tokenPath)
	if err != nil {
		return nil, fmt.Errorf("alertmanager: reading service account token from %s: %w", c.tokenPath, err)
	}

	active := true
	silenced := false
	inhibited := false
	params := alert.NewGetAlertsParamsWithContext(ctx).
		WithActive(&active).
		WithSilenced(&silenced).
		WithInhibited(&inhibited)
	bearerAuth := httptransport.BearerToken(strings.TrimSpace(string(token)))

	resp, err := c.api.Alert.GetAlerts(params, func(op *runtime.ClientOperation) {
		op.AuthInfo = bearerAuth
	})
	if err != nil {
		return nil, fmt.Errorf("alertmanager: querying alerts: %w", err)
	}

	return resp.GetPayload(), nil
}
