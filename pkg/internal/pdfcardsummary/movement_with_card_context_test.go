package pdfcardsummary

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestNewMovementsWithCardContext_WhenCardSummaryHasAllMovementTypes_ThenReturnsFactRowsInCSVOrder(t *testing.T) {
	// Given
	closeDate := time.Date(2025, 1, 25, 0, 0, 0, 0, time.UTC)
	expirationDate := time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC)
	statement := StatementContext{
		Bank:           Bank("DEMO_BANK"),
		CardCompany:    CardCompany("VISA"),
		CloseDate:      closeDate,
		ExpirationDate: expirationDate,
		TotalARS:       decimal.Zero,
		TotalUSD:       decimal.Zero,
	}

	pastPaymentMovement := Movement{
		Detail:    "SALDO ANTERIOR",
		AmountARS: decimal.RequireFromString("-10000"),
		AmountUSD: decimal.Zero,
	}
	taxMovement := Movement{
		Detail:    "IVA DEMO",
		AmountARS: decimal.RequireFromString("500"),
		AmountUSD: decimal.Zero,
	}

	receipt := "1001"
	cardMovement := Movement{
		ReceiptNumber: &receipt,
		Detail:        "SUPERMARKET DEMO",
		AmountARS:     decimal.RequireFromString("12345.67"),
		AmountUSD:     decimal.Zero,
	}

	cardNumber := "0000"
	card := Card{
		CardContext: CardContext{
			CardNumber:   &cardNumber,
			CardOwner:    "OWNER A",
			CardTotalARS: decimal.Zero,
			CardTotalUSD: decimal.Zero,
		},
		Movements: []Movement{cardMovement},
	}

	cs := CardSummary{
		StatementContext: statement,
		Table: Table{
			PastPaymentMovements: []Movement{pastPaymentMovement},
			Cards:                []Card{card},
			TaxesMovements:       []Movement{taxMovement},
		},
	}

	// When
	rows := NewMovementsWithCardContext(cs)

	// Then
	require.Equal(t, []MovementWithCardContext{
		{
			StatementContext: statement,
			CardContext:      nil,
			MovementType:     MovementTypePastPayment,
			Movement:         pastPaymentMovement,
		},
		{
			StatementContext: statement,
			CardContext:      nil,
			MovementType:     MovementTypeTax,
			Movement:         taxMovement,
		},
		{
			StatementContext: statement,
			CardContext:      &card.CardContext,
			MovementType:     MovementTypeCard,
			Movement:         cardMovement,
		},
	}, rows)
}
