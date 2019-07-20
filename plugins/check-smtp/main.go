package main

import (
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/fatih/color"
)

var (
	inittime = time.Now()
	user     = flag.String("user", "root", "User to login as")
	password = flag.String("password", "password123", "Password to use")
	host     = flag.String("host", "123.123.123.123", "IP address or FQDN to connect to")
	port     = flag.Int("port", 25, "Port of server")
	attempts = flag.Int("attempts", 3, "Amount of times to attempt login")
	timer    = flag.Duration("timer", 300*time.Millisecond, "Timeout between attempts")
)

type smtpServer struct {
	host string
	port string
}

func (s *smtpServer) ServerName() string {
	return s.host + ":" + s.port
}

// SMTPDialer - Attempt to authenticate with an SMTP server to confirm it is accessible
func SMTPDialer() (err error) {
	smtpServer := smtpServer{host: *host, port: strconv.Itoa(*port)}
	auth := smtp.PlainAuth("", *user, *password, smtpServer.host)

	client, err := smtp.Dial(smtpServer.host)
	if err != nil {
		log.Panic(err)
	}

	if err = client.Auth(auth); err == nil {
		end := time.Now()
		d := end.Sub(inittime)
		duration := d.Seconds()
		fmt.Fprintf(color.Output, "\n%s", color.YellowString("###########################"))
		fmt.Fprintf(color.Output, "\n%s", color.GreenString("Successful connection"))
		fmt.Fprintf(color.Output, "\n%s", color.YellowString("###########################"))
		fmt.Printf("\nCompleted in %v seconds\n", strconv.FormatFloat(duration, 'g', -1, 64))
		defer client.Quit()
	} else {
		fmt.Println(err)
		defer client.Quit()
	}

	return err
}

func main() {
	flag.Parse()

	for attempt := *attempts; attempt != 0; attempt-- {
		go func() {
			resp := SMTPDialer()
			if resp == nil {
				os.Exit(0)
			}
		}()

		fmt.Println("Attempt: ", attempt)
		time.Sleep(*timer)
	}
}
