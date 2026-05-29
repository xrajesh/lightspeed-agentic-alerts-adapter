package proposal

import (
	"context"
	"fmt"
	"log/slog"

	agenticv1alpha1 "github.com/openshift/lightspeed-agentic-operator/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client creates and lists Proposal resources in the cluster.
type Client struct {
	client.Client
	logger *slog.Logger
}

// NewClient creates a Client using in-cluster config with the Proposal scheme registered.
func NewClient(logger *slog.Logger) (*Client, error) {
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

	return &Client{Client: c, logger: logger}, nil
}

// ListProposals returns all Proposals created by this adapter, filtered by the
// source=alertmanager label.
func (c *Client) ListProposals(ctx context.Context) ([]agenticv1alpha1.Proposal, error) {
	var list agenticv1alpha1.ProposalList
	if err := c.List(ctx, &list, client.MatchingLabels{labelSource: sourceValue}); err != nil {
		return nil, fmt.Errorf("proposal: listing proposals: %w", err)
	}
	return list.Items, nil
}

// CreateProposal creates a Proposal resource in the cluster.
// If the Proposal already exists (409), it logs and returns nil.
func (c *Client) CreateProposal(ctx context.Context, p *agenticv1alpha1.Proposal) error {
	if err := c.Create(ctx, p); err != nil {
		if apierrors.IsAlreadyExists(err) {
			c.logger.Info("proposal already exists", "name", p.Name, "namespace", p.Namespace)
			return nil
		}
		return fmt.Errorf("proposal: creating %s/%s: %w", p.Namespace, p.Name, err)
	}
	return nil
}
