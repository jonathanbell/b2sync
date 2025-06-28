package notifier

import (
	"fmt"
	"os/exec"

	"b2sync/internal/logger"
	"b2sync/internal/sync"
)

type Notifier struct {
	logger *logger.Logger
}

func New(log *logger.Logger) *Notifier {
	return &Notifier{
		logger: log,
	}
}

func (n *Notifier) sendNotification(title, message string) error {
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`display notification "%s" with title "%s"`, message, title))
	output, err := cmd.CombinedOutput()

	if err != nil {
		n.logger.Errorf("Failed to send notification: %v, output: %s", err, string(output))
		return err
	}

	n.logger.Debugf("Notification sent: %s - %s", title, message)
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

func (n *Notifier) NotifySyncResults(results []sync.SyncResult, threshold int) {
	totalFiles := 0
	hasErrors := false

	for _, result := range results {
		if !result.Success {
			hasErrors = true
			n.NotifySyncError(result.Error)
		} else {
			totalFiles += result.FilesCount
		}
	}

	if !hasErrors && totalFiles >= threshold {
		title := "B2Sync Complete"
		message := fmt.Sprintf("Successfully synced %d files to Backblaze B2", totalFiles)
		n.sendNotification(title, message)
	}
}
