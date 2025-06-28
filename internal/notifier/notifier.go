package notifier

import (
	"fmt"
	"os/exec"
	"strings"

	"b2sync/internal/logger"
	"b2sync/internal/sync"
)

type Notifier struct {
	logger               *logger.Logger
	useTerminalNotifier  bool
}

func New(log *logger.Logger) *Notifier {
	n := &Notifier{
		logger: log,
	}
	n.checkNotificationMethod()
	return n
}

func (n *Notifier) checkNotificationMethod() {
	_, err := exec.LookPath("terminal-notifier")
	if err != nil {
		n.useTerminalNotifier = false
		n.logger.Warnf("terminal-notifier not found, falling back to osascript: %v", err)
	} else {
		n.useTerminalNotifier = true
		n.logger.Debugf("Using terminal-notifier for notifications")
	}
}

func (n *Notifier) sendNotification(title, message string) error {
	var cmd *exec.Cmd
	var method string

	if n.useTerminalNotifier {
		cmd = exec.Command("terminal-notifier", "-title", title, "-message", message, "-sender", "com.apple.finder")
		method = "terminal-notifier"
	} else {
		cmd = exec.Command("osascript", "-e", fmt.Sprintf(`display notification "%s" with title "%s"`, message, title))
		method = "osascript"
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		n.logger.Errorf("Failed to send notification via %s: %v, output: %s", method, err, string(output))
		return err
	}

	n.logger.Debugf("Notification sent via %s: %s - %s", method, title, message)
	return nil
}

func (n *Notifier) NotifyB2NotInstalled() {
	title := "B2Sync Error"
	message := "Backblaze B2 CLI is not installed. Please install it first."
	n.sendNotification(title, message)
}

func (n *Notifier) NotifySyncError(err error) {
	title := "B2Sync Error"
	message := fmt.Sprintf("Sync failed: %v", err)
	n.sendNotification(title, message)
}

func (n *Notifier) NotifySyncSkipped() {
	title := "B2Sync Info"
	message := "Sync skipped - another sync is already in progress"
	n.sendNotification(title, message)
}

func (n *Notifier) NotifySyncResults(results []sync.SyncResult, threshold int) {
	totalFiles := 0
	hasErrors := false
	hasSkippedSync := false

	for _, result := range results {
		if !result.Success {
			// Check if this is a "sync already running" error
			if result.Error != nil && strings.Contains(result.Error.Error(), "sync already running") {
				hasSkippedSync = true
				n.NotifySyncSkipped()
			} else {
				hasErrors = true
				n.NotifySyncError(result.Error)
			}
		} else {
			totalFiles += result.FilesCount
		}
	}

	if !hasErrors && !hasSkippedSync && totalFiles >= threshold {
		title := "B2Sync Complete"
		message := fmt.Sprintf("Successfully synced %d files to Backblaze B2", totalFiles)
		n.sendNotification(title, message)
	}
}

func (n *Notifier) NotifyStartup() {
	title := "B2Sync Started"
	message := "B2Sync background service has started successfully"
	n.sendNotification(title, message)
}

func (n *Notifier) NotifyShutdown() {
	title := "B2Sync Stopped"
	message := "B2Sync background service has been stopped"
	n.sendNotification(title, message)
}
