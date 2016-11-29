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
	Method   uint
	Distinct bool
}

// Condition represents a where clause.
type Condition struct {
	Column
	Operator, Value string
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

// Statement represents a AWQL base statement.
type Statement struct {
	Fields    []Field
	TableName string
}

// Stmt represents a AWQL base statement.
// By design, only the SELECT statement is supported by Adwords.
// The AWQL command line tool extends it with others SQL grammar.
type Stmt interface {
	Columns() []Field
	SourceName() string
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
// SELECT ColumnList
// FROM TableName
// WHERE ConditionList
// DURING DateRange
// GROUP BY ColumnName (, ColumnName)*
// ORDER BY Ordering (, Ordering)*
// LIMIT StartIndex , PageSize
type SelectStatement struct {
	Statement
	Where   []Condition
	During  []string
	GroupBy []ColumnPosition
	OrderBy []Ordering
	Limit
}

// SelectStmt represents a AWQL select statement.
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
// CREATE [OR REPLACE] VIEW view_name [(column_list)] AS select_statement
type CreateViewStatement struct {
	Statement
	Replace bool
	View    SelectStatement
}

// CreateViewStmt represents a AWQL create view statement.
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

// DescribeStatement represents a AWQL DESC statement.
// DESC [FULL] TableName [ColumnName]
type DescribeStatement struct {
	Statement
	Full bool
}

// DescribeStmt represents a AWQL desc statement.
type DescribeStmt interface {
	Stmt
	FullMode() bool
}

// FullMode returns true if the full display is required.
// It implements the DescribeStmt interface.
func (s DescribeStatement) FullMode() bool {
	return s.Full
}

// ShowStatement represents a AWQL SHOW statement.
// SHOW [FULL] TABLES [LIKE 'Pattern']
// SHOW TABLES [WITH 'Pattern']
type ShowStatement struct {
	DescribeStatement
	Like, With string
}

// ShowStmt represents a AWQL show statement.
type ShowStmt interface {
	DescribeStmt
	LikePattern() string
	WithColumnName() string
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
