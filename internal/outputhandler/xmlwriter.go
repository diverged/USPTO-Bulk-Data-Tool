package outputhandler

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/diverged/uspt-go/types"
	"github.com/diverged/uspto-bulk-data-tool/internal/config"
)

// WriteXMLFiles is used to simply write the split bulk documents to individual well-formed XML files.
func WriteXMLFiles(cfg *config.Config, parsedDocs <-chan *types.USPTGoDoc, errorChan chan<- error, log *zap.Logger) {

	log.Info("WriteXMLFiles called")

	outputDir := cfg.OutputDir

	for doc := range parsedDocs {

		// Use the "DocIndexOfZip" from SplitterMetadata for the filename.
		filename := doc.USPTGoMetadata.OriginZip.IndexName
		log.Info("Writing XML documents for:", zap.Any("index", filename))

		// TEMP - for debugging

		if doc.RawSplitDoc[len(doc.RawSplitDoc)-1] != '>' {
			log.Error("Document does not end with a closing tag", zap.String("filename", filename))
		}

		// TEMP - for debugging

		if filename == "" {
			log.Error("Document does not have a 'DocIndexOfZip' in its metadata")
			// Handle the error appropriately, possibly continue to the next document.
			continue
		}

		fullPath := filepath.Join(outputDir, filename)

		err := os.WriteFile(fullPath, doc.RawSplitDoc, 0644)
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
