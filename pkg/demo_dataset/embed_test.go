package demodataset

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type mappingsV1 struct {
	OwnersByCardOwner  map[string]string `json:"ownersByCardOwner"`
	OwnersByCardNumber map[string]string `json:"ownersByCardNumber"`
	CategoryByDetail   map[string]string `json:"categoryByDetail"`
}

func TestDemoDataset_ExtractedCSV_IsNonEmptyAndHasExpectedHeader(t *testing.T) {
	// Given
	expectedHeader := strings.Join([]string{
		"Bank",
		"CardCompany",
		"CloseDate",
		"ExpirationDate",
		"TotalARS",
		"TotalUSD",
		"CardNumber",
		"CardOwner",
		"CardTotalARS",
		"CardTotalUSD",
		"MovementType",
		"OriginalDate",
		"ReceiptNumber",
		"Detail",
		"CurrentInstallment",
		"TotalInstallments",
		"AmountARS",
		"AmountUSD",
	}, ";")

	// When
	lines := strings.Split(strings.TrimSpace(ExtractedCSV), "\n")

	// Then
	require.GreaterOrEqual(t, len(lines), 2, "expected header + at least one row")
	require.Equal(t, expectedHeader, strings.TrimSpace(lines[0]))
}

func TestDemoDataset_MappingsV1JSON_IsValidJSON(t *testing.T) {
	// Given
	var m mappingsV1

	// When
	err := json.Unmarshal([]byte(MappingsV1JSON), &m)

	// Then
	require.NoError(t, err)
	require.NotEmpty(t, m.OwnersByCardOwner)
	require.NotEmpty(t, m.CategoryByDetail)
}
