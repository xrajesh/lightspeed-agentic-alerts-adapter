package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/openshift/lightspeed-agentic-alerts-adapter/internal/alertmanager"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	logger.Info("lightspeed-agentic-alerts-adapter starting")

	cfg := alertmanager.Config{
		URL: os.Getenv("ALERTMANAGER_URL"),
	}

	client, err := alertmanager.New(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	alerts, err := client.GetAlerts(ctx)
	if err != nil {
		return err
	}

	logger.Info("alerts retrieved", "count", len(alerts))
	for _, a := range alerts {
		state := "unknown"
		if a.Status != nil && a.Status.State != nil {
			state = *a.Status.State
		}
		logger.Info("alert",
			"name", a.Labels["alertname"],
			"severity", a.Labels["severity"],
			"state", state,
			"startsAt", a.StartsAt,
		)
	}

	return nil
}
