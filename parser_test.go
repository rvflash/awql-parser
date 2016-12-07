package awql

import (
	"fmt"
	"strings"
)

// Ensure the scanner can scan tokens correctly.
func ExampleParser_Parse() {
	q := `SELECT CampaignName FROM CAMPAIGN_PERFORMACE_REPORT ORDER BY 1 LIMIT 5\G`
	stmts, _ := NewParser(strings.NewReader(q)).Parse()
	if stmt, ok := stmts[0].(SelectStmt); ok {
		fmt.Println(stmt.SourceName())
		fmt.Println(stmt.OrderList()[0].ColumnName)
		// Output: CAMPAIGN_PERFORMACE_REPORT
		// CampaignName
	}
}
