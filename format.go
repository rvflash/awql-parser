package awqlparse

import "strconv"

// String formats a SelectStmt as expected by Google Adwords.
// Indeed, aggregate functions, ORDER BY, GROUP BY and LIMIT are not supported for reports.
// Implements fmt.Stringer interface.
func (s SelectStatement) String() (q string) {
	if len(s.Fields) == 0 || s.TableName == "" {
		return
	}
	// Concat selected fields.
	q = "SELECT "
	for i, c := range s.Fields {
		if i > 0 {
			q += ", "
		}
		q += c.ColumnName
	}
	// Data source
	q += " FROM " + s.TableName
	// Conditions
	if len(s.Where) > 0 {
		q += " WHERE "
		for i, c := range s.Where {
			if i > 0 {
				q += " AND "
			}
			q += c.ColumnName + " " + c.Operator
			if len(c.Value) > 1 {
				q += " ["
				for y, v := range c.Value {
					if y > 0 {
						q += " ,"
					}
					if c.IsValueLiteral {
						q += " " + v
					} else {
						q += " " + strconv.Quote(v)
					}
				}
				q += " ]"
			} else if c.IsValueLiteral {
				q += " " + c.Value[0]
			} else {
				q += " " + strconv.Quote(c.Value[0])
			}
		}
	}
	// Range date
	d := s.During
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
