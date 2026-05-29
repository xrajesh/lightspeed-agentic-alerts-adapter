package proposal

import (
	"context"
	"fmt"

	agenticv1alpha1 "github.com/openshift/lightspeed-agentic-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client creates Proposal resources in the cluster.
type Client struct {
	client.Client
}

// NewClient creates a Client using in-cluster config with the Proposal scheme registered.
func NewClient() (*Client, error) {
	scheme := runtime.NewScheme()
	if err := agenticv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("proposal: registering scheme: %w", err)
	}

	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("proposal: loading kubeconfig: %w", err)
	}

	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("proposal: creating client: %w", err)
	}

	return &Client{c}, nil
}

// CreateProposal creates a Proposal resource in the cluster.
func (c *Client) CreateProposal(ctx context.Context, p *agenticv1alpha1.Proposal) error {
	if err := c.Create(ctx, p); err != nil {
		return fmt.Errorf("proposal: creating %s/%s: %w", p.Namespace, p.Name, err)
	}
	return nil
}
