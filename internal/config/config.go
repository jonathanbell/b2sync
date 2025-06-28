package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type SyncPair struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// Duration wraps time.Duration to provide custom JSON marshaling
type Duration time.Duration

// UnmarshalJSON implements json.Unmarshaler for Duration
func (d *Duration) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	
	switch value := v.(type) {
	case float64:
		// Handle numeric values as nanoseconds
		*d = Duration(time.Duration(value))
		return nil
	case string:
		// Handle string values like "10m", "1h", "30s"
		duration, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid duration format: %s", value)
		}
		*d = Duration(duration)
		return nil
	default:
		return fmt.Errorf("invalid duration type: %T", value)
	}
}

// MarshalJSON implements json.Marshaler for Duration
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// String returns the string representation of the duration
func (d Duration) String() string {
	return time.Duration(d).String()
}

// ToDuration converts Duration to time.Duration
func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}

type Config struct {
	SyncPairs             []SyncPair `json:"sync_pairs"`
	SyncFrequency         Duration   `json:"sync_frequency"`
	NotificationThreshold int        `json:"notification_threshold"`
	LogLevel              string     `json:"log_level"`
	LogDir                string     `json:"log_dir"`
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		SyncPairs: []SyncPair{
			{
				Source:      filepath.Join(homeDir, "Pictures"),
				Destination: "b2://your-bucket-name/Pictures",
			},
		},
		SyncFrequency:         Duration(10 * time.Minute),
		NotificationThreshold: 5,
		LogLevel:              "INFO",
		LogDir:                filepath.Join(homeDir, "Library", "Logs", "b2sync"),
	}
}

func LoadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	if config.LogDir == "" {
		homeDir, _ := os.UserHomeDir()
		config.LogDir = filepath.Join(homeDir, "Library", "Logs", "b2sync")
	}

	return &config, nil
}

func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "b2sync", "config.json")
}