package csvale

// CombineCSVMatrices combines multiple CSV matrices into a single CSV matrix.
// The first matrix must include headers (first row), subsequent matrices should
// also include headers but they will be skipped (only data rows appended).
// All matrices are assumed to have identical column structures (guaranteed by
// CardSummary type system).
func CombineCSVMatrices(matrices [][][]string) [][]string {
	if len(matrices) == 0 {
		return nil
	}

	combined := make([][]string, 0)
	// First matrix: include all rows (headers + data)
	combined = append(combined, matrices[0]...)

	// Subsequent matrices: skip header row (index 0), append only data rows
	for i := 1; i < len(matrices); i++ {
		if len(matrices[i]) > 0 {
			combined = append(combined, matrices[i][1:]...)
		}
	}

	return combined
}
