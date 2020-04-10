package helpers

import (
	"os"

	"github.com/google/logger"
)

// StartLogger creates a logger named $name in /var/log/gogios/$filename.log
func StartLogger(name, filename string, verbose bool) *logger.Logger {
	logPath := "/var/log/gogios/" + filename + ".log"

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	goglog := logger.Init(name, verbose, true, file)

	return goglog
}
