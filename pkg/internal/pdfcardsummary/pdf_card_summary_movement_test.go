package pdfcardsummary

import (
	"testing"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/platform/testsale"
	"github.com/stretchr/testify/require"
)

func TestMovement_IdentifiableInfo(t *testing.T) {
	testCases := []struct {
		name           string
		movement       Movement
		expectedOutput string
	}{
		{
			name: "movement with all fields",
			movement: Movement{
				OriginalDate:       testsale.DatePtr(2024, time.August, 10),
				ReceiptNumber:      testsale.StrPtr("123456*"),
				Detail:             "PURCHASE AT STORE",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
			},
			expectedOutput: `date: 2024-08-10, detail: "PURCHASE AT STORE", receipt: 123456*`,
		},
		{
			name: "movement with nil date",
			movement: Movement{
				OriginalDate:       nil,
				ReceiptNumber:      testsale.StrPtr("789012*"),
				Detail:             "SALDO ANTERIOR",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
			},
			expectedOutput: `date: <nil>, detail: "SALDO ANTERIOR", receipt: 789012*`,
		},
		{
			name: "movement with nil receipt",
			movement: Movement{
				OriginalDate:       testsale.DatePtr(2024, time.January, 15),
				ReceiptNumber:      nil,
				Detail:             "TAX CHARGE",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
			},
			expectedOutput: `date: 2024-01-15, detail: "TAX CHARGE", receipt: <nil>`,
		},
		{
			name: "movement with nil date and receipt",
			movement: Movement{
				OriginalDate:       nil,
				ReceiptNumber:      nil,
				Detail:             "SOME DETAIL",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
			},
			expectedOutput: `date: <nil>, detail: "SOME DETAIL", receipt: <nil>`,
		},
		{
			name: "movement with empty detail",
			movement: Movement{
				OriginalDate:       testsale.DatePtr(2024, time.December, 25),
				ReceiptNumber:      testsale.StrPtr("111222*"),
				Detail:             "",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
			},
			expectedOutput: `date: 2024-12-25, detail: "", receipt: 111222*`,
		},
		{
			name: "movement with detail containing special characters",
			movement: Movement{
				OriginalDate:       testsale.DatePtr(2024, time.March, 8),
				ReceiptNumber:      testsale.StrPtr("333444*"),
				Detail:             "PAGO EN CUOTAS - MARÍA GARCÍA",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
			},
			expectedOutput: `date: 2024-03-08, detail: "PAGO EN CUOTAS - MARÍA GARCÍA", receipt: 333444*`,
		},
		{
			name: "movement with long detail",
			movement: Movement{
				OriginalDate:       testsale.DatePtr(2024, time.July, 4),
				ReceiptNumber:      testsale.StrPtr("555666*"),
				Detail:             "SUPERMERCADO DISCO SUCURSAL 123 AVENIDA CORRIENTES 1234",
				CurrentInstallment: nil,
				TotalInstallments:  nil,
			},
			expectedOutput: `date: 2024-07-04, detail: "SUPERMERCADO DISCO SUCURSAL 123 AVENIDA CORRIENTES 1234", receipt: 555666*`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When
			actual := tc.movement.IdentifiableInfo()

			// Then
			require.Equal(t, tc.expectedOutput, actual)
		})
	}
}
