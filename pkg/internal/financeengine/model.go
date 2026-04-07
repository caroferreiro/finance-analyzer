package financeengine

type ColumnType string

const (
	ColumnTypeString   ColumnType = "string"
	ColumnTypeNumber   ColumnType = "number"
	ColumnTypeDate     ColumnType = "date"
	ColumnTypeMoneyARS ColumnType = "money_ars"
	ColumnTypeMoneyUSD ColumnType = "money_usd"
	ColumnTypeShare    ColumnType = "share" // percentage (0–100), display with % suffix
)

type TableColumn struct {
	Key   string
	Label string
	Type  ColumnType
}

type Table struct {
	TableID     string
	Title       string
	Description string // Human-readable explanation. If you change the rows, update this.
	Columns     []TableColumn
	Rows        [][]string
}

type ComputeResult struct {
	Tables []Table
}
