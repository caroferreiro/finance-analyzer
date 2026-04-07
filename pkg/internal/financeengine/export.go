package financeengine

import (
	"bytes"
	"encoding/csv"
	"fmt"
)

func ExportTableCSV(table Table) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Comma = ';'

	header := make([]string, 0, len(table.Columns))
	for _, c := range table.Columns {
		header = append(header, c.Key)
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("error writing CSV header: %v", err)
	}

	for i, row := range table.Rows {
		if len(row) != len(table.Columns) {
			return "", fmt.Errorf("row %d has %d columns, expected %d", i+1, len(row), len(table.Columns))
		}
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("error writing CSV row %d: %v", i+1, err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("error flushing CSV writer: %v", err)
	}

	return buf.String(), nil
}
