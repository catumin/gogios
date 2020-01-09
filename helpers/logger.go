package helpers

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	// Log holds the needed information for the logging service
	Log *log.Logger
)

func init() {
	logPath := "/var/log/gogios/service_check.log"

	err := Copy(logPath, logPath+time.Now().Format(time.RFC822))
	if err != nil {
		fmt.Println(err.Error())
	}

	file, err := os.Create(logPath)
	if err != nil {
		panic(err)
	}

	Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
	Log.Println("LogFile : " + logPath)
}
