package main

import (
	"encoding/json"
	"fmt"
	"os"

	demodataset "github.com/Alechan/finance-analyzer/pkg/demo_dataset"
	"github.com/Alechan/finance-analyzer/pkg/internal/financeengine"
	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
)

func main() {
	rows, err := pdfcardsummary.ParseMovementsWithCardContextCSV([]byte(demodataset.ExtractedCSV))
	if err != nil {
		exitWithError("failed to parse demo extracted CSV", err)
	}

	var mappings financeengine.Mappings
	if err := json.Unmarshal([]byte(demodataset.MappingsV1JSON), &mappings); err != nil {
		exitWithError("failed to parse demo mappings", err)
	}

	engine := financeengine.New()
	result := engine.Compute(rows, mappings)

	fmt.Printf("table_count=%d\n", len(result.Tables))
	fmt.Println("tables:")
	for _, table := range result.Tables {
		fmt.Printf("- %s (%s): columns=%d rows=%d\n", table.TableID, table.Title, len(table.Columns), len(table.Rows))
	}

	table, found := findTableByID(result.Tables, financeengine.TableIDOverviewByMonth)
	if !found {
		exitWithError("overview table not found", fmt.Errorf("missing table id %q", financeengine.TableIDOverviewByMonth))
	}

	csvText, err := financeengine.ExportTableCSV(table)
	if err != nil {
		exitWithError("failed to export overview table CSV", err)
	}

	fmt.Println()
	fmt.Printf("csv_export_table=%s\n", table.TableID)
	fmt.Println(csvText)
}

func findTableByID(tables []financeengine.Table, tableID string) (financeengine.Table, bool) {
	for _, table := range tables {
		if table.TableID == tableID {
			return table, true
		}
	}
	return financeengine.Table{}, false
}

func exitWithError(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	os.Exit(1)
}
