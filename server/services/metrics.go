package services

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// SystemMetrics holds a snapshot of system resource usage.
type SystemMetrics struct {
	CPU         float64 `json:"cpu"`
	RAMUsedMB   int64   `json:"ramUsedMb"`
	RAMTotalMB  int64   `json:"ramTotalMb"`
	DiskUsedGB  float64 `json:"diskUsedGb"`
	DiskTotalGB float64 `json:"diskTotalGb"`
	TempC       float64 `json:"tempC"`
	UptimeS     int64   `json:"uptimeS"`
}

// Read collects a fresh SystemMetrics snapshot.
// CPU is sampled over a 100 ms window. Returns 0 for values that are
// not available on the current platform (e.g. temperature on non-Pi Linux).
func Read() (SystemMetrics, error) {
	cpu, err := readCPU()
	if err != nil {
		cpu = 0
	}

	ramUsed, ramTotal, err := readRAM()
	if err != nil {
		ramUsed, ramTotal = 0, 0
	}

	diskUsed, diskTotal, err := readDisk()
	if err != nil {
		diskUsed, diskTotal = 0, 0
	}

	temp := readTemp()
	uptime, _ := readUptime()

	return SystemMetrics{
		CPU:         cpu,
		RAMUsedMB:   ramUsed,
		RAMTotalMB:  ramTotal,
		DiskUsedGB:  diskUsed,
		DiskTotalGB: diskTotal,
		TempC:       temp,
		UptimeS:     uptime,
	}, nil
}

// cpuStat holds raw values from /proc/stat.
type cpuStat struct {
	idle  uint64
	total uint64
}

func readCPUStat() (cpuStat, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return cpuStat{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			return cpuStat{}, fmt.Errorf("metrics: unexpected /proc/stat format")
		}
		var vals [10]uint64
		for i := 1; i < len(fields) && i <= 10; i++ {
			v, _ := strconv.ParseUint(fields[i], 10, 64)
			vals[i-1] = v
		}
		// user, nice, system, idle, iowait, irq, softirq, steal, guest, guest_nice
		idle := vals[3] + vals[4] // idle + iowait
		total := uint64(0)
		for _, v := range vals {
			total += v
		}
		return cpuStat{idle: idle, total: total}, nil
	}
	return cpuStat{}, fmt.Errorf("metrics: cpu line not found in /proc/stat")
}

func readCPU() (float64, error) {
	s1, err := readCPUStat()
	if err != nil {
		return 0, err
	}
	time.Sleep(100 * time.Millisecond)
	s2, err := readCPUStat()
	if err != nil {
		return 0, err
	}

	totalDelta := float64(s2.total - s1.total)
	idleDelta := float64(s2.idle - s1.idle)
	if totalDelta == 0 {
		return 0, nil
	}
	return (1 - idleDelta/totalDelta) * 100, nil
}

func readRAM() (usedMB, totalMB int64, err error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	values := make(map[string]int64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		val, _ := strconv.ParseInt(parts[1], 10, 64)
		values[key] = val
	}

	total := values["MemTotal"]   // kB
	avail := values["MemAvailable"] // kB
	if total == 0 {
		return 0, 0, fmt.Errorf("metrics: MemTotal not found")
	}
	used := total - avail
	return used / 1024, total / 1024, nil
}

func readDisk() (usedGB, totalGB float64, err error) {
	var stat syscall.Statfs_t
	path := "/"
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, fmt.Errorf("metrics: statfs %q: %w", path, err)
	}
	blockSize := uint64(stat.Bsize)
	total := float64(stat.Blocks*blockSize) / (1 << 30)
	free := float64(stat.Bavail*blockSize) / (1 << 30)
	return total - free, total, nil
}

func readTemp() float64 {
	data, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0
	}
	raw := strings.TrimSpace(string(data))
	milliC, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0
	}
	return milliC / 1000.0
}

func readUptime() (int64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0, fmt.Errorf("metrics: empty /proc/uptime")
	}
	f, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, err
	}
	return int64(f), nil
}
