package host

import (
	"net"
	"os"
	"strings"
	"sync/atomic"
)

var (
	Name atomic.Value

	PrivateIpv4 atomic.Value
	PublicIpv4  atomic.Value
	PrivateIpv6 atomic.Value
	PublicIpv6  atomic.Value
)

func RefreshHost() {
	host, _ := os.Hostname()
	Name.Store(host)

	privateIpv4 := []string{}
	privateIpv6 := []string{}
	publicIpv4 := []string{}
	publicIpv6 := []string{}

	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if strings.HasPrefix(i.Name, "docker") || strings.HasPrefix(i.Name, "lo") {
				continue
			}

			addrs, err := i.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				ip, _, err := net.ParseCIDR(addr.String())
				if err != nil || !ip.IsGlobalUnicast() {
					continue
				}

				if ip4 := ip.To4(); ip4 != nil {
					if (ip4[0] == 10) || (ip4[0] == 192 && ip4[1] == 168) || (ip4[0] == 172 && ip4[1] > 15 && ip4[1] < 32) {
						privateIpv4 = append(privateIpv4, ip4.String())
					} else {
						publicIpv4 = append(publicIpv4, ip4.String())
					}
				} else if len(ip) == net.IPv6len {
					if ip[0] == 0xfd {
						privateIpv6 = append(privateIpv6, ip.String())
					} else {
						publicIpv6 = append(publicIpv6, ip.String())
					}
				}
			}
		}
	}

	if len(privateIpv4) > 5 {
		privateIpv4 = privateIpv4[:5]
	}
	if len(privateIpv6) > 5 {
		privateIpv6 = privateIpv6[:5]
	}

	PrivateIpv4.Store(privateIpv4)
	PrivateIpv6.Store(privateIpv6)
	PublicIpv4.Store(publicIpv4)
	PublicIpv6.Store(publicIpv6)
}

func init() {
	RefreshHost()
}
