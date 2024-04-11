package outputhandler

import (
	"sync"

	"github.com/diverged/uspt-go/types"
	"github.com/diverged/uspto-bulk-data-tool/internal/config"
	"go.uber.org/zap"
)

func HandleOutput(cfg *config.Config, inputChan <-chan *types.USPTGoDoc, errorChan chan<- error, log *zap.Logger, originZipName string) {

	log.Debug("Handling output", zap.String("originZipName", originZipName))

	var wg sync.WaitGroup

	// Handle output based on configuration
	if cfg.OutputMode == "xml" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			WriteXMLFiles(cfg, inputChan, errorChan, log)
		}()
		wg.Wait()
	} else if cfg.OutputMode == "json" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			WriteJSONFiles(cfg, inputChan, errorChan, log)
		}()
		wg.Wait()
	} else if cfg.OutputMode == "parquet" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			WriteParquetFile(cfg, originZipName, inputChan, errorChan, log)

		}()
		wg.Wait()
	} else {

		log.Debug("No output mode specified, skipping output handling")
	}
}
