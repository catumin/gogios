package helpers

import (
	"log"
	"os"
)

var (
	// Log holds the needed information for the logging service
	Log *log.Logger
)

func init() {
	logPath := "/var/log/gogios/service_check.log"

	file, err := os.Create(logPath)
	if err != nil {
		panic(err)
	}

	Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
	Log.Println("LogFile : " + logPath)
}
