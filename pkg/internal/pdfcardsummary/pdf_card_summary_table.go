package pdfcardsummary

// Table represents the generic structure of a credit card statement table
type Table struct {
	PastPaymentMovements []Movement
	Cards                []Card
	TaxesMovements       []Movement
}
