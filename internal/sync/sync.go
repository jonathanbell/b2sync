package sync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"b2sync/internal/config"
	"b2sync/internal/logger"
)

type SyncResult struct {
	Success    bool
	FilesCount int
	Duration   time.Duration
	Error      error
	Output     string
}

type Manager struct {
	config *config.Config
	logger *logger.Logger
	pidDir string
}

func New(cfg *config.Config, log *logger.Logger) *Manager {
	homeDir, _ := os.UserHomeDir()
	pidDir := filepath.Join(homeDir, ".config", "b2sync", "pids")
	os.MkdirAll(pidDir, 0755)

	return &Manager{
		config: cfg,
		logger: log,
		pidDir: pidDir,
	}
}

func (m *Manager) CheckB2Available() error {
	_, err := exec.LookPath("b2")
	if err != nil {
		return fmt.Errorf("b2 CLI not found in PATH")
	}
	return nil
}

func (m *Manager) IsB2SyncRunning() (bool, error) {
	pidFile := filepath.Join(m.pidDir, "b2sync.pid")

	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return false, nil
	}

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		os.Remove(pidFile)
		return false, nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		os.Remove(pidFile)
		return false, nil
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		os.Remove(pidFile)
		return false, nil
	}

	return true, nil
}

func (m *Manager) createPidFile() error {
	pidFile := filepath.Join(m.pidDir, "b2sync.pid")
	pid := os.Getpid()
	return os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}

func (m *Manager) removePidFile() {
	pidFile := filepath.Join(m.pidDir, "b2sync.pid")
	os.Remove(pidFile)
}

func (m *Manager) SyncAll() []SyncResult {
	if err := m.CheckB2Available(); err != nil {
		m.logger.Error(fmt.Sprintf("b2 CLI check failed: %v", err))
		return []SyncResult{{Success: false, Error: err}}
	}

	running, err := m.IsB2SyncRunning()
	if err != nil {
		m.logger.Warnf("Failed to check if sync is running: %v", err)
	}
	if running {
		m.logger.Info("Sync already running, skipping this cycle")
		return []SyncResult{{Success: false, Error: fmt.Errorf("sync already running")}}
	}

	if err := m.createPidFile(); err != nil {
		m.logger.Errorf("Failed to create PID file: %v", err)
		return []SyncResult{{Success: false, Error: err}}
	}
	defer m.removePidFile()

	var results []SyncResult
	for _, pair := range m.config.SyncPairs {
		result := m.syncPair(pair)
		results = append(results, result)

		if result.Success {
			m.logger.Infof("Sync completed: %s -> %s (%d files, %v)",
				pair.Source, pair.Destination, result.FilesCount, result.Duration)
		} else {
			if result.Output != "" {
				m.logger.Errorf("Sync failed: %s -> %s: %v\nB2 output: %s",
					pair.Source, pair.Destination, result.Error, strings.TrimSpace(result.Output))
			} else {
				m.logger.Errorf("Sync failed: %s -> %s: %v",
					pair.Source, pair.Destination, result.Error)
			}
		}
	}

	return results
}

func (m *Manager) syncPair(pair config.SyncPair) SyncResult {
	start := time.Now()

	if _, err := os.Stat(pair.Source); os.IsNotExist(err) {
		return SyncResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("source directory does not exist: %s", pair.Source),
		}
	}

	args := []string{"sync"}
	if m.config.KeepDays > 0 {
		args = append(args, "--keep-days", strconv.Itoa(m.config.KeepDays))
	}
	args = append(args, pair.Source, pair.Destination)
	
	cmd := exec.Command("b2", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return SyncResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("b2 sync failed: %v", err),
			Output:   string(output),
		}
	}

	filesCount := m.parseFilesCount(string(output))

	return SyncResult{
		Success:    true,
		FilesCount: filesCount,
		Duration:   time.Since(start),
		Output:     string(output),
	}
}

func (m *Manager) parseFilesCount(output string) int {
	// First try to parse summary patterns
	patterns := []string{
		`(\d+) files? uploaded`,
		`uploaded (\d+) files?`,
		`(\d+) files? transferred`,
		`transferred (\d+) files?`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			if count, err := strconv.Atoi(matches[1]); err == nil && count > 0 {
				return count
			}
		}
	}
	
	// If no summary found, count individual upload operations
	uploadRe := regexp.MustCompile(`^upload\s+\S+`)
	lines := strings.Split(output, "\n")
	uploadCount := 0
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if uploadRe.MatchString(line) {
			uploadCount++
		}
	}
	
	return uploadCount
}
