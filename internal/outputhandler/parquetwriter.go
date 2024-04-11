package outputhandler

import (
	"path/filepath"
	"strings"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"

	"go.uber.org/zap"

	"github.com/diverged/uspt-go/types"
	"github.com/diverged/uspto-bulk-data-tool/internal/config"
)

type ParquetPatentDocument struct {

	// Patent Metadata
	MetaFileName       string `parquet:"name=document_name, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	MetaFileType       string `parquet:"name=document_type, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	MetaDateProduced   string `parquet:"name=date_produced, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	MetaDatePubl       string `parquet:"name=date_publ, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	MetaCountry        string `parquet:"name=country, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	MetaInventionTitle string `parquet:"name=invention_title, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	MetaNumberOfClaims int    `parquet:"name=number_of_claims, type=INT64"`

	// Invention Contents
	Abstract    string `parquet:"name=abstract, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	Description string `parquet:"name=description, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	Claims      string `parquet:"name=claims, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`

	// PublicationReference fields
	PubRefCountry   string `parquet:"name=pub_ref_country, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	PubRefDocNumber string `parquet:"name=pub_ref_doc_number, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	PubRefKindCode  string `parquet:"name=pub_ref_kind_code, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	PubRefDate      string `parquet:"name=pub_ref_date, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`

	// ClassificationNational fields
	ClassNatCountry               string `parquet:"name=class_nat_country, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	ClassNatMainClassification    string `parquet:"name=class_nat_main_classification, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	ClassNatFurtherClassification string `parquet:"name=class_nat_further_classification, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
}

func parquetConvert(parsedDocIn <-chan *types.USPTGoDoc, parquetDocChan chan<- ParquetPatentDocument, log *zap.Logger) {

	log.Debug("ParquetConvert has been invoked")

	for doc := range parsedDocIn {
		// Convert XMLPatentDocument to ParquetFile
		mappedParquet := ParquetPatentDocument{

			MetaFileName:       doc.Patent.MetaFileName,
			MetaFileType:       doc.USPTGoMetadata.DocumentType,
			MetaDateProduced:   doc.Patent.MetaDateProduced,
			MetaDatePubl:       doc.Patent.MetaDatePubl,
			MetaCountry:        doc.Patent.MetaCountry,
			MetaInventionTitle: doc.Patent.UsBibliographicData.InventionTitle.Text,
			MetaNumberOfClaims: doc.Patent.UsBibliographicData.NumberOfClaims,

			// MainTextFields
			Abstract:    doc.Patent.Abstract.Content,
			Description: doc.Patent.Description.Content,
			Claims:      doc.Patent.Claims.Content,

			// Biblio Data
			PubRefCountry:                 doc.Patent.UsBibliographicData.PublicationReference.DocumentID.Country,
			PubRefDocNumber:               doc.Patent.UsBibliographicData.PublicationReference.DocumentID.DocNumber,
			PubRefKindCode:                doc.Patent.UsBibliographicData.PublicationReference.DocumentID.KindCode,
			PubRefDate:                    doc.Patent.UsBibliographicData.PublicationReference.DocumentID.Date,
			ClassNatCountry:               doc.Patent.UsBibliographicData.ClassificationNational.Country,
			ClassNatMainClassification:    doc.Patent.UsBibliographicData.ClassificationNational.MainClassification,
			ClassNatFurtherClassification: doc.Patent.UsBibliographicData.ClassificationNational.FurtherClassification,
		}

		// Send the ParquetFile to the parquetDocOut channel
		parquetDocChan <- mappedParquet
		log.Debug("ParquetConvert: doc => parquetDocChan")

	}
}

func WriteParquetFile(cfg *config.Config, originZipName string, inputChan <-chan *types.USPTGoDoc, errorChan chan<- error, log *zap.Logger) {

	log.Debug("WriteParquetFile has been invoked", zap.String("OriginZipName", originZipName))

	// Set output path and file name based on originating zip file name
	outputFileName := strings.TrimSuffix(originZipName, ".zip") + ".parquet"
	outputFilePath := filepath.Join(cfg.OutputDir, outputFileName)

	// * Initialize LOCAL file writer
	var err error
	fw, err := local.NewLocalFileWriter(outputFilePath)
	if err != nil {
		log.Error("Error initializing parquet file writer", zap.Error(err))
		errorChan <- &types.USPTGoError{
			Skipped: true,
			Name:    outputFileName,
			Type:    "parquet",
			Whence:  "initializing the local file writer",
			Err:     err,
		}
		return
	}
	defer fw.Close()

	// * Initialize PARQUET writer

	pw, err := writer.NewParquetWriter(fw, new(ParquetPatentDocument), 4) // Observe & experiment with the parallelism as needed
	if err != nil {
		log.Error("Error initializing parquet writer", zap.Error(err))
		errorChan <- &types.USPTGoError{
			Skipped: true,
			Name:    outputFileName,
			Type:    "parquet",
			Whence:  "initializing the Parquet writer",
			Err:     err,
		}
		return
	}
	defer pw.WriteStop()

	// Set Parquet writer properties as needed
	pw.RowGroupSize = 128 * 1024 * 1024 //128M
	pw.PageSize = 8 * 1024              //8K

	switch cfg.OutputConfig.ParquetCompression {
	case "snappy":
		pw.CompressionType = parquet.CompressionCodec_SNAPPY
	case "gzip":
		pw.CompressionType = parquet.CompressionCodec_GZIP
	case "no-compress":
		pw.CompressionType = parquet.CompressionCodec_UNCOMPRESSED
	case "lz4":
		pw.CompressionType = parquet.CompressionCodec_LZ4
	case "zstd":
		pw.CompressionType = parquet.CompressionCodec_ZSTD
	}

	// Convert incoming types.USPTGoDoc docs to ParquetPatentDocument
	parquetDocChan := make(chan ParquetPatentDocument, 100) // ! Experiment with different sizes relative to NewParquetWriter's value ()
	go func() {
		defer close(parquetDocChan)
		parquetConvert(inputChan, parquetDocChan, log)
	}()

	// Range over the channel and write to the parquet file
	for doc := range parquetDocChan {

		if err = pw.Write(doc); err != nil {
			log.Error("Error writing document to parquet file", zap.Error(err))
			errorChan <- err
			return
		}
		//log.Debug("Wrote doc to Parquet", zap.String("doc written", doc.Patent.MetaFileName), zap.String("to parquet file", outputFileName))
	}

	// Finalize writing and close the file outside the loop
	if err = pw.WriteStop(); err != nil {
		log.Error("Error finalizing parquet file", zap.Error(err))
		errorChan <- err
	}

	fw.Close()
}
