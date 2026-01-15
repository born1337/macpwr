package profiles

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/born1337/macpwr/internal/settings"
)

// Profile represents a saved power profile
type Profile struct {
	Name    string
	Created time.Time
	Battery Settings
	AC      Settings
}

// Settings for a power source
type Settings struct {
	DisplaySleep int
	SystemSleep  int
	DiskSleep    int
}

// ProfilesDir returns the profiles directory path
func ProfilesDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "macpwr", "profiles")
}

// EnsureDir creates the profiles directory if it doesn't exist
func EnsureDir() error {
	return os.MkdirAll(ProfilesDir(), 0755)
}

// List returns all saved profiles
func List() ([]Profile, error) {
	dir := ProfilesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var profiles []Profile
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".profile") {
			name := strings.TrimSuffix(entry.Name(), ".profile")
			p, err := Load(name)
			if err == nil {
				profiles = append(profiles, *p)
			}
		}
	}
	return profiles, nil
}

// Save saves the current settings as a profile
func Save(name string) error {
	// Validate name
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name) {
		return fmt.Errorf("invalid profile name: use only letters, numbers, hyphens, and underscores")
	}

	if err := EnsureDir(); err != nil {
		return err
	}

	// Get current settings
	current, err := settings.Get()
	if err != nil {
		return fmt.Errorf("failed to read current settings: %w", err)
	}

	// Create profile file
	path := filepath.Join(ProfilesDir(), name+".profile")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "# macpwr profile: %s\n", name)
	fmt.Fprintf(f, "# Created: %s\n\n", time.Now().Format(time.RFC1123))
	fmt.Fprintln(f, "[battery]")
	fmt.Fprintf(f, "displaysleep=%d\n", current.Battery.DisplaySleep)
	fmt.Fprintf(f, "sleep=%d\n", current.Battery.SystemSleep)
	fmt.Fprintf(f, "disksleep=%d\n\n", current.Battery.DiskSleep)
	fmt.Fprintln(f, "[ac]")
	fmt.Fprintf(f, "displaysleep=%d\n", current.AC.DisplaySleep)
	fmt.Fprintf(f, "sleep=%d\n", current.AC.SystemSleep)
	fmt.Fprintf(f, "disksleep=%d\n", current.AC.DiskSleep)

	return nil
}

// Load loads a profile from disk
func Load(name string) (*Profile, error) {
	path := filepath.Join(ProfilesDir(), name+".profile")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p := &Profile{Name: name}
	var currentSection string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "# Created:") {
			p.Created, _ = time.Parse(time.RFC1123, strings.TrimPrefix(line, "# Created: "))
		} else if line == "[battery]" {
			currentSection = "battery"
		} else if line == "[ac]" {
			currentSection = "ac"
		} else if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := parts[0]
			val, _ := strconv.Atoi(parts[1])

			switch currentSection {
			case "battery":
				switch key {
				case "displaysleep":
					p.Battery.DisplaySleep = val
				case "sleep":
					p.Battery.SystemSleep = val
				case "disksleep":
					p.Battery.DiskSleep = val
				}
			case "ac":
				switch key {
				case "displaysleep":
					p.AC.DisplaySleep = val
				case "sleep":
					p.AC.SystemSleep = val
				case "disksleep":
					p.AC.DiskSleep = val
				}
			}
		}
	}

	return p, scanner.Err()
}

// Apply applies a profile's settings
func Apply(p *Profile) error {
	// Apply battery settings
	cmd := exec.Command("sudo", "pmset", "-b",
		"displaysleep", strconv.Itoa(p.Battery.DisplaySleep),
		"sleep", strconv.Itoa(p.Battery.SystemSleep),
		"disksleep", strconv.Itoa(p.Battery.DiskSleep))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply battery settings: %w", err)
	}

	// Apply AC settings
	cmd = exec.Command("sudo", "pmset", "-c",
		"displaysleep", strconv.Itoa(p.AC.DisplaySleep),
		"sleep", strconv.Itoa(p.AC.SystemSleep),
		"disksleep", strconv.Itoa(p.AC.DiskSleep))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply AC settings: %w", err)
	}

	return nil
}

// Delete deletes a profile
func Delete(name string) error {
	path := filepath.Join(ProfilesDir(), name+".profile")
	return os.Remove(path)
}

// Exists checks if a profile exists
func Exists(name string) bool {
	path := filepath.Join(ProfilesDir(), name+".profile")
	_, err := os.Stat(path)
	return err == nil
}
