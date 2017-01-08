package awqlparse

import (
	"fmt"
	"strings"
)

// Ensure the parser can parse statements correctly.
func ExampleParser_Parse() {
	q := `SELECT CampaignName FROM CAMPAIGN_PERFORMANCE_REPORT ORDER BY 1 LIMIT 5\GDESC ADGROUP_PERFORMANCE_REPORT AdGroupName;`
	stmts, _ := NewParser(strings.NewReader(q)).Parse()
	for _, stmt := range stmts {
		switch stmt.(type) {
		case SelectStmt:
			fmt.Println(stmt.(SelectStmt).OrderList()[0].ColumnName)
		case DescribeStmt:
			fmt.Println(stmt.(DescribeStmt).SourceName())
		}
	}
	// Output:
	// CampaignName
	// ADGROUP_PERFORMANCE_REPORT
}

// Ensure the parser can parse statements correctly and return only the first.
func ExampleParser_ParseRow() {
	q := `SELECT CampaignName FROM CAMPAIGN_PERFORMANCE_REPORT;`
	stmt, _ := NewParser(strings.NewReader(q)).ParseRow()
	if stmt, ok := stmt.(SelectStmt); ok {
		fmt.Println(stmt.SourceName())
		// Output: CAMPAIGN_PERFORMANCE_REPORT
	}
}

// Ensure the parser can parse select statement.
func ExampleParser_ParseSelect() {
	q := `SELECT AdGroupName FROM ADGROUP_PERFORMANCE_REPORT;`
	stmt, _ := NewParser(strings.NewReader(q)).ParseSelect()
	fmt.Println(stmt.SourceName())
	// Output: ADGROUP_PERFORMANCE_REPORT
}
