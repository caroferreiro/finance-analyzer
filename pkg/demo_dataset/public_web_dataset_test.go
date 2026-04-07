package demodataset

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Alechan/finance-analyzer/pkg/internal/financeengine"
	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestPublicWebDemoDataset_CanBeParsedIntoRepoCompatibleStructs(t *testing.T) {
	csvPath := filepath.Join("..", "..", "web", "mockups_lab", "tmp_public_data", "current", "demo_extracted.csv")
	csvBytes, err := os.ReadFile(csvPath)
	require.NoError(t, err)

	rows, err := pdfcardsummary.ParseMovementsWithCardContextCSV(csvBytes)
	require.NoError(t, err)
	require.NotEmpty(t, rows)
}

func TestPublicWebDemoDataset_LatestMonthKeepsOverviewCoverage(t *testing.T) {
	csvPath := filepath.Join("..", "..", "web", "mockups_lab", "tmp_public_data", "current", "demo_extracted.csv")
	csvBytes, err := os.ReadFile(csvPath)
	require.NoError(t, err)

	rows, err := pdfcardsummary.ParseMovementsWithCardContextCSV(csvBytes)
	require.NoError(t, err)
	require.NotEmpty(t, rows)

	metricsRows := financeengine.New().OverviewMetricsByStatementMonth(rows)
	require.GreaterOrEqual(t, len(metricsRows), 2)

	previous := metricsRows[len(metricsRows)-2]
	latest := metricsRows[len(metricsRows)-1]

	require.True(t, previous.CarryOverDebtARS.GreaterThan(decimal.Zero) || previous.CarryOverDebtUSD.GreaterThan(decimal.Zero))
	require.True(t, latest.CarryOverDebtARS.GreaterThan(decimal.Zero) || latest.CarryOverDebtUSD.GreaterThan(decimal.Zero))

	require.True(t, previous.NextMonthDebtARS.GreaterThan(decimal.Zero) || previous.NextMonthDebtUSD.GreaterThan(decimal.Zero))
	require.True(t, latest.NextMonthDebtARS.GreaterThan(decimal.Zero) || latest.NextMonthDebtUSD.GreaterThan(decimal.Zero))

	require.True(t, previous.RemainingDebtARS.GreaterThan(decimal.Zero) || previous.RemainingDebtUSD.GreaterThan(decimal.Zero))
	require.True(t, latest.RemainingDebtARS.GreaterThan(decimal.Zero) || latest.RemainingDebtUSD.GreaterThan(decimal.Zero))

	require.True(t, previous.TaxesARS.GreaterThan(decimal.Zero) || previous.TaxesUSD.GreaterThan(decimal.Zero))
	require.True(t, latest.TaxesARS.GreaterThan(decimal.Zero) || latest.TaxesUSD.GreaterThan(decimal.Zero))

	require.True(t, previous.PastPaymentsARS.GreaterThan(decimal.Zero) || previous.PastPaymentsUSD.GreaterThan(decimal.Zero))
	require.True(t, latest.PastPaymentsARS.GreaterThan(decimal.Zero) || latest.PastPaymentsUSD.GreaterThan(decimal.Zero))

	require.True(t, previous.CardMovementsUSD.GreaterThan(decimal.Zero))
	require.True(t, latest.CardMovementsUSD.GreaterThan(decimal.Zero))
}
