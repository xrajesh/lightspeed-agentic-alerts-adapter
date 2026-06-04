package adapter

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	agenticv1alpha1 "github.com/openshift/lightspeed-agentic-operator/api/v1alpha1"
	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeAlertSource struct {
	alerts models.GettableAlerts
	err    error
}

func (f *fakeAlertSource) GetAlerts(_ context.Context) (models.GettableAlerts, error) {
	return f.alerts, f.err
}

type fakeProposalClient struct {
	proposals    []agenticv1alpha1.Proposal
	listErr      error
	createErr    error
	created      []*agenticv1alpha1.Proposal
	createCalls  int
	wasCreated   *bool
}

func (f *fakeProposalClient) ListProposals(_ context.Context) ([]agenticv1alpha1.Proposal, error) {
	return f.proposals, f.listErr
}

func (f *fakeProposalClient) CreateProposal(_ context.Context, p *agenticv1alpha1.Proposal) (bool, error) {
	f.createCalls++
	if f.createErr != nil {
		return false, f.createErr
	}
	if f.wasCreated != nil && !*f.wasCreated {
		return false, nil
	}
	f.created = append(f.created, p)
	return true, nil
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func ptr[T any](v T) *T { return &v }

func makeAlert(name, fingerprint string, startsAt time.Time) *models.GettableAlert {
	return makeAlertWithSeverity(name, fingerprint, startsAt, "warning")
}

func makeAlertWithSeverity(name, fingerprint string, startsAt time.Time, severity string) *models.GettableAlert {
	sa := strfmt.DateTime(startsAt)
	labels := models.LabelSet{"alertname": name}
	if severity != "" {
		labels["severity"] = severity
	}
	return &models.GettableAlert{
		Fingerprint: &fingerprint,
		StartsAt:    &sa,
		Alert: models.Alert{
			Labels: labels,
		},
		Annotations: models.LabelSet{"summary": "test alert"},
	}
}

func makeProposal(fingerprint string, conditions []metav1.Condition) agenticv1alpha1.Proposal {
	return agenticv1alpha1.Proposal{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-" + fingerprint,
			Namespace: "openshift-lightspeed",
			Labels: map[string]string{
				"agentic.openshift.io/alert-fingerprint": fingerprint,
				"agentic.openshift.io/source":            "alertmanager",
			},
		},
		Status: agenticv1alpha1.ProposalStatus{
			Conditions: conditions,
		},
	}
}

func TestReconcile(t *testing.T) {
	now := time.Now()
	oldEnough := now.Add(-10 * time.Minute)
	tooRecent := now.Add(-2 * time.Minute)
	withinCooldown := now.Add(-30 * time.Minute)
	pastCooldown := now.Add(-2 * time.Hour)

	tests := []struct {
		name            string
		alerts          models.GettableAlerts
		proposals       []agenticv1alpha1.Proposal
		wantCreated     int
		wantCreateCalls int
		alertsErr       error
		proposalsErr    error
		createErr       error
		wasCreated      *bool
	}{
		{
			name:            "new alert creates proposal",
			alerts:          models.GettableAlerts{makeAlert("HighCPU", "abcdef1234567890", oldEnough)},
			wantCreated:     1,
			wantCreateCalls: 1,
		},
		{
			name:        "transient alert skipped (initial delay)",
			alerts:      models.GettableAlerts{makeAlert("HighCPU", "abcdef1234567890", tooRecent)},
			wantCreated: 0,
		},
		{
			name:   "active proposal skipped",
			alerts: models.GettableAlerts{makeAlert("HighCPU", "abcdef1234567890", oldEnough)},
			proposals: []agenticv1alpha1.Proposal{
				makeProposal("abcdef12", []metav1.Condition{
					{Type: "Analyzed", Status: metav1.ConditionUnknown},
				}),
			},
			wantCreated: 0,
		},
		{
			name:   "terminal proposal within cooldown skipped",
			alerts: models.GettableAlerts{makeAlert("HighCPU", "abcdef1234567890", oldEnough)},
			proposals: []agenticv1alpha1.Proposal{
				makeProposal("abcdef12", []metav1.Condition{
					{Type: "Analyzed", Status: metav1.ConditionTrue},
					{Type: "Executed", Status: metav1.ConditionTrue},
					{
						Type:               "Verified",
						Status:             metav1.ConditionTrue,
						LastTransitionTime: metav1.NewTime(withinCooldown),
					},
				}),
			},
			wantCreated: 0,
		},
		{
			name:   "terminal proposal past cooldown creates new proposal",
			alerts: models.GettableAlerts{makeAlert("HighCPU", "abcdef1234567890", oldEnough)},
			proposals: []agenticv1alpha1.Proposal{
				makeProposal("abcdef12", []metav1.Condition{
					{Type: "Analyzed", Status: metav1.ConditionTrue},
					{Type: "Executed", Status: metav1.ConditionTrue},
					{
						Type:               "Verified",
						Status:             metav1.ConditionTrue,
						LastTransitionTime: metav1.NewTime(pastCooldown),
					},
				}),
			},
			wantCreated:     1,
			wantCreateCalls: 1,
		},
		{
			name:   "failed proposal within cooldown skipped",
			alerts: models.GettableAlerts{makeAlert("HighCPU", "abcdef1234567890", oldEnough)},
			proposals: []agenticv1alpha1.Proposal{
				makeProposal("abcdef12", []metav1.Condition{
					{
						Type:               "Analyzed",
						Status:             metav1.ConditionFalse,
						LastTransitionTime: metav1.NewTime(withinCooldown),
					},
				}),
			},
			wantCreated: 0,
		},
		{
			name:   "denied proposal within cooldown skipped",
			alerts: models.GettableAlerts{makeAlert("HighCPU", "abcdef1234567890", oldEnough)},
			proposals: []agenticv1alpha1.Proposal{
				makeProposal("abcdef12", []metav1.Condition{
					{
						Type:               "Denied",
						Status:             metav1.ConditionTrue,
						LastTransitionTime: metav1.NewTime(withinCooldown),
					},
				}),
			},
			wantCreated: 0,
		},
		{
			name:   "escalated proposal within cooldown skipped",
			alerts: models.GettableAlerts{makeAlert("HighCPU", "abcdef1234567890", oldEnough)},
			proposals: []agenticv1alpha1.Proposal{
				makeProposal("abcdef12", []metav1.Condition{
					{
						Type:               "Escalated",
						Status:             metav1.ConditionTrue,
						LastTransitionTime: metav1.NewTime(withinCooldown),
					},
				}),
			},
			wantCreated: 0,
		},
		{
			name:        "alertmanager error skips cycle",
			alertsErr:   errors.New("connection refused"),
			wantCreated: 0,
		},
		{
			name:         "kubernetes list error skips cycle",
			alerts:       models.GettableAlerts{makeAlert("HighCPU", "abcdef1234567890", oldEnough)},
			proposalsErr: errors.New("api server unavailable"),
			wantCreated:  0,
		},
		{
			name: "kubernetes create error does not block other alerts",
			alerts: models.GettableAlerts{
				makeAlert("HighCPU", "aaaa111122223333", oldEnough),
				makeAlert("HighMem", "bbbb444455556666", oldEnough),
			},
			createErr:       errors.New("api server error"),
			wantCreated:     0,
			wantCreateCalls: 2,
		},
		{
			name:        "nil startsAt skips alert",
			alerts:      models.GettableAlerts{
				{
					Fingerprint: ptr("abcdef1234567890"),
					Alert: models.Alert{
						Labels: models.LabelSet{"alertname": "HighCPU"},
					},
				},
			},
			wantCreated: 0,
		},
		{
			name:            "already exists proposal not counted as created",
			alerts:          models.GettableAlerts{makeAlert("HighCPU", "abcdef1234567890", oldEnough)},
			wasCreated:      ptr(false),
			wantCreated:     0,
			wantCreateCalls: 1,
		},
		{
			name: "nil fingerprint alert skipped with build error",
			alerts: models.GettableAlerts{
				{
					StartsAt: func() *strfmt.DateTime { dt := strfmt.DateTime(oldEnough); return &dt }(),
					Alert: models.Alert{
						Labels: models.LabelSet{"alertname": "HighCPU"},
					},
				},
			},
			wantCreated:     0,
			wantCreateCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := &fakeAlertSource{alerts: tt.alerts, err: tt.alertsErr}
			pc := &fakeProposalClient{proposals: tt.proposals, listErr: tt.proposalsErr, createErr: tt.createErr, wasCreated: tt.wasCreated}

			a := &Adapter{
				alerts:         as,
				proposals:      pc,
				initialDelay:   initialDelay,
				cooldownWindow: cooldownWindow,
				logger:         quietLogger(),
			}

			a.reconcile(context.Background())

			if len(pc.created) != tt.wantCreated {
				t.Errorf("created %d proposals, want %d", len(pc.created), tt.wantCreated)
			}
			if pc.createCalls != tt.wantCreateCalls {
				t.Errorf("CreateProposal called %d times, want %d", pc.createCalls, tt.wantCreateCalls)
			}
		})
	}
}

func TestSkipSeverity(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		severity string
		want     bool
	}{
		{name: "none is skipped", severity: "none", want: true},
		{name: "info is skipped", severity: "info", want: true},
		{name: "None mixed case is skipped", severity: "None", want: true},
		{name: "INFO uppercase is skipped", severity: "INFO", want: true},
		{name: "NONE uppercase is skipped", severity: "NONE", want: true},
		{name: "Info mixed case is skipped", severity: "Info", want: true},
		{name: "warning is not skipped", severity: "warning", want: false},
		{name: "critical is not skipped", severity: "critical", want: false},
		{name: "empty string is not skipped", severity: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert := makeAlertWithSeverity("TestAlert", "abc123", now, tt.severity)
			got := skipSeverity(alert)
			if got != tt.want {
				t.Errorf("skipSeverity(severity=%q) = %v, want %v", tt.severity, got, tt.want)
			}
		})
	}

	t.Run("missing severity label is not skipped", func(t *testing.T) {
		alert := &models.GettableAlert{
			Alert: models.Alert{
				Labels: models.LabelSet{"alertname": "TestAlert"},
			},
		}
		if skipSeverity(alert) {
			t.Error("skipSeverity() = true for missing severity label, want false")
		}
	})
}

func TestReconcileSkipsSeverity(t *testing.T) {
	now := time.Now()
	oldEnough := now.Add(-10 * time.Minute)

	tests := []struct {
		name            string
		severity        string
		wantCreateCalls int
	}{
		{name: "none severity not processed", severity: "none", wantCreateCalls: 0},
		{name: "info severity not processed", severity: "info", wantCreateCalls: 0},
		{name: "warning severity processed", severity: "warning", wantCreateCalls: 1},
		{name: "critical severity processed", severity: "critical", wantCreateCalls: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert := makeAlertWithSeverity("HighCPU", "abcdef1234567890", oldEnough, tt.severity)
			as := &fakeAlertSource{alerts: models.GettableAlerts{alert}}
			pc := &fakeProposalClient{}

			a := &Adapter{
				alerts:         as,
				proposals:      pc,
				initialDelay:   initialDelay,
				cooldownWindow: cooldownWindow,
				logger:         quietLogger(),
			}

			a.reconcile(context.Background())

			if pc.createCalls != tt.wantCreateCalls {
				t.Errorf("CreateProposal called %d times, want %d", pc.createCalls, tt.wantCreateCalls)
			}
		})
	}
}

func TestRunExitsOnContextCancel(t *testing.T) {
	as := &fakeAlertSource{}
	pc := &fakeProposalClient{}

	a := &Adapter{
		alerts:         as,
		proposals:      pc,
		pollInterval:   time.Hour,
		initialDelay:   initialDelay,
		cooldownWindow: cooldownWindow,
		logger:         quietLogger(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after context cancellation")
	}
}
