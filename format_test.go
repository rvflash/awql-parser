package awqlparse_test

import (
	"strings"
	"testing"

	awql "github.com/rvflash/awql-parser"
)

func TestSelectStmt_String(t *testing.T) {
	var tests = []struct {
		fq, tq string
	}{
		{
			fq: `SELECT CampaignName FROM CAMPAIGN_PERFORMANCE_REPORT`,
			tq: `SELECT CampaignName FROM CAMPAIGN_PERFORMANCE_REPORT`,
		},
		{
			fq: `SELECT SUM(Cost) AS c FROM CAMPAIGN_PERFORMANCE_REPORT WHERE CampaignStatus = 'ENABLED'`,
			tq: `SELECT Cost FROM CAMPAIGN_PERFORMANCE_REPORT WHERE CampaignStatus = "ENABLED"`,
		},
		{
			fq: `SELECT CampaignName, Cost FROM CAMPAIGN_PERFORMANCE_REPORT GROUP BY 1 ORDER BY 2 DESC`,
			tq: `SELECT CampaignName, Cost FROM CAMPAIGN_PERFORMANCE_REPORT`,
		},
		{
			fq: `SELECT CampaignName FROM CAMPAIGN_PERFORMANCE_REPORT DURING 20161224,20161225 LIMIT 10`,
			tq: `SELECT CampaignName FROM CAMPAIGN_PERFORMANCE_REPORT DURING 20161224,20161225`,
		},
	}

	for i, qt := range tests {
		stmts, _ := awql.NewParser(strings.NewReader(qt.fq)).Parse()
		if stmt, ok := stmts[0].(awql.SelectStmt); ok {
			if q := stmt.String(); q != qt.tq {
				t.Errorf("%d. Expected the query '%v' with '%s', received '%v'", i, qt.tq, qt.fq, q)
			}
		}
	}
}
