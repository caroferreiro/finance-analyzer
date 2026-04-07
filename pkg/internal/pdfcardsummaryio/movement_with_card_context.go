package pdfcardsummaryio

import "github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"

// Re-export the denormalized "fact row" shape from `pdfcardsummary`.
//
// Rationale:
// - `pdfcardsummaryio` is an IO-oriented package. It shouldn't own domain contracts.
// - Keeping a type alias here avoids churn for any downstream imports while we refactor.
type MovementWithCardContext = pdfcardsummary.MovementWithCardContext
