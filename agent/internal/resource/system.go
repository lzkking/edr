package resource

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"time"
)

func GetSystemInfo() (cpuName string, bootTime uint64, cpuPercent float64, memPercent float64) {
	info, err := cpu.Info()
	if err == nil && len(info) > 0 {
		cpuName = info[0].ModelName
	}

	boot, err := host.BootTime()
	if err == nil {
		bootTime = boot
	}

	cpuPercents, err := cpu.Percent(time.Second, false)
	if err == nil {
		cpuPercent = cpuPercents[0]
	}

	vmStat, err := mem.VirtualMemory()
	if err == nil {
		memPercent = vmStat.UsedPercent
	}

	return
}
