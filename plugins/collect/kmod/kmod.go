package kmod

import (
	"bufio"
	"github.com/lzkking/edr/plugins/collect/engine"
	plugins "github.com/lzkking/edr/plugins/lib"
	"os"
	"strings"
	"time"
)

type KmodHandler struct{}

func (h *KmodHandler) Name() string {
	return "kmod"
}

func (h *KmodHandler) DataType() int {
	return 7320
}

func (h *KmodHandler) Handle(c *plugins.Client, cache *engine.Cache, seq string) {
	f, err := os.Open("/proc/modules")
	if err != nil {
		return
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) > 5 {
			c.SendRecord(&plugins.Record{
				DataType:  int32(h.DataType()),
				Timestamp: time.Now().Unix(),
				Data: &plugins.Payload{Fields: map[string]string{
					"name":        fields[0],
					"size":        fields[1],
					"refcount":    fields[2],
					"used_by":     fields[3],
					"state":       fields[4],
					"addr":        fields[5],
					"package_seq": seq,
				}},
			})
		}
	}
}
