package logger

import (
	"context"
	"log/slog"
	"shminer/backend/types"
	"sync/atomic"
)

type UIHandler struct {
	slog.Handler
	LogCallback func(types.LogEntry)
	logSeq      int64
}

func (h *UIHandler) Handle(ctx context.Context, r slog.Record) error {
	entry := types.LogEntry{
		ID:      atomic.AddInt64(&h.logSeq, 1),
		Time:    r.Time.Format("15:04:05"),
		Message: r.Message,
		Type:    r.Level.String(),
	}
	if h.LogCallback != nil {
		h.LogCallback(entry)
	}

	return h.Handler.Handle(ctx, r)
}
