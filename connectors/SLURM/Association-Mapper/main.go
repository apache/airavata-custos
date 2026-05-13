// Package main is the entry point for the SLURM Association-Mapper connector.
//
// It consumes allocation events released by the allocation manager and
// materializes them as SLURM associations via slurmrestd.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("association-mapper started")
	<-ctx.Done()
	logger.Info("association-mapper stopped")
}
