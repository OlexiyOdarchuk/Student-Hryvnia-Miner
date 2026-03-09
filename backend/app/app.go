package app

import (
	"context"
	"log/slog"
	"os"
	"shminer/backend/internal/logger"
	"shminer/backend/types"
)

func StartApp(ctx context.Context, logCallback func(entry types.LogEntry)) {
	uiLogger := slog.New(&logger.UIHandler{LogCallback: logCallback, Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})})
	slog.SetDefault(uiLogger)
}
