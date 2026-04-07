package pdfcardsummary

import (
	"strings"
	"testing"

	demodataset "github.com/Alechan/finance-analyzer/pkg/demo_dataset"
	"github.com/stretchr/testify/require"
)

func TestParseMovementsWithCardContextCSV_WhenDemoDataset_ThenParses(t *testing.T) {
	// Given
	csvBytes := []byte(demodataset.ExtractedCSV)

	// When
	rows, err := ParseMovementsWithCardContextCSV(csvBytes)

	// Then
	require.NoError(t, err)
	require.NotEmpty(t, rows)

	hasCardMovement := false
	hasNonCardMovement := false
	for _, row := range rows {
		if row.MovementType == MovementTypeCard {
			hasCardMovement = true
			require.NotNil(t, row.CardContext)
		} else {
			hasNonCardMovement = true
			require.Nil(t, row.CardContext)
		}
	}
	require.True(t, hasCardMovement)
	require.True(t, hasNonCardMovement)
}

func TestParseMovementsWithCardContextCSV_WhenHeaderMismatch_ThenReturnsError(t *testing.T) {
	// Given
	badCSV := strings.Replace(demodataset.ExtractedCSV, "AmountUSD", "AmountUSD_WRONG", 1)

	// When
	rows, err := ParseMovementsWithCardContextCSV([]byte(badCSV))

	// Then
	require.Nil(t, rows)
	require.EqualError(t, err, "invalid CSV header: expected [Bank CardCompany CloseDate ExpirationDate TotalARS TotalUSD CardNumber CardOwner CardTotalARS CardTotalUSD MovementType OriginalDate ReceiptNumber Detail CurrentInstallment TotalInstallments AmountARS AmountUSD], got [Bank CardCompany CloseDate ExpirationDate TotalARS TotalUSD CardNumber CardOwner CardTotalARS CardTotalUSD MovementType OriginalDate ReceiptNumber Detail CurrentInstallment TotalInstallments AmountARS AmountUSD_WRONG]")
}

func TestParseMovementsWithCardContextCSV_WhenInvalidMovementType_ThenReturnsError(t *testing.T) {
	// Given
	badCSV := strings.Replace(demodataset.ExtractedCSV, "PastPayment", "UNKNOWN_MOVEMENT", 1)

	// When
	rows, err := ParseMovementsWithCardContextCSV([]byte(badCSV))

	// Then
	require.Nil(t, rows)
	require.EqualError(t, err, "row 1 col \"MovementType\": invalid movement type \"UNKNOWN_MOVEMENT\"")
}

func TestParseMovementsWithCardContextCSV_WhenInvalidCloseDate_ThenReturnsError(t *testing.T) {
	// Given
	badCSV := strings.Replace(demodataset.ExtractedCSV, "2025-01-25", "2025/01/25", 1)

	// When
	rows, err := ParseMovementsWithCardContextCSV([]byte(badCSV))

	// Then
	require.Nil(t, rows)
	require.EqualError(t, err, "row 1 col \"CloseDate\": invalid date \"2025/01/25\": parsing time \"2025/01/25\" as \"2006-01-02\": cannot parse \"/01/25\" as \"-\"")
}

func TestParseMovementsWithCardContextCSV_WhenInvalidOriginalDate_ThenReturnsError(t *testing.T) {
	// Given
	badCSV := strings.Replace(demodataset.ExtractedCSV, "2025-01-24", "2025/01/24", 1)

	// When
	rows, err := ParseMovementsWithCardContextCSV([]byte(badCSV))

	// Then
	require.Nil(t, rows)
	require.EqualError(t, err, "row 2 col \"OriginalDate\": invalid date \"2025/01/24\": parsing time \"2025/01/24\" as \"2006-01-02\": cannot parse \"/01/24\" as \"-\"")
}

func TestParseMovementsWithCardContextCSV_WhenInvalidDecimal_ThenReturnsError(t *testing.T) {
	// Given
	badCSV := strings.Replace(demodataset.ExtractedCSV, "12.345,67", "12,34,56", 1)

	// When
	rows, err := ParseMovementsWithCardContextCSV([]byte(badCSV))

	// Then
	require.Nil(t, rows)
	require.ErrorContains(t, err, "row 3 col \"AmountARS\": invalid decimal \"12,34,56\"")
}

func TestParseMovementsWithCardContextCSV_WhenInvalidInstallmentInteger_ThenReturnsError(t *testing.T) {
	// Given
	badCSV := strings.Replace(demodataset.ExtractedCSV, ";1;3;5.000,00", ";X;3;5.000,00", 1)

	// When
	rows, err := ParseMovementsWithCardContextCSV([]byte(badCSV))

	// Then
	require.Nil(t, rows)
	require.EqualError(t, err, "row 7 col \"CurrentInstallment\": invalid integer \"X\": strconv.Atoi: parsing \"X\": invalid syntax")
}

func TestParseMovementsWithCardContextCSV_WhenNonCardMovementHasCardFields_ThenReturnsError(t *testing.T) {
	// Given
	badCSV := strings.Replace(demodataset.ExtractedCSV, "0,00;0,00;;;;;PastPayment", "0,00;0,00;0000;OWNER A;0,00;0,00;PastPayment", 1)

	// When
	rows, err := ParseMovementsWithCardContextCSV([]byte(badCSV))

	// Then
	require.Nil(t, rows)
	require.EqualError(t, err, "row 1: non-card movement type \"PastPayment\" must have empty card columns, got CardNumber=\"0000\" CardOwner=\"OWNER A\" CardTotalARS=\"0,00\" CardTotalUSD=\"0,00\"")
}
