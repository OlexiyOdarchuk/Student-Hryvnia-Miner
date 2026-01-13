package main

import "sync"

var (
	logsBuffer []LogEntry
	logsMutex  sync.Mutex
	lastLogID  int64
)
