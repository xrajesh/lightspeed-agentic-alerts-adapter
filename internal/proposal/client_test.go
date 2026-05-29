package proposal

import (
	"io"
	"log/slog"
	"testing"

	agenticv1alpha1 "github.com/openshift/lightspeed-agentic-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newTestClient(t *testing.T) *Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := agenticv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("registering scheme: %v", err)
	}

	fc := fake.NewClientBuilder().WithScheme(scheme).Build()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return &Client{Client: fc, logger: logger}
}

func TestCreateProposal(t *testing.T) {
	c := newTestClient(t)

	p := &agenticv1alpha1.Proposal{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-proposal-abcdef12",
			Namespace: proposalNamespace,
		},
		Spec: agenticv1alpha1.ProposalSpec{
			Request:  "test request",
			Analysis: agenticv1alpha1.ProposalStep{Agent: defaultAgent},
		},
	}

	if err := c.CreateProposal(t.Context(), p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got agenticv1alpha1.Proposal
	key := p.ObjectMeta
	if err := c.Get(t.Context(), toObjectKey(key), &got); err != nil {
		t.Fatalf("failed to get created proposal: %v", err)
	}

	if got.Spec.Request != "test request" {
		t.Errorf("request = %q, want %q", got.Spec.Request, "test request")
	}
}

func TestCreateProposalAlreadyExists(t *testing.T) {
	c := newTestClient(t)

	p := &agenticv1alpha1.Proposal{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-proposal-abcdef12",
			Namespace: proposalNamespace,
		},
		Spec: agenticv1alpha1.ProposalSpec{
			Request:  "test request",
			Analysis: agenticv1alpha1.ProposalStep{Agent: defaultAgent},
		},
	}

	if err := c.CreateProposal(t.Context(), p); err != nil {
		t.Fatalf("first create: %v", err)
	}

	duplicate := &agenticv1alpha1.Proposal{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-proposal-abcdef12",
			Namespace: proposalNamespace,
		},
		Spec: agenticv1alpha1.ProposalSpec{
			Request:  "test request",
			Analysis: agenticv1alpha1.ProposalStep{Agent: defaultAgent},
		},
	}
	if err := c.CreateProposal(t.Context(), duplicate); err != nil {
		t.Fatalf("duplicate create should succeed (409 swallowed), got: %v", err)
	}
}

func TestListProposals(t *testing.T) {
	c := newTestClient(t)

	matching := &agenticv1alpha1.Proposal{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "matching-abcdef12",
			Namespace: proposalNamespace,
			Labels:    map[string]string{labelSource: sourceValue},
		},
		Spec: agenticv1alpha1.ProposalSpec{
			Request:  "matching",
			Analysis: agenticv1alpha1.ProposalStep{Agent: defaultAgent},
		},
	}
	unrelated := &agenticv1alpha1.Proposal{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "unrelated-12345678",
			Namespace: proposalNamespace,
			Labels:    map[string]string{labelSource: "other"},
		},
		Spec: agenticv1alpha1.ProposalSpec{
			Request:  "unrelated",
			Analysis: agenticv1alpha1.ProposalStep{Agent: defaultAgent},
		},
	}

	if err := c.Create(t.Context(), matching); err != nil {
		t.Fatalf("creating matching proposal: %v", err)
	}
	if err := c.Create(t.Context(), unrelated); err != nil {
		t.Fatalf("creating unrelated proposal: %v", err)
	}

	proposals, err := c.ListProposals(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(proposals) != 1 {
		t.Fatalf("got %d proposals, want 1", len(proposals))
	}
	if proposals[0].Name != "matching-abcdef12" {
		t.Errorf("name = %q, want %q", proposals[0].Name, "matching-abcdef12")
	}
}

func TestListProposalsEmpty(t *testing.T) {
	c := newTestClient(t)

	proposals, err := c.ListProposals(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(proposals) != 0 {
		t.Errorf("got %d proposals, want 0", len(proposals))
	}
}

func toObjectKey(meta metav1.ObjectMeta) types.NamespacedName {
	return types.NamespacedName{Name: meta.Name, Namespace: meta.Namespace}
}
