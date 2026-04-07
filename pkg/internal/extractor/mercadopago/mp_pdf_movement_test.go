package mercadopago

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMovementChunk(t *testing.T) {
	closeDate := time.Date(2025, time.October, 12, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		chunk       string
		wantDetail  string
		wantCurrent *int
		wantTotal   *int
		wantReceipt string
		wantARS     string
	}{
		{
			name:       "simple without installments",
			chunk:      "12/sepSOME MERCHANT$ 3.954,65",
			wantDetail: "SOME MERCHANT",
			wantARS:    "3954.65",
		},
		{
			name:        "with installments, no operation",
			chunk:       "23/agoMERPAGO*SHOP2 de 3$ 56.333,33",
			wantDetail:  "MERPAGO*SHOP",
			wantCurrent: intPtr(2),
			wantTotal:   intPtr(3),
			wantARS:     "56333.33",
		},
		{
			name:        "with installments and operation glued",
			chunk:       "19/octMERPAGO*PLACE3 de 3292153$ 12.666,66",
			wantDetail:  "MERPAGO*PLACE",
			wantCurrent: intPtr(3),
			wantTotal:   intPtr(3),
			wantReceipt: "292153",
			wantARS:     "12666.66",
		},
		{
			name:        "standalone operation",
			chunk:       "14/sepMERPAGO*RESTO461159$ 22.199,40",
			wantDetail:  "MERPAGO*RESTO",
			wantReceipt: "461159",
			wantARS:     "22199.40",
		},
		{
			name:       "detail with spaces",
			chunk:      "6/octESTACION 59 SA$ 30.000,43",
			wantDetail: "ESTACION 59 SA",
			wantARS:    "30000.43",
		},
		{
			name:       "detail with dash",
			chunk:      "10/octSOME BAR-ALIAS$ 50.500,00",
			wantDetail: "SOME BAR-ALIAS",
			wantARS:    "50500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mov := ParseMovementChunk(tt.chunk, closeDate)
			require.NotNil(t, mov, "movement should not be nil")

			assert.Equal(t, tt.wantDetail, mov.Detail)

			expectedARS, _ := decimal.NewFromString(tt.wantARS)
			assert.True(t, mov.AmountARS.Equal(expectedARS), "ARS: got %s, want %s", mov.AmountARS, expectedARS)

			if tt.wantCurrent != nil {
				require.NotNil(t, mov.CurrentInstallment)
				assert.Equal(t, *tt.wantCurrent, *mov.CurrentInstallment)
			} else {
				assert.Nil(t, mov.CurrentInstallment)
			}

			if tt.wantTotal != nil {
				require.NotNil(t, mov.TotalInstallments)
				assert.Equal(t, *tt.wantTotal, *mov.TotalInstallments)
			} else {
				assert.Nil(t, mov.TotalInstallments)
			}

			if tt.wantReceipt != "" {
				require.NotNil(t, mov.ReceiptNumber)
				assert.Equal(t, tt.wantReceipt, *mov.ReceiptNumber)
			}
		})
	}
}

func TestParseMovementChunk_SkipsSectionHeaders(t *testing.T) {
	closeDate := time.Date(2025, time.October, 12, 0, 0, 0, 0, time.UTC)

	skippable := []string{
		"12/sepResumen de cualquier mes$ 159.387,70",
		"17/sepPago de tarjeta-$ 159.387,70",
		"12/sepSubtotal$ 159.387,70",
	}

	for _, chunk := range skippable {
		t.Run(chunk, func(t *testing.T) {
			mov := ParseMovementChunk(chunk, closeDate)
			assert.Nil(t, mov)
		})
	}
}

func TestSplitMovementLines_ConcatenatedEntries(t *testing.T) {
	closeDate := time.Date(2026, time.January, 12, 0, 0, 0, 0, time.UTC)

	text := "19/octMERPAGO*A3 de 3292153$ 12.666,661/novMERPAGO*B3 de 3180225$ 6.070,00"
	movements := SplitMovementLines(text, closeDate)

	require.Len(t, movements, 2)
	assert.Equal(t, "MERPAGO*A", movements[0].Detail)
	assert.Equal(t, "MERPAGO*B", movements[1].Detail)
}

func TestSplitMovementLines_PageBreak(t *testing.T) {
	closeDate := time.Date(2025, time.October, 12, 0, 0, 0, 0, time.UTC)

	text := "9/octSOME STORE$ 3.180,00\n10/octANOTHER STORE$ 50.500,00"
	movements := SplitMovementLines(text, closeDate)

	require.Len(t, movements, 2)
	assert.Equal(t, "SOME STORE", movements[0].Detail)
	assert.Equal(t, "ANOTHER STORE", movements[1].Detail)
}

func intPtr(i int) *int {
	return &i
}
