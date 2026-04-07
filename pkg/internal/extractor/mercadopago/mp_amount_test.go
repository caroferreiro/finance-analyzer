package mercadopago

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseARSAmount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "large positive", input: "$ 354.174,29", expected: "354174.29"},
		{name: "negative", input: "-$ 159.387,70", expected: "-159387.70"},
		{name: "small with thousands", input: "$ 3.954,65", expected: "3954.65"},
		{name: "no thousands separator", input: "$ 500,00", expected: "500"},
		{name: "zero", input: "$ 0,00", expected: "0"},
		{name: "millions", input: "$ 1.200.000,00", expected: "1200000"},
		{name: "cents only", input: "$ 0,99", expected: "0.99"},
		{name: "no space after $", input: "$100,50", expected: "100.50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseARSAmount(tt.input)
			expected, err := decimal.NewFromString(tt.expected)
			require.NoError(t, err)
			assert.True(t, got.Equal(expected), "got %s, want %s", got, expected)
		})
	}
}
