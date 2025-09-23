package net_interface

import (
	"github.com/lzkking/edr/plugins/collect/engine"
	plugins "github.com/lzkking/edr/plugins/lib"
	"net"
	"strconv"
	"strings"
	"time"
)

type NetInterfaceHandler struct{}

func (h *NetInterfaceHandler) Name() string {
	return "net_interface"
}

func (h *NetInterfaceHandler) DataType() int {
	return 7318
}

func (h *NetInterfaceHandler) Handle(c *plugins.Client, cache *engine.Cache, seq string) {
	nfs, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, nf := range nfs {
		if addrs, err := nf.Addrs(); err == nil {
			c.SendRecord(&plugins.Record{
				DataType:  int32(h.DataType()),
				Timestamp: time.Now().Unix(),
				Data: &plugins.Payload{Fields: map[string]string{
					"name":          nf.Name,
					"hardware_addr": nf.HardwareAddr.String(),
					"addrs":         joinAddrs(addrs, ","),
					"index":         strconv.Itoa(nf.Index),
					"mtu":           strconv.Itoa(nf.MTU),
					"package_seq":   seq,
				}},
			})
		}
	}
}

func joinAddrs(addrs []net.Addr, sep string) string {
	strs := []string{}
	for _, addr := range addrs {
		strs = append(strs, addr.String())
	}
	return strings.Join(strs, sep)
}
