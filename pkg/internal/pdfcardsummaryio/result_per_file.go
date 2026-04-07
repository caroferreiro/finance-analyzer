package pdfcardsummaryio

type ResultPerFile struct {
	// Original PDF file path
	PDFPath string
	// Generated CSV file path (empty if processing failed early)
	CSVPath string
	// Processing error (nil if successful)
	Err error
}
