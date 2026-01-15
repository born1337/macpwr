package thermal

import (
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// Info contains thermal and CPU information
type Info struct {
	CPUModel      string
	CPUCores      int
	Architecture  string
	LoadAverage   string
	MemoryFree    int // percentage
	PowerSource   string
	CPULimit      int // percentage (100 = no throttling)
	FansAvailable bool
}

// Get retrieves thermal information
func Get() (*Info, error) {
	info := &Info{}

	// CPU Model
	if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
		info.CPUModel = strings.TrimSpace(string(out))
	}

	// CPU Cores
	if out, err := exec.Command("sysctl", "-n", "hw.ncpu").Output(); err == nil {
		info.CPUCores, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}

	// Architecture
	info.Architecture = runtime.GOARCH
	if info.Architecture == "arm64" {
		info.Architecture = "Apple Silicon"
	} else if info.Architecture == "amd64" {
		info.Architecture = "Intel"
	}

	// Load Average
	if out, err := exec.Command("sysctl", "-n", "vm.loadavg").Output(); err == nil {
		load := strings.TrimSpace(string(out))
		load = strings.Trim(load, "{}")
		info.LoadAverage = strings.TrimSpace(load)
	}

	// Memory pressure
	if out, err := exec.Command("memory_pressure").Output(); err == nil {
		re := regexp.MustCompile(`System-wide memory free percentage:\s*(\d+)`)
		matches := re.FindStringSubmatch(string(out))
		if len(matches) >= 2 {
			info.MemoryFree, _ = strconv.Atoi(matches[1])
		}
	}

	// Power source
	if out, err := exec.Command("pmset", "-g", "batt").Output(); err == nil {
		data := string(out)
		if strings.Contains(data, "AC Power") {
			info.PowerSource = "AC Power"
		} else if strings.Contains(data, "Battery") {
			info.PowerSource = "Battery"
		} else {
			info.PowerSource = "Unknown"
		}
	}

	// CPU thermal limit
	if out, err := exec.Command("pmset", "-g", "therm").Output(); err == nil {
		re := regexp.MustCompile(`CPU_Scheduler_Limit\s*=\s*(\d+)`)
		matches := re.FindStringSubmatch(string(out))
		if len(matches) >= 2 {
			info.CPULimit, _ = strconv.Atoi(matches[1])
		}
	}

	// Check for fans
	if out, err := exec.Command("ioreg", "-rc", "AppleSMCACPIPlatformPlugin").Output(); err == nil {
		info.FansAvailable = strings.Contains(strings.ToLower(string(out)), "fan")
	}
	if !info.FansAvailable {
		if out, err := exec.Command("ioreg", "-rc", "AppleFanCtrl").Output(); err == nil {
			info.FansAvailable = len(out) > 0
		}
	}

	return info, nil
}
