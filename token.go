package awql

/*
AWQL Statement

SelectClause     : SELECT ColumnList
FromClause       : FROM SourceName
WhereClause      : WHERE ConditionList
DuringClause     : DURING DateRange
GroupByClause    : GROUP BY ColumnName(, ColumnName)*
OrderByClause    : ORDER BY Ordering (, Ordering)*
LimitClause      : LIMIT StartIndex , PageSize

ConditionList    : Condition (AND Condition)*
Condition        : ColumnName Operator Value
Value            : ValueLiteral | String | ValueLiteralList | StringList
Ordering         : ColumnName (DESC | ASC)?
DateRange        : DateRangeLiteral | Date,Date
ColumnList       : ColumnName (, ColumnName)*
ColumnName       : Literal
TableName        : Literal
StartIndex       : Non-negative integer
PageSize         : Non-negative integer

Operator         : = | != | > | >= | < | <= | IN | NOT_IN | STARTS_WITH | STARTS_WITH_IGNORE_CASE |
									CONTAINS | CONTAINS_IGNORE_CASE | DOES_NOT_CONTAIN | DOES_NOT_CONTAIN_IGNORE_CASE
String           : StringSingleQ | StringDoubleQ
StringSingleQ    : '(char)'
StringDoubleQ    : "(char)"
StringList       : [ String (, String)* ]
ValueLiteral     : [a-zA-Z0-9_.]*
ValueLiteralList : [ ValueLiteral (, ValueLiteral)* ]3
Literal          : [a-zA-Z0-9_]*
DateRangeLiteral : TODAY | YESTERDAY | LAST_7_DAYS | THIS_WEEK_SUN_TODAY | THIS_WEEK_MON_TODAY | LAST_WEEK |
									 LAST_14_DAYS | LAST_30_DAYS | LAST_BUSINESS_WEEK | LAST_WEEK_SUN_SAT | THIS_MONTH
Date             : 8-digit integer: YYYYMMDD

https://developers.google.com/adwords/api/docs/guides/awql#grammar
*/

// Token represents a lexical token.
type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF

	// Literals
	IDENTIFIER  // base element
	WHITE_SPACE // white space

	// Misc characters
	ASTERISK              // *
	COMMA                 // ,
	SINGLE_QUOTE          // '
	DOUBLE_QUOTE          // "
	LEFT_PARENTHESIS      // (
	RIGHT_PARENTHESIS     // )
	LEFT_SQUARE_BRACKETS  //[
	RIGHT_SQUARE_BRACKETS //[

	// Operator
	EQUAL             // =
	DIFFERENT         // !=
	SUPERIOR          // >
	SUPERIOR_OR_EQUAL // >=
	INFERIOR          // <
	INFERIOR_OR_EQUAL // <=
	IN
	NOT_IN
	STARTS_WITH
	STARTS_WITH_IGNORE_CASE
	CONTAINS
	CONTAINS_IGNORE_CASE
	DOES_NOT_CONTAIN
	DOES_NOT_CONTAIN_IGNORE_CASE

	// Aggregate functions
	AVG
	COUNT
	MAX
	MIN
	SUM

	// During date range
	TODAY
	YESTERDAY
	LAST_7_DAYS
	THIS_WEEK_SUN_TODAY
	THIS_WEEK_MON_TODAY
	LAST_WEEK
	LAST_14_DAYS
	LAST_30_DAYS
	LAST_BUSINESS_WEEK
	LAST_WEEK_SUN_SAT
	THIS_MONTH

	// Base keywords
	SELECT
	DISTINCT
	AS
	FROM
	WHERE
	AND
	DURING
	ORDER
	GROUP
	BY
	ASC
	DESC
	LIMIT
)
