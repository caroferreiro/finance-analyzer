package pdfcardsummary

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetectCardCompanyFromText(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected CardCompany
	}{
		{
			name:     "VISA uppercase",
			text:     "VISA SIGNATURE",
			expected: CardCompany("VISA"),
		},
		{
			name:     "VISA lowercase",
			text:     "visa signature",
			expected: CardCompany("VISA"),
		},
		{
			name:     "VISA mixed case",
			text:     "Visa Signature card",
			expected: CardCompany("VISA"),
		},
		{
			name:     "AMEX uppercase",
			text:     "AMEX card",
			expected: CardCompany("AMEX"),
		},
		{
			name:     "American Express",
			text:     "American Express card",
			expected: CardCompany("AMEX"),
		},
		{
			name:     "AMEX lowercase",
			text:     "amex card",
			expected: CardCompany("AMEX"),
		},
		{
			name:     "Mastercard uppercase",
			text:     "MASTERCARD",
			expected: CardCompany("Mastercard"),
		},
		{
			name:     "Mastercard lowercase",
			text:     "mastercard",
			expected: CardCompany("Mastercard"),
		},
		{
			name:     "VISA priority over AMEX",
			text:     "VISA and AMEX cards",
			expected: CardCompany("VISA"),
		},
		{
			name:     "VISA priority over Mastercard",
			text:     "VISA and Mastercard",
			expected: CardCompany("VISA"),
		},
		{
			name:     "AMEX priority over Mastercard",
			text:     "AMEX and Mastercard",
			expected: CardCompany("AMEX"),
		},
		{
			name:     "No card company found",
			text:     "Some random text without card companies",
			expected: CardCompany("?"),
		},
		{
			name:     "Empty text",
			text:     "",
			expected: CardCompany("?"),
		},
		{
			name:     "VISA in Plan V text",
			text:     "Plan V de Visa",
			expected: CardCompany("VISA"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			text := tc.text

			// When
			result := DetectCardCompanyFromText(text)

			// Then
			require.Equal(t, tc.expected, result)
		})
	}
}
