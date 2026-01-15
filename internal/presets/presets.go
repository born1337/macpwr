package presets

import (
	"fmt"
	"os/exec"
	"strconv"
)

// Preset defines a power preset
type Preset struct {
	Name        string
	Description string
	AC          Settings
	Battery     Settings
}

// Settings for a power source
type Settings struct {
	DisplaySleep int
	SystemSleep  int
	DiskSleep    int
}

// All available presets
var All = []Preset{
	{
		Name:        "default",
		Description: "Restore macOS default power settings",
		AC:          Settings{DisplaySleep: 10, SystemSleep: 0, DiskSleep: 10},
		Battery:     Settings{DisplaySleep: 2, SystemSleep: 10, DiskSleep: 10},
	},
	{
		Name:        "presentation",
		Description: "Keep screen on, prevent sleep (ideal for presentations)",
		AC:          Settings{DisplaySleep: 0, SystemSleep: 0, DiskSleep: 0},
		Battery:     Settings{DisplaySleep: 0, SystemSleep: 0, DiskSleep: 0},
	},
	{
		Name:        "battery-saver",
		Description: "Aggressive power saving to extend battery life",
		AC:          Settings{DisplaySleep: 5, SystemSleep: 10, DiskSleep: 5},
		Battery:     Settings{DisplaySleep: 1, SystemSleep: 2, DiskSleep: 2},
	},
	{
		Name:        "performance",
		Description: "Maximum performance, no sleep restrictions",
		AC:          Settings{DisplaySleep: 0, SystemSleep: 0, DiskSleep: 0},
		Battery:     Settings{DisplaySleep: 10, SystemSleep: 0, DiskSleep: 0},
	},
	{
		Name:        "movie",
		Description: "Screen stays on, system can sleep normally",
		AC:          Settings{DisplaySleep: 0, SystemSleep: 0, DiskSleep: 10},
		Battery:     Settings{DisplaySleep: 0, SystemSleep: 30, DiskSleep: 10},
	},
}

// Get returns a preset by name
func Get(name string) *Preset {
	// Handle alias
	if name == "batterysaver" {
		name = "battery-saver"
	}

	for _, p := range All {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

// Apply applies a preset
func Apply(p *Preset) error {
	// Apply AC settings
	if err := applySettings("-c", p.AC); err != nil {
		return fmt.Errorf("failed to apply AC settings: %w", err)
	}

	// Apply battery settings
	if err := applySettings("-b", p.Battery); err != nil {
		return fmt.Errorf("failed to apply battery settings: %w", err)
	}

	return nil
}

func applySettings(powerSource string, s Settings) error {
	cmd := exec.Command("sudo", "pmset", powerSource,
		"displaysleep", strconv.Itoa(s.DisplaySleep),
		"sleep", strconv.Itoa(s.SystemSleep),
		"disksleep", strconv.Itoa(s.DiskSleep))
	return cmd.Run()
}
