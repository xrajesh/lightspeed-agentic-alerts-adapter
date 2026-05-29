package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/openshift/lightspeed-agentic-alerts-adapter/internal/adapter"
	"github.com/openshift/lightspeed-agentic-alerts-adapter/internal/alertmanager"
	"github.com/openshift/lightspeed-agentic-alerts-adapter/internal/proposal"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	amClient, err := alertmanager.New(alertmanager.Config{
		URL: os.Getenv("ALERTMANAGER_URL"),
	})
	if err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}

	propClient, err := proposal.NewClient(logger)
	if err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}

	a := adapter.New(amClient, propClient, logger)
	if err := a.Run(ctx); err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}
}
