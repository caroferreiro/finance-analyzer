package pdfcardsummary

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetectBankFromText(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected Bank
	}{
		{
			name:     "Santander with Banco prefix",
			text:     "BANCO SANTANDER RIO S.A",
			expected: Bank("Santander"),
		},
		{
			name:     "Santander without prefix",
			text:     "Santander queremos ayudarte",
			expected: Bank("Santander"), // Word-based matching: "Santander" matches
		},
		{
			name:     "Galicia with Banco prefix",
			text:     "Banco Galicia",
			expected: Bank("Galicia"),
		},
		{
			name:     "Galicia without prefix",
			text:     "Galicia bank information",
			expected: Bank("Galicia"), // Word-based matching: "Galicia" matches
		},
		{
			name:     "HSBC",
			text:     "HSBC Bank Argentina",
			expected: Bank("HSBC"),
		},
		{
			name:     "HSBC lowercase",
			text:     "hsbc information",
			expected: Bank("HSBC"),
		},
		{
			name:     "Multiple banks - first match",
			text:     "Banco Santander and Banco Galicia banks",
			expected: Bank("Santander"), // Word-based matching: "banco" matches first (maps to Santander)
		},
		{
			name:     "No bank found",
			text:     "Some random text without bank names",
			expected: Bank("?"),
		},
		{
			name:     "Empty text",
			text:     "",
			expected: Bank("?"),
		},
		{
			name:     "Case insensitive with Banco prefix",
			text:     "banco SANTANDER Banco Santander",
			expected: Bank("Santander"),
		},
		{
			name:     "BBVA acronym",
			text:     "BBVA information",
			expected: Bank("BBVA"),
		},
		{
			name:     "BBVA with Banco prefix",
			text:     "Banco BBVA Frances",
			expected: Bank("BBVA"),
		},
		{
			name:     "BBVA uppercase",
			text:     "BANCO BBVA FRANCES S.A",
			expected: Bank("BBVA"),
		},
		{
			name:     "BBVA lowercase",
			text:     "banco bbva frances",
			expected: Bank("BBVA"),
		},
		{
			name:     "Macro single word",
			text:     "Banco Macro",
			expected: Bank("Macro"),
		},
		{
			name:     "Macro with additional text",
			text:     "Banco Macro Argentina",
			expected: Bank("Macro"),
		},
		{
			name:     "Bank not in whitelist returns ?",
			text:     "Banco Nuevo Banco",
			expected: Bank("?"),
		},
		{
			name:     "Banco Central returns ? (not in whitelist)",
			text:     "Banco Central de la República Argentina",
			expected: Bank("?"),
		},
		{
			name:     "Santander with Rio suffix",
			text:     "Banco Santander Rio",
			expected: Bank("Santander"),
		},
		{
			name:     "ICBC acronym",
			text:     "ICBC bank",
			expected: Bank("ICBC"),
		},
		{
			name:     "BIND acronym",
			text:     "BIND information",
			expected: Bank("BIND"),
		},
		{
			name:     "Brubank",
			text:     "Brubank bank",
			expected: Bank("Brubank"),
		},
		{
			name:     "Naranja X",
			text:     "Naranja X bank",
			expected: Bank("Naranja X"),
		},
		{
			name:     "Mercado Pago",
			text:     "Mercado Pago information",
			expected: Bank("Mercado Pago"),
		},
		{
			name:     "Ualá",
			text:     "Ualá bank",
			expected: Bank("Ualá"),
		},
		{
			name:     "Mercado Pago - full phrase",
			text:     "Mercado Pago information",
			expected: Bank("Mercado Pago"),
		},
		{
			name:     "Naranja X - single word",
			text:     "Naranja information",
			expected: Bank("Naranja X"),
		},
		{
			name:     "Naranja X - full phrase",
			text:     "Naranja X information",
			expected: Bank("Naranja X"),
		},
		{
			name:     "Case insensitive - lowercase santander",
			text:     "santander queremos ayudarte",
			expected: Bank("Santander"),
		},
		{
			name:     "Punctuation handling - santander with punctuation",
			text:     "Santander, queremos ayudarte!",
			expected: Bank("Santander"),
		},
		{
			name:     "Punctuation handling - bbva with dots",
			text:     "B.B.V.A. information",
			expected: Bank("BBVA"),
		},
		{
			name:     "First match wins - santander before galicia",
			text:     "Santander and Galicia banks",
			expected: Bank("Santander"),
		},
		{
			name:     "Mercado Pago - concatenated PDF text (deMercado Pago.)",
			text:     "sección Actividad, en la app deMercado Pago.Este resumen pertenece",
			expected: Bank("Mercado Pago"),
		},
		{
			name:     "Mercado Pago - concatenated PDF text (Pago.2)",
			text:     "Ayuda, en la web o la appde Mercado Pago.2- Desconocer consumos",
			expected: Bank("Mercado Pago"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			text := tc.text

			// When
			result := DetectBankFromText(text)

			// Then
			require.Equal(t, tc.expected, result)
		})
	}
}
