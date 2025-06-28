#!/bin/bash

set -e

PLIST_NAME="com.b2sync.agent.plist"
PLIST_PATH="$HOME/Library/LaunchAgents/$PLIST_NAME"
BINARY_PATH="/usr/local/bin/b2sync"

echo "B2Sync LaunchAgent Installation Script"
echo "======================================"

if [ ! -f "./b2sync" ]; then
    echo "Error: b2sync binary not found in current directory"
    echo "Please build the binary first: go build -o b2sync cmd/b2sync/main.go"
    exit 1
fi

echo "Installing b2sync binary to $BINARY_PATH..."
sudo cp ./b2sync "$BINARY_PATH"
sudo chmod +x "$BINARY_PATH"

echo "Creating LaunchAgent plist at $PLIST_PATH..."
cat > "$PLIST_PATH" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.b2sync.agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>$BINARY_PATH</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardErrorPath</key>
    <string>$HOME/Library/Logs/b2sync/launchd.error.log</string>
    <key>StandardOutPath</key>
    <string>$HOME/Library/Logs/b2sync/launchd.out.log</string>
    <key>WorkingDirectory</key>
    <string>$HOME</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin</string>
    </dict>
</dict>
</plist>
EOF

mkdir -p "$HOME/Library/Logs/b2sync"

echo "Loading LaunchAgent..."
launchctl load "$PLIST_PATH"

echo ""
echo "Installation complete!"
echo ""
echo "The b2sync service will now start automatically when you log in."
echo "To manage the service:"
echo "  Start:   launchctl load $PLIST_PATH"
echo "  Stop:    launchctl unload $PLIST_PATH"
echo "  Status:  launchctl list | grep b2sync"
echo ""
echo "Logs are available at: $HOME/Library/Logs/b2sync/"
echo ""
echo "Don't forget to:"
echo "1. Install and configure the b2 CLI tool"
echo "2. Create your configuration file at ~/.config/b2sync/config.json"
echo "3. Use the example config at configs/config.example.json as a template"