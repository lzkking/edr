package port

import (
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

func (h *PortHandler) Handle(c *plugins.Client, seq string) {
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
		rec.Data.Fields["package_seq"] = seq
		c.SendRecord(rec)
	}
}
