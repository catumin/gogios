package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	inittime = time.Now()
	password = flag.String("password", "password123", "Password to use")
	ip       = flag.String("ip", "123.123.123.123", "IP address connect to")
	port     = flag.Int("port", 3306, "Port of server")
	user     = flag.String("user", "root", "User to login as")
	database = flag.String("database", "", "Database to attempt connection to")
	attempts = flag.Int("attempts", 3, "Amount of times to attempt login")
	timer    = flag.Duration("timer", 300*time.Millisecond, "Timeout between attempts")
)

type resp struct {
	Error error
	mu    sync.Mutex
}

// MySQLDialer - Attempt to make a connection with an MySQL server
func MySQLDialer() *resp {
	exitcode := &resp{}
	var err error
	var db *sql.DB

	// Attempt MySQL connection
	if db, err = sql.Open("mysql", *user+":"+*password+"@tcp("+*ip+":"+strconv.Itoa(*port)+")/"+*database); err == nil {
		end := time.Now()
		d := end.Sub(inittime)
		duration := d.Seconds()
		fmt.Printf("\nCompleted in %v seconds\n", strconv.FormatFloat(duration, 'g', -1, 64))

		defer db.Close()

		fmt.Println(db.Stats())
	} else {
		fmt.Println(err)
		defer db.Close()
	}

	exitcode.Error = err
	return exitcode
}

func main() {
	flag.Parse()

	log.Panicln("This one does not currently fail correctly")

	for attempt := *attempts; attempt != 0; attempt-- {
		go func() {
			resp := MySQLDialer()
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
