package assertions

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Summary contains assertion counts
type Summary struct {
	PreventUserIdleDisplay int
	PreventSystemSleep     int
	PreventDisplaySleep    int
	ExternalMedia          int
	NetworkClientActive    int
}

// Assertion represents a single power assertion
type Assertion struct {
	PID     int
	Process string
	Type    string
}

// ScheduledEvent represents a scheduled wake/sleep event
type ScheduledEvent struct {
	Description string
}

// Info contains all assertion information
type Info struct {
	Summary   Summary
	Active    []Assertion
	Scheduled []ScheduledEvent
}

// Get retrieves power assertions
func Get() (*Info, error) {
	cmd := exec.Command("pmset", "-g", "assertions")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	data := string(output)
	info := &Info{}

	// Parse summary counts
	info.Summary = Summary{
		PreventUserIdleDisplay: parseCount(data, "PreventUserIdleDisplaySleep"),
		PreventSystemSleep:     parseCount(data, "PreventSystemSleep"),
		PreventDisplaySleep:    parseCount(data, "PreventDisplaySleep"),
		ExternalMedia:          parseCount(data, "ExternalMedia"),
		NetworkClientActive:    parseCount(data, "NetworkClientActive"),
	}

	// Parse active assertions
	info.Active = parseAssertions(data)

	// Get scheduled events
	info.Scheduled = getScheduledEvents()

	return info, nil
}

func parseCount(data, key string) int {
	re := regexp.MustCompile(key + `\s+(\d+)`)
	matches := re.FindStringSubmatch(data)
	if len(matches) >= 2 {
		val, _ := strconv.Atoi(matches[1])
		return val
	}
	return 0
}

func parseAssertions(data string) []Assertion {
	var assertions []Assertion

	// Find the "Listed by owning process" section
	idx := strings.Index(data, "Listed by owning process:")
	if idx == -1 {
		return nil
	}

	section := data[idx:]
	lines := strings.Split(section, "\n")

	// Parse lines like: pid 99(powerd): [0x000904b20001822d] ...
	re := regexp.MustCompile(`pid\s+(\d+)\(([^)]+)\):\s+\[([^\]]+)\]\s+.*?named:\s+"([^"]+)"`)

	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 5 {
			pid, _ := strconv.Atoi(matches[1])
			assertions = append(assertions, Assertion{
				PID:     pid,
				Process: matches[2],
				Type:    matches[4],
			})
		}
	}

	return assertions
}

func getScheduledEvents() []ScheduledEvent {
	cmd := exec.Command("pmset", "-g", "sched")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	data := strings.TrimSpace(string(output))
	if strings.Contains(data, "No scheduled") || len(data) == 0 {
		return nil
	}

	var events []ScheduledEvent
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if len(line) > 0 && !strings.HasPrefix(line, "Scheduled") {
			events = append(events, ScheduledEvent{Description: line})
		}
	}

	return events
}
