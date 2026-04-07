//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	demodataset "github.com/Alechan/finance-analyzer/pkg/demo_dataset"
	"github.com/Alechan/finance-analyzer/pkg/internal/financeengine"
	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
)

func main() {
	js.Global().Set("computeFromCSV", js.FuncOf(computeFromCSV))
	js.Global().Set("exportTableCSVFromResult", js.FuncOf(exportTableCSVFromResult))
	js.Global().Set("demoCSV", js.FuncOf(demoCSV))
	js.Global().Set("demoMappingsJSON", js.FuncOf(demoMappingsJSON))

	select {}
}

func computeFromCSV(_ js.Value, args []js.Value) any {
	if len(args) < 2 {
		return errorResult("computeFromCSV expects 2 args: csvText, mappingsJSON")
	}

	csvText := args[0].String()
	mappingsJSON := args[1].String()

	rows, err := pdfcardsummary.ParseMovementsWithCardContextCSV([]byte(csvText))
	if err != nil {
		return errorResult(fmt.Sprintf("error parsing CSV: %v", err))
	}

	var mappings financeengine.Mappings
	if err := json.Unmarshal([]byte(mappingsJSON), &mappings); err != nil {
		return errorResult(fmt.Sprintf("error parsing mappings JSON: %v", err))
	}

	engine := financeengine.New()
	result := engine.Compute(rows, mappings)

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return errorResult(fmt.Sprintf("error serializing compute result: %v", err))
	}

	return okResult(string(resultBytes))
}

func exportTableCSVFromResult(_ js.Value, args []js.Value) any {
	if len(args) < 2 {
		return errorResult("exportTableCSVFromResult expects 2 args: computeResultJSON, tableID")
	}

	computeResultJSON := args[0].String()
	tableID := args[1].String()

	var result financeengine.ComputeResult
	if err := json.Unmarshal([]byte(computeResultJSON), &result); err != nil {
		return errorResult(fmt.Sprintf("error parsing compute result JSON: %v", err))
	}

	table, err := financeengine.FindTableByID(result.Tables, tableID)
	if err != nil {
		return errorResult(err.Error())
	}

	csvText, err := financeengine.ExportTableCSV(table)
	if err != nil {
		return errorResult(fmt.Sprintf("error exporting table CSV: %v", err))
	}

	return okResult(csvText)
}

func demoCSV(_ js.Value, _ []js.Value) any {
	return okResult(demodataset.ExtractedCSV)
}

func demoMappingsJSON(_ js.Value, _ []js.Value) any {
	return okResult(demodataset.MappingsV1JSON)
}

func okResult(value string) map[string]any {
	return map[string]any{
		"ok":    true,
		"value": value,
	}
}

func errorResult(message string) map[string]any {
	return map[string]any{
		"ok":    false,
		"error": message,
	}
}
