package port

import (
	"github.com/lzkking/edr/plugins/collect/engine"
	plugins "github.com/lzkking/edr/plugins/lib"
	"github.com/mitchellh/mapstructure"
	"time"
)

type PortHandler struct{}

func (h *PortHandler) Name() string {
	return "port"
}

func (h *PortHandler) DataType() int {
	return 7311
}

func (h *PortHandler) Handle(c *plugins.Client, cache *engine.Cache, seq string) {
	ports, err := ListeningPorts()
	if err != nil {
		return
	}

	for _, port := range ports {
		rec := &plugins.Record{
			DataType:  int32(h.DataType()),
			Timestamp: time.Now().Unix(),
			Data: &plugins.Payload{
				Fields: make(map[string]string, 15),
			},
		}

		mapstructure.Decode(port, &rec.Data.Fields)
		m, _ := cache.Get(7314, port.Sport)
		rec.Data.Fields["container_id"] = m["container_id"]
		rec.Data.Fields["container_name"] = m["container_name"]
		rec.Data.Fields["package_seq"] = seq
		c.SendRecord(rec)
	}
}
