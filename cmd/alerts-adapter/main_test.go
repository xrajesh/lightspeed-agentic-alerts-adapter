package main

import (
	"io"
	"log/slog"
	"testing"
)

func TestRun(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	if err := run(logger); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
