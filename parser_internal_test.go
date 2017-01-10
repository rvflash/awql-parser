package awqlparse

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Ensure the parser can parse strings into CREATE VIEW Statement.
func TestParser_ParseCreateView(t *testing.T) {
	var queryTests = []struct {
		q    string
		stmt *CreateViewStatement
		err  string
	}{
		// Simple statement.
		{
			q: `CREATE VIEW CAMPAIGN_DAILY AS SELECT SUM(DISTINCT Cost) FROM CAMPAIGN_PERFORMANCE_REPORT`,
			stmt: &CreateViewStatement{
				DataStatement: DataStatement{
					TableName: "CAMPAIGN_DAILY",
				},
				View: &SelectStatement{
					DataStatement: DataStatement{
						Fields: []DynamicField{
							&DynamicColumn{Column: &Column{ColumnName: "Cost"}, Method: "SUM", Unique: true},
						},
						TableName: "CAMPAIGN_PERFORMANCE_REPORT",
					},
				},
			},
		},

		// Replace statement with explicit column names.
		{
			q: `CREATE OR REPLACE VIEW CAMPAIGN_DAILY (Date, Adspend) AS SELECT Date, SUM(DISTINCT Cost) FROM CAMPAIGN_PERFORMANCE_REPORT GROUP BY 1`,
			stmt: &CreateViewStatement{
				DataStatement: DataStatement{
					TableName: "CAMPAIGN_DAILY",
					Fields: []DynamicField{
						&DynamicColumn{&Column{ColumnName: "Date"}, "", false},
						&DynamicColumn{&Column{ColumnName: "Adspend"}, "", false},
					},
				},
				View: &SelectStatement{
					DataStatement: DataStatement{
						Fields: []DynamicField{
							&DynamicColumn{&Column{ColumnName: "Date"}, "", false},
							&DynamicColumn{&Column{ColumnName: "Cost"}, "SUM", true},
						},
						TableName: "CAMPAIGN_PERFORMANCE_REPORT",
					},
					GroupBy: []FieldPosition{
						&ColumnPosition{&Column{ColumnName: "Date"}, 1},
					},
				},
				Replace: true,
			},
		},

		// Errors
		{q: `SELECT`, err: fmt.Sprintf(ErrMsgBadMethod, "SELECT")},
		{q: `CREATE VIEW !`, err: fmt.Sprintf(ErrMsgBadSrc, "!")},
		{q: `CREATE VIEW CAMPAIGN_DAILY (Name, Cost) AS SELECT SUM(DISTINCT Cost) FROM CAMPAIGN_PERFORMANCE_REPORT`, err: ErrMsgColumnsNotMatch},
	}

	for i, qt := range queryTests {
		stmt, err := NewParser(strings.NewReader(qt.q)).ParseCreateView()
		if err != nil {
			if qt.err != err.Error() {
				t.Errorf("%d. Expected the error message %v with %s, received %v", i, qt.err, qt.q, err.Error())
			}
		} else if qt.err != "" {
			t.Errorf("%d. Expected the error message %v with %s, received no error", i, qt.err, qt.q)
		} else if !reflect.DeepEqual(qt.stmt, stmt) {
			t.Errorf("%d. Expected %#v, received %#v", i, qt.stmt, stmt)
		}
	}
}

// Ensure the parser can parse strings into DESCRIBE Statement.
func TestParser_ParseDescribe(t *testing.T) {
	var queryTests = []struct {
		q    string
		stmt *DescribeStatement
		err  string
	}{
		// Simple statement.
		{
			q: `DESC CAMPAIGN_PERFORMANCE_REPORT`,
			stmt: &DescribeStatement{
				DataStatement: DataStatement{
					TableName: "CAMPAIGN_PERFORMANCE_REPORT",
				},
			},
		},

		// Simple statement with alias of the method.
		{
			q: `DESCRIBE CAMPAIGN_PERFORMANCE_REPORT`,
			stmt: &DescribeStatement{
				DataStatement: DataStatement{
					TableName: "CAMPAIGN_PERFORMANCE_REPORT",
				},
			},
		},

		// Full statement.
		{
			q: `DESC FULL CAMPAIGN_PERFORMANCE_REPORT CampaignName\G`,
			stmt: &DescribeStatement{
				FullStatement: FullStatement{Full: true},
				DataStatement: DataStatement{
					Fields: []DynamicField{
						&DynamicColumn{&Column{ColumnName: "CampaignName"}, "", false},
					},
					TableName: "CAMPAIGN_PERFORMANCE_REPORT",
					Statement: Statement{GModifier: true},
				},
			},
		},

		// Errors
		{q: `SELECT`, err: fmt.Sprintf(ErrMsgBadMethod, "SELECT")},
		{q: `DESC !`, err: fmt.Sprintf(ErrMsgBadSrc, "!")},
	}

	for i, qt := range queryTests {
		stmt, err := NewParser(strings.NewReader(qt.q)).ParseDescribe()
		if err != nil {
			if qt.err != err.Error() {
				t.Errorf("%d. Expected the error message %v with %s, received %v", i, qt.err, qt.q, err.Error())
			}
		} else if qt.err != "" {
			t.Errorf("%d. Expected the error message %v with %s, received no error", i, qt.err, qt.q)
		} else if !reflect.DeepEqual(qt.stmt, stmt) {
			t.Errorf("%d. Expected %#v, received %#v", i, qt.stmt, stmt)
		}
	}
}

// Ensure the parser can parse strings into SHOW Statement.
func TestParser_ParseShow(t *testing.T) {
	var queryTests = []struct {
		q    string
		stmt *ShowStatement
		err  string
	}{
		// Simple statement.
		{
			q:    `SHOW TABLES`,
			stmt: &ShowStatement{},
		},

		// Full statement.
		{
			q: `SHOW FULL TABLES\G`,
			stmt: &ShowStatement{
				FullStatement: FullStatement{Full: true},
				Statement:     Statement{GModifier: true},
			},
		},

		// Show statement like something as prefix.
		{
			q: `SHOW TABLES LIKE 'CAMPAIGN%'\G`,
			stmt: &ShowStatement{
				Statement: Statement{GModifier: true},
				Like:      Pattern{Prefix: "CAMPAIGN"},
			},
		},

		// Show statement like something as suffix.
		{
			q: `SHOW TABLES LIKE '%REPORT'\G`,
			stmt: &ShowStatement{
				Statement: Statement{GModifier: true},
				Like:      Pattern{Suffix: "REPORT"},
			},
		},

		// Show statement like something.
		{
			q: `SHOW TABLES LIKE '%NEGATIVE%'`,
			stmt: &ShowStatement{
				Like: Pattern{Contains: "NEGATIVE"},
			},
		},

		// Show statement named something.
		{
			q: `SHOW TABLES LIKE 'LABEL';`,
			stmt: &ShowStatement{
				Like: Pattern{Equal: "LABEL"},
			},
		},

		// Show statement with a specific column.
		{
			q: `SHOW TABLES WITH CampaignName;`,
			stmt: &ShowStatement{
				With:    "CampaignName",
				UseWith: true,
			},
		},

		// Show statement with a specific column.
		{
			q: `SHOW TABLES WITH "CampaignName";`,
			stmt: &ShowStatement{
				With:    "CampaignName",
				UseWith: true,
			},
		},

		// Show statement with no column.
		{
			q: `SHOW TABLES WITH "";`,
			stmt: &ShowStatement{
				With:    "",
				UseWith: true,
			},
		},

		// Errors
		{q: `SELECT`, err: fmt.Sprintf(ErrMsgBadMethod, "SELECT")},
		{q: `SHOW`, err: fmt.Sprintf(ErrMsgSyntax, "")},
		{q: `SHOW TABLES LIKE rv`, err: fmt.Sprintf(ErrMsgSyntax, "rv")},
		{q: `SHOW TABLES LABEL`, err: fmt.Sprintf(ErrMsgSyntax, "LABEL")},
	}

	for i, qt := range queryTests {
		stmt, err := NewParser(strings.NewReader(qt.q)).ParseShow()
		if err != nil {
			if qt.err != err.Error() {
				t.Errorf("%d. Expected the error message %v with %s, received %v", i, qt.err, qt.q, err.Error())
			}
		} else if qt.err != "" {
			t.Errorf("%d. Expected the error message %v with %s, received no error", i, qt.err, qt.q)
		} else if !reflect.DeepEqual(qt.stmt, stmt) {
			t.Errorf("%d. Expected %#v, received %#v", i, qt.stmt, stmt)
		}
	}
}

// Ensure the parser can parse strings into SELECT Statement.
func TestParser_ParseSelect(t *testing.T) {
	var queryTests = []struct {
		q    string
		stmt *SelectStatement
		err  string
	}{
		// Single field statement.
		{
			q: `SELECT CampaignName FROM CAMPAIGN_PERFORMANCE_REPORT`,
			stmt: &SelectStatement{
				DataStatement: DataStatement{
					Fields: []DynamicField{
						&DynamicColumn{&Column{ColumnName: "CampaignName"}, "", false},
					},
					TableName: "CAMPAIGN_PERFORMANCE_REPORT",
				},
			},
		},

		// Multi-fields statement with vertical display.
		{
			q: `SELECT CampaignId, CampaignName, Cost FROM CAMPAIGN_PERFORMANCE_REPORT\G`,
			stmt: &SelectStatement{
				DataStatement: DataStatement{
					Fields: []DynamicField{
						&DynamicColumn{&Column{ColumnName: "CampaignId"}, "", false},
						&DynamicColumn{&Column{ColumnName: "CampaignName"}, "", false},
						&DynamicColumn{&Column{ColumnName: "Cost"}, "", false},
					},
					TableName: "CAMPAIGN_PERFORMANCE_REPORT",
					Statement: Statement{GModifier: true},
				},
			},
		},

		// Select all statement on view with condition and range date.
		{
			q: `SELECT * FROM CAMPAIGN_DAILY WHERE CampaignId = 12345678 DURING YESTERDAY;`,
			stmt: &SelectStatement{
				DataStatement: DataStatement{
					Fields: []DynamicField{
						&DynamicColumn{&Column{ColumnName: "*"}, "", false},
					},
					TableName: "CAMPAIGN_DAILY",
				},
				Where: []Condition{
					&Where{&Column{ColumnName: "CampaignId"}, "=", []string{"12345678"}, true},
				},
				During: []string{"YESTERDAY"},
			},
		},

		// Select statement with aggregate function and alias with row count limit.
		{
			q: `SELECT MAX(Cost) as max FROM CAMPAIGN_PERFORMANCE_REPORT LIMIT 5\G`,
			stmt: &SelectStatement{
				DataStatement: DataStatement{
					Fields: []DynamicField{
						&DynamicColumn{&Column{ColumnName: "Cost", ColumnAlias: "max"}, "MAX", false},
					},
					TableName: "CAMPAIGN_PERFORMANCE_REPORT",
					Statement: Statement{GModifier: true},
				},
				Limit: Limit{0, 5, true},
			},
		},

		// Select statement with aggregate function with distinct inside.
		{
			q: `SELECT SUM(distinct Cost) FROM CAMPAIGN_PERFORMANCE_REPORT`,
			stmt: &SelectStatement{
				DataStatement: DataStatement{
					Fields: []DynamicField{
						&DynamicColumn{&Column{ColumnName: "Cost"}, "SUM", true},
					},
					TableName: "CAMPAIGN_PERFORMANCE_REPORT",
				},
			},
		},

		// Select statement with distinct column with alias, ordering and limit with offset and row count.
		{
			q: `SELECT DISTINCT Cost as c FROM CAMPAIGN_PERFORMANCE_REPORT DURING 20161224,20161224 ORDER BY 1 DESC LIMIT 15, 5;`,
			stmt: &SelectStatement{
				DataStatement: DataStatement{
					Fields: []DynamicField{
						&DynamicColumn{&Column{ColumnName: "Cost", ColumnAlias: "c"}, "", true},
					},
					TableName: "CAMPAIGN_PERFORMANCE_REPORT",
				},
				During: []string{"20161224", "20161224"},
				OrderBy: []Orderer{
					&Order{&ColumnPosition{&Column{ColumnName: "Cost", ColumnAlias: "c"}, 1}, true},
				},
				Limit: Limit{15, 5, true},
			},
		},

		// Select statement with group by and string value list.
		{
			q: `SELECT Date, Cost FROM CAMPAIGN_PERFORMANCE_REPORT WHERE CampaignStatus IN ["ENABLED","PAUSED"] DURING LAST_WEEK GROUP BY 1;`,
			stmt: &SelectStatement{
				DataStatement: DataStatement{
					Fields: []DynamicField{
						&DynamicColumn{&Column{ColumnName: "Date"}, "", false},
						&DynamicColumn{&Column{ColumnName: "Cost"}, "", false},
					},
					TableName: "CAMPAIGN_PERFORMANCE_REPORT",
				},
				Where: []Condition{
					&Where{&Column{ColumnName: "CampaignStatus"}, "IN", []string{"ENABLED", "PAUSED"}, false},
				},
				During: []string{"LAST_WEEK"},
				GroupBy: []FieldPosition{
					&ColumnPosition{&Column{ColumnName: "Date"}, 1},
				},
			},
		},

		// Select statement with value literal list and EOF as ending.
		{
			q: `SELECT Cost FROM CAMPAIGN_PERFORMANCE_REPORT WHERE CampaignId IN [123456789,987654321]`,
			stmt: &SelectStatement{
				DataStatement: DataStatement{
					Fields: []DynamicField{
						&DynamicColumn{&Column{ColumnName: "Cost"}, "", false},
					},
					TableName: "CAMPAIGN_PERFORMANCE_REPORT",
				},
				Where: []Condition{
					&Where{&Column{ColumnName: "CampaignId"}, "IN", []string{"123456789", "987654321"}, true},
				},
			},
		},

		// Errors
		{q: `DELETE`, err: fmt.Sprintf(ErrMsgBadMethod, "DELETE")},
		{q: `SELECT !`, err: fmt.Sprintf(ErrMsgBadField, "!")},
		{q: `SELECT CampaignId Impressions`, err: ErrMsgMissingSrc},
		{q: `SELECT CampaignId FROM`, err: fmt.Sprintf(ErrMsgBadSrc, "")},
		{q: `SELECT CampaignId FROM REPORT WHERE`, err: fmt.Sprintf(ErrMsgBadField, "")},
		{q: `SELECT CampaignId FROM REPORT GROUP`, err: fmt.Sprintf(ErrMsgBadGroup, "")},
		{q: `SELECT CampaignId FROM REPORT GROUP BY ,`, err: fmt.Sprintf(ErrMsgBadGroup, ",")},
		{q: `SELECT CampaignId FROM REPORT GROUP BY 2`, err: fmt.Sprintf(ErrMsgBadGroup, fmt.Sprintf(ErrMsgBadColumn, "2"))},
		{q: `SELECT CampaignId FROM REPORT ORDER 1`, err: fmt.Sprintf(ErrMsgBadOrder, "1")},
		{q: `SELECT CampaignId FROM REPORT LIMIT 1 SELECT`, err: fmt.Sprintf(ErrMsgSyntax, "SELECT")},
		{q: `SELECT CampaignId FROM REPORT LIMIT`, err: fmt.Sprintf(ErrMsgBadLimit, "")},
		{q: `SELECT DISTINCT 1 FROM CAMPAIGN_PERFORMANCE_REPORT`, err: fmt.Sprintf(ErrMsgBadField, "1")},
		{q: `SELECT rv(Cost) FROM CAMPAIGN_PERFORMANCE_REPORT`, err: fmt.Sprintf(ErrMsgBadFunc, "rv")},
		{q: `SELECT CampaignId FROM CAMPAIGN_PERFORMANCE_REPORT WHERE CampaignName ! "rv"`, err: fmt.Sprintf(ErrMsgSyntax, "!")},
		{q: `SELECT CampaignId FROM CAMPAIGN_PERFORMANCE_REPORT WHERE CampaignName = !`, err: fmt.Sprintf(ErrMsgSyntax, "!")},
		{q: `SELECT CampaignId FROM CAMPAIGN_PERFORMANCE_REPORT WHERE CampaignName IN [ !`, err: fmt.Sprintf(ErrMsgSyntax, "[")},
		{q: `SELECT CampaignId FROM CAMPAIGN_PERFORMANCE_REPORT DURING`, err: fmt.Sprintf(ErrMsgBadDuring, "")},
		{q: `SELECT CampaignId FROM CAMPAIGN_PERFORMANCE_REPORT DURING RV`, err: fmt.Sprintf(ErrMsgBadDuring, "RV")},
		{q: `SELECT CampaignId FROM CAMPAIGN_PERFORMANCE_REPORT DURING TODAY, YESTERDAY`, err: fmt.Sprintf(ErrMsgBadDuring, IErrMsgDuringDateSize)},
		{q: `SELECT CampaignId FROM CAMPAIGN_PERFORMANCE_REPORT DURING 201612`, err: fmt.Sprintf(ErrMsgBadDuring, "201612")},
		{q: `SELECT CampaignId FROM CAMPAIGN_PERFORMANCE_REPORT DURING 20161224`, err: fmt.Sprintf(ErrMsgBadDuring, IErrMsgDuringLitSize)},
		{q: `SELECT CampaignId FROM CAMPAIGN_PERFORMANCE_REPORT DURING 20161224,20161225,20161226`, err: fmt.Sprintf(ErrMsgBadDuring, IErrMsgDuringSize)},
		{q: `SELECT Cost FROM CAMPAIGN_PERFORMANCE_REPORT WHERE CampaignStatus IN ["ENABLED",PAUSED];`, err: fmt.Sprintf(ErrMsgSyntax, "[")},
		{q: `SELECT Cost FROM CAMPAIGN_PERFORMANCE_REPORT WHERE CampaignStatus IN [PAUSED,"ENABLED"];`, err: fmt.Sprintf(ErrMsgSyntax, "[")},
	}

	for i, qt := range queryTests {
		stmt, err := NewParser(strings.NewReader(qt.q)).ParseSelect()
		if err != nil {
			if qt.err != err.Error() {
				t.Errorf("%d. Expected the error message %v with %s, received %v", i, qt.err, qt.q, err.Error())
			}
		} else if qt.err != "" {
			t.Errorf("%d. Expected the error message %v with %s, received no error", i, qt.err, qt.q)
		} else if !reflect.DeepEqual(qt.stmt, stmt) {
			t.Errorf("%d. Expected %#v, received %#v", i, qt.stmt, stmt)
		}
	}
}
