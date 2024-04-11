package controller

import (
	"runtime"

	"go.uber.org/zap"

	"github.com/diverged/uspto-bulk-data-tool/internal/config"
	"github.com/shirou/gopsutil/mem"
)

// CalcMaxConcurrent checks if the MaxConcurrentZips is set in the config, if not it calculates an appropriate value
func SetConcurrency(cfg *config.Config, log *zap.Logger) (int, error) {

	configMaxConcZips := cfg.TuningConfig.MaxConcurrentZips

	if configMaxConcZips != 0 {
		log.Info("MaxConcurrentZips set", zap.Int("value", configMaxConcZips))
		return configMaxConcZips, nil
	}

	log.Info("MaxCurrentZips not set in config, calculating based on system CPU count")

	numCPU := runtime.NumCPU()

	// TODO: Make this a function that takes into account the available memory
	// * I.e. if writing Parquet, 1GB RAM = 1 Zip, else 1GB = 5 Zips

	// If concurrency is being calculated, set it based on CPU cores - 1
	configMaxConcZips = numCPU - 1

	v, err := mem.VirtualMemory()
	if err != nil {
		log.Error("unable to profile system memory resources")
		// Return error to controller
	}
	// Total and available system memory
	log.Info("System Memory Profiled", zap.Uint64("Total RAM", v.Total), zap.Uint64("Available RAM", v.Available))

	return configMaxConcZips, nil
}
