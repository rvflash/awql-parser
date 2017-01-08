package awqlparse

import (
	"fmt"
	"strings"
)

// Ensure the parser can parse statements correctly.
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

// Ensure the parser can parse statements correctly and return only the first.
func ExampleParser_ParseRow() {
	q := `SELECT CampaignName FROM CAMPAIGN_PERFORMACE_REPORT;`
	stmt, _ := NewParser(strings.NewReader(q)).ParseRow()
	if stmt, ok := stmt.(SelectStmt); ok {
		fmt.Println(stmt.SourceName())
		// Output: CAMPAIGN_PERFORMACE_REPORT
	}
}
