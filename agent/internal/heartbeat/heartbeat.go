package heartbeat

import (
	"context"
	"github.com/lzkking/edr/agent/assets"
	"github.com/lzkking/edr/agent/internal/buffer"
	"github.com/lzkking/edr/agent/internal/host"
	"github.com/lzkking/edr/agent/internal/plugin"
	"github.com/lzkking/edr/agent/internal/resource"
	pb "github.com/lzkking/edr/edrproto"
	"go.uber.org/zap"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// getAgentStat - 填充心跳包并传递到服务端
func getAgentStat(now time.Time) {
	rec := &pb.Record{
		DataType:  2025,
		Timestamp: now.Unix(),
		Data: &pb.Payload{Fields: make(
			map[string]string),
		},
	}

	rec.Data.Fields["kernel_version"] = host.KernelVersion
	rec.Data.Fields["arch"] = host.Arch
	rec.Data.Fields["platform"] = host.Platform
	rec.Data.Fields["platform_family"] = host.PlatformFamily
	rec.Data.Fields["platform_version"] = host.PlatformVersion

	//自身进程的信息
	cpuPercent, rss, readSpeed, writeSpeed, fds, startAt, err := resource.GetPidInfo(int32(os.Getpid()))
	if err != nil {
		zap.S().Warnf("获取进程的负载信息失败")
	} else {
		rec.Data.Fields["cpu_percent"] = strconv.FormatFloat(cpuPercent, 'f', 2, 64)
		rec.Data.Fields["rss"] = strconv.FormatUint(rss, 10)
		rec.Data.Fields["read_speed"] = strconv.FormatFloat(readSpeed, 'f', 2, 64)
		rec.Data.Fields["write_speed"] = strconv.FormatFloat(writeSpeed, 'f', 2, 64)
		rec.Data.Fields["fds"] = strconv.FormatInt(int64(fds), 10)
		rec.Data.Fields["start_at"] = strconv.FormatInt(startAt, 10)
	}

	// linux下的平均负载
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/proc/loadavg"); err == nil {
			fields := strings.Fields(string(data))
			if len(fields) >= 5 {
				rec.Data.Fields["load_1"] = fields[0]
				rec.Data.Fields["load_5"] = fields[1]
				rec.Data.Fields["load_15"] = fields[2]
				subFields := strings.Split(fields[3], "/")
				if len(subFields) > 1 {
					rec.Data.Fields["running_procs"] = subFields[0]
					rec.Data.Fields["total_procs"] = subFields[1]
				}
			}
		}
	}

	// 硬件信息
	hostSerial, hostID, hostModel, hostVendor := resource.GetHostInfo()
	rec.Data.Fields["host_serial"] = hostSerial
	rec.Data.Fields["host_id"] = hostID
	rec.Data.Fields["host_model"] = hostModel
	rec.Data.Fields["host_vendor"] = hostVendor

	// DNS信息
	dnss, err := resource.GetDNS()
	if err == nil {
		rec.Data.Fields["DNSs"] = strings.Join(dnss, ",")
	}

	//gateway信息
	gateway, err := resource.GetGateway()
	if err == nil {
		rec.Data.Fields["gateway"] = gateway
	}

	//工作路径下的信息
	rootDir := assets.GetAgentRootAppDir()
	rootDirSize, err := resource.DirSize(rootDir)
	if err == nil {
		rec.Data.Fields["work_dir_size"] = strconv.FormatInt(rootDirSize, 10)
	}

	rec.Data.Fields["num_goroutine"] = strconv.FormatInt(int64(runtime.NumGoroutine()), 10)
	rec.Data.Fields["num_max_procs"] = strconv.FormatInt(int64(runtime.GOMAXPROCS(0)), 10)

	// CPU信息
	cpuName, bootTime, cpuSystemPercent, memSysPercent := resource.GetSystemInfo()
	rec.Data.Fields["cpu_name"] = cpuName
	rec.Data.Fields["boot_time"] = strconv.FormatUint(bootTime, 10)
	rec.Data.Fields["sys_cpu_percent"] = strconv.FormatFloat(cpuSystemPercent, 'f', 2, 64)
	rec.Data.Fields["sys_mem_percent"] = strconv.FormatFloat(memSysPercent, 'f', 2, 64)

	buffer.WriteRecord(rec)
}

func getPlgStat(now time.Time) {
	plgs := plugin.GetAll()
	for _, plg := range plgs {
		if !plg.IsExited() {
			rec := &pb.Record{
				DataType:  2026,
				Timestamp: now.Unix(),
				Data: &pb.Payload{
					Fields: map[string]string{
						"name":     plg.Name(),
						"pversion": plg.Version(),
					},
				},
			}
			cpuPercent, rss, readSpeed, writeSpeed, fds, startAt, err := resource.GetPidInfo(int32(plg.Pid()))
			if err != nil {
				zap.S().Error(err)
			} else {
				rec.Data.Fields["cpu"] = strconv.FormatFloat(cpuPercent, 'f', 8, 64)
				rec.Data.Fields["rss"] = strconv.FormatUint(rss, 10)
				rec.Data.Fields["read_speed"] = strconv.FormatFloat(readSpeed, 'f', 8, 64)
				rec.Data.Fields["write_speed"] = strconv.FormatFloat(writeSpeed, 'f', 8, 64)
				rec.Data.Fields["pid"] = strconv.Itoa(plg.Pid())
				rec.Data.Fields["nfd"] = strconv.FormatInt(int64(fds), 10)
				rec.Data.Fields["start_time"] = strconv.FormatInt(startAt, 10)
			}
			plgDirSize, _ := resource.DirSize(plg.GetWorkingDirectory())
			rec.Data.Fields["du"] = strconv.FormatUint(uint64(plgDirSize), 10)
			RxSpeed, TxSpeed, RxTPS, TxTPS := plg.GetState()
			rec.Data.Fields["rx_tps"] = strconv.FormatFloat(RxTPS, 'f', 8, 64)
			rec.Data.Fields["tx_tps"] = strconv.FormatFloat(TxTPS, 'f', 8, 64)
			rec.Data.Fields["rx_speed"] = strconv.FormatFloat(RxSpeed, 'f', 8, 64)
			rec.Data.Fields["tx_speed"] = strconv.FormatFloat(TxSpeed, 'f', 8, 64)

			buffer.WriteRecord(rec)
		}
	}
}

func StartUp(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	getAgentStat(time.Now())
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			{
				zap.S().Infof("向agent-center发送心跳包")
				host.RefreshHost()
				getAgentStat(t)
				getPlgStat(t)
			}
		}
	}
}
