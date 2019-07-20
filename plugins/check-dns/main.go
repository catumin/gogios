package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/miekg/dns"
)

var (
	domain = flag.String("domain", "angrysysadmins.tech", "Domain to lookup")
	server = flag.String("server", "8.8.8.8", "DNS server to use")
	record = flag.String("record", "A", "DNS record to lookup. A, AAAA, NS, MX, TXT.")
	port   = flag.Int("port", 53, "Port to use")
)

// DNSLookup - Check *record from *domain using *server to lookup
func DNSLookup(recordType uint16) {
	c := dns.Client{}
	m := dns.Msg{}

	m.SetQuestion(*domain+".", recordType)
	r, t, err := c.Exchange(&m, *server+":"+strconv.Itoa(*port))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Took %v", t)
	if len(r.Answer) == 0 {
		log.Fatal("No results")
	}

	switch record := *record; record {
	case "A":
		for _, ans := range r.Answer {
			recordAnswer := ans.(*dns.A)
			fmt.Printf("%s\n", recordAnswer.A)
		}
	case "AAAA":
		for _, ans := range r.Answer {
			recordAnswer := ans.(*dns.AAAA)
			fmt.Printf("%s\n", recordAnswer.AAAA)
		}
	case "NS":
		for _, ans := range r.Answer {
			recordAnswer := ans.(*dns.NS)
			fmt.Printf("%s\n", recordAnswer.Ns)
		}
	case "MX":
		for _, ans := range r.Answer {
			recordAnswer := ans.(*dns.MX)
			fmt.Printf("%s\n", recordAnswer.Mx)
		}
	case "TXT":
		for _, ans := range r.Answer {
			recordAnswer := ans.(*dns.TXT)
			fmt.Printf("%s\n", recordAnswer.Txt)
		}
	default:
		log.Fatalln("Please enter a supported record type.")
	}
}

func main() {
	flag.Parse()

	switch record := *record; record {
	case "A":
		DNSLookup(dns.TypeA)
	case "AAAA":
		DNSLookup(dns.TypeAAAA)
	case "NS":
		DNSLookup(dns.TypeNS)
	case "MX":
		DNSLookup(dns.TypeMX)
	case "TXT":
		DNSLookup(dns.TypeTXT)
	default:
		log.Fatalln("Please enter a supported record type.")
	}
}
