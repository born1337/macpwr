package settings

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// PowerSettings contains power settings for a power source
type PowerSettings struct {
	DisplaySleep  int
	SystemSleep   int
	DiskSleep     int
	PowerNap      bool
	WakeOnLAN     bool
	LowPowerMode  bool
	TCPKeepAlive  bool
}

// AllSettings contains settings for both power sources
type AllSettings struct {
	Battery *PowerSettings
	AC      *PowerSettings
}

// Get retrieves current power settings
func Get() (*AllSettings, error) {
	cmd := exec.Command("pmset", "-g", "custom")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	data := string(output)

	// Split into battery and AC sections
	parts := strings.Split(data, "AC Power:")
	if len(parts) < 2 {
		return nil, fmt.Errorf("unexpected pmset output format")
	}

	batterySection := parts[0]
	acSection := parts[1]

	return &AllSettings{
		Battery: parseSection(batterySection),
		AC:      parseSection(acSection),
	}, nil
}

func parseSection(section string) *PowerSettings {
	return &PowerSettings{
		DisplaySleep:  parseIntSetting(section, "displaysleep"),
		SystemSleep:   parseIntSetting(section, "sleep"),
		DiskSleep:     parseIntSetting(section, "disksleep"),
		PowerNap:      parseBoolSetting(section, "powernap"),
		WakeOnLAN:     parseBoolSetting(section, "womp"),
		LowPowerMode:  parseBoolSetting(section, "lowpowermode"),
		TCPKeepAlive:  parseBoolSetting(section, "tcpkeepalive"),
	}
}

func parseIntSetting(section, key string) int {
	re := regexp.MustCompile(`(?m)^\s*` + key + `\s+(\d+)`)
	matches := re.FindStringSubmatch(section)
	if len(matches) >= 2 {
		val, _ := strconv.Atoi(matches[1])
		return val
	}
	return 0
}

func parseBoolSetting(section, key string) bool {
	re := regexp.MustCompile(`(?m)^\s*` + key + `\s+(\d+)`)
	matches := re.FindStringSubmatch(section)
	if len(matches) >= 2 {
		return matches[1] == "1"
	}
	return false
}

// SetOptions contains options for setting power settings
type SetOptions struct {
	PowerSource  string // "-c" for AC, "-b" for battery
	DisplaySleep *int
	SystemSleep  *int
	DiskSleep    *int
}

// Set applies power settings
func Set(opts SetOptions) error {
	args := []string{opts.PowerSource}

	if opts.DisplaySleep != nil {
		args = append(args, "displaysleep", strconv.Itoa(*opts.DisplaySleep))
	}
	if opts.SystemSleep != nil {
		args = append(args, "sleep", strconv.Itoa(*opts.SystemSleep))
	}
	if opts.DiskSleep != nil {
		args = append(args, "disksleep", strconv.Itoa(*opts.DiskSleep))
	}

	cmd := exec.Command("sudo", append([]string{"pmset"}, args...)...)
	cmd.Stdin = nil
	return cmd.Run()
}

// GetPowerSource returns the current power source
func GetPowerSource() string {
	cmd := exec.Command("pmset", "-g", "batt")
	output, err := cmd.Output()
	if err != nil {
		return "Unknown"
	}

	data := string(output)
	if strings.Contains(data, "AC Power") {
		return "AC Power"
	}
	if strings.Contains(data, "Battery") {
		return "Battery"
	}
	return "Unknown"
}
