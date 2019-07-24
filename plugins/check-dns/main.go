package main

import (
	"errors"
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

// DNSLookup - Check record from domain using server to lookup
func DNSLookup(recordType uint16, domain string, server string) (dns.RR, error) {
	c := dns.Client{}
	m := dns.Msg{}

	m.SetQuestion(domain+".", recordType)
	r, t, err := c.Exchange(&m, server+":"+strconv.Itoa(*port))
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
			return ans, nil
		}
	case "AAAA":
		for _, ans := range r.Answer {
			recordAnswer := ans.(*dns.AAAA)
			fmt.Printf("%s\n", recordAnswer.AAAA)
			return ans, nil
		}
	case "NS":
		for _, ans := range r.Answer {
			recordAnswer := ans.(*dns.NS)
			fmt.Printf("%s\n", recordAnswer.Ns)
			return ans, nil
		}
	case "MX":
		for _, ans := range r.Answer {
			recordAnswer := ans.(*dns.MX)
			fmt.Printf("%s\n", recordAnswer.Mx)
			return ans, nil
		}
	case "TXT":
		for _, ans := range r.Answer {
			recordAnswer := ans.(*dns.TXT)
			fmt.Printf("%s\n", recordAnswer.Txt)
			return ans, nil
		}
	default:
		log.Fatalln("Please enter a supported record type.")
		return nil, errors.New("Record type not supported")
	}

	return nil, errors.New("Received invalid resposne")
}

func main() {
	flag.Parse()

	switch record := *record; record {
	case "A":
		DNSLookup(dns.TypeA, *domain, *server)
	case "AAAA":
		DNSLookup(dns.TypeAAAA, *domain, *server)
	case "NS":
		DNSLookup(dns.TypeNS, *domain, *server)
	case "MX":
		DNSLookup(dns.TypeMX, *domain, *server)
	case "TXT":
		DNSLookup(dns.TypeTXT, *domain, *server)
	default:
		log.Fatalln("Please enter a supported record type.")
	}
}
