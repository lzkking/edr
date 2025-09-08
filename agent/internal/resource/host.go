package resource

import (
	"github.com/jaypipes/ghw"
	"os"
	"runtime"
	"strings"
)

func trim(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\x00", ""))
}

func readSysFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return trim(string(b))
}

func GetHostInfo() (hostSerial, hostID, hostModel, hostVendor string) {
	if prod, err := ghw.Product(); err == nil && prod != nil {
		hostSerial = trim(prod.SerialNumber)
		hostID = trim(prod.UUID)
		hostModel = trim(prod.Name)
		hostVendor = trim(prod.Vendor)
	}

	if runtime.GOOS == "linux" {
		if hostSerial == "" {
			hostSerial = readSysFile("/sys/class/dmi/id/product_serial")
		}
		if hostID == "" {
			if id := readSysFile("/sys/class/dmi/id/product_uuid"); id != "" {
				hostID = id
			}
		}
		if hostModel == "" {
			hostModel = readSysFile("/sys/class/dmi/id/product_name")
		}
		if hostVendor == "" {
			hostVendor = readSysFile("/sys/class/dmi/id/sys_vendor")
		}
	}

	return
}
