package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/sparrc/go-ping"
)

var (
	host      = flag.String("host", "123.123.123.123", "IP address or FQDN to connect to")
	attempts  = flag.Int("attempts", 3, "Amount of times to attempt ping")
	threshold = flag.Int("threshold", 3, "Amount of pings that need to be successful to be considered alive")
)

// SendPing attempts to ping a host $count times
func SendPing(target string, threshold, count int) (bool, error) {
	alive := false
	pinger, err := ping.NewPinger(*host)
	if err != nil {
		fmt.Println("If you receive a privilege error, try either allowing unprivileged UDP ping with:")
		fmt.Println(`sudo sysctl -w net.ipv4.ping_group_range="0   2147483647"	`)
		fmt.Println("Or by allowing just this binary to bind to raw sockets with:")
		fmt.Println(`setcap cap_net_raw=+ep /usr/lib/gogios/plugins/check-host-alive`)
		fmt.Println("Or ignore this problem and do a counted ping. It won't give stats but it will work:")
		fmt.Println(`ping -c 3 $target`)
		panic(err)
	}

	pinger.Count = count
	pinger.Run()
	stats := pinger.Statistics()

	if stats.PacketsRecv >= threshold {
		alive = true
		fmt.Fprintf(color.Output, "\n%s", color.YellowString("###########################"))
		fmt.Fprintf(color.Output, "\n%s", color.GreenString("Successful connection"))
		fmt.Fprintf(color.Output, "\n%s\n", color.YellowString("###########################"))
		fmt.Printf("Packets sent: %d\n", stats.PacketsSent)
		fmt.Printf("Packets received: %d\n", stats.PacketsRecv)
		fmt.Printf("Average rate: %v\n", stats.AvgRtt)
	}

	return alive, err
}

func main() {
	flag.Parse()

	log.Printf("Starting %s\n", time.Now())
	_, err := SendPing(*host, *threshold, *attempts)
	if err != nil {
		fmt.Println("If you receive a privilege error, try either allowing unprivileged UDP ping with:")
		fmt.Println(`sudo sysctl -w net.ipv4.ping_group_range="0   2147483647"	`)
		fmt.Println("Or by allowing just this binary to bind to raw sockets with:")
		fmt.Println(`setcap cap_net_raw=+ep /usr/lib/gogios/plugins/check-host-alive`)
		panic(err)
	}
}
