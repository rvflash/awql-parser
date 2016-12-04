package awql

// Column represents a column.
type Column struct {
	ColumnName, ColumnAlias string
}

// ColumnPosition represents a column with it position in the query.
type ColumnPosition struct {
	Column
	Position uint
}

// Field represents a field.
type Field struct {
	Column
	Method   string
	Distinct bool
}

// Condition represents a where clause.
type Condition struct {
	Column
	Operator, Value string
}

// Pattern represents a LIKE clause.
type Pattern struct {
	Equal, Prefix, Contains, Suffix string
}

// Ordering represents an order by clause.
type Ordering struct {
	ColumnPosition
	SortDesc bool
}

// Limit represents a limit clause.
type Limit struct {
	Offset, RowCount uint
	WithRowCount     bool
}

// OutputFormat enables to format the query output.
type OutputFormat struct {
	VerticalOutput bool
}

// OutputStmt formats the query output.
type OutputFmt interface {
	VerticalOutput() bool
}

// Statement represents a AWQL base statement.
type Statement struct {
	Fields    []Field
	TableName string
	OutputFormat
}

// Stmt represents a AWQL base statement.
// By design, only the SELECT statement is supported by Adwords.
// The AWQL command line tool extends it with others SQL grammar.
type Stmt interface {
	Columns() []Field
	SourceName() string
	OutputFmt
}

// Columns returns the list of table fields.
// It implements the Stmt interface.
func (s Statement) Columns() []Field {
	return s.Fields
}

// SourceName returns the table's name.
// It implements the Stmt interface.
func (s Statement) SourceName() string {
	return s.TableName
}

// SelectStatement represents a AWQL SELECT statement.
// SELECT...FROM...WHERE...DURING...GROUP BY...ORDER BY...LIMIT...
type SelectStatement struct {
	Statement
	Where   []Condition
	During  []string
	GroupBy []ColumnPosition
	OrderBy []Ordering
	Limit
}

/*
SelectStmt exposes the interface of AWQL Select Statement

This is a extended version of the original grammar in order to manage all
the possibilities of the AWQL command line tool.

SelectClause     : SELECT ColumnList
FromClause       : FROM SourceName
WhereClause      : WHERE ConditionList
DuringClause     : DURING DateRange
GroupByClause    : GROUP BY Grouping (, Grouping)*
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
ValueLiteralList : [ ValueLiteral (, ValueLiteral)* ]
Literal          : [a-zA-Z0-9_]*
DateRangeLiteral : TODAY | YESTERDAY | LAST_7_DAYS | THIS_WEEK_SUN_TODAY | THIS_WEEK_MON_TODAY | LAST_WEEK |
									 LAST_14_DAYS | LAST_30_DAYS | LAST_BUSINESS_WEEK | LAST_WEEK_SUN_SAT | THIS_MONTH
Date             : 8-digit integer: YYYYMMDD
*/
type SelectStmt interface {
	Stmt
	ConditionList() []Condition
	DuringList() []string
	GroupList() []ColumnPosition
	OrderList() []Ordering
	StartIndex() uint
	PageSize() (uint, bool)
}

// ConditionList returns the condition list.
// It implements the SelectStmt interface.
func (s SelectStatement) ConditionList() []Condition {
	return s.Where
}

// DuringList returns the during (date range).
// It implements the SelectStmt interface.
func (s SelectStatement) DuringList() []string {
	return s.During
}

// GroupList returns the group by columns.
// It implements the SelectStmt interface.
func (s SelectStatement) GroupList() []ColumnPosition {
	return s.GroupBy
}

// OrderList returns the order by columns.
// It implements the SelectStmt interface.
func (s SelectStatement) OrderList() []Ordering {
	return s.OrderBy
}

// StartIndex returns the start index.
// It implements the SelectStmt interface.
func (s SelectStatement) StartIndex() uint {
	return s.Offset
}

// PageSize returns the row count.
// It implements the SelectStmt interface.
func (s SelectStatement) PageSize() (uint, bool) {
	return s.RowCount, s.WithRowCount
}

// CreateViewStatement represents a AWQL CREATE VIEW statement.
// CREATE...OR REPLACE...VIEW...AS
type CreateViewStatement struct {
	Statement
	Replace bool
	View    SelectStatement
}

/*
CreateViewStmt exposes the interface of AWQL Create View Statement

Not supported natively by Adwords API. Used by the following AWQL command line tool:
https://github.com/rvflash/awql/

CreateClause     : CREATE (OR REPLACE)* VIEW DestinationName (**(**ColumnList**)**)*
FromClause       : AS SelectClause
*/
type CreateViewStmt interface {
	Stmt
	ReplaceMode() bool
	SourceQuery() Stmt
}

// ReplaceMode returns true if it is required to replace the existing view.
// It implements the CreateViewStmt interface.
func (s CreateViewStatement) ReplaceMode() bool {
	return s.Replace
}

// SourceQuery returns the source query, base of the view to create.
// It implements the CreateViewStmt interface.
func (s CreateViewStatement) SourceQuery() Stmt {
	return s.View
}

// FullStatement enables a AWQL FULL mode.
type FullStatement struct {
	Full bool
}

// DescribeStatement represents a AWQL DESC statement.
// DESC...FULL
type DescribeStatement struct {
	FullStatement
	Statement
}

// FullStmt proposes the full statement mode.
type FullStmt interface {
	FullMode() bool
}

/*
DescribeStmt exposes the interface of AWQL Describe Statement

Not supported natively by Adwords API. Used by the following AWQL command line tool:
https://github.com/rvflash/awql/

DescribeClause   : (DESCRIBE | DESC) (FULL)* SourceName (ColumnName)*
*/
type DescribeStmt interface {
	Stmt
	FullStmt
}

// FullMode returns true if the full display is required.
// It implements the DescribeStmt interface.
func (s DescribeStatement) FullMode() bool {
	return s.Full
}

// ShowStatement represents a AWQL SHOW statement.
// SHOW...FULL...TABLES...LIKE...WITH
type ShowStatement struct {
	FullStatement
	Like Pattern
	With string
	OutputFormat
}

/*
ShowStmt exposes the interface of AWQL Show Statement

Not supported natively by Adwords API. Used by the following AWQL command line tool:
https://github.com/rvflash/awql/

ShowClause   : SHOW (FULL)* TABLES
WithClause   : WITH ColumnName
LikeClause   : LIKE String
*/
type ShowStmt interface {
	FullStmt
	LikePattern() Pattern
	WithColumnName() string
	OutputFmt
}

// LikePattern returns the pattern used for a like query on the table list.
// It implements the ShowStmt interface.
func (s ShowStatement) LikePattern() string {
	return s.Like
}

// WithColumnName returns the column name used to search table with this column.
// It implements the ShowStmt interface.
func (s ShowStatement) WithColumnName() string {
	return s.With
}
