package battery

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Info contains battery information
type Info struct {
	CurrentCapacity    int
	MaxCapacity        int
	DesignCapacity     int
	CycleCount         int
	IsCharging         bool
	ExternalConnected  bool
	FullyCharged       bool
	Temperature        int // in centi-degrees
	TimeRemaining      int // in minutes
	RawCurrentCapacity int // AppleRawCurrentCapacity for Apple Silicon
	RawMaxCapacity     int // AppleRawMaxCapacity for Apple Silicon
}

// GetInfo retrieves battery information from IOKit
func GetInfo() (*Info, error) {
	cmd := exec.Command("ioreg", "-rc", "AppleSmartBattery")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	data := string(output)
	if len(data) == 0 {
		return nil, nil // No battery
	}

	info := &Info{}

	// Parse values using regex for top-level keys
	info.CurrentCapacity = parseIntValue(data, "CurrentCapacity")
	info.MaxCapacity = parseIntValue(data, "MaxCapacity")
	info.DesignCapacity = parseIntValue(data, "DesignCapacity")
	info.CycleCount = parseIntValue(data, "CycleCount")
	info.Temperature = parseIntValue(data, "Temperature")
	info.TimeRemaining = parseIntValue(data, "TimeRemaining")
	info.RawCurrentCapacity = parseIntValue(data, "AppleRawCurrentCapacity")
	info.RawMaxCapacity = parseIntValue(data, "AppleRawMaxCapacity")

	info.IsCharging = parseBoolValue(data, "IsCharging")
	info.ExternalConnected = parseBoolValue(data, "ExternalConnected")
	info.FullyCharged = parseBoolValue(data, "FullyCharged")

	return info, nil
}

// ActualCurrentCapacity returns the actual current capacity (prefers raw values)
func (i *Info) ActualCurrentCapacity() int {
	if i.RawCurrentCapacity > 0 {
		return i.RawCurrentCapacity
	}
	return i.CurrentCapacity
}

// ActualMaxCapacity returns the actual max capacity (prefers raw values)
func (i *Info) ActualMaxCapacity() int {
	if i.RawMaxCapacity > 0 {
		return i.RawMaxCapacity
	}
	return i.MaxCapacity
}

// ChargePercent calculates the charge percentage
func (i *Info) ChargePercent() int {
	max := i.ActualMaxCapacity()
	if max <= 0 {
		return 0
	}
	return (i.ActualCurrentCapacity() * 100) / max
}

// HealthPercent calculates the battery health percentage (capped at 100%)
func (i *Info) HealthPercent() int {
	if i.DesignCapacity <= 0 {
		return 0
	}
	health := (i.ActualMaxCapacity() * 100) / i.DesignCapacity
	if health > 100 {
		health = 100
	}
	return health
}

// TemperatureCelsius returns the temperature in Celsius
func (i *Info) TemperatureCelsius() int {
	return i.Temperature / 100
}

// Status returns a human-readable status string
func (i *Info) Status() string {
	if i.FullyCharged {
		return "Fully Charged"
	}
	if i.IsCharging {
		return "Charging"
	}
	if i.ExternalConnected {
		return "On AC Power"
	}
	return "On Battery"
}

// TimeRemainingFormatted returns formatted time remaining
func (i *Info) TimeRemainingFormatted() string {
	if i.TimeRemaining <= 0 {
		if i.FullyCharged {
			return "—"
		}
		return "Calculating..."
	}
	hours := i.TimeRemaining / 60
	mins := i.TimeRemaining % 60
	if i.IsCharging {
		return formatDuration(hours, mins) + " until full"
	}
	return formatDuration(hours, mins) + " remaining"
}

func formatDuration(hours, mins int) string {
	return strconv.Itoa(hours) + "h " + strconv.Itoa(mins) + "m"
}

func parseIntValue(data, key string) int {
	// Match lines like:   "Key" = value (with leading whitespace)
	re := regexp.MustCompile(`(?m)^\s+"` + key + `"\s*=\s*(\d+)`)
	matches := re.FindStringSubmatch(data)
	if len(matches) >= 2 {
		val, _ := strconv.Atoi(matches[1])
		return val
	}
	return 0
}

func parseBoolValue(data, key string) bool {
	re := regexp.MustCompile(`(?m)^\s+"` + key + `"\s*=\s*(\w+)`)
	matches := re.FindStringSubmatch(data)
	if len(matches) >= 2 {
		return strings.ToLower(matches[1]) == "yes"
	}
	return false
}
