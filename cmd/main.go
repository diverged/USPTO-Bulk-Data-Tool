package main

import (
	"fmt"
	"os"
	"time"

	"github.com/diverged/uspto-bulk-data-tool/internal/config"
	"github.com/diverged/uspto-bulk-data-tool/internal/controller"
	"github.com/diverged/uspto-bulk-data-tool/internal/logger"

	"go.uber.org/zap"
)

func main() {
	// Timestamp the start of runtime
	startTime := time.Now()

	// Load the configuration file.
	var cliArgConfigPath string
	if len(os.Args) > 1 {
		cliArgConfigPath = os.Args[1]
	}

	cfg, err := config.LoadConfig(cliArgConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %s\n", err)
		os.Exit(1)
	}

	// * Initialize global logger
	log, err := logger.InitLogger(cfg.LoggerConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := log.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flushing log: %v\n", err)
		}
	}()

	// * Initialize the controller
	if err := controller.Controller(cfg, log); err != nil {
		log.Error("Error in controller", zap.Error(err))
		os.Exit(1)
	}
	// Clean output directorty if required
	if cfg.CleanOutput {
		if err := os.RemoveAll(cfg.OutputDir); err != nil {
			log.Error("Error in cleaning output directory", zap.Error(err))

		}
		log.Debug("Output directory cleaned")
	}

	// Log total runtime
	elapsedTime := time.Since(startTime)
	log.Info("\nExecution Completed", zap.String("Execution Time", elapsedTime.String()))
}
