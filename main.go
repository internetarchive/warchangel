package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/internetarchive/warchangel/pkg/warchangel"
)

var (
	logger *slog.Logger
)

func signalHandler(doneChan chan struct{}) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sigCount := 0
		for {
			sig := <-sigChan
			sigCount++
			if sigCount == 1 {
				logger.Info("Received signal, initiating graceful shutdown", "signal", sig)
				close(doneChan)
			} else {
				logger.Info("Received second signal, forcing shutdown", "signal", sig)
				os.Exit(1)
			}
		}
	}()
}

func main() {
	argumentParsing(os.Args)
	doneChan := make(chan struct{})

	// Create a new logger
	level := slog.LevelInfo
	addSource := false
	if arguments.Debug {
		addSource = true
		level = slog.LevelDebug
	}

	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: addSource,
		Level:     level,
	}))

	logger.Info("starting warchangel")
	logger.Debug("config",
		"threads", arguments.Threads,
		"s3-access-key", arguments.S3AccessKey,
		"s3-secret-key", arguments.S3SecretKey,
		"s3-creds-file", arguments.S3CredsFile,
		"config", arguments.Config,
		"debug", arguments.Debug,
	)

	// Loading config file, if the file is YAML then we use the legacy draintasker config loader
	// else it's a warchangel config file and it will be loaded as such
	configType, err := warchangel.DetectConfigFormat(arguments.Config)
	if err != nil {
		logger.Error("error detecting config format", "err", err)
		os.Exit(1)
	}

	var config *warchangel.Config
	switch configType {
	case warchangel.FormatYAML:
		logger.Info("loading legacy Draintasker configuration")
		config, err = warchangel.LoadDraintaskerConfig(arguments.Config)
	case warchangel.FormatJSON:
		logger.Info("loading warchangel configuration")
		config, err = warchangel.LoadConfig(arguments.Config)
	}
	if err != nil {
		logger.Error("error loading config", "err", err)
		os.Exit(1)
	}

	// Start the watcher
	go func() {
		if err := warchangel.NewWatcher(config, logger, arguments.Threads, arguments.S3AccessKey, arguments.S3SecretKey, doneChan); err != nil {
			logger.Error("watcher error", "err", err)
		}
	}()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for a termination signal
	sig := <-sigChan
	logger.Info("received signal, shutting down", "signal", sig)
	close(doneChan)
}
