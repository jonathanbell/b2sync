package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func ParseLevel(s string) Level {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

type Logger struct {
	logger   *log.Logger
	level    Level
	logFile  *os.File
	logDir   string
	filename string
}

func New(logDir string, level Level) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	filename := fmt.Sprintf("b2sync-%s.log", time.Now().Format("2006-01-02"))
	logPath := filepath.Join(logDir, filename)

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	multiWriter := io.MultiWriter(logFile)
	logger := log.New(multiWriter, "", 0)

	return &Logger{
		logger:   logger,
		level:    level,
		logFile:  logFile,
		logDir:   logDir,
		filename: filename,
	}, nil
}

func (l *Logger) log(level Level, msg string) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMsg := fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), msg)
	l.logger.Println(logMsg)
}

func (l *Logger) Debug(msg string) {
	l.log(DEBUG, msg)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, fmt.Sprintf(format, args...))
}

func (l *Logger) Info(msg string) {
	l.log(INFO, msg)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(msg string) {
	l.log(WARN, msg)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, fmt.Sprintf(format, args...))
}

func (l *Logger) Error(msg string) {
	l.log(ERROR, msg)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, fmt.Sprintf(format, args...))
}

func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

func (l *Logger) RotateIfNeeded() error {
	currentDate := time.Now().Format("2006-01-02")
	expectedFilename := fmt.Sprintf("b2sync-%s.log", currentDate)

	if l.filename != expectedFilename {
		l.Close()

		logPath := filepath.Join(l.logDir, expectedFilename)
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to rotate log file: %w", err)
		}

		multiWriter := io.MultiWriter(logFile)
		l.logger = log.New(multiWriter, "", 0)
		l.logFile = logFile
		l.filename = expectedFilename

		l.Info("Log file rotated")
	}

	return nil
}
