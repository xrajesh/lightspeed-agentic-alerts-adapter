package main

import (
	"io"
	"log/slog"
	"testing"
)

func TestRunFailsOutsideCluster(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	err := run(logger)
	if err == nil {
		t.Fatal("expected error when running outside cluster, got nil")
	}
}
