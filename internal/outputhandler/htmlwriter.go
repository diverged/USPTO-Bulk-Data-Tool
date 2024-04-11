package outputhandler

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/diverged/uspt-go/types"
	"go.uber.org/zap"

	"github.com/diverged/uspto-bulk-data-tool/internal/config"
)

// Do not use
func WriteHtmlFiles(cfg *config.Config, parsedDocs <-chan *types.USPTGoDoc, errorChan chan<- error, log *zap.Logger) {

	log.Info("WriteHtmlFiles called")
	outputDir := cfg.OutputDir

	for doc := range parsedDocs {

		// Use the "DocIndexOfZip" from SplitterMetadata for the filename.
		filename := doc.USPTGoMetadata.OriginZip.IndexName
		log.Info("Writing HTML documents for:", zap.Any("index", filename))

		if filename == "" {
			log.Error("Document does not have a 'DocIndexOfZip' in its metadata")
			// Handle the error appropriately, possibly continue to the next document.
			continue
		}

		outputFileName := strings.TrimSuffix(strings.TrimSuffix(filename, ".XML"), ".xml") + ".html"

		outputFilePath := filepath.Join(outputDir, outputFileName)

		err := os.WriteFile(outputFilePath, []byte(doc.Patent.Description.Content), 0644)
		if err != nil {
			log.Error("Failed to save document to disk", zap.String("filename",
				filename), zap.Error(err))
			/* 			errorChan <- fmt.Errorf("failed to save document %s: %w", filename,
			err) */
			continue
		}
		log.Debug("Document saved", zap.String("filename", filename))
	}
}
