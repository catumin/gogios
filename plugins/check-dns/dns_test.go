package main

import (
	"testing"

	"github.com/miekg/dns"
)

func TestDNSLookup(t *testing.T) {
	ip, _ := DNSLookup(dns.TypeA, "angrysysadmins.tech", "8.8.8.8")

	if ip.(*dns.A).A.String() != "45.33.54.48" {
		t.Errorf("A record lookup failed, got: %s, want: %s.", ip.(*dns.A).A.String(), "45.33.54.48")
	}
}
