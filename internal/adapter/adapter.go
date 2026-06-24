// Package adapter implements the poll loop that connects AlertManager alerts
// to Proposal creation with stateless deduplication.
package adapter

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"time"

	agenticv1alpha1 "github.com/openshift/lightspeed-agentic-operator/api/v1alpha1"
	"github.com/prometheus/alertmanager/api/v2/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/lightspeed-agentic-alerts-adapter/internal/config"
	"github.com/openshift/lightspeed-agentic-alerts-adapter/internal/proposal"
)

// AlertSource retrieves firing alerts from an external alerting system.
type AlertSource interface {
	GetAlerts(ctx context.Context) (models.GettableAlerts, error)
}

// ProposalClient manages Proposal custom resources in the cluster.
type ProposalClient interface {
	ListProposals(ctx context.Context) ([]agenticv1alpha1.Proposal, error)
	CreateProposal(ctx context.Context, p *agenticv1alpha1.Proposal) (bool, error)
}

// ConfigSource provides runtime configuration for each poll cycle.
type ConfigSource interface {
	Load(ctx context.Context) config.Config
}

// Adapter polls AlertManager for firing alerts and creates Proposal CRs,
// applying stateless deduplication (initial delay, active-proposal check,
// and cooldown window) on each cycle.
type Adapter struct {
	alerts    AlertSource
	proposals ProposalClient
	config    ConfigSource
	logger    *slog.Logger
}

// New creates an Adapter with the given alert source, proposal client,
// config source, and logger.
func New(alerts AlertSource, proposals ProposalClient, cfg ConfigSource, logger *slog.Logger) *Adapter {
	return &Adapter{
		alerts:    alerts,
		proposals: proposals,
		config:    cfg,
		logger:    logger,
	}
}

// Run starts the poll loop, blocking until the context is cancelled.
func (a *Adapter) Run(ctx context.Context) error {
	cfg := a.config.Load(ctx)
	a.logger.Info("adapter started",
		"pollInterval", cfg.PollInterval,
		"initialDelay", cfg.InitialDelay,
		"cooldownWindow", cfg.CooldownWindow,
		"allowedReceivers", cfg.AllowedReceivers,
	)

	a.reconcile(ctx)

	currentInterval := cfg.PollInterval
	ticker := time.NewTicker(currentInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("adapter stopping")
			return nil
		case <-ticker.C:
			a.reconcile(ctx)

			cfg = a.config.Load(ctx)
			if cfg.PollInterval != currentInterval {
				a.logger.Info("poll interval changed", "old", currentInterval, "new", cfg.PollInterval)
				currentInterval = cfg.PollInterval
				ticker.Reset(currentInterval)
			}
		}
	}
}

func (a *Adapter) reconcile(ctx context.Context) {
	a.logger.Debug("poll cycle start")

	cfg := a.config.Load(ctx)

	alerts, err := a.alerts.GetAlerts(ctx)
	if err != nil {
		a.logger.Error("failed to get alerts", "error", err)
		return
	}

	proposals, err := a.proposals.ListProposals(ctx)
	if err != nil {
		a.logger.Error("failed to list proposals", "error", err)
		return
	}

	now := time.Now()
	var created, skipped int

	for i := range alerts {
		if ctx.Err() != nil {
			return
		}

		alert := alerts[i]

		fingerprint := ""
		if alert.Fingerprint != nil {
			fingerprint = *alert.Fingerprint
		}
		alertName := alert.Labels["alertname"]

		if skipReceiver(alert, cfg.AllowedReceivers) {
			a.logger.Debug("alert skipped: no matching receiver",
				"alertname", alertName,
				"fingerprint", fingerprint,
				"receivers", receiverNames(alert),
			)
			skipped++
			continue
		}

		if skipSeverity(alert) {
			a.logger.Debug("alert skipped: low severity",
				"alertname", alertName,
				"fingerprint", fingerprint,
				"severity", alert.Labels["severity"],
			)
			skipped++
			continue
		}

		if skipInitialDelay(alert, now, cfg.InitialDelay) {
			a.logger.Debug("alert skipped: initial delay",
				"alertname", alertName,
				"fingerprint", fingerprint,
				"startsAt", alert.StartsAt,
				"threshold", cfg.InitialDelay,
			)
			skipped++
			continue
		}

		if hasActiveProposal(alert, proposals) {
			a.logger.Debug("alert skipped: active proposal exists",
				"alertname", alertName,
				"fingerprint", fingerprint,
			)
			skipped++
			continue
		}

		if inCooldown(alert, proposals, now, cfg.CooldownWindow) {
			a.logger.Debug("alert skipped: cooldown window",
				"alertname", alertName,
				"fingerprint", fingerprint,
				"cooldown", cfg.CooldownWindow,
			)
			skipped++
			continue
		}

		p, err := proposal.Build(alert, cfg.Tools)
		if err != nil {
			a.logger.Error("failed to build proposal",
				"alertname", alertName,
				"fingerprint", fingerprint,
				"error", err,
			)
			continue
		}

		wasCreated, err := a.proposals.CreateProposal(ctx, p)
		if err != nil {
			a.logger.Error("failed to create proposal",
				"alertname", alertName,
				"fingerprint", fingerprint,
				"proposal", p.Name,
				"error", err,
			)
			continue
		}

		if wasCreated {
			a.logger.Info("proposal created",
				"alertname", alertName,
				"fingerprint", fingerprint,
				"proposal", p.Name,
			)
			created++
		}
	}

	a.logger.Info("poll cycle complete",
		"alertsTotal", len(alerts),
		"skipped", skipped,
		"created", created,
	)
}

func skipReceiver(alert *models.GettableAlert, allowed []string) bool {
	for _, r := range alert.Receivers {
		if r == nil || r.Name == nil {
			continue
		}
		if slices.Contains(allowed, strings.ToLower(*r.Name)) {
			return false
		}
	}
	return true
}

func receiverNames(alert *models.GettableAlert) []string {
	names := make([]string, 0, len(alert.Receivers))
	for _, r := range alert.Receivers {
		if r != nil && r.Name != nil {
			names = append(names, *r.Name)
		}
	}
	return names
}

func skipSeverity(alert *models.GettableAlert) bool {
	sev := strings.ToLower(string(alert.Labels["severity"]))
	return sev == "none" || sev == "info"
}

func skipInitialDelay(alert *models.GettableAlert, now time.Time, threshold time.Duration) bool {
	if alert.StartsAt == nil {
		return true
	}
	return now.Sub(time.Time(*alert.StartsAt)) < threshold
}

func hasActiveProposal(alert *models.GettableAlert, proposals []agenticv1alpha1.Proposal) bool {
	fp := fingerprintPrefix(alert)
	if fp == "" {
		return false
	}

	for i := range proposals {
		if proposals[i].Labels["agentic.openshift.io/alert-fingerprint"] != fp {
			continue
		}
		phase := agenticv1alpha1.DerivePhase(proposals[i].Status.Conditions)
		if !isTerminal(phase) {
			return true
		}
	}
	return false
}

func inCooldown(alert *models.GettableAlert, proposals []agenticv1alpha1.Proposal, now time.Time, window time.Duration) bool {
	fp := fingerprintPrefix(alert)
	if fp == "" {
		return false
	}

	for i := range proposals {
		if proposals[i].Labels["agentic.openshift.io/alert-fingerprint"] != fp {
			continue
		}
		tt := terminalTime(&proposals[i])
		if tt != nil && now.Sub(*tt) < window {
			return true
		}
	}
	return false
}

// terminalTime returns the LastTransitionTime of the condition that caused
// the proposal to reach a terminal phase, or nil if the proposal is not terminal.
func terminalTime(p *agenticv1alpha1.Proposal) *time.Time {
	phase := agenticv1alpha1.DerivePhase(p.Status.Conditions)

	var condType string
	switch phase {
	case agenticv1alpha1.ProposalPhaseCompleted:
		condType = agenticv1alpha1.ProposalConditionVerified
	case agenticv1alpha1.ProposalPhaseFailed:
		condType = findFailedConditionType(p.Status.Conditions)
	case agenticv1alpha1.ProposalPhaseDenied:
		condType = agenticv1alpha1.ProposalConditionDenied
	case agenticv1alpha1.ProposalPhaseEscalated:
		condType = agenticv1alpha1.ProposalConditionEscalated
	default:
		return nil
	}

	if condType == "" {
		return nil
	}

	for i := range p.Status.Conditions {
		if p.Status.Conditions[i].Type == condType {
			t := p.Status.Conditions[i].LastTransitionTime.Time
			return &t
		}
	}
	return nil
}

func findFailedConditionType(conditions []metav1.Condition) string {
	for i := range conditions {
		if conditions[i].Status == metav1.ConditionFalse {
			return conditions[i].Type
		}
	}
	return ""
}

func isTerminal(phase agenticv1alpha1.ProposalPhase) bool {
	switch phase {
	case agenticv1alpha1.ProposalPhaseCompleted,
		agenticv1alpha1.ProposalPhaseFailed,
		agenticv1alpha1.ProposalPhaseDenied,
		agenticv1alpha1.ProposalPhaseEscalated:
		return true
	}
	return false
}

func fingerprintPrefix(alert *models.GettableAlert) string {
	if alert.Fingerprint == nil {
		return ""
	}
	fp := *alert.Fingerprint
	if len(fp) > proposal.FingerprintLen {
		fp = fp[:proposal.FingerprintLen]
	}
	return fp
}
