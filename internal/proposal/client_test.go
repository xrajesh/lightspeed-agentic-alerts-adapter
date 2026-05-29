package proposal

import (
	"strings"
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
	return &Client{fc}
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

func TestCreateProposalDuplicate(t *testing.T) {
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
	err := c.CreateProposal(t.Context(), duplicate)
	if err == nil {
		t.Fatal("expected error for duplicate, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error %q does not mention already exists", err.Error())
	}
}

func toObjectKey(meta metav1.ObjectMeta) types.NamespacedName {
	return types.NamespacedName{Name: meta.Name, Namespace: meta.Namespace}
}
