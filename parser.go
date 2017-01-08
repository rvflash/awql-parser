package awqlparse

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Like with %
const wildcard = "%"

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		t Token  // last read token
		l string // last read literal
		n int    // buffer size, char by char, maximum value: 1
	}
}

// Error messages
var (
	ErrMsgBadStmt        = "ParserError.UNKWOWN_STATEMENT"
	ErrMsgMissingSrc     = "ParserError.MISSING_SOURCE"
	ErrMsgBadColumn      = "ParserError.UNKNOWN_COLUMN (%s)"
	ErrMsgBadMethod      = "ParserError.INVALID_METHOD (%s)"
	ErrMsgBadField       = "ParserError.INVALID_FIELD (%s)"
	ErrMsgBadFunc        = "ParserError.INVALID_FUNCTION (%s)"
	ErrMsgBadSrc         = "ParserError.INVALID_SOURCE (%s)"
	ErrMsgBadDuring      = "ParserError.INVALID_DURING (%s)"
	ErrMsgBadGroup       = "ParserError.INVALID_GROUP_BY (%s)"
	ErrMsgBadOrder       = "ParserError.INVALID_ORDER_BY (%s)"
	ErrMsgBadLimit       = "ParserError.INVALID_LIMIT (%s)"
	ErrMsgSyntax         = "ParserError.SYNTAX_NEAR (%s)"
	ErrMsgDuringSize     = "unexpected number of date range"
	ErrMsgDuringLitSize  = "expected date range literal"
	ErrMsgDuringDateSize = "expected no literal date"
)

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// Parse parses a AWQL statement.
func (p *Parser) Parse() (statements []Stmt, err error) {
	for {
		var stmt Stmt
		// Retrieve the first token of the statement.
		tk, _ := p.scanIgnoreWhitespace()
		switch tk {
		case DESC, DESCRIBE:
			p.unscan()
			stmt, err = p.ParseDescribe()
		case CREATE:
			p.unscan()
			stmt, err = p.ParseCreateView()
		case SELECT:
			p.unscan()
			stmt, err = p.ParseSelect()
		case SHOW:
			p.unscan()
			stmt, err = p.ParseShow()
		default:
			err = errors.New(ErrMsgBadStmt)
		}
		if err != nil {
			return
		}
		statements = append(statements, stmt)

		// If the next token is EOF, break the loop.
		if tk, _ := p.scanIgnoreWhitespace(); tk == EOF {
			break
		} else {
			p.unscan()
		}
	}
	return
}

// ParseRow parses a AWQL statement and returns only the first.
func (p *Parser) ParseRow() (Stmt, error) {
	stmts, err := p.Parse()
	if err != nil {
		return nil, err
	}
	return stmts[0], nil
}

// ParseDescribe parses a AWQL DESCRIBE statement.
func (p *Parser) ParseDescribe() (DescribeStmt, error) {
	// First token should be a "DESC" keyword.
	if tk, literal := p.scanIgnoreWhitespace(); tk != DESC && tk != DESCRIBE {
		return nil, fmt.Errorf(ErrMsgBadMethod, literal)
	}
	stmt := &DescribeStatement{}

	// Next we may see the "FULL" keyword.
	if tk, _ := p.scanIgnoreWhitespace(); tk == FULL {
		stmt.Full = true
	} else {
		p.unscan()
	}

	// Next we should read the table name.
	if tk, literal := p.scanIgnoreWhitespace(); tk == IDENTIFIER {
		stmt.TableName = literal
	} else {
		return nil, fmt.Errorf(ErrMsgBadSrc, literal)
	}

	// Next we may see a column name.
	if tk, literal := p.scanIgnoreWhitespace(); tk == IDENTIFIER {
		field := Field{Column{ColumnName: literal}, "", false}
		stmt.Fields = append(stmt.Fields, field)
	} else {
		p.unscan()
	}

	// Finally, we should find the end of the query.
	var err error
	if stmt.GModifier, err = p.scanQueryEnding(); err != nil {
		return nil, err
	}
	return stmt, nil
}

// ParseCreateView parses a AWQL CREATE VIEW statement.
func (p *Parser) ParseCreateView() (CreateViewStmt, error) {
	// First token should be a "CREATE" keyword.
	if tk, literal := p.scanIgnoreWhitespace(); tk != CREATE {
		return nil, fmt.Errorf(ErrMsgBadMethod, literal)
	}
	stmt := &CreateViewStatement{}

	// Next we may see the "OR" keyword.
	if tk, _ := p.scanIgnoreWhitespace(); tk == OR {
		if tk, literal := p.scanIgnoreWhitespace(); tk != REPLACE {
			return nil, fmt.Errorf(ErrMsgSyntax, literal)
		}
		stmt.Replace = true
	} else {
		p.unscan()
	}

	// Next we should see the "VIEW" keyword.
	if tk, literal := p.scanIgnoreWhitespace(); tk != VIEW {
		return nil, fmt.Errorf(ErrMsgSyntax, literal)
	}

	// Next we should read the view name.
	tk, literal := p.scanIgnoreWhitespace()
	if tk != IDENTIFIER {
		return nil, fmt.Errorf(ErrMsgBadSrc, literal)
	}
	stmt.TableName = literal

	// Next we may see columns names.
	if tk, _ := p.scanIgnoreWhitespace(); tk == LEFT_PARENTHESIS {
		for {
			if tk, literal := p.scanIgnoreWhitespace(); tk == RIGHT_PARENTHESIS {
				break
			} else if tk == IDENTIFIER {
				stmt.Fields = append(stmt.Fields, Field{Column: Column{ColumnName: literal}})
			} else if tk == COMMA {
				// If the next token is not an "COMMA" then break the loop.
				continue
			} else {
				return nil, fmt.Errorf(ErrMsgBadField, literal)
			}
		}
	} else {
		p.unscan()
	}

	// Next we should see the "AS" keyword.
	if tk, literal := p.scanIgnoreWhitespace(); tk != AS {
		return nil, fmt.Errorf(ErrMsgSyntax, literal)
	}

	// And finally, the query source of the view.
	if selectStmt, err := p.ParseSelect(); err != nil {
		return nil, err
	} else {
		stmt.View = selectStmt.(*SelectStatement)
	}
	return stmt, nil
}

// ParseShow parses a AWQL SHOW statement.
func (p *Parser) ParseShow() (ShowStmt, error) {
	// First token should be a "SHOW" keyword.
	if tk, literal := p.scanIgnoreWhitespace(); tk != SHOW {
		return nil, fmt.Errorf(ErrMsgBadMethod, literal)
	}
	stmt := &ShowStatement{}

	// Next we may see the "FULL" keyword.
	if tk, _ := p.scanIgnoreWhitespace(); tk == FULL {
		stmt.Full = true
	} else {
		p.unscan()
	}

	// Next we should see the "TABLES" keyword.
	if tk, literal := p.scanIgnoreWhitespace(); tk != TABLES {
		return nil, fmt.Errorf(ErrMsgSyntax, literal)
	}

	// Next we may find a LIKE or WITH keyword.
	if clause, _ := p.scanIgnoreWhitespace(); clause == LIKE || clause == WITH {
		// And then, the search pattern.
		tk, pattern := p.scanIgnoreWhitespace()
		switch tk {
		case IDENTIFIER:
			if clause == LIKE {
				return nil, fmt.Errorf(ErrMsgSyntax, pattern)
			}
			stmt.With = pattern
		case STRING:
			if clause == LIKE {
				// Like clause can have a wildcard characters in the pattern.
				wl := strings.HasPrefix(pattern, wildcard)
				wr := strings.HasSuffix(pattern, wildcard)
				like := Pattern{}
				if wl == wr && wl {
					like.Contains = strings.Trim(pattern, wildcard)
				} else if wl == wr && !wl {
					like.Equal = pattern
				} else if wl {
					like.Suffix = strings.TrimPrefix(pattern, wildcard)
				} else if wr {
					like.Prefix = strings.TrimSuffix(pattern, wildcard)
				}
				stmt.Like = like
			} else {
				stmt.With = pattern
			}
		default:
			return nil, fmt.Errorf(ErrMsgSyntax, pattern)
		}
	} else {
		p.unscan()
	}

	// Finally, we should find the end of the query.
	var err error
	if stmt.GModifier, err = p.scanQueryEnding(); err != nil {
		return nil, err
	}
	return stmt, nil
}

// Parse parses a AWQL SELECT statement.
func (p *Parser) ParseSelect() (SelectStmt, error) {
	// First token should be a "SELECT" keyword.
	if tk, literal := p.scanIgnoreWhitespace(); tk != SELECT {
		return nil, fmt.Errorf(ErrMsgBadMethod, literal)
	}
	stmt := &SelectStatement{}

	// Next we should loop over all our comma-delimited fields.
	for {
		// Read a field.
		field := Field{}
		tk, literal := p.scanIgnoreWhitespace()
		switch tk {
		case ASTERISK:
			field.ColumnName = literal
		case DISTINCT:
			if err := p.scanDistinct(&field); err != nil {
				return nil, err
			}
		case IDENTIFIER:
			// Next we may find a function declaration.
			if tk, _ := p.scan(); tk != LEFT_PARENTHESIS {
				// Just a column name.
				field.ColumnName = literal
				p.unscan()
			} else if !isFunction(literal) {
				// This function does not exist.
				return nil, fmt.Errorf(ErrMsgBadFunc, literal)
			} else {
				// It is an aggregate function.
				field.Method = strings.ToUpper(literal)

				// Next we may read a distinct clause, a column position or just a column name.
				tk, literal = p.scanIgnoreWhitespace()
				switch tk {
				case ASTERISK:
					// Accept the rune '*' only with the count function.
					if field.Method != "COUNT" {
						return nil, fmt.Errorf(ErrMsgSyntax, literal)
					}
					field.ColumnName = literal
				case DISTINCT:
					if err := p.scanDistinct(&field); err != nil {
						return nil, err
					}
				case DIGIT:
					digit, _ := strconv.Atoi(literal)
					column, err := stmt.searchColumnByPosition(digit)
					if err != nil {
						return nil, fmt.Errorf(ErrMsgSyntax, literal)
					}
					field.Column = column.Column
				case IDENTIFIER:
					field.ColumnName = literal
				default:
					return nil, fmt.Errorf(ErrMsgBadFunc, literal)
				}

				// Next, we expect the end of the function.
				if tk, _ := p.scanIgnoreWhitespace(); tk != RIGHT_PARENTHESIS {
					return nil, fmt.Errorf(ErrMsgBadFunc, literal)
				}
			}
		default:
			return nil, fmt.Errorf(ErrMsgBadField, literal)
		}

		// Next we may find an alias name for the column.
		if tk, _ := p.scanIgnoreWhitespace(); tk == AS {
			// By using the "AS" keyword.
			if tk, literal := p.scanIgnoreWhitespace(); tk != IDENTIFIER {
				return nil, fmt.Errorf(ErrMsgBadField, literal)
			} else {
				field.ColumnAlias = literal
			}
		} else if tk == IDENTIFIER {
			// Or without keyword.
			field.ColumnAlias = literal
		} else {
			p.unscan()
		}
		// Finally, add this field with the others.
		stmt.Fields = append(stmt.Fields, field)

		// If the next token is not a comma then break the loop.
		if tk, _ := p.scanIgnoreWhitespace(); tk != COMMA {
			p.unscan()
			break
		}
	}

	// Next we should see the "FROM" keyword.
	if tk, _ := p.scanIgnoreWhitespace(); tk != FROM {
		return nil, errors.New(ErrMsgMissingSrc)
	}

	// Next we should read the table name.
	if tk, literal := p.scanIgnoreWhitespace(); tk != IDENTIFIER {
		return nil, fmt.Errorf(ErrMsgBadSrc, literal)
	} else {
		stmt.TableName = literal
	}

	// Newt we may read a "WHERE" keyword.
	if tk, _ := p.scanIgnoreWhitespace(); tk == WHERE {
		for {
			// Parse each condition, begin by the column name.
			cond := Condition{}
			if tk, literal := p.scanIgnoreWhitespace(); tk != IDENTIFIER {
				return nil, fmt.Errorf(ErrMsgBadField, literal)
			} else {
				cond.ColumnName = literal
			}
			// Expects the operator.
			if tk, literal := p.scanIgnoreWhitespace(); !isOperator(tk) {
				return nil, fmt.Errorf(ErrMsgSyntax, literal)
			} else {
				cond.Operator = literal
			}
			// And the value of the condition.ValueLiteral | String | ValueLiteralList | StringList
			tk, literal := p.scanIgnoreWhitespace()
			switch tk {
			case DECIMAL, DIGIT, VALUE_LITERAL:
				cond.IsValueLiteral = true
				fallthrough
			case STRING:
				cond.Value = append(cond.Value, literal)
			case LEFT_SQUARE_BRACKETS:
				p.unscan()
				if tk, cond.Value = p.scanValueList(); tk != VALUE_LITERAL_LIST && tk != STRING_LIST {
					return nil, fmt.Errorf(ErrMsgSyntax, literal)
				} else if tk == VALUE_LITERAL_LIST {
					cond.IsValueLiteral = true
				}
			default:
				return nil, fmt.Errorf(ErrMsgSyntax, literal)
			}
			stmt.Where = append(stmt.Where, cond)

			// If the next token is not an "AND" keyword then break the loop.
			if tk, _ := p.scanIgnoreWhitespace(); tk != AND {
				p.unscan()
				break
			}
		}
	} else {
		// No where clause.
		p.unscan()
	}

	// Next we may read a "DURING" keyword.
	if tk, _ := p.scanIgnoreWhitespace(); tk == DURING {
		var dateLiteral bool
		for {
			// Read the field used to group.
			tk, literal := p.scanIgnoreWhitespace()
			if tk == DIGIT && isDate(literal) {
				stmt.During = append(stmt.During, literal)
			} else if tk == IDENTIFIER && isDateRangeLiteral(literal) {
				stmt.During = append(stmt.During, literal)
				dateLiteral = true
			} else {
				return nil, fmt.Errorf(ErrMsgBadDuring, literal)
			}
			// If the next token is not a comma then break the loop.
			if tk, _ := p.scanIgnoreWhitespace(); tk != COMMA {
				p.unscan()
				break
			}
		}
		// Checks expected bounds.
		if rangeSize := len(stmt.During); rangeSize > 2 {
			return nil, fmt.Errorf(ErrMsgBadDuring, ErrMsgDuringSize)
		} else if rangeSize == 1 && !dateLiteral {
			return nil, fmt.Errorf(ErrMsgBadDuring, ErrMsgDuringLitSize)
		} else if rangeSize == 2 && dateLiteral {
			return nil, fmt.Errorf(ErrMsgBadDuring, ErrMsgDuringDateSize)
		}
	} else {
		// No during clause.
		p.unscan()
	}

	// Next we may see a "GROUP" keyword.
	if tk, _ := p.scanIgnoreWhitespace(); tk == GROUP {
		if tk, literal := p.scanIgnoreWhitespace(); tk != BY {
			return nil, fmt.Errorf(ErrMsgBadGroup, literal)
		}
		for {
			// Read the field used to group.
			tk, literal := p.scanIgnoreWhitespace()
			if tk != IDENTIFIER && tk != DIGIT {
				return nil, fmt.Errorf(ErrMsgBadGroup, literal)
			}
			// Check if the column exists as field.
			if groupBy, err := stmt.searchColumn(literal); err != nil {
				return nil, fmt.Errorf(ErrMsgBadGroup, err.Error())
			} else {
				stmt.GroupBy = append(stmt.GroupBy, groupBy)
			}
			// If the next token is not a comma then break the loop.
			if tk, _ := p.scanIgnoreWhitespace(); tk != COMMA {
				p.unscan()
				break
			}
		}
	} else {
		// No grouping clause.
		p.unscan()
	}

	// Next we may see a "ORDER" keyword.
	if tk, _ := p.scanIgnoreWhitespace(); tk == ORDER {
		if tk, literal := p.scanIgnoreWhitespace(); tk != BY {
			return nil, fmt.Errorf(ErrMsgBadOrder, literal)
		}
		for {
			// Read the field used to order.
			tk, literal := p.scanIgnoreWhitespace()
			if tk != IDENTIFIER && tk != DIGIT {
				return nil, fmt.Errorf(ErrMsgBadOrder, literal)
			}
			// Check if the column exists as field.
			orderBy := &Ordering{}
			if column, err := stmt.searchColumn(literal); err != nil {
				return nil, err
			} else {
				orderBy.ColumnPosition = *column
			}
			// Then, we may find a DESC or ASC keywords.
			if tk, _ = p.scanIgnoreWhitespace(); tk == DESC {
				orderBy.SortDesc = true
			} else if tk != ASC {
				p.unscan()
			}
			stmt.OrderBy = append(stmt.OrderBy, orderBy)

			// If the next token is not a comma then break the loop.
			if tk, _ := p.scanIgnoreWhitespace(); tk != COMMA {
				p.unscan()
				break
			}
		}
	} else {
		// No ordering clause.
		p.unscan()
	}

	// Next we may see a "LIMIT" keyword.
	if tk, _ := p.scanIgnoreWhitespace(); tk == LIMIT {
		var literal string
		if tk, literal = p.scanIgnoreWhitespace(); tk != DIGIT {
			return nil, fmt.Errorf(ErrMsgBadLimit, literal)
		}
		offset, _ := strconv.Atoi(literal)
		stmt.WithRowCount = true

		// If the next token is a comma then we should get the row count.
		if tk, _ := p.scanIgnoreWhitespace(); tk == COMMA {
			if tk, literal := p.scanIgnoreWhitespace(); tk != DIGIT {
				return nil, fmt.Errorf(ErrMsgBadLimit, stmt.RowCount)
			} else {
				stmt.Offset = offset
				stmt.RowCount, _ = strconv.Atoi(literal)
			}
		} else {
			// No row count value, so the offset is finally the row count.
			stmt.RowCount = offset
			p.unscan()
		}
	} else {
		// No limit clause.
		p.unscan()
	}

	// Finally, we should find the end of the query.
	var err error
	if stmt.GModifier, err = p.scanQueryEnding(); err != nil {
		return nil, err
	}
	return stmt, nil
}

// searchColumn returns the column matching the search expression.
func (s SelectStatement) searchColumn(expr string) (*ColumnPosition, error) {
	// If expr is a digit, search column by position.
	if pos, err := strconv.Atoi(expr); err == nil {
		if column, err := s.searchColumnByPosition(pos); err == nil {
			return column, nil
		}
		return nil, fmt.Errorf(ErrMsgBadColumn, expr)
	}
	// Otherwise fetch each column to find it by name or alias.
	for i, field := range s.Fields {
		if field.ColumnName == expr || field.ColumnAlias == expr {
			return &ColumnPosition{Column: field.Column, Position: (i + 1)}, nil
		}
	}
	return nil, fmt.Errorf(ErrMsgBadColumn, expr)
}

// searchColumnByPosition returns the column matching the search position.
func (s DataStatement) searchColumnByPosition(pos int) (*ColumnPosition, error) {
	if pos < 1 || pos > len(s.Fields) {
		return nil, fmt.Errorf(ErrMsgBadColumn, pos)
	}
	return &ColumnPosition{Column: s.Fields[(pos - 1)].Column, Position: pos}, nil
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (Token, string) {
	if p.buf.n != 0 {
		p.buf.n = 0
	} else {
		// No token in the buffer so, read the next token from the scanner.
		p.buf.t, p.buf.l = p.s.Scan()
	}
	return p.buf.t, p.buf.l
}

// scanDistinct scans the next runes as column to use to group.
func (p *Parser) scanDistinct(field *Field) error {
	tk, literal := p.scanIgnoreWhitespace()
	if tk != IDENTIFIER {
		return fmt.Errorf(ErrMsgBadField, literal)
	}
	field.Distinct = true
	field.ColumnName = literal

	return nil
}

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tk Token, literal string) {
	tk, literal = p.scan()
	if tk == WHITE_SPACE {
		return p.scan()
	}
	return
}

// scanList consumes all runes between left and right square brackets.
// Use comma as separator to return a list of string or literal value.
func (p *Parser) scanValueList() (tk Token, list []string) {
	// A list must begin with a left square brackets.
	if ctk, _ := p.scanIgnoreWhitespace(); ctk != LEFT_SQUARE_BRACKETS {
		return
	}
	// Get all values of the list and names the loop on it: L
L:
	for {
		ctk, literal := p.scanIgnoreWhitespace()
		switch ctk {
		case EOF:
			tk = ILLEGAL
			break L
		case RIGHT_SQUARE_BRACKETS:
			// End of the list.
			break L
		case VALUE_LITERAL, IDENTIFIER, DECIMAL, DIGIT:
			// A list can only be string list or a value literal list but not the both.
			if tk == STRING_LIST {
				tk = ILLEGAL
				break L
			}
			// Consume as value literal.
			tk = VALUE_LITERAL_LIST
		case STRING:
			// A list can only be string list or a value literal list but not the both.
			if tk == VALUE_LITERAL_LIST {
				tk = ILLEGAL
				break L
			}
			tk = STRING_LIST
		case COMMA:
			continue L
		default:
			tk = ILLEGAL
			break L
		}
		list = append(list, literal)
	}
	return
}

// scanQueryEnding scans the next runes as query ending.
// Return true if vertical output is required or error if it is not the end of the query.
func (p *Parser) scanQueryEnding() (bool, error) {
	tk, literal := p.scanIgnoreWhitespace()
	switch tk {
	case G_MODIFIER:
		return true, nil
	case SEMICOLON, EOF:
		return false, nil
	default:
		p.unscan()
	}
	return false, fmt.Errorf(ErrMsgSyntax, literal)
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() {
	p.buf.n = 1
}
