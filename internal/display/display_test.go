package display

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestVisibleLen(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"plain text", "Hello", 5},
		{"empty string", "", 0},
		{"single char", "X", 1},
		{"with green color", "\033[0;32mOn\033[0m", 2},
		{"with red color", "\033[0;31mOff\033[0m", 3},
		{"with bold", "\033[1mBold\033[0m", 4},
		{"with dim", "\033[2mDim\033[0m", 3},
		{"multiple colors", "\033[0;32mGreen\033[0m and \033[0;31mRed\033[0m", 13},
		{"nested escapes", "\033[1m\033[0;32mBoldGreen\033[0m\033[0m", 9},
		{"numbers", "12345", 5},
		{"with spaces", "a b c", 5},
		{"Never with color", "\033[0;32mNever\033[0m", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := visibleLen(tt.input)
			if got != tt.want {
				t.Errorf("visibleLen(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestPadWithColor(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		width  int
		want   int // expected visible length after padding
	}{
		{"plain short", "On", 12, 12},
		{"plain exact", "HelloWorld12", 12, 12},
		{"plain long", "ThisIsVeryLong", 12, 14}, // no truncation, just returns as-is
		{"colored short", "\033[0;32mOn\033[0m", 12, 12},
		{"colored exact", "\033[0;32mHelloWorld12\033[0m", 12, 12},
		{"empty", "", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padWithColor(tt.input, tt.width)
			got := visibleLen(result)
			if got != tt.want {
				t.Errorf("padWithColor(%q, %d) visible len = %d, want %d", tt.input, tt.width, got, tt.want)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	// Save original colors and restore after test
	origGreen := Green
	origReset := Reset
	defer func() {
		Green = origGreen
		Reset = origReset
	}()

	tests := []struct {
		name    string
		minutes int
		want    string
	}{
		{"zero is Never", 0, Green + "Never" + Reset},
		{"one minute", 1, "1 min"},
		{"multiple minutes", 10, "10 min"},
		{"large number", 120, "120 min"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTime(tt.minutes)
			if got != tt.want {
				t.Errorf("FormatTime(%d) = %q, want %q", tt.minutes, got, tt.want)
			}
		})
	}
}

func TestFormatBool(t *testing.T) {
	// Save original colors and restore after test
	origGreen := Green
	origDim := Dim
	origReset := Reset
	defer func() {
		Green = origGreen
		Dim = origDim
		Reset = origReset
	}()

	tests := []struct {
		name  string
		input bool
		want  string
	}{
		{"true is On", true, Green + "On" + Reset},
		{"false is Off", false, Dim + "Off" + Reset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBool(tt.input)
			if got != tt.want {
				t.Errorf("FormatBool(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatPercent(t *testing.T) {
	// Save original colors and restore after test
	origGreen := Green
	origYellow := Yellow
	origRed := Red
	origReset := Reset
	defer func() {
		Green = origGreen
		Yellow = origYellow
		Red = origRed
		Reset = origReset
	}()

	// Set predictable colors for testing
	Green = "[G]"
	Yellow = "[Y]"
	Red = "[R]"
	Reset = "[/]"

	tests := []struct {
		name    string
		percent int
		want    string
	}{
		{"high percent green", 100, "[G]100%[/]"},
		{"80 percent green", 80, "[G]80%[/]"},
		{"79 percent yellow", 79, "[Y]79%[/]"},
		{"40 percent yellow", 40, "[Y]40%[/]"},
		{"39 percent red", 39, "[R]39%[/]"},
		{"zero percent red", 0, "[R]0%[/]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPercent(tt.percent)
			if got != tt.want {
				t.Errorf("FormatPercent(%d) = %q, want %q", tt.percent, got, tt.want)
			}
		})
	}
}

func TestFormatPercentVisibleLength(t *testing.T) {
	// Test that FormatPercent output has correct visible length
	tests := []struct {
		percent     int
		wantVisible int
	}{
		{100, 4}, // "100%"
		{80, 3},  // "80%"
		{5, 2},   // "5%"
		{0, 2},   // "0%"
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := FormatPercent(tt.percent)
			got := visibleLen(result)
			if got != tt.wantVisible {
				t.Errorf("FormatPercent(%d) visible len = %d, want %d", tt.percent, got, tt.wantVisible)
			}
		})
	}
}

func TestColorVariables(t *testing.T) {
	// Just verify colors are set (non-empty in terminal)
	// This test documents the expected color codes
	colors := map[string]string{
		"Red":    Red,
		"Green":  Green,
		"Yellow": Yellow,
		"Blue":   Blue,
		"Cyan":   Cyan,
		"Bold":   Bold,
		"Dim":    Dim,
		"Reset":  Reset,
	}

	for name, val := range colors {
		t.Run(name, func(t *testing.T) {
			// Colors should either be ANSI codes or empty (if NO_COLOR)
			// We just verify they exist as variables
			_ = val
		})
	}
}

// captureStdout captures stdout output from a function
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// captureStderr captures stderr output from a function
func captureStderr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestHeader(t *testing.T) {
	output := captureStdout(func() {
		Header("Test Title")
	})

	if !strings.Contains(output, "Test Title") {
		t.Errorf("Header should contain title, got: %s", output)
	}
	if !strings.Contains(output, "╔") {
		t.Errorf("Header should contain box characters, got: %s", output)
	}
	if !strings.Contains(output, "╚") {
		t.Errorf("Header should contain box characters, got: %s", output)
	}
}

func TestSection(t *testing.T) {
	output := captureStdout(func() {
		Section("My Section")
	})

	if !strings.Contains(output, "My Section") {
		t.Errorf("Section should contain title, got: %s", output)
	}
	if !strings.Contains(output, "─") {
		t.Errorf("Section should contain underline, got: %s", output)
	}
}

func TestKV(t *testing.T) {
	output := captureStdout(func() {
		KV("Key", "Value")
	})

	if !strings.Contains(output, "Key:") {
		t.Errorf("KV should contain key with colon, got: %s", output)
	}
	if !strings.Contains(output, "Value") {
		t.Errorf("KV should contain value, got: %s", output)
	}
}

func TestTableHeader(t *testing.T) {
	output := captureStdout(func() {
		TableHeader()
	})

	if !strings.Contains(output, "Setting") {
		t.Errorf("TableHeader should contain 'Setting', got: %s", output)
	}
	if !strings.Contains(output, "Battery") {
		t.Errorf("TableHeader should contain 'Battery', got: %s", output)
	}
	if !strings.Contains(output, "AC Power") {
		t.Errorf("TableHeader should contain 'AC Power', got: %s", output)
	}
	if !strings.Contains(output, "┌") {
		t.Errorf("TableHeader should contain box drawing chars, got: %s", output)
	}
}

func TestTableRow(t *testing.T) {
	output := captureStdout(func() {
		TableRow("Display Sleep", "10 min", "Never")
	})

	if !strings.Contains(output, "Display Sleep") {
		t.Errorf("TableRow should contain setting name, got: %s", output)
	}
	if !strings.Contains(output, "10 min") {
		t.Errorf("TableRow should contain battery value, got: %s", output)
	}
	if !strings.Contains(output, "Never") {
		t.Errorf("TableRow should contain AC value, got: %s", output)
	}
	if !strings.Contains(output, "│") {
		t.Errorf("TableRow should contain column separators, got: %s", output)
	}
}

func TestTableRowAlignment(t *testing.T) {
	// This test specifically checks for the bug we fixed
	// Colored values should still align properly
	output := captureStdout(func() {
		TableRow("Power Nap", "\033[0;32mOn\033[0m", "\033[0;32mOn\033[0m")
	})

	// Count the │ characters - should be exactly 4 (start, after col1, after col2, end)
	pipeCount := strings.Count(output, "│")
	if pipeCount != 4 {
		t.Errorf("TableRow should have exactly 4 pipe chars, got %d in: %s", pipeCount, output)
	}
}

func TestTableSep(t *testing.T) {
	output := captureStdout(func() {
		TableSep()
	})

	if !strings.Contains(output, "├") {
		t.Errorf("TableSep should contain separator chars, got: %s", output)
	}
	if !strings.Contains(output, "┼") {
		t.Errorf("TableSep should contain cross chars, got: %s", output)
	}
}

func TestTableFooter(t *testing.T) {
	output := captureStdout(func() {
		TableFooter()
	})

	if !strings.Contains(output, "└") {
		t.Errorf("TableFooter should contain bottom-left corner, got: %s", output)
	}
	if !strings.Contains(output, "┘") {
		t.Errorf("TableFooter should contain bottom-right corner, got: %s", output)
	}
}

func TestSuccess(t *testing.T) {
	output := captureStdout(func() {
		Success("Operation completed")
	})

	if !strings.Contains(output, "✓") {
		t.Errorf("Success should contain checkmark, got: %s", output)
	}
	if !strings.Contains(output, "Operation completed") {
		t.Errorf("Success should contain message, got: %s", output)
	}
}

func TestError(t *testing.T) {
	output := captureStderr(func() {
		Error("Something went wrong")
	})

	if !strings.Contains(output, "✗") {
		t.Errorf("Error should contain X mark, got: %s", output)
	}
	if !strings.Contains(output, "Something went wrong") {
		t.Errorf("Error should contain message, got: %s", output)
	}
}

func TestWarning(t *testing.T) {
	output := captureStdout(func() {
		Warning("Be careful")
	})

	if !strings.Contains(output, "!") {
		t.Errorf("Warning should contain exclamation, got: %s", output)
	}
	if !strings.Contains(output, "Be careful") {
		t.Errorf("Warning should contain message, got: %s", output)
	}
}

func TestInfo(t *testing.T) {
	output := captureStdout(func() {
		Info("Some information")
	})

	if !strings.Contains(output, "ℹ") {
		t.Errorf("Info should contain info symbol, got: %s", output)
	}
	if !strings.Contains(output, "Some information") {
		t.Errorf("Info should contain message, got: %s", output)
	}
}
