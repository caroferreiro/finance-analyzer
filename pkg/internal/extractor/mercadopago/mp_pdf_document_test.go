package mercadopago

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractTotals(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		expectedARS string
		expectedUSD string
	}{
		{
			name:        "ARS and USD",
			text:        "Total a pagar$ 71.085,96US$ 0,00",
			expectedARS: "71085.96",
			expectedUSD: "0",
		},
		{
			name:        "ARS only",
			text:        "Total a pagar$ 354.174,29",
			expectedARS: "354174.29",
			expectedUSD: "0",
		},
		{
			name:        "uses last occurrence",
			text:        "Total a pagar$ 71.08596...Resumen anterior$ 306.675,21...Total a pagar$ 71.085,96US$ 0,00",
			expectedARS: "71085.96",
			expectedUSD: "0",
		},
		{
			name:        "not found",
			text:        "no total here",
			expectedARS: "0",
			expectedUSD: "0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ars, usd := extractTotals(tt.text)
			expectedARS, _ := decimal.NewFromString(tt.expectedARS)
			expectedUSD, _ := decimal.NewFromString(tt.expectedUSD)
			assert.True(t, ars.Equal(expectedARS), "ARS: got %s, want %s", ars, expectedARS)
			assert.True(t, usd.Equal(expectedUSD), "USD: got %s, want %s", usd, expectedUSD)
		})
	}
}

func TestExtractCardOwner(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "found",
			text:     "Este resumen pertenece a John Doe, CUIL 20123456789, DNI 12345678",
			expected: "John Doe",
		},
		{
			name:     "with extra whitespace",
			text:     "pertenece a  Jane Smith , CUIL 27111222333",
			expected: "Jane Smith",
		},
		{
			name:     "not found",
			text:     "no owner in this text",
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractCardOwner(tt.text))
		})
	}
}

func TestExtractPastPayments(t *testing.T) {
	text := "12/sepResumen de septiembre$ 159.387,70Subtotal$ 159.387,70Pagos realizados17/sepPago de tarjeta-$ 159.387,70"
	closeDate := mustDate(2025, 10, 12)

	payments := extractPastPayments(text, closeDate)
	require.Len(t, payments, 2)

	assert.Equal(t, "SALDO ANTERIOR", payments[0].Detail)
	expected, _ := decimal.NewFromString("159387.70")
	assert.True(t, payments[0].AmountARS.Equal(expected))

	assert.Equal(t, "PAGO DE TARJETA", payments[1].Detail)
	expectedNeg, _ := decimal.NewFromString("-159387.70")
	assert.True(t, payments[1].AmountARS.Equal(expectedNeg))
}

func TestExtractPastPayments_NoPastPayments(t *testing.T) {
	text := "ConsumosCon tarjeta virtual23/ago..."
	closeDate := mustDate(2025, 10, 12)

	payments := extractPastPayments(text, closeDate)
	assert.Empty(t, payments)
}

func TestExtractConsumosSection(t *testing.T) {
	text := `PrevStuffCon tarjeta virtualFechaDescripciónPesosDólares5/octMERCHANT$ 100,00Subtotal$ 100,00Impuestos`

	section := extractConsumosSection(text)
	assert.Contains(t, section, "5/oct")
	assert.Contains(t, section, "MERCHANT")
	assert.NotContains(t, section, "Subtotal")
	assert.NotContains(t, section, "Con tarjeta virtual")
}

func TestExtractConsumosSection_MultiPage(t *testing.T) {
	text := `Con tarjeta virtualFechaDescripciónPesos1/octA$ 100,00
DETALLE DE MOVIMIENTOSFechaDescripciónCuotaPesos2/octB$ 200,00Subtotal$ 300,00`

	section := extractConsumosSection(text)
	assert.Contains(t, section, "1/oct")
	assert.Contains(t, section, "2/oct")
	assert.NotContains(t, section, "DETALLE DE MOVIMIENTOS")
}

func TestExtractConsumosSection_NotFound(t *testing.T) {
	assert.Empty(t, extractConsumosSection("no consumos here"))
}
