# B2Sync - Automated Backblaze Backup Utility

A background service for macOS that automatically syncs local folders to Backblaze B2 buckets using the B2 command-line tool.

## Features

- **Automated syncing**: Syncs local directories to B2 buckets every 10 minutes (configurable)
- **Background operation**: Runs silently as a macOS LaunchAgent
- **Process management**: Prevents concurrent sync operations using PID files
- **Comprehensive logging**: Structured logs with rotation and multiple levels
- **Native notifications**: macOS notifications for sync results and errors
- **Flexible configuration**: JSON-based configuration for multiple sync pairs

## Prerequisites

1. **Backblaze B2 CLI**: Install and configure the B2 command-line tool so as to have [the `sync` command](https://b2-command-line-tool.readthedocs.io/en/master/subcommands/sync.html) available.

   ```bash
   # Install via Homebrew
   brew install b2-tools

   # Or download from: https://github.com/Backblaze/B2_Command_Line_Tool

   # Configure with your credentials
   b2 authorize-account <applicationKeyId> <applicationKey>
   ```

2. **terminal-notifier**: Required for macOS notifications.

   ```bash
   # Install via Homebrew
   brew install terminal-notifier
   ```

3. **Go 1.19+**: Required for building the application

4. **Backblaze B2 Account**: Active account with configured buckets

## Installation

### Step 1: Build the Application

```bash
# Clone or download the source code
cd b2sync

# Build the binary
go build -o b2sync cmd/b2sync/main.go
```

### Step 2: Configure

1. Copy the example configuration:

   ```bash
   mkdir -p ~/.config/b2sync
   cp configs/config.example.json ~/.config/b2sync/config.json
   ```

2. Edit the configuration file:

   ```bash
   nano ~/.config/b2sync/config.json
   ```

3. Update the sync pairs with your actual directories and B2 bucket paths:

   ```json
   {
     "sync_pairs": [
       {
         "source": "/Users/yourusername/Pictures",
         "destination": "b2://your-bucket-name/Pictures"
       },
       {
         "source": "/Users/yourusername/Documents",
         "destination": "b2://your-bucket-name/Documents"
       }
     ],
     "sync_frequency": "10m",
     "notification_threshold": 5,
     "log_level": "INFO",
     "log_dir": "/Users/yourusername/Library/Logs/b2sync"
   }
   ```

### Step 3: Install as LaunchAgent (Auto-start)

Use the provided installation script:

```bash
# Make the script executable
chmod +x scripts/install-launchd.sh

# Run the installation script
./scripts/install-launchd.sh
```

The script will:

- Install the binary to `/usr/local/bin/b2sync`
- Create a LaunchAgent plist file
- Load the service to start automatically

## Configuration Options

Configuration options for `~/.config/b2sync/config.json`

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `sync_pairs` | Array | - | Source/destination folder pairs to sync |
| `sync_frequency` | Duration | `"10m"` | How often to run sync (e.g., "5m", "1h") |
| `notification_threshold` | Number | `5` | Minimum files to trigger success notification |
| `log_level` | String | `"INFO"` | Log level: DEBUG, INFO, WARN, ERROR |
| `log_dir` | String | `~/Library/Logs/b2sync` | Directory for log files |

### Sync Pair Format

```json
{
  "source": "/path/to/local/folder",
  "destination": "b2://bucket-name/remote/path"
}
```

- **Source**: Absolute path to local directory
- **Destination**: B2 bucket path in format `b2://bucket-name/path`

## Usage

### Running Manually

```bash
# Run once
./b2sync

# Run in foreground with logs
./b2sync 2>&1 | tee ~/b2sync.log
```

Remember you can always `tail` the logs while the app is running: `tail -n 50 -f /path/to/your/logfile.log`

### Managing the Service

```bash
# Start the service
launchctl load ~/Library/LaunchAgents/com.b2sync.agent.plist

# Stop the service
launchctl unload ~/Library/LaunchAgents/com.b2sync.agent.plist

# Check if running
launchctl list | grep b2sync

# Kill all b2sync processes (handles both built binary and go run)
pkill -f "b2sync\|cmd/b2sync/main.go"

# View service logs
tail -f ~/Library/Logs/b2sync/b2sync-$(date +%Y-%m-%d).log
```

## Logging

B2Sync creates detailed logs in `~/Library/Logs/b2sync/`:

- **Application logs**: `b2sync-YYYY-MM-DD.log` (rotated daily)
- **LaunchAgent logs**: `launchd.out.log` and `launchd.error.log`

Log levels available: DEBUG, INFO, WARN, ERROR

## Notifications

B2Sync sends macOS notifications for:

- **Sync completion**: When files synced ≥ notification threshold
- **Sync errors**: Any failed sync operations
- **Missing B2 CLI**: When b2 command is not found

## Troubleshooting

### B2 CLI Not Found

If you see "b2 CLI not found" errors:

1. Ensure B2 CLI is installed and in PATH
2. Check your PATH in the LaunchAgent environment
3. Verify B2 CLI works manually: `b2 version`

### Notifications Not Appearing

If macOS notifications are not showing:

1. Ensure terminal-notifier is installed: `brew install terminal-notifier`
2. Test notifications manually: `terminal-notifier -message 'Test' -title 'B2Sync Test'`
3. Check macOS notification permissions for the application

### Permission Issues

If sync fails with permission errors:

1. Ensure the source directories are readable
2. Verify B2 credentials are configured: `b2 get-account-info`
3. Test manual sync: `b2 sync /source/path b2://bucket/path --dryRun`

### High CPU Usage

If the service uses too much CPU:

1. Increase sync frequency in config
2. Reduce the number of sync pairs
3. Check for large files or many small files

### Logs Not Appearing

If logs are missing:

1. Check log directory permissions
2. Verify log_dir path in config is writable
3. Check LaunchAgent stderr: `~/Library/Logs/b2sync/launchd.error.log`

## Uninstallation

To completely remove B2Sync:

```bash
# Stop and unload the service
sudo launchctl unload ~/Library/LaunchAgents/com.b2sync.agent.plist

# Remove files
rm ~/Library/LaunchAgents/com.b2sync.agent.plist
sudo rm /usr/local/bin/b2sync
rm -rf ~/.config/b2sync
rm -rf ~/Library/Logs/b2sync
```

## Security Notes

- Configuration and log files are stored in user directories with appropriate permissions
- B2 credentials are managed by the B2 CLI, not stored by B2Sync
- PID files prevent multiple concurrent sync operations
- No network credentials are stored in configuration files

## Development

### Running Without Building

For development, you can run the project directly without building each time:

```bash
# Run directly from source (recommended for development)
go run cmd/b2sync/main.go

# Run in background for testing
go run cmd/b2sync/main.go &

# Stop background process
pkill -f "go run.*b2sync"

# Run with custom config path
CONFIG_PATH=/path/to/test-config.json go run cmd/b2sync/main.go
```

### Building

Only build when you need an actual binary for deployment:

```bash
# Build for macOS (current platform)
go build -o b2sync cmd/b2sync/main.go

# Build with optimizations for production
go build -ldflags="-s -w" -o b2sync cmd/b2sync/main.go

# Install to GOPATH/bin for system-wide access
go install cmd/b2sync/main.go

# Build for specific macOS architectures
GOOS=darwin GOARCH=amd64 go build -o b2sync-intel cmd/b2sync/main.go # Intel Macs
GOOS=darwin GOARCH=arm64 go build -o b2sync-m1 cmd/b2sync/main.go # Apple Silicon

# Build universal binary for macOS (both Intel and Apple Silicon)
GOOS=darwin GOARCH=amd64 go build -o b2sync-intel cmd/b2sync/main.go
GOOS=darwin GOARCH=arm64 go build -o b2sync-arm64 cmd/b2sync/main.go
lipo -create -output b2sync-universal b2sync-intel b2sync-arm64

# Verify the binary
file b2sync
./b2sync --version || echo "Binary created successfully"
```

### Development Workflow

```bash
# 1. Set up development config
mkdir -p ~/.config/b2sync-dev
cp configs/config.example.json ~/.config/b2sync-dev/config.json

# 2. Edit config for safe testing (use test directories/buckets)
nano ~/.config/b2sync-dev/config.json

# 3. Run with development config
CONFIG_PATH=~/.config/b2sync-dev/config.json go run cmd/b2sync/main.go

# 4. Test individual components
go run -tags debug cmd/b2sync/main.go
```

### Testing

```bash
# Basic functionality test
go run cmd/b2sync/main.go --help

# Verify configuration has loaded
go run cmd/b2sync/main.go --config-test

# Run with debug logging
LOG_LEVEL=DEBUG go run cmd/b2sync/main.go

# Test with minimal sync frequency for faster iteration
echo '{"sync_pairs":[{"source":"~/test","destination":"b2://test-bucket/test"}],"sync_frequency":"30s","log_level":"DEBUG"}' > /tmp/test-config.json
CONFIG_PATH=/tmp/test-config.json go run cmd/b2sync/main.go
```

### Project Structure

```
b2sync/
├── cmd/b2sync/main.go           # Main entry point
├── internal/
│   ├── config/config.go         # Configuration management
│   ├── sync/sync.go             # Sync logic and b2 command execution
│   ├── logger/logger.go         # Structured logging
│   └── notifier/notifier.go     # macOS notifications
├── configs/
│   └── config.example.json      # Example configuration
├── scripts/
│   └── install-launchd.sh       # Installation script for launchd
└── README.md                    # This file
```

## TODO

- [ ] Tests!
