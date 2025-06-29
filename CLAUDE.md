# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

B2Sync is a macOS background service written in Go that automatically syncs local directories to Backblaze B2 buckets using the B2 CLI. It runs as a LaunchAgent and provides structured logging, notifications, and process management.

## Development Commands

### Primary Development Workflow

```bash
# Development (recommended - no build required)
go run cmd/b2sync/main.go

# Development with custom config
CONFIG_PATH=/path/to/test-config.json go run cmd/b2sync/main.go

# Build for testing
go build -o b2sync cmd/b2sync/main.go

# Production build with optimizations
go build -ldflags="-s -w" -o b2sync cmd/b2sync/main.go
```

### macOS Service Management

```bash
# Install as LaunchAgent
chmod +x scripts/install-launchd.sh && ./scripts/install-launchd.sh

# Manual service control
launchctl load ~/Library/LaunchAgents/com.b2sync.agent.plist
launchctl unload ~/Library/LaunchAgents/com.b2sync.agent.plist
launchctl list | grep b2sync

# Kill all b2sync processes (handles both built binary and go run)
pkill -f "b2sync\|cmd/b2sync/main.go"
```

### Testing and Verification

```bash
# Basic functionality test
go run cmd/b2sync/main.go --help

# Test with debug logging
LOG_LEVEL=DEBUG go run cmd/b2sync/main.go

# View logs while running
tail -f ~/Library/Logs/b2sync/b2sync-$(date +%Y-%m-%d).log
```

## Architecture Overview

### Core Modules (`/internal/`)

1. **Config** (`config/config.go`): JSON configuration management with custom Duration type, default generation, and path resolution
2. **Sync** (`sync/sync.go`): B2 CLI wrapper with PID-based concurrency prevention, command execution, and result parsing
3. **Logger** (`logger/logger.go`): Structured logging with daily rotation, multiple levels (DEBUG/INFO/WARN/ERROR), and multi-writer support
4. **Notifier** (`notifier/notifier.go`): macOS native notifications via terminal-notifier for sync results and errors

### Entry Point

- **Main** (`cmd/b2sync/main.go`): Signal handling, ticker-based scheduling, and command-line interface

### Key Design Patterns

- **Service-oriented architecture** with clean separation of concerns
- **Dependency injection** between modules for testability
- **Process safety** via PID files to prevent concurrent syncs
- **Graceful shutdown** with proper signal handling
- **Fault tolerance** - continues operation despite individual sync failures

## Configuration

### Default Config Location

- Primary: `~/.config/b2sync/config.json`
- Override via: `CONFIG_PATH` environment variable
- Example: `/configs/config.example.json`

### Key Configuration Options

```json5
{
  "sync_pairs": [{"source": "/absolute/path/to/folder", "destination": "b2://bucket/path/to/folder"}],
  "sync_frequency": "10m",  // Go duration string
  "notification_threshold": 5, // Minimum files for success notification
  "log_level": "INFO", // DEBUG, INFO, WARN, ERROR
  "log_dir": "~/Library/Logs/b2sync"
}
```

## Dependencies and Requirements

### External Dependencies

- **Backblaze B2 CLI**: Must be installed and configured (`b2` command in `PATH`)
- **terminal-notifier**: Required for macOS notifications (`brew install terminal-notifier`)
- **macOS**: Platform-specific (LaunchAgent)
- **Go 1.24.4+**: For building

**Important**: If using Homebrew, ensure both `b2` and `terminal-notifier` are installed via `brew`. The LaunchAgent is configured to include `/opt/homebrew/bin` in its PATH to locate these dependencies.

### Runtime Behavior

- Runs in the background as a user LaunchAgent with configurable sync intervals
- Automatic startup via macOS LaunchAgent
- PID-based process safety prevents overlapping syncs
- Daily log rotation with structured logging
- Native macOS notifications for user feedback

## Development Notes

### Security Considerations

- No credentials stored in config (relies on B2 CLI configuration)
- User-directory based storage with appropriate permissions
- PID files prevent multiple concurrent operations

### Potential Areas for Enhancement

- Linting configuration (golangci-lint)
- Unit test framework setup
- Performance optimization for large file sets
- Cross-platform support beyond macOS

## Troubleshooting Commands

```bash
# Check B2 CLI availability
b2 version

# Verify B2 configuration
b2 get-account-info

# Check terminal-notifier availability
terminal-notifier -help

# Check LaunchAgent status
launchctl list | grep b2sync

# View error logs
cat ~/Library/Logs/b2sync/launchd.error.log
```
