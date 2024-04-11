package controller

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"

	usptgo "github.com/diverged/uspt-go"
	"github.com/diverged/uspt-go/types"

	"github.com/diverged/uspto-bulk-data-tool/internal/config"
	"github.com/diverged/uspto-bulk-data-tool/internal/logger"
	"github.com/diverged/uspto-bulk-data-tool/internal/outputhandler"
)

func Controller(cfg *config.Config, log *zap.Logger) error {

	if err := os.MkdirAll(cfg.OutputDir, os.ModePerm); err != nil {
		log.Fatal("Failed to create output directory", zap.String("directory", cfg.OutputDir), zap.Error(err))
		return err
	}

	// Set max concurrency, i.e. number of concurrent zip files being processed
	maxConcurrentZips, err := SetConcurrency(cfg, log)
	if err != nil {
		log.Fatal("Error setting max concurrent zips", zap.Error(err))
		return err
	}

	// Initiate the errorChan & ErrorHandler() to monitor the error channel
	errorChan := make(chan error, maxConcurrentZips*100)
	go ErrorHandler(errorChan, cfg, log)

	// Initialize a semaphore channel to limit the number of concurrent bulk zip files being processed
	semaphore := make(chan struct{}, maxConcurrentZips)

	// Intitialize a wait group to manage concurrent processing
	var wg sync.WaitGroup

	// Walk the input directory of bulk files to be processed
	err = filepath.WalkDir(cfg.InputDir, func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			log.Error("Error walking the input directory", zap.String("Error path", path), zap.String("Input directory", cfg.InputDir), zap.Error(err))
			return nil // Consider expanding error handling capabilities here to differentiate between a fatal on the directory being walked, or a non-fatal on a subdirectory
		}

		// Process only non-directory files with a .zip extension.
		if !dirEntry.IsDir() && filepath.Ext(dirEntry.Name()) == ".zip" {

			bulkZipName := dirEntry.Name()

			// Acquire a semaphore token and increment the wait group counter
			wg.Add(1)
			semaphore <- struct{}{}

			// Initiate go routine to process the zip file
			go func() {
				defer wg.Done()
				defer func() { <-semaphore }()

				var subwg sync.WaitGroup

				// * Initialize the USPTGoConfig struct

				usptgoConfig := &types.USPTGoConfig{}

				if cfg.OutputMode == "xml" {
					usptgoConfig = &types.USPTGoConfig{
						InputPath:         filepath.Join(cfg.InputDir, bulkZipName),
						Logger:            logger.NewZapLoggerAdapter(log),
						ReturnRawSplitDoc: true,
					}
				} else {
					usptgoConfig = &types.USPTGoConfig{
						InputPath:         filepath.Join(cfg.InputDir, bulkZipName),
						Logger:            logger.NewZapLoggerAdapter(log),
						ReturnRawSplitDoc: cfg.DevConfig.ParserReturnsRaw,
					}
				}

				// * Call USPT-Go parser
				parsedDocs, parserErr, err := usptgo.USPTGo(usptgoConfig)
				if err != nil {
					log.Error("error when calling usptgo.USPTGo(parserConfig)")
					os.Exit(1)
				}
				log.Info("usptgo parser called from controller.go")
				// Redirect the error channel contents
				go func() {
					for err := range parserErr {
						errorChan <- err
					}
				}()

				// * Call outputhandler.HandleOutput()
				subwg.Add(1)
				go func() {
					defer subwg.Done()
					log.Debug("Calling outputhandler.HandleOutput()")
					outputhandler.HandleOutput(cfg, parsedDocs, errorChan, log, bulkZipName)
				}()
				subwg.Wait()

			}()
		}

		return nil
	})
	// Wait for all go routines to complete
	wg.Wait()

	// Close the error channel after all go routines have finished, signaling the error handler to exit
	close(errorChan)
	// Need to revisit this return err to investigate whether a nonfatal error could be returned to main, thereby causing main to believe a fatal error happened even if it did not.
	if err != nil {
		log.Error("Error encountered in filepath.WalkDir()", zap.Error(err))
		return err
	}
	return nil
}
