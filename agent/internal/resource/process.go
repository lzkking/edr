package resource

import (
	"github.com/shirou/gopsutil/v3/process"
	"time"
)

func GetPidInfo(pid int32) (cpuPercent float64, rss uint64, readSpeed float64, writeSpeed float64, fds int32, startAt int64, err error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return
	}
	io0, _ := p.IOCounters()
	t0 := time.Now()

	cpuPercent, _ = p.Percent(time.Second)

	io1, _ := p.IOCounters()
	elapsed := time.Since(t0).Seconds()
	if io0 != nil && io1 != nil && elapsed > 0 {
		readSpeed = float64(io1.ReadBytes-io0.ReadBytes) / elapsed
		writeSpeed = float64(io1.WriteBytes-io0.WriteBytes) / elapsed
	}

	mem, _ := p.MemoryInfo()
	if mem != nil {
		rss = mem.RSS
	}

	fds, _ = p.NumFDs()
	startAt, _ = p.CreateTime()

	return
}
