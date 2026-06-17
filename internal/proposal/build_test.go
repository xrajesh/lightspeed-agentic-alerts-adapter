package proposal

import (
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	agenticv1alpha1 "github.com/openshift/lightspeed-agentic-operator/api/v1alpha1"
	"github.com/prometheus/alertmanager/api/v2/models"
)

func strPtr(s string) *string { return &s }

func makeAlert(alertName, namespace, fingerprint, severity string) *models.GettableAlert {
	startsAt := strfmt.DateTime(time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC))
	return &models.GettableAlert{
		Alert: models.Alert{
			Labels: models.LabelSet{
				"alertname": alertName,
				"namespace": namespace,
				"severity":  severity,
			},
			GeneratorURL: "http://prometheus:9090/graph",
		},
		Annotations: models.LabelSet{
			"summary":     "Pod is crash looping",
			"description": "Pod my-pod has restarted 5 times in the last hour",
		},
		Fingerprint: strPtr(fingerprint),
		StartsAt:    &startsAt,
		Status: &models.AlertStatus{
			State: strPtr("active"),
		},
	}
}

func TestBuild(t *testing.T) {
	tests := []struct {
		name              string
		alert             *models.GettableAlert
		expectedName      string
		expectedNamespace string
		expectedTargetNS  []string
		expectedLabels    map[string]string
	}{
		{
			name:              "namespaced alert",
			alert:             makeAlert("KubePodCrashLooping", "production", "abcdef1234567890", "critical"),
			expectedName:      "kubepodcrashlooping-production-abcdef12",
			expectedNamespace: proposalNamespace,
			expectedTargetNS:  []string{"production"},
			expectedLabels: map[string]string{
				labelSource:      sourceValue,
				labelFingerprint: "abcdef12",
				labelAlertName:   "kubepodcrashlooping",
				labelSeverity:    "critical",
			},
		},
		{
			name: "cluster-scoped alert (no namespace)",
			alert: func() *models.GettableAlert {
				a := makeAlert("ClusterVersionAvailable", "", "ff00ff00ff00ff00", "info")
				delete(a.Labels, "namespace")
				return a
			}(),
			expectedName:      "clusterversionavailable-ff00ff00",
			expectedNamespace: proposalNamespace,
			expectedTargetNS:  nil,
			expectedLabels: map[string]string{
				labelSource:      sourceValue,
				labelFingerprint: "ff00ff00",
				labelAlertName:   "clusterversionavailable",
				labelSeverity:    "info",
			},
		},
		{
			name:              "short fingerprint (less than 8 chars)",
			alert:             makeAlert("TestAlert", "ns", "abc", "warning"),
			expectedName:      "testalert-ns-abc",
			expectedNamespace: proposalNamespace,
			expectedTargetNS:  []string{"ns"},
			expectedLabels: map[string]string{
				labelSource:      sourceValue,
				labelFingerprint: "abc",
				labelAlertName:   "testalert",
				labelSeverity:    "warning",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Build(tt.alert, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if p.Name != tt.expectedName {
				t.Errorf("name = %q, want %q", p.Name, tt.expectedName)
			}
			if p.Namespace != tt.expectedNamespace {
				t.Errorf("namespace = %q, want %q", p.Namespace, tt.expectedNamespace)
			}

			if len(tt.expectedTargetNS) == 0 && len(p.Spec.TargetNamespaces) != 0 {
				t.Errorf("targetNamespaces = %v, want empty", p.Spec.TargetNamespaces)
			}
			if len(tt.expectedTargetNS) > 0 {
				if len(p.Spec.TargetNamespaces) != len(tt.expectedTargetNS) {
					t.Fatalf("targetNamespaces length = %d, want %d", len(p.Spec.TargetNamespaces), len(tt.expectedTargetNS))
				}
				for i, ns := range tt.expectedTargetNS {
					if p.Spec.TargetNamespaces[i] != ns {
						t.Errorf("targetNamespaces[%d] = %q, want %q", i, p.Spec.TargetNamespaces[i], ns)
					}
				}
			}

			for k, v := range tt.expectedLabels {
				if got := p.Labels[k]; got != v {
					t.Errorf("label %q = %q, want %q", k, got, v)
				}
			}
		})
	}
}

func TestBuildNilFingerprint(t *testing.T) {
	a := makeAlert("TestAlert", "ns", "abcdef12", "warning")
	a.Fingerprint = nil

	_, err := Build(a, nil)
	if err == nil {
		t.Fatal("expected error for nil fingerprint, got nil")
	}

	want := "fingerprint is nil"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err, want)
	}
}

func TestBuildWorkflowSteps(t *testing.T) {
	a := makeAlert("TestAlert", "ns", "abcdef12", "warning")
	p, err := Build(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Spec.Analysis.Agent != defaultAgent {
		t.Errorf("analysis.agent = %q, want %q", p.Spec.Analysis.Agent, defaultAgent)
	}
	if p.Spec.Execution.Agent != defaultAgent {
		t.Errorf("execution.agent = %q, want %q", p.Spec.Execution.Agent, defaultAgent)
	}
	if p.Spec.Verification.Agent != defaultAgent {
		t.Errorf("verification.agent = %q, want %q", p.Spec.Verification.Agent, defaultAgent)
	}
}

func TestBuildDeterministicNaming(t *testing.T) {
	a := makeAlert("KubePodCrashLooping", "production", "abcdef1234567890", "critical")

	p1, err := Build(a, nil)
	if err != nil {
		t.Fatalf("first build: %v", err)
	}
	p2, err := Build(a, nil)
	if err != nil {
		t.Fatalf("second build: %v", err)
	}

	if p1.Name != p2.Name {
		t.Errorf("names differ: %q vs %q", p1.Name, p2.Name)
	}
}

func TestBuildAnnotations(t *testing.T) {
	t.Run("starts-at is RFC3339 UTC", func(t *testing.T) {
		a := makeAlert("TestAlert", "ns", "abcdef12", "warning")
		p, err := Build(a, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got := p.Annotations[annotStartsAt]
		expected := "2026-01-15T10:30:00Z"
		if got != expected {
			t.Errorf("starts-at = %q, want %q", got, expected)
		}
	})

	t.Run("summary is included", func(t *testing.T) {
		a := makeAlert("TestAlert", "ns", "abcdef12", "warning")
		p, err := Build(a, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := p.Annotations[annotSummary]; got != "Pod is crash looping" {
			t.Errorf("summary = %q, want %q", got, "Pod is crash looping")
		}
	})

	t.Run("missing annotations produce empty map entries", func(t *testing.T) {
		a := makeAlert("TestAlert", "ns", "abcdef12", "warning")
		a.Annotations = models.LabelSet{}
		startsAt := strfmt.DateTime(time.Time{})
		a.StartsAt = &startsAt

		p, err := Build(a, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := p.Annotations[annotStartsAt]; ok {
			t.Error("expected no starts-at annotation for zero time")
		}
		if _, ok := p.Annotations[annotSummary]; ok {
			t.Error("expected no summary annotation when empty")
		}
	})

	t.Run("long summary is truncated", func(t *testing.T) {
		a := makeAlert("TestAlert", "ns", "abcdef12", "warning")
		a.Annotations["summary"] = strings.Repeat("x", 300)

		p, err := Build(a, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := len(p.Annotations[annotSummary]); got != maxSummaryLen {
			t.Errorf("summary length = %d, want %d", got, maxSummaryLen)
		}
	})
}

func TestBuildRequest(t *testing.T) {
	a := makeAlert("KubePodCrashLooping", "production", "abcdef12", "critical")
	a.Annotations["runbook_url"] = "https://runbooks.example.com/KubePodCrashLooping"

	p, err := Build(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{
		"Alert: KubePodCrashLooping (critical)",
		"Namespace: production",
		"Runbook URL: https://runbooks.example.com/KubePodCrashLooping",
		"Description: Pod my-pod has restarted 5 times in the last hour",
		"alertname: KubePodCrashLooping",
	} {
		if !strings.Contains(p.Spec.Request, want) {
			t.Errorf("request does not contain %q\nfull request:\n%s", want, p.Spec.Request)
		}
	}
}

func TestBuildRequestWithoutRunbook(t *testing.T) {
	a := makeAlert("TestAlert", "ns", "abcdef12", "warning")
	p, err := Build(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(p.Spec.Request, "Runbook URL:") {
		t.Error("request should not contain Runbook URL when not set")
	}
}

func TestBuildName(t *testing.T) {
	tests := []struct {
		name        string
		alertName   string
		namespace   string
		fingerprint string
		expected    string
	}{
		{
			name:        "standard namespaced",
			alertName:   "KubePodCrashLooping",
			namespace:   "production",
			fingerprint: "abcdef1234567890",
			expected:    "kubepodcrashlooping-production-abcdef12",
		},
		{
			name:        "cluster-scoped",
			alertName:   "ClusterReady",
			namespace:   "",
			fingerprint: "ff00ff00ff00ff00",
			expected:    "clusterready-ff00ff00",
		},
		{
			name:        "invalid characters replaced",
			alertName:   "Alert_With:Special/Chars",
			namespace:   "my ns!",
			fingerprint: "aabbccdd",
			expected:    "alert-with-special-chars-my-ns--aabbccdd",
		},
		{
			name:        "long alert name truncated with namespace",
			alertName:   strings.Repeat("a", 250),
			namespace:   "ns",
			fingerprint: "12345678",
			expected:    strings.Repeat("a", 253-len("-ns-12345678")) + "-ns-12345678",
		},
		{
			name:        "long alert name truncated without namespace",
			alertName:   strings.Repeat("a", 250),
			namespace:   "",
			fingerprint: "12345678",
			expected:    strings.Repeat("a", 253-len("-12345678")) + "-12345678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildName(tt.alertName, tt.namespace, tt.fingerprint)
			if got != tt.expected {
				t.Errorf("buildName() = %q, want %q", got, tt.expected)
			}
			if len(got) > maxNameLen {
				t.Errorf("name length %d exceeds max %d", len(got), maxNameLen)
			}
		})
	}
}

func TestSanitizeLabelValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid value unchanged",
			input:    "my-valid-label",
			expected: "my-valid-label",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "special characters replaced",
			input:    "alert/name:with@chars",
			expected: "alert-name-with-chars",
		},
		{
			name:     "truncated to 63 chars",
			input:    strings.Repeat("x", 70),
			expected: strings.Repeat("x", 63),
		},
		{
			name:     "leading non-alphanumeric trimmed",
			input:    "---value",
			expected: "value",
		},
		{
			name:     "trailing non-alphanumeric trimmed",
			input:    "value---",
			expected: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeLabelValue(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeLabelValue(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestBuildTypeMeta(t *testing.T) {
	a := makeAlert("TestAlert", "ns", "abcdef12", "warning")
	p, err := Build(a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.APIVersion != "agentic.openshift.io/v1alpha1" {
		t.Errorf("apiVersion = %q, want %q", p.APIVersion, "agentic.openshift.io/v1alpha1")
	}
	if p.Kind != "Proposal" {
		t.Errorf("kind = %q, want %q", p.Kind, "Proposal")
	}
}

func TestBuildWithSkills(t *testing.T) {
	a := makeAlert("TestAlert", "ns", "abcdef12", "warning")

	t.Run("nil skills omits tools", func(t *testing.T) {
		p, err := Build(a, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !p.Spec.Tools.IsZero() {
			t.Errorf("expected zero tools, got %+v", p.Spec.Tools)
		}
	})

	t.Run("empty skills slice omits tools", func(t *testing.T) {
		p, err := Build(a, []agenticv1alpha1.SkillsSource{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !p.Spec.Tools.IsZero() {
			t.Errorf("expected zero tools, got %+v", p.Spec.Tools)
		}
	})

	t.Run("single skill sets tools", func(t *testing.T) {
		skills := []agenticv1alpha1.SkillsSource{
			{Image: "registry.example.com/skills:latest", Paths: []string{"/skills/prometheus"}},
		}
		p, err := Build(a, skills)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(p.Spec.Tools.Skills) != 1 {
			t.Fatalf("tools.skills length = %d, want 1", len(p.Spec.Tools.Skills))
		}
		if p.Spec.Tools.Skills[0].Image != "registry.example.com/skills:latest" {
			t.Errorf("tools.skills[0].image = %q, want %q", p.Spec.Tools.Skills[0].Image, "registry.example.com/skills:latest")
		}
		if len(p.Spec.Tools.Skills[0].Paths) != 1 || p.Spec.Tools.Skills[0].Paths[0] != "/skills/prometheus" {
			t.Errorf("tools.skills[0].paths = %v, want [/skills/prometheus]", p.Spec.Tools.Skills[0].Paths)
		}
	})

	t.Run("multiple skills sets tools", func(t *testing.T) {
		skills := []agenticv1alpha1.SkillsSource{
			{Image: "registry.example.com/skills:latest", Paths: []string{"/skills/prometheus"}},
			{Image: "registry.example.com/acs-skills:latest", Paths: []string{"/skills/acs", "/skills/cve"}},
		}
		p, err := Build(a, skills)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(p.Spec.Tools.Skills) != 2 {
			t.Fatalf("tools.skills length = %d, want 2", len(p.Spec.Tools.Skills))
		}
		if p.Spec.Tools.Skills[1].Image != "registry.example.com/acs-skills:latest" {
			t.Errorf("tools.skills[1].image = %q, want %q", p.Spec.Tools.Skills[1].Image, "registry.example.com/acs-skills:latest")
		}
		if len(p.Spec.Tools.Skills[1].Paths) != 2 {
			t.Errorf("tools.skills[1].paths length = %d, want 2", len(p.Spec.Tools.Skills[1].Paths))
		}
	})
}
