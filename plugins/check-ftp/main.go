package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/dutchcoders/goftp"
	"github.com/fatih/color"
)

var (
	inittime = time.Now()
	host     = flag.String("host", "123.123.123.123", "IP address or FQDN to connect to")
	port     = flag.Int("port", 21, "Port of server")
	attempts = flag.Int("attempts", 3, "Amount of times to attempt login")
	timer    = flag.Duration("timer", 300*time.Millisecond, "Timeout between attempts")
)

type resp struct {
	Error error
	mu    sync.Mutex
}

// FTPDialer - Attempt to make a connection with an FTP server
func FTPDialer() *resp {
	exitcode := &resp{}
	var err error
	var ftp *goftp.FTP

	// Attempt FTP connection
	if ftp, err = goftp.Connect(*host + ":" + strconv.Itoa(*port)); err == nil {
		end := time.Now()
		d := end.Sub(inittime)
		duration := d.Seconds()
		fmt.Fprintf(color.Output, "\n%s", color.YellowString("###########################"))
		fmt.Fprintf(color.Output, "\n%s", color.GreenString("Successful connection"))
		fmt.Fprintf(color.Output, "\n%s", color.YellowString("###########################"))
		fmt.Printf("\nCompleted in %v seconds\n", strconv.FormatFloat(duration, 'g', -1, 64))
	} else {
		fmt.Println(err)
	}
	defer ftp.Close()

	exitcode.Error = err
	return exitcode
}

func main() {
	flag.Parse()

	for attempt := *attempts; attempt != 0; attempt-- {
		go func() {
			resp := FTPDialer()
			resp.mu.Lock()
			if resp.Error == nil {
				resp.mu.Unlock()
				os.Exit(0)
			}
		}()

		fmt.Println("Attempt: ", attempt)
		time.Sleep(*timer)
	}
}
