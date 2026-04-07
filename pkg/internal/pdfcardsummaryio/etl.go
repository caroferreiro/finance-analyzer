package pdfcardsummaryio

import (
	"fmt"
	"os"
	"strings"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/csvale"
)

func NewPDFCardSummaryETL(extractor Extractor, validator Validator) *PDFCardSummaryETL {
	return &PDFCardSummaryETL{
		extractor: extractor,
		validator: validator,
	}
}

// PDFCardSummaryETL is a struct that extracts, transforms and loads (ETL) PDF card summaries
type PDFCardSummaryETL struct {
	extractor Extractor
	validator Validator
}

// TODO: use ResultPerFile as the first return value and a "unrecoverable error" as the second return value
func (etl *PDFCardSummaryETL) ETLFilesIndependently(paths []string) error {
	var errs []error
	for _, path := range paths {
		err := etl.ETLFile(path)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("errors per file: %v", errs)
}

func (etl *PDFCardSummaryETL) ETLFile(path string) error {
	fileData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", path, err)
	}

	cardSummary, err := etl.extractor.ExtractFromBytes(fileData)
	if err != nil {
		return fmt.Errorf("error parsing file %s: %w", path, err)
	}

	if err := etl.validator.Validate(cardSummary); err != nil {
		return fmt.Errorf("error validating file %s: %w", path, err)
	}

	csv, err := cardSummary.ToCSVBytes()
	if err != nil {
		return fmt.Errorf("error converting file %s to CSV: %w", path, err)
	}

	csvPath := path + ".csv"
	err = os.WriteFile(csvPath, csv, 0644)
	if err != nil {
		return fmt.Errorf("error writing CSV to file %s: %w", csvPath, err)
	}

	return nil
}

// ETLFilesWithJoinedCSV processes multiple PDF files, extracts CardSummary structs,
// creates individual CSV files for each PDF, and also combines their CSV outputs into
// a single file. If any PDF fails during extraction or validation, no output is written
// and an error is returned.
func (etl *PDFCardSummaryETL) ETLFilesWithJoinedCSV(pdfs []string, outputPath string) error {
	// Step 1: Extract and validate all CardSummary structs (fail fast)
	summaries := make([]pdfcardsummary.CardSummary, 0, len(pdfs))
	for _, path := range pdfs {
		fileData, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", path, err)
		}

		cardSummary, err := etl.extractor.ExtractFromBytes(fileData)
		if err != nil {
			return fmt.Errorf("error parsing file %s: %w", path, err)
		}

		if err := etl.validator.Validate(cardSummary); err != nil {
			return fmt.Errorf("error validating file %s: %w", path, err)
		}

		summaries = append(summaries, cardSummary)
	}

	// Step 2: Write individual CSV files (reuse existing ETLFile logic)
	for i, path := range pdfs {
		csvBytes, err := summaries[i].ToCSVBytes()
		if err != nil {
			return fmt.Errorf("error converting file %s to CSV: %w", path, err)
		}

		csvPath := path + ".csv"
		err = os.WriteFile(csvPath, csvBytes, 0644)
		if err != nil {
			return fmt.Errorf("error writing CSV to file %s: %w", csvPath, err)
		}
	}

	// Step 3: Convert all CardSummary structs to CSV matrices
	matrices := make([][][]string, 0, len(summaries))
	for _, summary := range summaries {
		matrix, err := summary.ToCSVMatrix()
		if err != nil {
			return fmt.Errorf("error converting to CSV matrix: %w", err)
		}
		matrices = append(matrices, matrix)
	}

	// Step 4: Combine matrices (first includes headers, rest skip headers)
	combinedMatrix := csvale.CombineCSVMatrices(matrices)

	// Step 5: Convert combined matrix to bytes (same format as ToCSVBytes)
	var rowsAsStrings []string
	for _, row := range combinedMatrix {
		rowsAsStrings = append(rowsAsStrings, strings.Join(row, ";"))
	}
	finalString := strings.Join(rowsAsStrings, "\n")
	csvBytes := []byte(finalString)

	// Step 6: Write combined CSV to output path (silently overwrite if exists)
	err := os.WriteFile(outputPath, csvBytes, 0644)
	if err != nil {
		return fmt.Errorf("error writing combined CSV to file %s: %w", outputPath, err)
	}

	return nil
}
