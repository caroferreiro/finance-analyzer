package financeengine

import "fmt"

func FindTableByID(tables []Table, tableID string) (Table, error) {
	for _, table := range tables {
		if table.TableID == tableID {
			return table, nil
		}
	}

	return Table{}, fmt.Errorf("table %q not found", tableID)
}
