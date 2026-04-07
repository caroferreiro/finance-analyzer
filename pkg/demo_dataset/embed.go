package demodataset

import _ "embed"

var (
	// ExtractedCSV is a synthetic/anonymized extracted CSV fixture.
	//
	//go:embed extracted.csv
	ExtractedCSV string

	// MappingsV1JSON is a minimal mappings fixture consumed by the dashboard.
	//
	//go:embed mappings.v1.json
	MappingsV1JSON string
)
