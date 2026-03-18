package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/orvice/sandbox-mcp/internal/app"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if err := app.Run(context.Background(), logger); err != nil {
		logger.Error("sandbox mcp server exited", slog.Any("error", err))
		os.Exit(1)
	}
}
