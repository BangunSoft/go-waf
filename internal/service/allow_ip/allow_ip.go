package service_allow_ip

import (
	"go-waf/config"
	"go-waf/internal/interface/service"
	"net/netip"
	"strings"
)

type AllowIP struct {
	prefixes []netip.Prefix
}

func NewAllowIP(config *config.Config) service.AllowIPInterface {
	s := &AllowIP{}
	s.loadIPs(config.CACHE_REMOVE_ALLOW_IP)

	return s
}

func (s *AllowIP) loadIPs(ipv4s string) {
	// List of IP prefixes to parse
	ips := strings.Split(ipv4s, ",")

	// Parse the IP prefixes and store them in the struct
	for _, ip := range ips {
		if !strings.Contains(ip, "/") {
			ip += "/32"
		}
		prefix := netip.MustParsePrefix(ip)
		s.prefixes = append(s.prefixes, prefix)
	}
}

func (s *AllowIP) Check(ipv4 string) bool {
	ip, err := netip.ParseAddr(ipv4)
	if err != nil {
		return false
	}

	// Check if the IP is contained in any of the prefixes
	for _, prefix := range s.prefixes {
		if prefix.Contains(ip) {
			return true
		}
	}

	return false
}
