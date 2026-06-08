package config

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

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

			defaults := Default()
			if cfg != defaults {
				t.Errorf("expected defaults on invalid input, got %+v", cfg)
			}
		})
	}
}

func TestLoadConfigMapNotFound(t *testing.T) {
	src := newTestSource(t)

	cfg := src.Load(context.Background())

	defaults := Default()
	if cfg != defaults {
		t.Errorf("expected defaults when ConfigMap not found, got %+v", cfg)
	}
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

	defaults := Default()
	if cfg != defaults {
		t.Errorf("expected defaults when config.yaml key missing, got %+v", cfg)
	}
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

	defaults := Default()
	if cfg != defaults {
		t.Errorf("expected defaults when ConfigMap has no data, got %+v", cfg)
	}
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
