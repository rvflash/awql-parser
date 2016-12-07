# Awql Parser

 Parser for parsing AWQL SELECT statements.
 
 ## Example
 
 ```golang
 q := `SELECT CampaignName FROM CAMPAIGN_PERFORMACE_REPORT ORDER BY 1 LIMIT 5\G`
 	stmts, _ := NewParser(strings.NewReader(q)).Parse()
 	if stmt, ok := stmts[0].(SelectStmt); ok {
 		fmt.Println(stmt.OrderList()[0].ColumnName)
 		// Output: CampaignName
 	}
 ```