package stringsale

import (
	"fmt"
	"strings"
	"time"
)

func IntPtrToString(i *int, nilReplacement string) string {
	if i == nil {
		return nilReplacement
	}
	return fmt.Sprintf("%d", *i)
}

func StringPtrToString(i *string, nilReplacement string) string {
	if i == nil {
		return nilReplacement
	}
	return *i
}

func TimePtrToYearMonthAndDateString(t *time.Time, nilReplacement string) string {
	if t == nil {
		return nilReplacement
	}
	return t.Format("2006-01-02")
}

func SliceOfMapsToSliceOfStrings(rows []map[string]string, keysToKeep []string) ([][]string, error) {
	var innerRecords [][]string
	for i, rowMap := range rows {
		row := make([]string, 0, len(keysToKeep))
		for _, colName := range keysToKeep {
			e, ok := rowMap[colName]
			if !ok {
				return nil, fmt.Errorf("missing column %q in CSV row %d", colName, i)
			}
			row = append(row, e)
		}
		innerRecords = append(innerRecords, row)
	}
	return innerRecords, nil
}

// RemoveDuplicateSpaces removes duplicate whitespaces from the input string.
func RemoveDuplicateSpaces(input string) string {
	// Split the string into fields, which removes extra spaces
	parts := strings.Fields(input)

	// Join the fields back together with a single space
	return strings.Join(parts, " ")
}
