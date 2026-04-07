package pdfcardsummaryio

import (
	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/pdfwrapper"
)

//go:generate mockgen -source interfaces.go -destination=./mocks_test.go -package=pdfcardsummaryio

type Extractor interface {
	ExtractFromBytes(rawBytes []byte) (pdfcardsummary.CardSummary, error)
	ExtractFromDocument(pdfDoc pdfwrapper.Document) (pdfcardsummary.CardSummary, error)
}

type Validator interface {
	Validate(cs pdfcardsummary.CardSummary) error
}
