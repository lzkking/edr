package process

import (
	plugins "github.com/lzkking/edr/plugins/lib"
	"github.com/mitchellh/mapstructure"
	"time"
)

type ProcessHandler struct{}

func (h *ProcessHandler) Name() string {
	return "process"
}

func (h *ProcessHandler) DataType() int {
	return 7312
}

func (h *ProcessHandler) Handle(c *plugins.Client, seq string) {
	procs, err := Processes(false)
	if err != nil {
		return
	} else {
		for _, p := range procs {
			time.Sleep(TraversalInterval)
			cmdline, err := p.Cmdline()
			if err != nil {
				continue
			}
			stat, err := p.Stat()
			if err != nil {
				continue
			}
			status, err := p.Status()
			if err != nil {
				continue
			}
			ns, _ := p.Namespaces()
			rec := &plugins.Record{
				DataType:  int32(h.DataType()),
				Timestamp: time.Now().Unix(),
				Data: &plugins.Payload{
					Fields: make(map[string]string, 40),
				},
			}
			rec.Data.Fields["cmdline"] = cmdline
			rec.Data.Fields["cwd"], _ = p.Cwd()
			rec.Data.Fields["checksum"], _ = p.ExeChecksum()
			rec.Data.Fields["exe_hash"], _ = p.ExeHash()
			rec.Data.Fields["exe"], _ = p.Exe()
			rec.Data.Fields["pid"] = p.Pid()
			mapstructure.Decode(stat, &rec.Data.Fields)
			mapstructure.Decode(status, &rec.Data.Fields)
			mapstructure.Decode(ns, &rec.Data.Fields)
			//缺少对容器进程的映射

			rec.Data.Fields["package_seq"] = seq

			c.SendRecord(rec)
		}
	}
}
