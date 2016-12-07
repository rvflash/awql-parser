package awql

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Ensure the parser can parse strings into Statement ASTs.
func TestParser_parseSelect(t *testing.T) {
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
					Fields: []Field{
						{Column{ColumnName: "CampaignName"}, "", false},
					},
					TableName: "CAMPAIGN_PERFORMANCE_REPORT",
					Statement: Statement{GModifier: false},
				},
				Where:   []Condition(nil),
				During:  []string(nil),
				GroupBy: []*ColumnPosition(nil),
				OrderBy: []*Ordering(nil),
				Limit:   Limit{},
			},
		},

		// Multi-fields statement with vertical display.
		{
			q: `SELECT CampaignId, CampaignName, Cost FROM CAMPAIGN_PERFORMANCE_REPORT\G`,
			stmt: &SelectStatement{
				DataStatement: DataStatement{
					Fields: []Field{
						{Column{ColumnName: "CampaignId"}, "", false},
						{Column{ColumnName: "CampaignName"}, "", false},
						{Column{ColumnName: "Cost"}, "", false},
					},
					TableName: "CAMPAIGN_PERFORMANCE_REPORT",
					Statement: Statement{GModifier: true},
				},
				Where:   []Condition(nil),
				During:  []string(nil),
				GroupBy: []*ColumnPosition(nil),
				OrderBy: []*Ordering(nil),
				Limit:   Limit{},
			},
		},

		// Select all statement on view with condition and range date.
		{
			q: `SELECT * FROM CAMPAIGN_DAILY WHERE CampaignId = 12345678 DURING YESTERDAY;\G`,
			stmt: &SelectStatement{
				DataStatement: DataStatement{
					Fields: []Field{
						{Column{ColumnName: "*"}, "", false},
					},
					TableName: "CAMPAIGN_DAILY",
					Statement: Statement{GModifier: false},
				},
				Where: []Condition{
					{Column{ColumnName: "CampaignId"}, "=", []string{"12345678"}, true},
				},
				During:  []string{"YESTERDAY"},
				GroupBy: []*ColumnPosition(nil),
				OrderBy: []*Ordering(nil),
				Limit:   Limit{},
			},
		},

		// Errors
		{q: `DELETE`, err: fmt.Sprintf(ErrMsgBadMethod, "DELETE")},
		{q: `SELECT !`, err: fmt.Sprintf(ErrMsgBadField, "!")},
		{q: `SELECT CampaignName Impressions`, err: ErrMsgMissingSrc},
		{q: `SELECT CampaignName FROM`, err: fmt.Sprintf(ErrMsgBadSrc, "")},
		{q: `SELECT CampaignName FROM REPORT WHERE`, err: fmt.Sprintf(ErrMsgBadField, "")},
		{q: `SELECT CampaignName FROM REPORT LIMIT`, err: fmt.Sprintf(ErrMsgBadLimit, "")},
	}

	for i, qt := range queryTests {
		stmt, err := NewParser(strings.NewReader(qt.q)).parseSelect()
		if err != nil {
			if qt.err != err.Error() {
				t.Errorf("%d. Expected the error message %v with %s, received %v", i, qt.err, qt.q, err.Error())
			}
			continue
		} else if qt.err != "" {
			t.Errorf("%d. Expected the error message %v with %s, received no error", i, qt.err, qt.q)
		} else if !reflect.DeepEqual(qt.stmt, stmt) {
			t.Errorf("%d. Expected %#v, received %#v", i, qt.stmt, stmt)
		}
		fmt.Println(stmt)
	}
}
