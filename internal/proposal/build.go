// Package proposal translates Alertmanager alerts into Proposal custom resources.
package proposal

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"

	agenticv1alpha1 "github.com/openshift/lightspeed-agentic-operator/api/v1alpha1"
	"github.com/prometheus/alertmanager/api/v2/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	proposalNamespace = "openshift-lightspeed"
	defaultAgent      = "default"

	labelSource      = "agentic.openshift.io/source"
	labelFingerprint = "agentic.openshift.io/alert-fingerprint"
	labelAlertName   = "agentic.openshift.io/alert-name"
	labelSeverity    = "agentic.openshift.io/alert-severity"
	annotStartsAt    = "agentic.openshift.io/alert-starts-at"
	annotSummary     = "agentic.openshift.io/alert-summary"

	sourceValue = "alertmanager"

	maxLabelValueLen = 63
	maxNameLen       = 253
	// FingerprintLen is the number of characters used from the alert fingerprint
	// for labels and dedup matching. Exported for use by the adapter package.
	FingerprintLen = 8
	maxSummaryLen    = 256
)

var (
	invalidLabelChars = regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	invalidDNSChars   = regexp.MustCompile(`[^a-z0-9-]`)

	//go:embed request.tmpl
	requestTemplateStr string
	requestTemplate    = template.Must(template.New("request").Parse(requestTemplateStr))
)

type requestData struct {
	AlertName   string
	Severity    string
	RunbookURL  string
	Namespace   string
	Summary     string
	Description string
	Labels      map[string]string
}

// Build constructs a Proposal from a single Alertmanager GettableAlert.
// The Proposal name is deterministic based on the alert's identity (alertname,
// namespace, fingerprint), making repeated calls for the same alert safe
// against duplicate creation via Kubernetes 409 AlreadyExists.
// When skills is non-empty, spec.tools.skills is set on the Proposal.
func Build(a *models.GettableAlert, skills []agenticv1alpha1.SkillsSource) (*agenticv1alpha1.Proposal, error) {
	if a.Fingerprint == nil {
		return nil, fmt.Errorf("proposal: alert fingerprint is nil")
	}

	alertName := a.Labels["alertname"]
	namespace := a.Labels["namespace"]
	fingerprint := *a.Fingerprint
	severity := a.Labels["severity"]

	request, err := buildRequest(a)
	if err != nil {
		return nil, err
	}

	p := &agenticv1alpha1.Proposal{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "agentic.openshift.io/v1alpha1",
			Kind:       "Proposal",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        buildName(alertName, namespace, fingerprint),
			Namespace:   proposalNamespace,
			Labels:      buildLabels(alertName, severity, fingerprint),
			Annotations: buildAnnotations(a),
		},
		Spec: agenticv1alpha1.ProposalSpec{
			Request:      request,
			Analysis:     agenticv1alpha1.ProposalStep{Agent: defaultAgent},
			Execution:    agenticv1alpha1.ProposalStep{Agent: defaultAgent},
			Verification: agenticv1alpha1.ProposalStep{Agent: defaultAgent},
		},
	}

	if namespace != "" {
		p.Spec.TargetNamespaces = []string{namespace}
	}

	if len(skills) > 0 {
		p.Spec.Tools = agenticv1alpha1.ToolsSpec{
			Skills: skills,
		}
	}

	return p, nil
}

// buildName produces a deterministic DNS-compatible name: {alertname}-{namespace}-{fingerprint[:8]}
// or {alertname}-{fingerprint[:8]} for cluster-scoped alerts.
func buildName(alertName, namespace, fingerprint string) string {
	fp := fingerprint
	if len(fp) > FingerprintLen {
		fp = fp[:FingerprintLen]
	}

	name := strings.ToLower(alertName)
	name = invalidDNSChars.ReplaceAllString(name, "-")

	if namespace != "" {
		ns := strings.ToLower(namespace)
		ns = invalidDNSChars.ReplaceAllString(ns, "-")
		combined := name + "-" + ns + "-" + fp
		if len(combined) <= maxNameLen {
			return combined
		}
		available := max(maxNameLen-len(ns)-len(fp)-2, 1)
		return truncateDNS(name, available) + "-" + ns + "-" + fp
	}

	combined := name + "-" + fp
	if len(combined) <= maxNameLen {
		return combined
	}
	available := maxNameLen - len(fp) - 1
	return truncateDNS(name, available) + "-" + fp
}

// buildLabels sets Kubernetes labels for alert traceability and filtering.
func buildLabels(alertName, severity, fingerprint string) map[string]string {
	labels := map[string]string{
		labelSource:      sourceValue,
		labelFingerprint: sanitizeLabelValue(fingerprint[:min(len(fingerprint), FingerprintLen)]),
		labelAlertName:   sanitizeLabelValue(strings.ToLower(alertName)),
		labelSeverity:    sanitizeLabelValue(severity),
	}
	return labels
}

// buildAnnotations sets Kubernetes annotations with non-indexed alert metadata.
func buildAnnotations(a *models.GettableAlert) map[string]string {
	annots := map[string]string{}

	if a.StartsAt != nil && !time.Time(*a.StartsAt).IsZero() {
		annots[annotStartsAt] = time.Time(*a.StartsAt).UTC().Format(time.RFC3339)
	}

	if summary, ok := a.Annotations["summary"]; ok && summary != "" {
		if len(summary) > maxSummaryLen {
			summary = summary[:maxSummaryLen]
		}
		annots[annotSummary] = summary
	}

	return annots
}

// buildRequest renders the embedded template with alert data for the analysis agent.
func buildRequest(a *models.GettableAlert) (string, error) {
	data := requestData{
		AlertName:   a.Labels["alertname"],
		Severity:    a.Labels["severity"],
		RunbookURL:  a.Annotations["runbook_url"],
		Namespace:   a.Labels["namespace"],
		Summary:     a.Annotations["summary"],
		Description: a.Annotations["description"],
		Labels:      a.Labels,
	}

	var buf bytes.Buffer
	if err := requestTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("proposal: rendering request template: %w", err)
	}
	return buf.String(), nil
}

// sanitizeLabelValue ensures a string conforms to Kubernetes label value rules:
// max 63 chars, alphanumeric/hyphens/underscores/dots, must start and end with alphanumeric.
func sanitizeLabelValue(s string) string {
	if s == "" {
		return s
	}

	s = invalidLabelChars.ReplaceAllString(s, "-")

	if len(s) > maxLabelValueLen {
		s = s[:maxLabelValueLen]
	}

	s = strings.TrimLeft(s, "-_.")
	s = strings.TrimRight(s, "-_.")

	return s
}

func truncateDNS(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	s = s[:maxLen]
	s = strings.TrimRight(s, "-_.")
	return s
}
