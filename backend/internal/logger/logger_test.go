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
		expectedType string
		logFunc      func(msg string, args ...any)
		expectLog    bool
	}{
		{
			name:         "Info",
			expectedMsg:  "My Info",
			expectedType: slog.LevelInfo.String(),
			logFunc:      slog.Info,
			expectLog:    true,
		},
		{
			name:         "Debug",
			expectedMsg:  "My Debug",
			expectedType: slog.LevelDebug.String(),
			logFunc:      slog.Debug,
			expectLog:    false,
		},
		{
			name:         "Error",
			expectedMsg:  "My Error",
			expectedType: slog.LevelError.String(),
			logFunc:      slog.Error,
			expectLog:    true,
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

		if tt.expectLog {
			assert.Equal(t, 1, len(logs))
			assert.Equal(t, tt.expectedType, logs[0].Type)
			assert.Equal(t, tt.expectedMsg, logs[0].Message)
		} else {
			assert.Equal(t, 0, len(logs))
		}
	}
}
