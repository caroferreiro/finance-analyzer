package pdfcardsummary

// MovementType represents the type of a statement line item.
// These values are part of the extracted CSV contract (see `CardSummary.ToCSVMatrix`).
type MovementType string

const (
	MovementTypePastPayment MovementType = "PastPayment"
	MovementTypeTax         MovementType = "Tax"
	MovementTypeCard        MovementType = "CardMovement"
)

// NewMovementsWithCardContext is a factory that flattens a CardSummary into "fact rows":
// one row per statement line item (movement) with the statement/card context attached.
//
// Output ordering matches the current extracted CSV ordering:
// PastPayment rows → Tax rows → CardMovement rows (cards in table order, movements in card order).
func NewMovementsWithCardContext(cs CardSummary) []MovementWithCardContext {
	rows := make([]MovementWithCardContext, 0,
		len(cs.Table.PastPaymentMovements)+len(cs.Table.TaxesMovements),
	)

	for _, m := range cs.Table.PastPaymentMovements {
		rows = append(rows, MovementWithCardContext{
			StatementContext: cs.StatementContext,
			CardContext:      nil,
			MovementType:     MovementTypePastPayment,
			Movement:         m,
		})
	}

	for _, m := range cs.Table.TaxesMovements {
		rows = append(rows, MovementWithCardContext{
			StatementContext: cs.StatementContext,
			CardContext:      nil,
			MovementType:     MovementTypeTax,
			Movement:         m,
		})
	}

	for i := range cs.Table.Cards {
		card := &cs.Table.Cards[i]
		for _, m := range card.Movements {
			rows = append(rows, MovementWithCardContext{
				StatementContext: cs.StatementContext,
				CardContext:      &card.CardContext,
				MovementType:     MovementTypeCard,
				Movement:         m,
			})
		}
	}

	return rows
}

// MovementWithCardContext is a denormalized fact row representing one extracted CSV row:
// a Movement plus the statement/card context required to analyze it.
//
// This struct is intentionally "flat-ish": the extracted CSV is a fact table (similar to the spreadsheet's
// canonical `data` tab), and analysis is easiest when consuming one line item at a time.
type MovementWithCardContext struct {
	StatementContext
	*CardContext
	MovementType MovementType
	Movement     Movement
}
