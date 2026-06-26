package config

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	agenticv1alpha1 "github.com/openshift/lightspeed-agentic-operator/api/v1alpha1"
)

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	return path
}

func TestLoadFromFile(t *testing.T) {
	tests := []struct {
		name               string
		yaml               string
		wantPollInterval   time.Duration
		wantInitialDelay   time.Duration
		wantCooldownWindow time.Duration
	}{
		{
			name:               "valid full config",
			yaml:               "pollInterval: 45s\ninitialDelay: 10m\ncooldownWindow: 30m\n",
			wantPollInterval:   45 * time.Second,
			wantInitialDelay:   10 * time.Minute,
			wantCooldownWindow: 30 * time.Minute,
		},
		{
			name:               "partial config - only poll interval",
			yaml:               "pollInterval: 1m\n",
			wantPollInterval:   time.Minute,
			wantInitialDelay:   DefaultInitialDelay,
			wantCooldownWindow: DefaultCooldownWindow,
		},
		{
			name:               "partial config - only cooldown window",
			yaml:               "cooldownWindow: 2h\n",
			wantPollInterval:   DefaultPollInterval,
			wantInitialDelay:   DefaultInitialDelay,
			wantCooldownWindow: 2 * time.Hour,
		},
		{
			name:               "empty yaml",
			yaml:               "",
			wantPollInterval:   DefaultPollInterval,
			wantInitialDelay:   DefaultInitialDelay,
			wantCooldownWindow: DefaultCooldownWindow,
		},
		{
			name:               "empty document",
			yaml:               "---\n",
			wantPollInterval:   DefaultPollInterval,
			wantInitialDelay:   DefaultInitialDelay,
			wantCooldownWindow: DefaultCooldownWindow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfigFile(t, tt.yaml)

			cfg := LoadFromFile(path, quietLogger())

			if cfg.PollInterval != tt.wantPollInterval {
				t.Errorf("PollInterval = %v, want %v", cfg.PollInterval, tt.wantPollInterval)
			}
			if cfg.InitialDelay != tt.wantInitialDelay {
				t.Errorf("InitialDelay = %v, want %v", cfg.InitialDelay, tt.wantInitialDelay)
			}
			if cfg.CooldownWindow != tt.wantCooldownWindow {
				t.Errorf("CooldownWindow = %v, want %v", cfg.CooldownWindow, tt.wantCooldownWindow)
			}
		})
	}
}

func TestLoadFromFileInvalid(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "invalid yaml",
			yaml: ":::not yaml:::",
		},
		{
			name: "invalid duration value",
			yaml: "pollInterval: not-a-duration\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfigFile(t, tt.yaml)

			cfg := LoadFromFile(path, quietLogger())

			assertDefaults(t, cfg)
		})
	}
}

func TestNonPositiveDurationsFallBackToDefaults(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "zero poll interval",
			yaml: "pollInterval: 0s",
		},
		{
			name: "negative initial delay",
			yaml: "initialDelay: -5m",
		},
		{
			name: "negative cooldown window",
			yaml: "cooldownWindow: -1h",
		},
		{
			name: "all zero",
			yaml: "pollInterval: 0s\ninitialDelay: 0s\ncooldownWindow: 0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfigFile(t, tt.yaml)

			cfg := LoadFromFile(path, quietLogger())

			assertDefaults(t, cfg)
		})
	}
}

func TestLoadFromFileMissing(t *testing.T) {
	cfg := LoadFromFile("/nonexistent/path/config.yaml", quietLogger())

	assertDefaults(t, cfg)
}

func assertDefaults(t *testing.T, cfg Config) {
	t.Helper()
	defaults := Default()
	if cfg.PollInterval != defaults.PollInterval {
		t.Errorf("PollInterval = %v, want %v", cfg.PollInterval, defaults.PollInterval)
	}
	if cfg.InitialDelay != defaults.InitialDelay {
		t.Errorf("InitialDelay = %v, want %v", cfg.InitialDelay, defaults.InitialDelay)
	}
	if cfg.CooldownWindow != defaults.CooldownWindow {
		t.Errorf("CooldownWindow = %v, want %v", cfg.CooldownWindow, defaults.CooldownWindow)
	}
	assertReceiversEqual(t, cfg.AllowedReceivers, defaults.AllowedReceivers)
	assertEmptyTools(t, cfg.Tools)
}

func assertEmptyTools(t *testing.T, tc ToolsConfig) {
	t.Helper()
	if len(tc.Shared) != 0 {
		t.Errorf("Tools.Shared = %v, want empty", tc.Shared)
	}
	if len(tc.Analysis) != 0 {
		t.Errorf("Tools.Analysis = %v, want empty", tc.Analysis)
	}
	if len(tc.Execution) != 0 {
		t.Errorf("Tools.Execution = %v, want empty", tc.Execution)
	}
	if len(tc.Verification) != 0 {
		t.Errorf("Tools.Verification = %v, want empty", tc.Verification)
	}
}

func TestParseToolsConfig(t *testing.T) {
	tests := []struct {
		name      string
		yaml      string
		wantTools ToolsConfig
	}{
		{
			name: "shared skills only",
			yaml: `
tools:
  skills:
    - image: registry.example.com/skills:latest
      paths:
        - /skills/prometheus
`,
			wantTools: ToolsConfig{
				Shared: []agenticv1alpha1.SkillsSource{
					{Image: "registry.example.com/skills:latest", Paths: []string{"/skills/prometheus"}},
				},
			},
		},
		{
			name: "per-step skills only",
			yaml: `
analysis:
  tools:
    skills:
      - image: registry.example.com/analysis:latest
        paths:
          - /skills/diagnostic
execution:
  tools:
    skills:
      - image: registry.example.com/exec:latest
        paths:
          - /skills/remediation
verification:
  tools:
    skills:
      - image: registry.example.com/verify:latest
        paths:
          - /skills/validation
`,
			wantTools: ToolsConfig{
				Analysis: []agenticv1alpha1.SkillsSource{
					{Image: "registry.example.com/analysis:latest", Paths: []string{"/skills/diagnostic"}},
				},
				Execution: []agenticv1alpha1.SkillsSource{
					{Image: "registry.example.com/exec:latest", Paths: []string{"/skills/remediation"}},
				},
				Verification: []agenticv1alpha1.SkillsSource{
					{Image: "registry.example.com/verify:latest", Paths: []string{"/skills/validation"}},
				},
			},
		},
		{
			name: "shared and per-step skills combined",
			yaml: `
tools:
  skills:
    - image: registry.example.com/shared:latest
      paths:
        - /skills/common
analysis:
  tools:
    skills:
      - image: registry.example.com/analysis:latest
        paths:
          - /skills/diagnostic
`,
			wantTools: ToolsConfig{
				Shared: []agenticv1alpha1.SkillsSource{
					{Image: "registry.example.com/shared:latest", Paths: []string{"/skills/common"}},
				},
				Analysis: []agenticv1alpha1.SkillsSource{
					{Image: "registry.example.com/analysis:latest", Paths: []string{"/skills/diagnostic"}},
				},
			},
		},
		{
			name: "multiple skills in shared",
			yaml: `
tools:
  skills:
    - image: registry.example.com/skills:latest
      paths:
        - /skills/prometheus
        - /skills/cluster-diagnostics
    - image: registry.example.com/acs:latest
      paths:
        - /skills/acs
`,
			wantTools: ToolsConfig{
				Shared: []agenticv1alpha1.SkillsSource{
					{Image: "registry.example.com/skills:latest", Paths: []string{"/skills/prometheus", "/skills/cluster-diagnostics"}},
					{Image: "registry.example.com/acs:latest", Paths: []string{"/skills/acs"}},
				},
			},
		},
		{
			name:      "no tools key",
			yaml:      "pollInterval: 30s",
			wantTools: ToolsConfig{},
		},
		{
			name: "empty skills list",
			yaml: `
tools:
  skills: []
`,
			wantTools: ToolsConfig{},
		},
		{
			name: "shared skills entry with empty image skipped",
			yaml: `
tools:
  skills:
    - image: ""
      paths:
        - /skills/prometheus
`,
			wantTools: ToolsConfig{},
		},
		{
			name: "per-step skills entry with empty paths skipped",
			yaml: `
analysis:
  tools:
    skills:
      - image: registry.example.com/skills:latest
        paths: []
`,
			wantTools: ToolsConfig{},
		},
		{
			name: "mix of valid and invalid entries across levels",
			yaml: `
tools:
  skills:
    - image: ""
      paths:
        - /skills/bad
    - image: registry.example.com/good:latest
      paths:
        - /skills/good
execution:
  tools:
    skills:
      - image: registry.example.com/no-paths:latest
        paths: []
      - image: registry.example.com/exec:latest
        paths:
          - /skills/remediation
`,
			wantTools: ToolsConfig{
				Shared: []agenticv1alpha1.SkillsSource{
					{Image: "registry.example.com/good:latest", Paths: []string{"/skills/good"}},
				},
				Execution: []agenticv1alpha1.SkillsSource{
					{Image: "registry.example.com/exec:latest", Paths: []string{"/skills/remediation"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfigFile(t, tt.yaml)

			cfg := LoadFromFile(path, quietLogger())

			assertSkillsEqual(t, "Tools.Shared", cfg.Tools.Shared, tt.wantTools.Shared)
			assertSkillsEqual(t, "Tools.Analysis", cfg.Tools.Analysis, tt.wantTools.Analysis)
			assertSkillsEqual(t, "Tools.Execution", cfg.Tools.Execution, tt.wantTools.Execution)
			assertSkillsEqual(t, "Tools.Verification", cfg.Tools.Verification, tt.wantTools.Verification)
		})
	}
}

func TestParseAllowedReceivers(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want []string
	}{
		{
			name: "single receiver",
			yaml: "allowedReceivers:\n  - Critical\n",
			want: []string{"critical"},
		},
		{
			name: "multiple receivers",
			yaml: "allowedReceivers:\n  - Critical\n  - PagerDuty\n",
			want: []string{"critical", "pagerduty"},
		},
		{
			name: "mixed case normalized to lowercase",
			yaml: "allowedReceivers:\n  - CRITICAL\n  - Slack-OnCall\n",
			want: []string{"critical", "slack-oncall"},
		},
		{
			name: "field absent defaults to empty",
			yaml: "pollInterval: 30s\n",
			want: nil,
		},
		{
			name: "explicit empty list",
			yaml: "allowedReceivers: []\n",
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfigFile(t, tt.yaml)

			cfg := LoadFromFile(path, quietLogger())

			assertReceiversEqual(t, cfg.AllowedReceivers, tt.want)
		})
	}
}

func TestAllowedReceiversDefaultsOnMissingFile(t *testing.T) {
	cfg := LoadFromFile("/nonexistent/path/config.yaml", quietLogger())

	assertReceiversEqual(t, cfg.AllowedReceivers, nil)
}

func assertReceiversEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("AllowedReceivers length = %d, want %d (got %v)", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("AllowedReceivers[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func assertSkillsEqual(t *testing.T, field string, got, want []agenticv1alpha1.SkillsSource) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d", field, len(got), len(want))
	}
	for i, w := range want {
		if got[i].Image != w.Image {
			t.Errorf("%s[%d].image = %q, want %q", field, i, got[i].Image, w.Image)
		}
		if len(got[i].Paths) != len(w.Paths) {
			t.Fatalf("%s[%d].paths length = %d, want %d", field, i, len(got[i].Paths), len(w.Paths))
		}
		for j, p := range w.Paths {
			if got[i].Paths[j] != p {
				t.Errorf("%s[%d].paths[%d] = %q, want %q", field, i, j, got[i].Paths[j], p)
			}
		}
	}
}
