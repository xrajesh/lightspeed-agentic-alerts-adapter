package proposal

import (
	"context"
	"fmt"
	"log/slog"

	agenticv1alpha1 "github.com/openshift/lightspeed-agentic-operator/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client creates and lists Proposal resources in the cluster.
type Client struct {
	client.Client
	logger *slog.Logger
}

// NewClient creates a Client that wraps the given controller-runtime client.
func NewClient(c client.Client, logger *slog.Logger) *Client {
	return &Client{Client: c, logger: logger}
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
// It returns true if the Proposal was created, false if it already existed.
func (c *Client) CreateProposal(ctx context.Context, p *agenticv1alpha1.Proposal) (bool, error) {
	if err := c.Create(ctx, p); err != nil {
		if apierrors.IsAlreadyExists(err) {
			c.logger.Info("proposal already exists", "name", p.Name, "namespace", p.Namespace)
			return false, nil
		}
		return false, fmt.Errorf("proposal: creating %s/%s: %w", p.Namespace, p.Name, err)
	}
	return true, nil
}
