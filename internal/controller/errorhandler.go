package controller

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/diverged/uspt-go/types"
	"github.com/diverged/uspto-bulk-data-tool/internal/config"
)

// ErrorHandler centralizes handling of certain errors that may occur during processing.
func ErrorHandler(errorChan <-chan error, cfg *config.Config, log *zap.Logger) {

	log.Debug("ErrorHandler invoked")

	// TEMP - Runtime standin
	cfgRunTime := time.Now()

	// TODO - Move report subdirectory into output directory after adding additional subfolder for files out

	reportPath := "data/runreports"
	if err := os.MkdirAll(reportPath, 0755); err != nil {
		// Should use this instead of log.Fatal to allow for deferred cleanup
		log.Error("Failed to create report directory", zap.String("directory", reportPath), zap.Error(err))
		os.Exit(1)
	}

	// Create and open the skipReport file
	skipReportFileName := "SkippedFiles-" + cfgRunTime.Format("2006-01-02T15-04-05") + ".txt"
	skipReportPath := filepath.Join(reportPath, skipReportFileName)
	skipReport, err := os.OpenFile(skipReportPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error("Failed to open skipReport file", zap.String("report", reportPath), zap.Error(err))
		os.Exit(1)
	}
	defer skipReport.Close()

	// Monitors errorChan, taking additional action selectively on custom error types
	for err := range errorChan {
		// Type Assertion to determine if the error is of interest. (If err holds a *FileErr, then ok is true and fileErr contains the value.)
		fileErr, ok := err.(*types.USPTGoError)
		// TODO - Different message warning of partially processed zips

		if !ok || !fileErr.Skipped { // If err is not *FileErr *OR* it is a *FileErr with Skipped == false, then:

			continue

		} else if fileErr.Skipped {
			// If the error is a *FileErr with Skipped == true, then:
			if writeErr := appendSkippedReport(skipReport, fileErr, log); writeErr != nil {

				log.Error("Failed to append to skipReport", zap.String("report", skipReportFileName), zap.Error(err))
				os.Exit(1)
			}
		}

	}

}

func appendSkippedReport(reportFile *os.File, fileErr *types.USPTGoError, log *zap.Logger) error {

	skippedTime := time.Now().Format("T15-04-05")

	// Create the message to append to the file
	message := fmt.Sprintf(
		"%s file skipped [%s] due to error encountered while %s.\n    At Time: %s.\n    Error: %v\n\n",
		fileErr.Type, fileErr.Name, fileErr.Whence, skippedTime, fileErr.Err,
	)

	// Log the message
	log.Error(message, zap.Error(fileErr.Err))

	// Append the message to the report file
	if _, writeErr := reportFile.WriteString(message); writeErr != nil {
		return writeErr
	}
	return nil
}
