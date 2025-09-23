package container

import (
	"context"
	"github.com/lzkking/edr/plugins/collect/engine"
	"github.com/lzkking/edr/plugins/collect/process"
	plugins "github.com/lzkking/edr/plugins/lib"
	"time"
)

type ContainerHandler struct{}

func (h *ContainerHandler) Name() string {
	return "container"
}

func (h *ContainerHandler) DataType() int {
	return 7314
}

func (h *ContainerHandler) Handle(c *plugins.Client, cache *engine.Cache, seq string) {
	clients := NewClients()
	for _, client := range clients {
		containers, err := client.ListContainers(context.Background())
		client.Close()
		if err != nil {
			continue
		}

		for _, ctr := range containers {
			c.SendRecord(&plugins.Record{
				DataType:  int32(h.DataType()),
				Timestamp: time.Now().Unix(),
				Data: &plugins.Payload{
					Fields: map[string]string{
						"id":          ctr.ID,
						"name":        ctr.Name,
						"state":       ctr.State,
						"image_id":    ctr.ImageID,
						"image_name":  ctr.ImageName,
						"pid":         ctr.Pid,
						"pns":         ctr.Pns,
						"runtime":     ctr.State,
						"create_time": ctr.CreateTime,
						"package_seq": seq,
					},
				},
			})
			if ctr.State == StateName[int32(RUNNING)] && ctr.Pns != "" && process.PnsDiffWithRpns(ctr.Pns) {
				cache.Put(h.DataType(), ctr.Pns, map[string]string{
					"container_id":   ctr.ID,
					"container_name": ctr.Name,
				})
			}
		}
	}
}
