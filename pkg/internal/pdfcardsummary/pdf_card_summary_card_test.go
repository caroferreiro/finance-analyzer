package pdfcardsummary

import (
	"testing"

	"github.com/Alechan/finance-analyzer/pkg/internal/platform/testsale"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCard_IdentifiableInfo(t *testing.T) {
	testCases := []struct {
		name           string
		card           Card
		expectedOutput string
	}{
		{
			name: "card with number, owner, and movements",
			card: Card{
				CardContext: CardContext{
					CardNumber:   testsale.StrPtr("1234"),
					CardOwner:    "TEST USER",
					CardTotalARS: decimal.Zero,
					CardTotalUSD: decimal.Zero,
				},
				Movements: []Movement{{Detail: "MOVEMENT1"}, {Detail: "MOVEMENT2"}},
			},
			expectedOutput: `owner: "TEST USER", number: 1234, movements: 2`,
		},
		{
			name: "card with nil number",
			card: Card{
				CardContext: CardContext{
					CardNumber:   nil,
					CardOwner:    "JANE DOE",
					CardTotalARS: decimal.Zero,
					CardTotalUSD: decimal.Zero,
				},
				Movements: []Movement{{Detail: "MOVEMENT1"}},
			},
			expectedOutput: `owner: "JANE DOE", number: <nil>, movements: 1`,
		},
		{
			name: "card with no movements",
			card: Card{
				CardContext: CardContext{
					CardNumber:   testsale.StrPtr("5678"),
					CardOwner:    "BOB JONES",
					CardTotalARS: decimal.Zero,
					CardTotalUSD: decimal.Zero,
				},
				Movements: []Movement{},
			},
			expectedOutput: `owner: "BOB JONES", number: 5678, movements: 0`,
		},
		{
			name: "card with empty owner",
			card: Card{
				CardContext: CardContext{
					CardNumber:   testsale.StrPtr("9999"),
					CardOwner:    "",
					CardTotalARS: decimal.Zero,
					CardTotalUSD: decimal.Zero,
				},
				Movements: []Movement{{Detail: "MOVEMENT1"}},
			},
			expectedOutput: `owner: "", number: 9999, movements: 1`,
		},
		{
			name: "card with many movements",
			card: Card{
				CardContext: CardContext{
					CardNumber:   testsale.StrPtr("1111"),
					CardOwner:    "ALICE BROWN",
					CardTotalARS: decimal.Zero,
					CardTotalUSD: decimal.Zero,
				},
				Movements: []Movement{
					{Detail: "MOVEMENT1"},
					{Detail: "MOVEMENT2"},
					{Detail: "MOVEMENT3"},
					{Detail: "MOVEMENT4"},
					{Detail: "MOVEMENT5"},
				},
			},
			expectedOutput: `owner: "ALICE BROWN", number: 1111, movements: 5`,
		},
		{
			name: "card with owner containing special characters",
			card: Card{
				CardContext: CardContext{
					CardNumber:   testsale.StrPtr("2222"),
					CardOwner:    "MARÍA GARCÍA-LÓPEZ",
					CardTotalARS: decimal.Zero,
					CardTotalUSD: decimal.Zero,
				},
				Movements: []Movement{{Detail: "MOVEMENT1"}},
			},
			expectedOutput: `owner: "MARÍA GARCÍA-LÓPEZ", number: 2222, movements: 1`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When
			actual := tc.card.IdentifiableInfo()

			// Then
			require.Equal(t, tc.expectedOutput, actual)
		})
	}
}

func TestAddCardsAmounts(t *testing.T) {
	testCases := []struct {
		name        string
		cards       []Card
		expectedARS decimal.Decimal
		expectedUSD decimal.Decimal
	}{
		{
			name: "single card",
			cards: []Card{
				{
					CardContext: CardContext{
						CardTotalARS: testsale.AsDecimal(t, "1000.00"),
						CardTotalUSD: testsale.AsDecimal(t, "50.00"),
					},
				},
			},
			expectedARS: testsale.AsDecimal(t, "1000.00"),
			expectedUSD: testsale.AsDecimal(t, "50.00"),
		},
		{
			name: "multiple cards",
			cards: []Card{
				{
					CardContext: CardContext{
						CardTotalARS: testsale.AsDecimal(t, "1000.00"),
						CardTotalUSD: testsale.AsDecimal(t, "50.00"),
					},
				},
				{
					CardContext: CardContext{
						CardTotalARS: testsale.AsDecimal(t, "2000.00"),
						CardTotalUSD: testsale.AsDecimal(t, "100.00"),
					},
				},
				{
					CardContext: CardContext{
						CardTotalARS: testsale.AsDecimal(t, "500.00"),
						CardTotalUSD: testsale.AsDecimal(t, "25.00"),
					},
				},
			},
			expectedARS: testsale.AsDecimal(t, "3500.00"),
			expectedUSD: testsale.AsDecimal(t, "175.00"),
		},
		{
			name:        "empty cards slice",
			cards:       []Card{},
			expectedARS: decimal.Zero,
			expectedUSD: decimal.Zero,
		},
		{
			name: "cards with zero amounts",
			cards: []Card{
				{
					CardContext: CardContext{
						CardTotalARS: decimal.Zero,
						CardTotalUSD: decimal.Zero,
					},
				},
				{
					CardContext: CardContext{
						CardTotalARS: decimal.Zero,
						CardTotalUSD: decimal.Zero,
					},
				},
			},
			expectedARS: decimal.Zero,
			expectedUSD: decimal.Zero,
		},
		{
			name: "cards with negative amounts",
			cards: []Card{
				{
					CardContext: CardContext{
						CardTotalARS: testsale.AsDecimal(t, "-1000.00"),
						CardTotalUSD: testsale.AsDecimal(t, "-50.00"),
					},
				},
				{
					CardContext: CardContext{
						CardTotalARS: testsale.AsDecimal(t, "2000.00"),
						CardTotalUSD: testsale.AsDecimal(t, "100.00"),
					},
				},
			},
			expectedARS: testsale.AsDecimal(t, "1000.00"),
			expectedUSD: testsale.AsDecimal(t, "50.00"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When
			actualARS, actualUSD := AddCardsAmounts(tc.cards)

			// Then
			// For zero values, check IsZero() instead of exact equality due to different internal representations
			if tc.expectedARS.IsZero() {
				require.True(t, actualARS.IsZero(), "expected ARS to be zero, got %s", actualARS)
			} else {
				require.Equal(t, tc.expectedARS, actualARS)
			}
			if tc.expectedUSD.IsZero() {
				require.True(t, actualUSD.IsZero(), "expected USD to be zero, got %s", actualUSD)
			} else {
				require.Equal(t, tc.expectedUSD, actualUSD)
			}
		})
	}
}
