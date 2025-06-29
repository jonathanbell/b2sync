package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"b2sync/internal/config"
	"b2sync/internal/logger"
	"b2sync/internal/notifier"
	"b2sync/internal/sync"
)

func main() {
	var help = flag.Bool("help", false, "Show help information")
	flag.Parse()

	if *help {
		fmt.Println("B2Sync - Automated Backblaze Backup Utility")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  b2sync [options]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  --help    Show this help message")
		fmt.Println()
		fmt.Println("Configuration:")
		fmt.Printf("  Config file: %s\n", config.GetConfigPath())
		fmt.Println("  Supports duration formats: 1m, 5m, 1h, 30s, etc.")
		fmt.Println()
		fmt.Println("For more information, see README.md")
		return
	}

	configPath := config.GetConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(cfg.LogDir, logger.ParseLevel(cfg.LogLevel))
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("B2Sync started")

	notif := notifier.New(log)
	syncManager := sync.New(cfg, log)

	if err := syncManager.CheckB2Available(); err != nil {
		log.Error(fmt.Sprintf("B2 CLI not available: %v", err))
		notif.NotifyB2NotInstalled()
		os.Exit(1)
	}

	notif.NotifyStartup()

	ticker := time.NewTicker(cfg.SyncFrequency.ToDuration())
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	performSync := func() {
		log.Debug("Starting sync cycle")

		if err := log.RotateIfNeeded(); err != nil {
			log.Warnf("Log rotation failed: %v", err)
		}

		results := syncManager.SyncAll()
		notif.NotifySyncResults(results, cfg.NotificationThreshold)

		log.Debug("Sync cycle completed")
	}

	performSync()

	for {
		select {
		case <-ticker.C:
			performSync()
		case sig := <-sigChan:
			log.Info(fmt.Sprintf("Received signal %v, shutting down", sig))
			notif.NotifyShutdown()
			return
		}
	}
}
