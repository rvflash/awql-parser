package awql

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
	"time"
)

// eof represents a marker rune for the end of the reader.
var eof = rune(0)

// Scanner represents a lexical scanner.
type Scanner struct {
	r *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// Scan returns the next token and literal value.
func (s *Scanner) Scan() (Token, string) {
	// Get the next rune.
	r := s.read()
	if isWhitespace(r) {
		// Consume all contiguous whitespace.
		s.unread()
		return s.scanWhitespace()
	} else if isQuote(r) {
		// Consume as string.
		s.unread()
		return s.scanQuotedString()
	} else if r == '[' {
		// Consume as list of string or value literal.
		s.unread()
		return s.scanList()
	} else if isLetter(r) {
		// A keyword begins by a letter.
		// Consume as an identifier or reserved word.
		s.unread()
		return s.scanIdentifier()
	} else if isDigit(r) {
		// Consume as a number.
		s.unread()
		return s.scanNumber()
	}

	// Otherwise read the individual character.
	switch r {
	case eof:
		return EOF, ""
	case '*':
		return ASTERISK, string(r)
	case ',':
		return COMMA, string(r)
	case '(':
		return LEFT_PARENTHESIS, string(r)
	case ')':
		return RIGHT_PARENTHESIS, string(r)
	case '=':
		return EQUAL, string(r)
	case '!':
		// Deal with !=
		if r := s.read(); r == '=' {
			return DIFFERENT, "!="
		}
		s.unread()
	case '>':
		// Deal with >=
		if r := s.read(); r == '=' {
			return SUPERIOR_OR_EQUAL, ">="
		}
		s.unread()
		return SUPERIOR, string(r)
	case '<':
		// Deal with <=
		if r := s.read(); r == '=' {
			return INFERIOR_OR_EQUAL, "<="
		}
		s.unread()
		return INFERIOR, string(r)
	}
	return ILLEGAL, string(r)
}

// scanIdentifier consumes the current rune and all contiguous literal runes.
func (s *Scanner) scanIdentifier() (Token, string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent character of this token into the buffer.
	// Non-literal characters or EOF will cause the loop to exit.
	for {
		if r := s.read(); r == eof {
			break
		} else if !isLiteral(r) {
			s.unread()
			break
		} else {
			buf.WriteRune(r)
		}
	}

	// If the string matches a reserved keyword then return it.
	switch strings.ToUpper(buf.String()) {
	case "DESCRIBE":
		return DESCRIBE, buf.String()
	case "SELECT":
		return SELECT, buf.String()
	case "CREATE":
		return CREATE, buf.String()
	case "REPLACE":
		return REPLACE, buf.String()
	case "VIEW":
		return VIEW, buf.String()
	case "SHOW":
		return SHOW, buf.String()
	case "FULL":
		return FULL, buf.String()
	case "TABLES":
		return TABLES, buf.String()
	case "DISTINCT":
		return DISTINCT, buf.String()
	case "AS":
		return AS, buf.String()
	case "FROM":
		return FROM, buf.String()
	case "WHERE":
		return WHERE, buf.String()
	case "LIKE":
		return LIKE, buf.String()
	case "WITH":
		return WITH, buf.String()
	case "AND":
		return AND, buf.String()
	case "OR":
		return OR, buf.String()
	case "IN":
		return IN, buf.String()
	case "NOT_IN":
		return NOT_IN, buf.String()
	case "STARTS_WITH":
		return STARTS_WITH, buf.String()
	case "STARTS_WITH_IGNORE_CASE":
		return STARTS_WITH_IGNORE_CASE, buf.String()
	case "CONTAINS":
		return CONTAINS, buf.String()
	case "CONTAINS_IGNORE_CASE":
		return CONTAINS_IGNORE_CASE, buf.String()
	case "DOES_NOT_CONTAIN":
		return DOES_NOT_CONTAIN, buf.String()
	case "DOES_NOT_CONTAIN_IGNORE_CASE":
		return DOES_NOT_CONTAIN_IGNORE_CASE, buf.String()
	case "DURING":
		return DURING, buf.String()
	case "GROUP":
		return GROUP, buf.String()
	case "ORDER":
		return ORDER, buf.String()
	case "BY":
		return BY, buf.String()
	case "ASC":
		return ASC, buf.String()
	case "DESC":
		return DESC, buf.String()
	case "LIMIT":
		return LIMIT, buf.String()
	}
	return IDENTIFIER, buf.String()
}

// scanList consumes all runes between left and right square brackets.
// Use comma as separator to return a list of string or literal value.
func (s *Scanner) scanList() (tk Token, list []string) {
	// A list must begin with a left square brackets.
	if r := s.read(); r != '[' {
		return
	}
	for {
		if r := s.read(); r == eof {
			tk = ILLEGAL
			break
		} else if isQuote(r) {
			s.unread()
			// A list can only be string list or a value literal list but not the both.
			if tk == VALUE_LITERAL_LIST {
				tk = ILLEGAL
				break
			}
			// Consume as string.
			tk = STRING_LIST
			list = append(list, s.scanQuotedString())
		} else if isLiteral(r) {
			s.unread()
			// A list can only be string list or a value literal list but not the both.
			if tk == STRING_LIST {
				tk = ILLEGAL
				break
			}
			// Consume as value literal.
			tk = VALUE_LITERAL_LIST
			list = append(list, s.scanValueLiteral())
		} else if r == ']' {
			// End of the list.
			break
		} else if r != ',' && !isWhitespace(r) {
			// If the rune is not a comma or whitespace then break the loop.
			s.unread()
			tk = ILLEGAL
			break
		}
	}

	return
}

// scanNumber consumes all digit or dot runes.
func (s *Scanner) scanNumber() (Token, string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	for {
		if r := s.read(); r == eof {
			break
		} else if !isDigit(r) && r != '.' {
			s.unread()
			break
		} else {
			buf.WriteRune(r)
		}
	}
	// Check if it is a valid number.
	s := buf.String()
	if _, err := strconv.Atoi(s); err == nil {
		return DIGIT, s
	} else if _, err := strconv.ParseFloat(s, 64); err == nil {
		return DECIMAL, s
	}
	return
}

// scanQuotedString consumes the current rune and all runes after it
// until the next unprotected quote character.
func (s *Scanner) scanQuotedString() (Token, string) {
	// Create a buffer and add the single or double quote into it.
	if quote := s.read(); quote == '\'' || quote == '"' {
		var buf bytes.Buffer
		for {
			r := s.read()
			if r == eof {
				return
			}
			buf.WriteRune(r)

			if r == '\\' {
				// Only the character immediately after the escape can itself be a backslash or quote.
				// Thus, we only need to protect the first character after the backslash.
				buf.WriteRune(s.read())
			} else if r == quote {
				break
			}
		}
		return STRING, buf.String()
	}
	return
}

// scanValueLiteral consumes all value literal runes.
func (s *Scanner) scanValueLiteral() (Token, string) {
	var buf bytes.Buffer
	for {
		if r := s.read(); isValueLiteral(r) {
			buf.WriteRune(r)
		} else {
			break
		}
	}
	return VALUE_LITERAL, buf.String()
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (Token, string) {
	var buf bytes.Buffer
	for {
		if r := s.read(); r == eof {
			break
		} else if !isWhitespace(r) {
			s.unread()
			break
		} else {
			buf.WriteRune(r)
		}
	}
	return WHITE_SPACE, buf.String()
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
}

// isDate return true if the string is a date as expected by Adwords.
func isDate(s string) bool {
	if _, err := time.Parse("20060102", s); err == nil {
		return true
	}
	return false
}

// isDateRange return true if the string is a date range literal.
func isDateRangeLiteral(s string) bool {
	switch s {
	case "TODAY", "YESTERDAY",
		"THIS_WEEK_SUN_TODAY", "THIS_WEEK_MON_TODAY",
		"LAST_WEEK", "LAST_7_DAYS", "LAST_14_DAYS",
		"LAST_30_DAYS", "LAST_BUSINESS_WEEK",
		"LAST_WEEK_SUN_SAT", "THIS_MONTH":
		return true
	}
	return false
}

// isDigit returns true if the rune is a digit.
func isDigit(r rune) bool {
	return (r >= '0' && r <= '9')
}

// isFunction returns true if it is an aggregate function.
func isFunction(s string) bool {
	switch strings.ToUpper(s) {
	case "AVG", "COUNT", "MAX", "MIN", "SUM":
		return true
	}
	return false
}

// isLetter returns true if the rune is a letter.
func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// isLiteral returns true if the rune is a literal [a-zA-Z0-9_]
func isLiteral(r rune) bool {
	return r == '_' || isDigit(r) || isLetter(r)
}

// isOperator returns true if the token is an operator
func isOperator(tk Token) bool {
	switch tk {
	case EQUAL, DIFFERENT, SUPERIOR, SUPERIOR_OR_EQUAL,
		INFERIOR, INFERIOR_OR_EQUAL, IN, NOT_IN,
		STARTS_WITH, STARTS_WITH_IGNORE_CASE,
		CONTAINS, CONTAINS_IGNORE_CASE,
		DOES_NOT_CONTAIN, DOES_NOT_CONTAIN_IGNORE_CASE:
		return true
	}
	return false
}

// isQuote returns if the rune is a single quote or double quote.
func isQuote(r rune) bool {
	return r == '"' || r == '\''
}

// isValueLiteral returns true if the rune is a value literal [a-zA-Z0-9_.]
func isValueLiteral(r rune) bool {
	return r == '.' || isLiteral(r)
}

// isWhitespace returns true if the rune is a space, tab, or newline.
func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}
