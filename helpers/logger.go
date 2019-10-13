package helpers

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	Log *log.Logger
)

func init() {
	logPath := "/var/log/gingertechnology/service_check.log"

	err := Copy(logPath, logPath+time.Now().Format("20060102150405"))
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
