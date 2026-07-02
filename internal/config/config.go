package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	agenticv1alpha1 "github.com/openshift/lightspeed-agentic-operator/api/v1alpha1"
	"go.yaml.in/yaml/v3"
)

const (
	DefaultPollInterval   = 30 * time.Second
	DefaultInitialDelay   = 5 * time.Minute
	DefaultCooldownWindow = 1 * time.Hour

	DefaultConfigPath = "/etc/alerts-adapter/config.yaml"
)

// AgentConfig holds the agent name overrides for Proposal workflow steps.
type AgentConfig struct {
	Default      string
	Analysis     string
	Execution    string
	Verification string
}

// ToolsConfig holds shared and per-step skills configuration.
type ToolsConfig struct {
	Shared       []agenticv1alpha1.SkillsSource
	Analysis     []agenticv1alpha1.SkillsSource
	Execution    []agenticv1alpha1.SkillsSource
	Verification []agenticv1alpha1.SkillsSource
}

// Config holds the adapter's runtime-tunable parameters.
type Config struct {
	PollInterval     time.Duration
	InitialDelay     time.Duration
	CooldownWindow   time.Duration
	AllowedReceivers []string
	Tools            ToolsConfig
	Agent            AgentConfig
}

// Default returns a Config with the default values.
func Default() Config {
	return Config{
		PollInterval:     DefaultPollInterval,
		InitialDelay:     DefaultInitialDelay,
		CooldownWindow:   DefaultCooldownWindow,
		AllowedReceivers: nil,
	}
}

type configFile struct {
	PollInterval     Duration         `yaml:"pollInterval"`
	InitialDelay     Duration         `yaml:"initialDelay"`
	CooldownWindow   Duration         `yaml:"cooldownWindow"`
	AllowedReceivers *[]string        `yaml:"allowedReceivers"`
	Tools            toolsEntry       `yaml:"tools"`
	Analysis         stepEntry        `yaml:"analysis"`
	Execution        stepEntry        `yaml:"execution"`
	Verification     stepEntry        `yaml:"verification"`
	Agent            agentEntry       `yaml:"agent"`
}

type agentEntry struct {
	Default      string `yaml:"default"`
	Analysis     string `yaml:"analysis"`
	Execution    string `yaml:"execution"`
	Verification string `yaml:"verification"`
}

type toolsEntry struct {
	Skills []skillsEntry `yaml:"skills"`
}

type stepEntry struct {
	Tools toolsEntry `yaml:"tools"`
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

// LoadFromFile reads configuration from a YAML file at the given path.
// It never returns an error — on any failure it falls back to defaults and logs.
func LoadFromFile(path string, logger *slog.Logger) Config {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		logger.Error("failed to read config file, using defaults", "path", path, "error", err)
		return cfg
	}

	var cf configFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		logger.Error("failed to parse config file, using defaults", "path", path, "error", err)
		return cfg
	}

	if cf.PollInterval.isSet {
		if cf.PollInterval.Duration > 0 {
			cfg.PollInterval = cf.PollInterval.Duration
		} else {
			logger.Error("pollInterval must be positive, using default", "value", cf.PollInterval.Duration)
		}
	}
	if cf.InitialDelay.isSet {
		if cf.InitialDelay.Duration > 0 {
			cfg.InitialDelay = cf.InitialDelay.Duration
		} else {
			logger.Error("initialDelay must be positive, using default", "value", cf.InitialDelay.Duration)
		}
	}
	if cf.CooldownWindow.isSet {
		if cf.CooldownWindow.Duration > 0 {
			cfg.CooldownWindow = cf.CooldownWindow.Duration
		} else {
			logger.Error("cooldownWindow must be positive, using default", "value", cf.CooldownWindow.Duration)
		}
	}

	if cf.AllowedReceivers != nil {
		normalized := make([]string, 0, len(*cf.AllowedReceivers))
		for _, r := range *cf.AllowedReceivers {
			normalized = append(normalized, strings.ToLower(r))
		}
		cfg.AllowedReceivers = normalized
	}

	cfg.Tools = ToolsConfig{
		Shared:       parseSkills(cf.Tools.Skills, "shared", logger),
		Analysis:     parseSkills(cf.Analysis.Tools.Skills, "analysis", logger),
		Execution:    parseSkills(cf.Execution.Tools.Skills, "execution", logger),
		Verification: parseSkills(cf.Verification.Tools.Skills, "verification", logger),
	}

	cfg.Agent = AgentConfig{
		Default:      cf.Agent.Default,
		Analysis:     cf.Agent.Analysis,
		Execution:    cf.Agent.Execution,
		Verification: cf.Agent.Verification,
	}

	return cfg
}

func parseSkills(entries []skillsEntry, step string, logger *slog.Logger) []agenticv1alpha1.SkillsSource {
	var skills []agenticv1alpha1.SkillsSource
	for _, e := range entries {
		if e.Image == "" {
			logger.Warn("skills entry missing image, skipping", "step", step)
			continue
		}
		if len(e.Paths) == 0 {
			logger.Warn("skills entry missing paths, skipping", "step", step)
			continue
		}
		skills = append(skills, agenticv1alpha1.SkillsSource{
			Image: e.Image,
			Paths: e.Paths,
		})
	}
	return skills
}
