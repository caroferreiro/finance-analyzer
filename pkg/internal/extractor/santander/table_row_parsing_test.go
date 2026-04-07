package santander

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/santander/testdata"

	"github.com/Alechan/finance-analyzer/pkg/internal/extractor/pdftable"
	"github.com/stretchr/testify/require"
)

func TestSantanderTableRowParsing_WithRealisticStatement(t *testing.T) {
	cfg := DefaultConfig()

	// Given: realistic_statement.txt
	lines := strings.Split(string(testdata.RawRealisticStatementData), "\n")

	// When: Parse to rows (filtering out comments and empty lines)
	var actualRows []pdftable.Row
	factory := pdftable.NewRowFactory(cfg.TableColumnPositions)
	for _, line := range lines {
		row := factory.CreateRow(line)
		actualRows = append(actualRows, row)
	}

	// Then: Compare with RealisticStatementRows
	require.Equal(t, testdata.RealisticStatementRows, actualRows)
}
