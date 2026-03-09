package logger

import (
	"log/slog"
	"os"
	"shminer/backend/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandle_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		expectedMsg  string
		expectedType slog.Level
		logFunc      func(msg string, args ...any)
	}{
		{
			name:         "success",
			expectedMsg:  "My Info",
			expectedType: slog.LevelInfo,
			logFunc:      slog.Info,
		},
		{
			name:         "success",
			expectedMsg:  "My Debug",
			expectedType: slog.LevelDebug,
			logFunc:      slog.Debug,
		},
		{
			name:         "success",
			expectedMsg:  "My Warning",
			expectedType: slog.LevelWarn,
			logFunc:      slog.Warn,
		},
		{
			name:         "success",
			expectedMsg:  "My Error",
			expectedType: slog.LevelError,
			logFunc:      slog.Error,
		},
	}
	for _, tt := range tests {
		var logs []types.LogEntry
		logCallback := func(entry types.LogEntry) {
			logs = append(logs, entry)
		}
		uiLogger := slog.New(&UIHandler{LogCallback: logCallback, Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})})
		slog.SetDefault(uiLogger)

		tt.logFunc(tt.expectedMsg)

		assert.Equal(t, 1, len(logs))
		assert.Equal(t, tt.expectedType, logs[0].Type)
		assert.Equal(t, tt.expectedMsg, logs[0].Message)
	}
}
