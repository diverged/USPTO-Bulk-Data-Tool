package outputhandler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"github.com/diverged/uspt-go/types"

	"github.com/diverged/uspto-bulk-data-tool/internal/config"
)

func WriteJSONFiles(cfg *config.Config, parsedDocs <-chan *types.USPTGoDoc, errorChan chan<- error, log *zap.Logger) {

	log.Info("WriteJSONFiles called")
	outputDir := cfg.OutputDir

	for doc := range parsedDocs {

		//outputSubDir := filepath.Join(outputDir, doc.USPTGoMetadata.OriginZip.ZipName)
		filename := doc.Patent.MetaFileName
		log.Info("Writing JSON documents for:", zap.Any("index", filename))

		// TEMP - for debugging

		if filename == "" {
			log.Error("Document does not have a 'DocIndexOfZip' in its metadata")
			// TODO Handle the error appropriately, possibly continue to the next document.
			continue
		}

		outputFileName := strings.TrimSuffix(strings.TrimSuffix(filename, ".XML"), ".xml") + ".json"

		outputFilePath := filepath.Join(outputDir, outputFileName)

		// Marshall the JSON
		jsonData, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			log.Error("Failed to marshal document to JSON", zap.String("filename",
				filename), zap.Error(err))
			continue
		}

		err = os.WriteFile(outputFilePath, jsonData, 0644)
		if err != nil {
			log.Error("Failed to save document to disk", zap.String("filename",
				filename), zap.Error(err))
			continue
		}
		log.Debug("Document saved", zap.String("filename", filename))

	}
}
