package caffeinate

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

// Options for caffeinate
type Options struct {
	Duration     int      // Duration in minutes (0 = indefinite)
	DisplayOnly  bool     // Only prevent display sleep
	IdleOnly     bool     // Only prevent idle sleep
	SystemOnly   bool     // Only prevent system sleep
	Command      []string // Command to run while preventing sleep
}

// Run executes caffeinate with the given options
func Run(opts Options) error {
	args := buildArgs(opts)

	if len(opts.Command) > 0 {
		// Run command while preventing sleep
		args = append(args, opts.Command...)
		cmd := exec.Command("caffeinate", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	if opts.Duration > 0 {
		// Run for specified duration with countdown
		return runWithCountdown(args, opts.Duration)
	}

	// Run indefinitely
	return runIndefinitely(args)
}

func buildArgs(opts Options) []string {
	var args []string

	if opts.DisplayOnly {
		args = append(args, "-d")
	}
	if opts.IdleOnly {
		args = append(args, "-i")
	}
	if opts.SystemOnly {
		args = append(args, "-s")
	}

	// Default: prevent all sleep types
	if len(args) == 0 {
		args = []string{"-dims"}
	}

	if opts.Duration > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", opts.Duration*60))
	}

	return args
}

func runWithCountdown(args []string, minutes int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	cmd := exec.CommandContext(ctx, "caffeinate", args...)
	if err := cmd.Start(); err != nil {
		return err
	}

	// Show countdown
	remaining := minutes
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	fmt.Printf("\r  Time remaining: %d min  ", remaining)

	for remaining > 0 {
		select {
		case <-ctx.Done():
			fmt.Println("\n\nCaffeinate cancelled.")
			cmd.Process.Kill()
			return nil
		case <-ticker.C:
			remaining--
			fmt.Printf("\r  Time remaining: %d min  ", remaining)
		}
	}

	fmt.Println("\n")
	return cmd.Wait()
}

func runIndefinitely(args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	cmd := exec.CommandContext(ctx, "caffeinate", args...)
	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		<-sigChan
		fmt.Println("\nCaffeinate stopped.")
		cancel()
	}()

	return cmd.Wait()
}
