package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	agenticv1alpha1 "github.com/openshift/lightspeed-agentic-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"go.yaml.in/yaml/v3"
)

const (
	DefaultPollInterval   = 30 * time.Second
	DefaultInitialDelay   = 5 * time.Minute
	DefaultCooldownWindow = 1 * time.Hour

	configMapName    = "alerts-adapter-config"
	configMapDataKey = "config.yaml"
	defaultNamespace = "openshift-lightspeed"
)

// Config holds the adapter's runtime-tunable parameters.
type Config struct {
	PollInterval   time.Duration
	InitialDelay   time.Duration
	CooldownWindow time.Duration
	Skills         []agenticv1alpha1.SkillsSource
}

// Default returns a Config with the default values.
func Default() Config {
	return Config{
		PollInterval:   DefaultPollInterval,
		InitialDelay:   DefaultInitialDelay,
		CooldownWindow: DefaultCooldownWindow,
	}
}

type configFile struct {
	PollInterval   Duration      `yaml:"pollInterval"`
	InitialDelay   Duration      `yaml:"initialDelay"`
	CooldownWindow Duration      `yaml:"cooldownWindow"`
	Skills         []skillsEntry `yaml:"skills"`
}

type skillsEntry struct {
	Image string   `yaml:"image"`
	Paths []string `yaml:"paths"`
}

// Duration wraps time.Duration for YAML unmarshalling.
type Duration struct {
	time.Duration
	isSet bool
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("config: invalid duration %q: %w", s, err)
	}
	d.Duration = parsed
	d.isSet = true
	return nil
}

// ConfigMapSource reads configuration from a Kubernetes ConfigMap.
type ConfigMapSource struct {
	client    client.Reader
	namespace string
	logger    *slog.Logger
}

// NewConfigMapSource creates a ConfigMapSource that reads the alerts-adapter-config
// ConfigMap from the given namespace. If namespace is empty, it reads POD_NAMESPACE
// from the environment, falling back to openshift-lightspeed.
func NewConfigMapSource(c client.Reader, namespace string, logger *slog.Logger) *ConfigMapSource {
	if namespace == "" {
		namespace = os.Getenv("POD_NAMESPACE")
	}
	if namespace == "" {
		namespace = defaultNamespace
	}
	return &ConfigMapSource{
		client:    c,
		namespace: namespace,
		logger:    logger,
	}
}

// Load reads the ConfigMap and returns the current configuration.
// It never returns an error — on any failure it falls back to defaults.
func (s *ConfigMapSource) Load(ctx context.Context) Config {
	cfg := Default()

	var cm corev1.ConfigMap
	key := types.NamespacedName{Name: configMapName, Namespace: s.namespace}
	if err := s.client.Get(ctx, key, &cm); err != nil {
		if apierrors.IsNotFound(err) {
			s.logger.Info("config ConfigMap not found, using defaults", "name", configMapName, "namespace", s.namespace)
		} else {
			s.logger.Warn("failed to read config ConfigMap, using defaults", "error", err)
		}
		return cfg
	}

	data, ok := cm.Data[configMapDataKey]
	if !ok {
		s.logger.Info("config ConfigMap missing config.yaml key, using defaults", "name", configMapName)
		return cfg
	}

	var cf configFile
	if err := yaml.Unmarshal([]byte(data), &cf); err != nil {
		s.logger.Warn("failed to parse config.yaml, using defaults", "error", err)
		return cfg
	}

	if cf.PollInterval.isSet {
		if cf.PollInterval.Duration > 0 {
			cfg.PollInterval = cf.PollInterval.Duration
		} else {
			s.logger.Warn("pollInterval must be positive, using default", "value", cf.PollInterval.Duration)
		}
	}
	if cf.InitialDelay.isSet {
		if cf.InitialDelay.Duration > 0 {
			cfg.InitialDelay = cf.InitialDelay.Duration
		} else {
			s.logger.Warn("initialDelay must be positive, using default", "value", cf.InitialDelay.Duration)
		}
	}
	if cf.CooldownWindow.isSet {
		if cf.CooldownWindow.Duration > 0 {
			cfg.CooldownWindow = cf.CooldownWindow.Duration
		} else {
			s.logger.Warn("cooldownWindow must be positive, using default", "value", cf.CooldownWindow.Duration)
		}
	}

	cfg.Skills = s.parseSkills(cf.Skills)

	return cfg
}

func (s *ConfigMapSource) parseSkills(entries []skillsEntry) []agenticv1alpha1.SkillsSource {
	var skills []agenticv1alpha1.SkillsSource
	for i, e := range entries {
		if e.Image == "" {
			s.logger.Warn("skills entry has empty image, skipping", "index", i)
			continue
		}
		if len(e.Paths) == 0 {
			s.logger.Warn("skills entry has empty paths, skipping", "index", i, "image", e.Image)
			continue
		}
		skills = append(skills, agenticv1alpha1.SkillsSource{
			Image: e.Image,
			Paths: e.Paths,
		})
	}
	return skills
}
