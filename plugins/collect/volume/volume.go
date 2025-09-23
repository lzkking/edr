package volume

import (
	"github.com/lzkking/edr/plugins/collect/engine"
	plugins "github.com/lzkking/edr/plugins/lib"
	"github.com/shirou/gopsutil/v3/disk"
	"strconv"
	"time"
)

type VolumeHandler struct{}

func (h *VolumeHandler) Name() string {
	return "volume"
}

func (h *VolumeHandler) DataType() int {
	return 7319
}

func (h *VolumeHandler) Handle(c *plugins.Client, cache *engine.Cache, seq string) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return
	}

	for _, part := range parts {
		if usage, err := disk.Usage(part.Mountpoint); err == nil {
			c.SendRecord(&plugins.Record{
				DataType:  int32(h.DataType()),
				Timestamp: time.Now().Unix(),
				Data: &plugins.Payload{Fields: map[string]string{
					"name":        part.Device,
					"fstype":      part.Fstype,
					"mount_point": part.Mountpoint,
					"total":       strconv.FormatUint(usage.Total, 10),
					"used":        strconv.FormatUint(usage.Used, 10),
					"free":        strconv.FormatUint(usage.Free, 10),
					"usage":       strconv.FormatFloat(usage.UsedPercent/100, 'f', 8, 64),
					"package_seq": seq,
				}},
			})
		}
	}
}
