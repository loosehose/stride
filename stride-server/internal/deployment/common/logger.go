package common

import (
	"encoding/json"
	"sync"
	"time"
)

// Define log levels
const (
	Exec    = "EXEC"
	Info    = "INFO"
	Success = "SUCCESS"
	Warning = "WARNING"
	Error   = "ERROR"
)

type LogMessage struct {
	Level     string    `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	ToastID   string    `json:"toastId"`
}

// LogStorage stores log messages.
type LogStorage struct {
	messages []LogMessage
	lock     sync.RWMutex
}

var logStorage = LogStorage{
	messages: make([]LogMessage, 0),
}

var recentLogs []LogMessage
var lock sync.Mutex

// NewLogMessage adds a new log message to the storage and broadcasts it to all connected clients.
func NewLogMessage(level, message, toastId string, wsm *WebSocketManager) {
	logMessage := LogMessage{
		Level:     level,
		Timestamp: time.Now(),
		Message:   message,
		ToastID:   toastId,
	}

	logStorage.lock.Lock()
	defer logStorage.lock.Unlock()
	logStorage.messages = append(logStorage.messages, logMessage)

	lock.Lock()
	defer lock.Unlock()
	// Assume we keep the latest 100 logs
	if len(recentLogs) >= 100 {
		recentLogs = recentLogs[1:]
	}
	recentLogs = append(recentLogs, logMessage)

	// Broadcast the log message to all connected clients
	messageBytes, _ := json.Marshal(logMessage)
	wsm.BroadcastMessage(messageBytes)
}

// GetRecentLogs returns the recent logs.
func GetRecentLogs() []LogMessage {
	lock.Lock()
	defer lock.Unlock()
	// Return a copy of the logs to avoid race conditions
	logsCopy := make([]LogMessage, len(recentLogs))
	copy(logsCopy, recentLogs)
	return logsCopy
}