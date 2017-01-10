package awqlparse

import "strconv"

// String formats a SelectStmt as expected by Google Adwords.
// Indeed, aggregate functions, ORDER BY, GROUP BY and LIMIT are not supported for reports.
// Implements fmt.Stringer interface.
func (s SelectStatement) String() (q string) {
	if len(s.Columns()) == 0 || s.SourceName() == "" {
		return
	}

	// Concats selected fields.
	q = "SELECT "
	for i, c := range s.Columns() {
		if i > 0 {
			q += ", "
		}
		q += c.Name()
	}

	// Adds data source name.
	q += " FROM " + s.SourceName()

	// Conditions.
	if len(s.ConditionList()) > 0 {
		q += " WHERE "
		for i, c := range s.ConditionList() {
			if i > 0 {
				q += " AND "
			}
			q += c.Name() + " " + c.Operator()
			val, lit := c.Value()
			if len(val) > 1 {
				q += " ["
				for y, v := range val {
					if y > 0 {
						q += " ,"
					}
					if lit {
						q += " " + v
					} else {
						q += " " + strconv.Quote(v)
					}
				}
				q += " ]"
			} else if lit {
				q += " " + val[0]
			} else {
				q += " " + strconv.Quote(val[0])
			}
		}
	}

	// Range date
	d := s.DuringList()
	if ds := len(d); ds > 0 {
		q += " DURING "
		if ds == 2 {
			q += d[0] + "," + d[1]
		} else {
			// Literal range date
			q += d[0]
		}
	}

	return
}
