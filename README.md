# macpwr

A fast, native CLI tool for managing macOS power settings, battery info, sleep behavior, and power profiles.

**This is the Go implementation** - significantly faster than the bash version (~10-30x speedup).

## Requirements

- macOS 10.15 or later
- Go 1.21+ (for building)

## Installation

### Build from source

```bash
# Clone the repository
git clone https://github.com/born1337/macpwr.git
cd macpwr

# Build
make build

# Install system-wide
sudo make install
```

### Or install directly with Go

```bash
go install github.com/born1337/macpwr/cmd/macpwr@latest
```

## Usage

### Quick Status

```bash
macpwr              # Show quick power status (default)
```

### View Settings

```bash
macpwr show         # Display detailed power settings table
```

### Change Settings

```bash
# Set AC power display and system sleep to 60 minutes
macpwr set -a -d 60 -s 60

# Set battery display sleep to 10 minutes
macpwr set -b -d 10

# Set AC power to never sleep
macpwr set -a -d 0 -s 0
```

### Battery Information

```bash
macpwr battery      # Show detailed battery info
```

### Presets

```bash
macpwr preset list              # List available presets
macpwr preset presentation      # Keep screen on for presentations
macpwr preset battery-saver     # Maximize battery life
macpwr preset performance       # No sleep restrictions
macpwr preset movie             # Screen on, system can sleep
macpwr preset default           # Restore macOS defaults
```

### Custom Profiles

```bash
macpwr profile save work        # Save current settings as 'work'
macpwr profile load work        # Load the 'work' profile
macpwr profile list             # List all saved profiles
macpwr profile delete work      # Delete 'work' profile
```

### Prevent Sleep (Caffeinate)

```bash
macpwr caffeinate               # Prevent sleep until Ctrl+C
macpwr caffeinate -t 60         # Prevent sleep for 60 minutes
macpwr caffeinate -d            # Only prevent display sleep
macpwr caffeinate -- make build # Run command without sleeping
```

### Power Assertions

```bash
macpwr assertions   # Show what's preventing sleep
```

### Thermal Information

```bash
macpwr thermal      # Show CPU and thermal info
```

## Commands Reference

| Command | Description |
|---------|-------------|
| `status` | Show quick power status (default) |
| `show` | Display detailed power settings table |
| `set` | Change power settings |
| `battery` | Show detailed battery information |
| `preset` | Apply built-in power presets |
| `profile` | Save/load custom power profiles |
| `caffeinate` | Prevent system from sleeping |
| `assertions` | Show what's preventing sleep |
| `thermal` | Show thermal and CPU information |
| `help` | Show help message |
| `version` | Show version |

## Set Options

| Option | Description |
|--------|-------------|
| `-a, --ac` | Apply to AC power (default) |
| `-b, --battery` | Apply to battery power |
| `-d, --display <min>` | Set display sleep (0 = never) |
| `-s, --sleep <min>` | Set system sleep (0 = never) |
| `-k, --disk <min>` | Set disk sleep (0 = never) |

## Performance

The Go implementation is significantly faster than the bash version:

| Command | Bash | Go | Speedup |
|---------|------|-----|---------|
| `status` | ~80ms | ~8ms | **10x** |
| `show` | ~100ms | ~5ms | **20x** |
| `battery` | ~100ms | ~4ms | **25x** |
| `help` | ~10ms | ~1ms | **10x** |

## Project Structure

```
macpwr/
├── cmd/macpwr/          # Main CLI application
├── internal/
│   ├── display/         # Terminal formatting
│   ├── battery/         # Battery info
│   ├── settings/        # Power settings
│   ├── presets/         # Built-in presets
│   ├── profiles/        # Custom profiles
│   ├── caffeinate/      # Sleep prevention
│   ├── assertions/      # Power assertions
│   └── thermal/         # Thermal info
├── completions/         # Shell completions
├── go.mod
├── Makefile
└── README.md
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
