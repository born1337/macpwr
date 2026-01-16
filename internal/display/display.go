package display

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// ANSI color codes
var (
	Red    = "\033[0;31m"
	Green  = "\033[0;32m"
	Yellow = "\033[0;33m"
	Blue   = "\033[0;34m"
	Cyan   = "\033[0;36m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
	Reset  = "\033[0m"
)

func init() {
	// Disable colors if not a terminal or NO_COLOR is set
	if !term.IsTerminal(int(os.Stdout.Fd())) || os.Getenv("NO_COLOR") != "" {
		Red = ""
		Green = ""
		Yellow = ""
		Blue = ""
		Cyan = ""
		Bold = ""
		Dim = ""
		Reset = ""
	}
}

// Header prints a boxed header
func Header(title string) {
	width := 58
	fmt.Println()
	fmt.Printf("╔%s╗\n", strings.Repeat("═", width))
	fmt.Printf("║ %s%-*s%s║\n", Bold, width-1, title, Reset)
	fmt.Printf("╚%s╝\n", strings.Repeat("═", width))
	fmt.Println()
}

// Section prints a section header
func Section(title string) {
	fmt.Println()
	fmt.Printf("%s%s%s\n", Bold, title, Reset)
	fmt.Println(strings.Repeat("─", len(title)))
}

// KV prints a key-value pair
func KV(key, value string) {
	fmt.Printf("  %s%-24s%s %s\n", Dim, key+":", Reset, value)
}

// TableHeader prints the table header
func TableHeader() {
	fmt.Println("┌─────────────────────────┬──────────────┬──────────────┐")
	fmt.Printf("│ %s%-23s%s │ %s%-12s%s │ %s%-12s%s │\n",
		Bold, "Setting", Reset,
		Bold, "Battery", Reset,
		Bold, "AC Power", Reset)
	fmt.Println("├─────────────────────────┼──────────────┼──────────────┤")
}

// TableRow prints a table row
func TableRow(setting, battery, ac string) {
	// Pad values accounting for invisible ANSI codes
	fmt.Printf("│ %-23s │ %s │ %s │\n", setting, padWithColor(battery, 12), padWithColor(ac, 12))
}

// padWithColor pads a string to width, accounting for ANSI escape codes
func padWithColor(s string, width int) string {
	// Calculate visible length (without ANSI codes)
	visible := visibleLen(s)
	if visible >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visible)
}

// visibleLen returns the visible length of a string (excluding ANSI codes)
func visibleLen(s string) int {
	inEscape := false
	count := 0
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		count++
	}
	return count
}

// TableSep prints a table separator
func TableSep() {
	fmt.Println("├─────────────────────────┼──────────────┼──────────────┤")
}

// TableFooter prints the table footer
func TableFooter() {
	fmt.Println("└─────────────────────────┴──────────────┴──────────────┘")
}

// FormatTime formats minutes as a readable time
func FormatTime(minutes int) string {
	if minutes == 0 {
		return Green + "Never" + Reset
	}
	if minutes == 1 {
		return "1 min"
	}
	return fmt.Sprintf("%d min", minutes)
}

// FormatBool formats a boolean as On/Off
func FormatBool(val bool) string {
	if val {
		return Green + "On" + Reset
	}
	return Dim + "Off" + Reset
}

// FormatPercent formats a percentage with color
func FormatPercent(val int) string {
	if val >= 80 {
		return fmt.Sprintf("%s%d%%%s", Green, val, Reset)
	} else if val >= 40 {
		return fmt.Sprintf("%s%d%%%s", Yellow, val, Reset)
	}
	return fmt.Sprintf("%s%d%%%s", Red, val, Reset)
}

// Success prints a success message
func Success(msg string) {
	fmt.Printf("%s✓%s %s\n", Green, Reset, msg)
}

// Error prints an error message
func Error(msg string) {
	fmt.Fprintf(os.Stderr, "%s✗%s %s\n", Red, Reset, msg)
}

// Warning prints a warning message
func Warning(msg string) {
	fmt.Printf("%s!%s %s\n", Yellow, Reset, msg)
}

// Info prints an info message
func Info(msg string) {
	fmt.Printf("%sℹ%s %s\n", Blue, Reset, msg)
}
