package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/born1337/macpwr/internal/assertions"
	"github.com/born1337/macpwr/internal/battery"
	"github.com/born1337/macpwr/internal/caffeinate"
	"github.com/born1337/macpwr/internal/display"
	"github.com/born1337/macpwr/internal/presets"
	"github.com/born1337/macpwr/internal/profiles"
	"github.com/born1337/macpwr/internal/settings"
	"github.com/born1337/macpwr/internal/thermal"

	"github.com/spf13/cobra"
)

const version = "1.0.0"

func main() {
	rootCmd := &cobra.Command{
		Use:     "macpwr",
		Short:   "macOS Power Management CLI",
		Long:    "A comprehensive tool for managing macOS power settings, battery info, sleep behavior, and power profiles.",
		Version: version,
		Run:     runStatus,
	}

	// Add subcommands
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(showCmd())
	rootCmd.AddCommand(setCmd())
	rootCmd.AddCommand(batteryCmd())
	rootCmd.AddCommand(presetCmd())
	rootCmd.AddCommand(profileCmd())
	rootCmd.AddCommand(caffeinateCmd())
	rootCmd.AddCommand(assertionsCmd())
	rootCmd.AddCommand(thermalCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show quick power status",
		Run:   runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) {
	fmt.Printf("\n%smacpwr%s %sv%s%s\n\n", display.Bold, display.Reset, display.Dim, version, display.Reset)

	// Power source
	source := settings.GetPowerSource()
	if source == "AC Power" {
		fmt.Printf("  Power: %s⚡ AC Power%s\n", display.Green, display.Reset)
	} else {
		fmt.Printf("  Power: %s🔋 Battery%s\n", display.Yellow, display.Reset)
	}

	// Battery info
	if info, err := battery.GetInfo(); err == nil && info != nil {
		percent := info.ChargePercent()
		health := info.HealthPercent()

		statusIcon := ""
		if info.IsCharging {
			statusIcon = " (charging)"
		}

		fmt.Printf("  Battery: %s%s\n", display.FormatPercent(percent), statusIcon)
		fmt.Printf("  Health: %s (%d cycles)\n", display.FormatPercent(health), info.CycleCount)
	}

	// AC settings summary
	if s, err := settings.Get(); err == nil {
		fmt.Printf("\n  %sAC Settings:%s Display %s, Sleep %s\n",
			display.Dim, display.Reset,
			display.FormatTime(s.AC.DisplaySleep),
			display.FormatTime(s.AC.SystemSleep))
	}

	fmt.Printf("\n%sCommands: show, set, battery, preset, profile, caffeinate, assertions, thermal%s\n", display.Dim, display.Reset)
	fmt.Printf("%sRun 'macpwr help' for more information%s\n\n", display.Dim, display.Reset)
}

func showCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display detailed power settings table",
		Run: func(cmd *cobra.Command, args []string) {
			display.Header("macOS Power Settings")

			s, err := settings.Get()
			if err != nil {
				display.Error("Failed to read power settings: " + err.Error())
				return
			}

			display.TableHeader()
			display.TableRow("Display Sleep",
				display.FormatTime(s.Battery.DisplaySleep),
				display.FormatTime(s.AC.DisplaySleep))
			display.TableRow("System Sleep",
				display.FormatTime(s.Battery.SystemSleep),
				display.FormatTime(s.AC.SystemSleep))
			display.TableRow("Disk Sleep",
				display.FormatTime(s.Battery.DiskSleep),
				display.FormatTime(s.AC.DiskSleep))
			display.TableSep()
			display.TableRow("Power Nap",
				display.FormatBool(s.Battery.PowerNap),
				display.FormatBool(s.AC.PowerNap))
			display.TableRow("Wake on LAN",
				display.FormatBool(s.Battery.WakeOnLAN),
				display.FormatBool(s.AC.WakeOnLAN))
			display.TableRow("Low Power Mode",
				display.FormatBool(s.Battery.LowPowerMode),
				display.FormatBool(s.AC.LowPowerMode))
			display.TableRow("TCP Keep Alive",
				display.FormatBool(s.Battery.TCPKeepAlive),
				display.FormatBool(s.AC.TCPKeepAlive))
			display.TableFooter()
			fmt.Println()
		},
	}
}

func setCmd() *cobra.Command {
	var ac, bat bool
	var displaySleep, systemSleep, diskSleep int

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Change power settings",
		Long: `Change power settings for AC or battery power.

Examples:
  macpwr set -a -d 60 -s 60     Set AC: display & sleep to 60 min
  macpwr set -b -d 10 -s 5      Set Battery: display 10, sleep 5 min
  macpwr set --ac --display 0   Set AC display to never sleep`,
		Run: func(cmd *cobra.Command, args []string) {
			// Default to AC if neither specified
			if !ac && !bat {
				ac = true
			}

			dFlag := cmd.Flags().Changed("display")
			sFlag := cmd.Flags().Changed("sleep")
			kFlag := cmd.Flags().Changed("disk")

			if !dFlag && !sFlag && !kFlag {
				display.Error("No settings specified. Use -d, -s, or -k to set values.")
				cmd.Help()
				return
			}

			opts := settings.SetOptions{}
			if ac {
				opts.PowerSource = "-c"
			} else {
				opts.PowerSource = "-b"
			}

			if dFlag {
				opts.DisplaySleep = &displaySleep
			}
			if sFlag {
				opts.SystemSleep = &systemSleep
			}
			if kFlag {
				opts.DiskSleep = &diskSleep
			}

			sourceName := "AC Power"
			if bat {
				sourceName = "Battery"
			}

			fmt.Printf("\nApplying to %s%s%s:\n", display.Bold, sourceName, display.Reset)
			if dFlag {
				fmt.Printf("  • Display sleep: %s\n", display.FormatTime(displaySleep))
			}
			if sFlag {
				fmt.Printf("  • System sleep: %s\n", display.FormatTime(systemSleep))
			}
			if kFlag {
				fmt.Printf("  • Disk sleep: %s\n", display.FormatTime(diskSleep))
			}
			fmt.Println()

			if err := settings.Set(opts); err != nil {
				display.Error("Failed to apply settings: " + err.Error())
				return
			}

			display.Success("Settings applied successfully!")
			fmt.Println()
		},
	}

	cmd.Flags().BoolVarP(&ac, "ac", "a", false, "Apply to AC power")
	cmd.Flags().BoolVarP(&bat, "battery", "b", false, "Apply to battery power")
	cmd.Flags().IntVarP(&displaySleep, "display", "d", 0, "Set display sleep (0 = never)")
	cmd.Flags().IntVarP(&systemSleep, "sleep", "s", 0, "Set system sleep (0 = never)")
	cmd.Flags().IntVarP(&diskSleep, "disk", "k", 0, "Set disk sleep (0 = never)")

	return cmd
}

func batteryCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "battery",
		Aliases: []string{"batt"},
		Short:   "Show detailed battery information",
		Run: func(cmd *cobra.Command, args []string) {
			display.Header("Battery Information")

			info, err := battery.GetInfo()
			if err != nil {
				display.Error("Failed to read battery info: " + err.Error())
				return
			}
			if info == nil {
				display.Error("No battery found (desktop Mac?)")
				return
			}

			display.Section("Charge")
			display.KV("Level", display.FormatPercent(info.ChargePercent()))

			status := info.Status()
			switch status {
			case "Fully Charged":
				display.KV("Status", display.Green+status+display.Reset)
			case "Charging":
				display.KV("Status", display.Yellow+status+display.Reset)
			case "On AC Power":
				display.KV("Status", display.Cyan+status+display.Reset)
			default:
				display.KV("Status", display.Blue+status+display.Reset)
			}

			display.KV("Time", info.TimeRemainingFormatted())
			display.KV("Current Capacity", fmt.Sprintf("%d mAh", info.ActualCurrentCapacity()))
			display.KV("Max Capacity", fmt.Sprintf("%d mAh", info.ActualMaxCapacity()))

			display.Section("Health")
			display.KV("Health", display.FormatPercent(info.HealthPercent()))
			display.KV("Cycle Count", strconv.Itoa(info.CycleCount))
			display.KV("Design Capacity", fmt.Sprintf("%d mAh", info.DesignCapacity))

			display.Section("Details")
			display.KV("Temperature", fmt.Sprintf("%d°C", info.TemperatureCelsius()))
			display.KV("AC Connected", display.FormatBool(info.ExternalConnected))

			fmt.Println()
		},
	}
}

func presetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preset [name]",
		Short: "Apply built-in power presets",
		Long: `Apply a built-in power preset.

Available presets:
  default         Restore macOS default settings
  presentation    Keep screen on, prevent sleep
  battery-saver   Aggressive power saving
  performance     Maximum performance, no sleep
  movie           Screen stays on, system can sleep`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 || args[0] == "list" {
				fmt.Printf("\n%sAvailable Presets%s\n\n", display.Bold, display.Reset)
				for _, p := range presets.All {
					fmt.Printf("  %s%-16s%s %s\n", display.Cyan, p.Name, display.Reset, p.Description)
				}
				fmt.Println()
				return
			}

			name := args[0]
			p := presets.Get(name)
			if p == nil {
				display.Error("Unknown preset: " + name)
				fmt.Println()
				cmd.Run(cmd, []string{"list"})
				return
			}

			fmt.Printf("\nApplying preset: %s%s%s\n", display.Bold, p.Name, display.Reset)
			fmt.Printf("%s%s%s\n\n", display.Dim, p.Description, display.Reset)

			fmt.Println("Setting AC Power...")
			fmt.Println("Setting Battery Power...")

			if err := presets.Apply(p); err != nil {
				display.Error("Failed to apply preset: " + err.Error())
				return
			}

			fmt.Println()
			display.Success(fmt.Sprintf("Preset '%s' applied successfully!", p.Name))
			fmt.Println()
		},
	}

	return cmd
}

func profileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Save/load custom power profiles",
	}

	saveCmd := &cobra.Command{
		Use:   "save [name]",
		Short: "Save current settings as a profile",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			if err := profiles.Save(name); err != nil {
				display.Error(err.Error())
				return
			}
			display.Success(fmt.Sprintf("Profile '%s' saved to %s/%s.profile", name, profiles.ProfilesDir(), name))
		},
	}

	loadCmd := &cobra.Command{
		Use:   "load [name]",
		Short: "Load and apply a saved profile",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			p, err := profiles.Load(name)
			if err != nil {
				display.Error("Profile not found: " + name)
				return
			}

			fmt.Printf("\nLoading profile: %s%s%s\n\n", display.Bold, name, display.Reset)
			fmt.Println("Setting Battery Power...")
			fmt.Println("Setting AC Power...")

			if err := profiles.Apply(p); err != nil {
				display.Error(err.Error())
				return
			}

			fmt.Println()
			display.Success(fmt.Sprintf("Profile '%s' loaded successfully!", name))
			fmt.Println()
		},
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all saved profiles",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("\n%sSaved Profiles%s\n\n", display.Bold, display.Reset)

			list, _ := profiles.List()
			if len(list) == 0 {
				fmt.Printf("  %sNo profiles saved yet%s\n\n", display.Dim, display.Reset)
				fmt.Println("  Save current settings with: macpwr profile save <name>")
			} else {
				for _, p := range list {
					fmt.Printf("  %s%-20s%s %s%s%s\n",
						display.Cyan, p.Name, display.Reset,
						display.Dim, p.Created.Format(time.RFC1123), display.Reset)
				}
			}
			fmt.Println()
		},
	}

	deleteCmd := &cobra.Command{
		Use:     "delete [name]",
		Aliases: []string{"rm"},
		Short:   "Delete a saved profile",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			if !profiles.Exists(name) {
				display.Error("Profile not found: " + name)
				return
			}
			if err := profiles.Delete(name); err != nil {
				display.Error(err.Error())
				return
			}
			display.Success(fmt.Sprintf("Profile '%s' deleted", name))
		},
	}

	cmd.AddCommand(saveCmd, loadCmd, listCmd, deleteCmd)
	return cmd
}

func caffeinateCmd() *cobra.Command {
	var duration int
	var displayOnly, idleOnly, systemOnly bool

	cmd := &cobra.Command{
		Use:   "caffeinate [-- command]",
		Aliases: []string{"cafe"},
		Short: "Prevent system from sleeping",
		Long: `Prevent the system from sleeping.

Examples:
  macpwr caffeinate              Prevent sleep until Ctrl+C
  macpwr caffeinate -t 60        Prevent sleep for 60 minutes
  macpwr caffeinate -d           Only prevent display sleep
  macpwr caffeinate -- make      Run 'make' without sleeping`,
		Run: func(cmd *cobra.Command, args []string) {
			opts := caffeinate.Options{
				Duration:    duration,
				DisplayOnly: displayOnly,
				IdleOnly:    idleOnly,
				SystemOnly:  systemOnly,
				Command:     args,
			}

			fmt.Println()
			if len(args) > 0 {
				fmt.Printf("%sRunning command while preventing sleep...%s\n", display.Bold, display.Reset)
				fmt.Printf("%sCommand: %s%s\n\n", display.Dim, args, display.Reset)
			} else if duration > 0 {
				fmt.Printf("%sPreventing sleep for %d minutes%s\n", display.Bold, duration, display.Reset)
				fmt.Printf("%sPress Ctrl+C to cancel early%s\n\n", display.Dim, display.Reset)
			} else {
				fmt.Printf("%sPreventing sleep indefinitely%s\n", display.Bold, display.Reset)
				fmt.Printf("%sPress Ctrl+C to stop%s\n\n", display.Dim, display.Reset)
			}

			if err := caffeinate.Run(opts); err != nil {
				display.Error(err.Error())
				return
			}

			if len(args) > 0 {
				display.Success("Command completed")
			} else {
				display.Success("Caffeinate completed!")
			}
			fmt.Println()
		},
	}

	cmd.Flags().IntVarP(&duration, "time", "t", 0, "Prevent sleep for specified minutes")
	cmd.Flags().BoolVarP(&displayOnly, "display", "d", false, "Only prevent display sleep")
	cmd.Flags().BoolVarP(&idleOnly, "idle", "i", false, "Only prevent idle sleep")
	cmd.Flags().BoolVarP(&systemOnly, "system", "s", false, "Only prevent system sleep")

	return cmd
}

func assertionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "assertions",
		Aliases: []string{"assert"},
		Short:   "Show what's preventing sleep",
		Run: func(cmd *cobra.Command, args []string) {
			display.Header("Power Assertions")

			info, err := assertions.Get()
			if err != nil {
				display.Error("Failed to read assertions: " + err.Error())
				return
			}

			display.Section("Assertion Summary")
			display.KV("Prevent User Idle Display", fmt.Sprintf("%d active", info.Summary.PreventUserIdleDisplay))
			display.KV("Prevent System Sleep", fmt.Sprintf("%d active", info.Summary.PreventSystemSleep))
			display.KV("Prevent Display Sleep", fmt.Sprintf("%d active", info.Summary.PreventDisplaySleep))
			display.KV("External Media", fmt.Sprintf("%d active", info.Summary.ExternalMedia))
			display.KV("Network Client Active", fmt.Sprintf("%d active", info.Summary.NetworkClientActive))

			display.Section("Active Assertions")
			if len(info.Active) == 0 {
				fmt.Printf("  %sNo active power assertions%s\n", display.Green, display.Reset)
				fmt.Printf("  %sSystem is free to sleep normally%s\n", display.Dim, display.Reset)
			} else {
				for _, a := range info.Active {
					fmt.Printf("  %s%-20s%s %s(PID %d)%s\n",
						display.Cyan, a.Process, display.Reset,
						display.Dim, a.PID, display.Reset)
					fmt.Printf("    └─ %s\n", a.Type)
				}
			}

			display.Section("Scheduled Events")
			if len(info.Scheduled) == 0 {
				fmt.Printf("  %sNo scheduled wake/sleep events%s\n", display.Dim, display.Reset)
			} else {
				for _, e := range info.Scheduled {
					fmt.Printf("  • %s\n", e.Description)
				}
			}

			fmt.Println()
		},
	}
}

func thermalCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "thermal",
		Aliases: []string{"temp"},
		Short:   "Show thermal and CPU information",
		Run: func(cmd *cobra.Command, args []string) {
			display.Header("Thermal Information")

			info, err := thermal.Get()
			if err != nil {
				display.Error("Failed to read thermal info: " + err.Error())
				return
			}

			display.Section("CPU")
			display.KV("Model", info.CPUModel)
			display.KV("Cores", strconv.Itoa(info.CPUCores))
			display.KV("Architecture", info.Architecture)

			display.Section("Fans")
			if info.FansAvailable {
				fmt.Println("  Fans detected (detailed RPM requires powermetrics)")
			} else {
				fmt.Printf("  %sFan information not available%s\n", display.Dim, display.Reset)
				fmt.Printf("  %s(May require additional tools or root access)%s\n", display.Dim, display.Reset)
			}

			display.Section("System Load")
			display.KV("Load Average", info.LoadAverage)
			if info.MemoryFree > 0 {
				display.KV("Memory Free", fmt.Sprintf("%d%%", info.MemoryFree))
			}

			display.Section("Power")
			if info.PowerSource == "AC Power" {
				display.KV("Power Source", display.Green+"AC Power"+display.Reset)
			} else if info.PowerSource == "Battery" {
				display.KV("Power Source", display.Yellow+"Battery"+display.Reset)
			} else {
				display.KV("Power Source", info.PowerSource)
			}

			if info.CPULimit > 0 {
				if info.CPULimit == 100 {
					display.KV("CPU Limit", display.Green+"100% (No throttling)"+display.Reset)
				} else {
					display.KV("CPU Limit", fmt.Sprintf("%s%d%% (Throttled)%s", display.Yellow, info.CPULimit, display.Reset))
				}
			}

			fmt.Printf("\n%sNote: Detailed thermal data requires 'sudo powermetrics'%s\n\n", display.Dim, display.Reset)
		},
	}
}
