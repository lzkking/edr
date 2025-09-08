package resource

import (
	"bufio"
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// GetDNS 跨平台获取 DNS 服务器列表
func GetDNS() ([]string, error) {
	if runtime.GOOS == "windows" {
		// Windows: 调用 netsh
		out, err := exec.Command("netsh", "interface", "ip", "show", "dnsservers").CombinedOutput()
		if err != nil {
			return nil, err
		}
		var servers []string
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if ip := net.ParseIP(line); ip != nil {
				servers = append(servers, ip.String())
			}
		}
		return servers, nil
	}

	// Linux/macOS: 解析 /etc/resolv.conf
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var servers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				servers = append(servers, fields[1])
			}
		}
	}
	return servers, scanner.Err()
}

// GetGateway 跨平台获取默认网关
func GetGateway() (string, error) {
	switch runtime.GOOS {
	case "linux":
		routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
		if err != nil {
			return "", err
		}
		for _, r := range routes {
			if r.Dst == nil && r.Gw != nil {
				return r.Gw.String(), nil
			}
		}
		return "", fmt.Errorf("no default gateway found")

	case "darwin":
		// macOS: 使用 route
		out, err := exec.Command("route", "-n", "get", "default").CombinedOutput()
		if err != nil {
			return "", err
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "gateway:") {
				return strings.TrimSpace(strings.TrimPrefix(line, "gateway:")), nil
			}
		}
		return "", fmt.Errorf("gateway not found")

	case "windows":
		// Windows: netstat
		out, err := exec.Command("route", "print", "0.0.0.0").CombinedOutput()
		if err != nil {
			return "", err
		}
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 4 && fields[0] == "0.0.0.0" {
				return fields[2], nil // gateway 列
			}
		}
		return "", fmt.Errorf("gateway not found")

	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}
