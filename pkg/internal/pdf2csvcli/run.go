package pdf2csvcli

import (
	"fmt"
	"io"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummaryio"
	"github.com/Alechan/finance-analyzer/pkg/internal/validation"
)

const (
	ExitSuccess = 0
	ExitFailure = 1
)

// Run executes the CLI logic and writes user-facing failures to stderr.
func Run(rawArgs []string, stderr io.Writer) int {
	args, err := ParseArgs(rawArgs)
	if err != nil {
		fmt.Fprintln(stderr, "Error parsing arguments:", err)
		return ExitFailure
	}

	reader, err := ExtractorFactory(args.Bank)
	if err != nil {
		fmt.Fprintln(stderr, "Error creating reader:", err)
		return ExitFailure
	}

	validator := validation.NewValidator()
	etl := pdfcardsummaryio.NewPDFCardSummaryETL(reader, validator)

	if args.JoinCSV != nil {
		err = etl.ETLFilesWithJoinedCSV(args.PDFs, *args.JoinCSV)
	} else {
		err = etl.ETLFilesIndependently(args.PDFs)
	}

	if err != nil {
		fmt.Fprintln(stderr, fmt.Sprintf("Error processing files: %v", err))
		return ExitFailure
	}

	return ExitSuccess
}
