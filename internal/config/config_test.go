package config

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	agenticv1alpha1 "github.com/openshift/lightspeed-agentic-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestParseConfigFile(t *testing.T) {
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
			cm := configMapWith(tt.yaml)
			src := newTestSource(t, cm)

			cfg := src.Load(context.Background())

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

func TestParseConfigFileInvalid(t *testing.T) {
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
			cm := configMapWith(tt.yaml)
			src := newTestSource(t, cm)

			cfg := src.Load(context.Background())

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
			cm := configMapWith(tt.yaml)
			src := newTestSource(t, cm)

			cfg := src.Load(context.Background())

			assertDefaults(t, cfg)
		})
	}
}

func TestLoadConfigMapNotFound(t *testing.T) {
	src := newTestSource(t)

	cfg := src.Load(context.Background())

	assertDefaults(t, cfg)
}

func TestLoadConfigMapMissingKey(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: defaultNamespace,
		},
		Data: map[string]string{
			"other-key": "value",
		},
	}
	src := newTestSource(t, cm)

	cfg := src.Load(context.Background())

	assertDefaults(t, cfg)
}

func TestLoadConfigMapEmptyData(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: defaultNamespace,
		},
	}
	src := newTestSource(t, cm)

	cfg := src.Load(context.Background())

	assertDefaults(t, cfg)
}

func configMapWith(yamlData string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: defaultNamespace,
		},
		Data: map[string]string{
			configMapDataKey: yamlData,
		},
	}
}

func newTestSource(t *testing.T, objs ...runtime.Object) *ConfigMapSource {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add corev1 to scheme: %v", err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...).Build()
	return NewConfigMapSource(c, defaultNamespace, quietLogger())
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
	if len(cfg.Skills) != 0 {
		t.Errorf("Skills = %v, want empty", cfg.Skills)
	}
}

func TestParseSkillsConfig(t *testing.T) {
	tests := []struct {
		name       string
		yaml       string
		wantSkills []agenticv1alpha1.SkillsSource
	}{
		{
			name: "valid single skill",
			yaml: `
skills:
  - image: registry.example.com/skills:latest
    paths:
      - /skills/prometheus
`,
			wantSkills: []agenticv1alpha1.SkillsSource{
				{Image: "registry.example.com/skills:latest", Paths: []string{"/skills/prometheus"}},
			},
		},
		{
			name: "valid multiple skills",
			yaml: `
skills:
  - image: registry.example.com/skills:latest
    paths:
      - /skills/prometheus
      - /skills/cluster-diagnostics
  - image: registry.example.com/acs:latest
    paths:
      - /skills/acs
`,
			wantSkills: []agenticv1alpha1.SkillsSource{
				{Image: "registry.example.com/skills:latest", Paths: []string{"/skills/prometheus", "/skills/cluster-diagnostics"}},
				{Image: "registry.example.com/acs:latest", Paths: []string{"/skills/acs"}},
			},
		},
		{
			name:       "no skills key",
			yaml:       "pollInterval: 30s",
			wantSkills: nil,
		},
		{
			name:       "empty skills list",
			yaml:       "skills: []",
			wantSkills: nil,
		},
		{
			name: "skills entry with empty image skipped",
			yaml: `
skills:
  - image: ""
    paths:
      - /skills/prometheus
`,
			wantSkills: nil,
		},
		{
			name: "skills entry with empty paths skipped",
			yaml: `
skills:
  - image: registry.example.com/skills:latest
    paths: []
`,
			wantSkills: nil,
		},
		{
			name: "mix of valid and invalid entries",
			yaml: `
skills:
  - image: ""
    paths:
      - /skills/bad
  - image: registry.example.com/good:latest
    paths:
      - /skills/good
  - image: registry.example.com/no-paths:latest
    paths: []
`,
			wantSkills: []agenticv1alpha1.SkillsSource{
				{Image: "registry.example.com/good:latest", Paths: []string{"/skills/good"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := configMapWith(tt.yaml)
			src := newTestSource(t, cm)

			cfg := src.Load(context.Background())

			if len(cfg.Skills) != len(tt.wantSkills) {
				t.Fatalf("skills length = %d, want %d", len(cfg.Skills), len(tt.wantSkills))
			}
			for i, want := range tt.wantSkills {
				got := cfg.Skills[i]
				if got.Image != want.Image {
					t.Errorf("skills[%d].image = %q, want %q", i, got.Image, want.Image)
				}
				if len(got.Paths) != len(want.Paths) {
					t.Fatalf("skills[%d].paths length = %d, want %d", i, len(got.Paths), len(want.Paths))
				}
				for j, p := range want.Paths {
					if got.Paths[j] != p {
						t.Errorf("skills[%d].paths[%d] = %q, want %q", i, j, got.Paths[j], p)
					}
				}
			}
		})
	}
}
