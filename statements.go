package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	useTemplate    string = "USE `%s`;"
	insertTemplate string = "INSERT INTO `%s`.`%s` VALUES (%s);"
	updateTemplate string = "UPDATE `%s`.`%s` SET %s WHERE %s;"
	deleteTemplate string = "DELETE FROM `%s`.`%s` WHERE %s;"
	selectTemplate string = "SELECT * FROM `%s`.`%s` WHERE %s;"

	createTableTemplate string = "CREATE TABLE `%s`.`%s` (%s);"
	alterTableTemplate  string = "ALTER TABLE `%s`.`%s` %s;"

	modifyColumnTemplate string = "MODIFY COLUMN `%s` %s"
	newColumnTemplate    string = "ADD COLUMN `%s` %s"
	dropColumnTemplate   string = "DROP COLUMN `%s`"
)

const (
	mySQLDecimal    uint8 = 0x00
	mySQLTiny       uint8 = 0x01
	mySQLShort      uint8 = 0x02
	mySQLLong       uint8 = 0x03
	mySQLFloat      uint8 = 0x04
	mySQLDouble     uint8 = 0x05
	mySQLNull       uint8 = 0x06
	mySQLTimestamp  uint8 = 0x07
	mySQLLongLong   uint8 = 0x08
	mySQLInt24      uint8 = 0x09
	mySQLDate       uint8 = 0x0a
	mySQLTime       uint8 = 0x0b
	mySQLDateTime   uint8 = 0x0c
	mySQLYear       uint8 = 0x0d
	mySQLNewDate    uint8 = 0x0e
	mySQLVarChar    uint8 = 0x0f
	mySQLBit        uint8 = 0x10
	mySQLTimestamp2 uint8 = 0x11
	mySQLDateTime2  uint8 = 0x12
	mySQLTime2      uint8 = 0x13
	mySQLNewDecimal uint8 = 0xf6
	mySQLEnum       uint8 = 0xf7
	mySQLSet        uint8 = 0xf8
	mySQLTinyBlob   uint8 = 0xf9
	mySQLMediumBlob uint8 = 0xfa
	mySQLLongBlob   uint8 = 0xfb
	mySQLBlob       uint8 = 0xfc
	mySQLVarString  uint8 = 0xfd
	mySQLString     uint8 = 0xfe
	mySQLGeometry   uint8 = 0xff
)

// Column details a column as seen from querying information_schema.columns
type Column struct {
	TableSchema     string
	TableName       string
	Name            string
	Type            string
	Position        int
	IsAutoIncrement bool
	IsPrimary       bool
	IsNullable      bool
	DefaultValue    string
}

// Statement represents an SQL statement
type Statement interface {
	fmt.Stringer
}

// UseStatement represents an SQL USE statement
type UseStatement struct {
	Database string
}

// NewUseStatement creates a new USE statement with a given database
func NewUseStatement(database string) UseStatement {
	return UseStatement{
		Database: database,
	}
}

// String returns a formatted SQL USE statement
func (us UseStatement) String() string {
	return fmt.Sprintf(useTemplate, escapeSQLChars(us.Database))
}

// InsertStatement represents an SQL INSERT statement
type InsertStatement struct {
	SchemaName string
	TableName  string
	Values     []string
}

// NewInsertStatement creates a new INSERT statement with a schema, table,
// and values to be inserted
func NewInsertStatement(
	schemaName,
	tableName string,
	vals []string,
) InsertStatement {
	return InsertStatement{
		SchemaName: schemaName,
		TableName:  tableName,
		Values:     vals,
	}
}

// String returns a formatted SQL INSERT statement
func (is InsertStatement) String() string {
	var vals []string
	for _, v := range is.Values {
		if v == "nil" {
			v = "NULL"
		} else {
			v = "'" + escapeSQLChars(v) + "'"
		}
		vals = append(vals, v)
	}
	return fmt.Sprintf(
		insertTemplate,
		escapeSQLChars(is.SchemaName),
		escapeSQLChars(is.TableName),
		strings.Join(vals, ", "),
	)
}

// UpdateStatement requests an SQL UPDATE statement
type UpdateStatement struct {
	SchemaName string
	TableName  string
	Values     map[string]string
	Conditions map[string]string
}

// NewUpdateStatement creates a UPDATE statement with schema, table,
// values to be updated, and conditions to limit to. These conditions
// are ANDs only
func NewUpdateStatement(
	schemaName,
	tableName string,
	vals,
	conditions map[string]string,
) UpdateStatement {
	return UpdateStatement{
		SchemaName: schemaName,
		TableName:  tableName,
		Values:     vals,
		Conditions: conditions,
	}
}

// String returns a formatted SQL UPDATE statement
func (us UpdateStatement) String() string {
	return fmt.Sprintf(
		updateTemplate,
		escapeSQLChars(us.SchemaName),
		escapeSQLChars(us.TableName),
		mapSQLFields(us.Values),
		mapSQLConditions(us.Conditions),
	)
}

// DeleteStatement represents an SQL DELETE statement
type DeleteStatement struct {
	SchemaName string
	TableName  string
	Conditions map[string]string
}

// NewDeleteStatement creates a new DELETE statement with schema, table,
// and WHERE conditions. These conditions are ANDs only
func NewDeleteStatement(
	schemaName,
	tableName string,
	conditions map[string]string,
) DeleteStatement {
	return DeleteStatement{
		SchemaName: schemaName,
		TableName:  tableName,
		Conditions: conditions,
	}
}

// String returns a formatted SQL DELETE statement
func (ds DeleteStatement) String() string {
	return fmt.Sprintf(
		deleteTemplate,
		escapeSQLChars(ds.SchemaName),
		escapeSQLChars(ds.TableName),
		mapSQLConditions(ds.Conditions),
	)
}

// SelectStatement represents an SQL SELECT statement
type SelectStatement struct {
	SchemaName string
	TableName  string
	Conditions map[string]string
}

// NewSelectStatement recreates a new SELECT statement with schema, table,
// and WHERE conditions. These conditions are ANDs only
func NewSelectStatement(
	schemaName,
	tableName string,
	conditions map[string]string,
) SelectStatement {
	return SelectStatement{
		SchemaName: schemaName,
		TableName:  tableName,
		Conditions: conditions,
	}
}

// String returns a formatted SQL SELECT statement
func (s SelectStatement) String() string {
	return fmt.Sprintf(
		selectTemplate,
		escapeSQLChars(s.SchemaName),
		escapeSQLChars(s.TableName),
		mapSQLConditions(s.Conditions),
	)
}

// ColumnDef represents a column definition found in a
// CREATE/ALTER table statement
type ColumnDef struct {
	Name            string
	DataType        uint8
	IsNullable      bool
	IsAutoIncrement bool
	DefaultValue    string
	Meta            uint16

	after   *ColumnDef
	isFirst bool
}

// NewColumnDef creates a new ColumnDef
func NewColumnDef(name string, datatype uint8) *ColumnDef {
	return &ColumnDef{
		Name:     name,
		DataType: datatype,
	}
}

// Definition returns a formatted definition for use in CREATE/ALTER statements
func (c *ColumnDef) Definition() string {
	var b strings.Builder

	typename, err := getSQLTypeName(c.DataType)
	if err != nil {
		panic(err)
	}
	b.WriteString(typename)

	if c.Meta > 0 {
		b.WriteString(fmt.Sprintf("(%d)", c.Meta))
	}

	if c.IsNullable {
		b.WriteString(" NULL")
	}

	if c.DefaultValue != "" {
		var invalidDefault bool

		fmtStr := " DEFAULT '%s'"
		if c.DefaultValue == "NULL" {

			if !c.IsNullable {
				invalidDefault = true
			}

			fmtStr = " DEFAULT %s"
		}

		if !invalidDefault {
			b.WriteString(
				fmt.Sprintf(fmtStr, escapeSQLChars(c.DefaultValue)),
			)
		}
	}

	if c.IsAutoIncrement {
		b.WriteString(" AUTO_INCREMENT")
	}

	if c.IsFirst() {
		b.WriteString(" FIRST")
	} else if a := c.After(); a != nil {
		b.WriteString(
			fmt.Sprintf(" AFTER `%s`", a.Name),
		)
	}
	return b.String()
}

// SetAfter sets the column definition to follow after another
func (c *ColumnDef) SetAfter(other *ColumnDef) {
	c.after = other
	c.isFirst = false
}

// SetFirst changes if the column is first depending on flag
func (c *ColumnDef) SetFirst(flag bool) {
	c.isFirst = flag
	c.after = nil
}

// IsFirst returns whether the column definition is first
func (c *ColumnDef) IsFirst() bool {
	return c.isFirst
}

// After returns the column definition that it follows, if any
func (c *ColumnDef) After() *ColumnDef {
	return c.after
}

// Part is a Statement with a different name
type Part interface {
	Statement
}

// ModifyColumnPart is a column definiton for a column that has been changed.
// See MySQL ALTER statement, MODIFY changes the definition of a column
type ModifyColumnPart struct {
	*ColumnDef
}

// String returns a formatted MODIFY part
func (c ModifyColumnPart) String() string {
	return fmt.Sprintf(
		modifyColumnTemplate,
		c.Name,
		c.Definition(),
	)
}

// DropColumnPart is a column to be deleted from a table in an ALTER statement
type DropColumnPart struct {
	ColumnName string
}

// String returns a formatted DROP COLUMN part
func (c DropColumnPart) String() string {
	return fmt.Sprintf(
		dropColumnTemplate,
		c.ColumnName,
	)
}

// NewColumnPart is a column definition being added in a CREATE/ALTER table
// statement
type NewColumnPart struct {
	*ColumnDef
}

// String returns a formatted ADD COLUMN part
func (c NewColumnPart) String() string {
	return fmt.Sprintf(
		newColumnTemplate,
		c.Name,
		c.Definition(),
	)
}

// AlterTableStatement requests an SQL ALTER TABLE statement
type AlterTableStatement struct {
	SchemaName string
	TableName  string
	Parts      []Part
}

// NewAlterTableStatement creates a new ALTER TABLE statement with schema,
// table, and column parts
func NewAlterTableStatement(
	schemaName,
	tableName string,
	parts ...Part,
) AlterTableStatement {
	return AlterTableStatement{
		SchemaName: schemaName,
		TableName:  tableName,
		Parts:      parts,
	}
}

// String returns a formatted SQL ALTER TABLE statement
func (a AlterTableStatement) String() string {
	s := make([]string, len(a.Parts))
	for i, p := range a.Parts {
		s[i] = p.String()
	}
	return fmt.Sprintf(
		alterTableTemplate,
		a.SchemaName,
		a.TableName,
		strings.Join(s, ", "),
	)
}

// CreateTableStatement represents an SQL CREATE TABLE statement
type CreateTableStatement struct {
	SchemaName string
	TableName  string
	Parts      []Part
}

// NewCreateTableStatement creates a new CREATE TABLE statement with schema,
// table, and parts
func NewCreateTableStatement(
	schemaName,
	tableName string,
	parts ...Part,
) CreateTableStatement {
	return CreateTableStatement{
		SchemaName: schemaName,
		TableName:  tableName,
		Parts:      parts,
	}
}

// String returns a formatted SQL CREATE TABLE statement
func (c CreateTableStatement) String() string {
	s := make([]string, len(c.Parts))
	for i, p := range c.Parts {
		s[i] = p.String()
	}
	return fmt.Sprintf(
		createTableTemplate,
		c.SchemaName,
		c.TableName,
		strings.Join(s, ", "),
	)
}

// getColumnIsNullFromBitmask takes an array of bytes and an absolute offset
// and returns if that bit is 1 or 0
func getColumnIsNullFromBitmask(index uint, bitmask []byte) bool {
	byteOffset := byte(index / 8)
	bitOffset := byte(1 << (index % 8))

	return (bitmask[byteOffset] & bitOffset) == bitOffset
}

// getSQLTypeName takes a SQL type enum and returns the SQL name for it
func getSQLTypeName(t uint8) (typename string, err error) {
	switch t {
	case mySQLNewDecimal:
		fallthrough
	case mySQLDecimal:
		typename = "DECIMAL"
	case mySQLTiny:
		typename = "TINYINT"
	case mySQLYear:
		fallthrough
	case mySQLShort:
		typename = "SMALLINT"
	case mySQLInt24:
		fallthrough
	case mySQLLong:
		typename = "INT"
	case mySQLFloat:
		typename = "FLOAT"
	case mySQLDouble:
		typename = "DOUBLE"
	case mySQLNull:
		typename = "NULL"
	case mySQLTimestamp2:
		fallthrough
	case mySQLTimestamp:
		typename = "TIMESTAMP"
	case mySQLLongLong:
		typename = "BIGINT"
	case mySQLNewDate:
		fallthrough
	case mySQLDate:
		typename = "DATE"
	case mySQLTime:
		typename = "TIME"
	case mySQLDateTime2:
		fallthrough
	case mySQLDateTime:
		typename = "DATETIME"
	case mySQLVarChar:
		typename = "VARCHAR"
	case mySQLBit:
		typename = "BIT"
	case mySQLEnum:
		typename = "ENUM"
	case mySQLSet:
		typename = "SET"
	case mySQLTinyBlob:
		typename = "TINYBLOB"
	case mySQLMediumBlob:
		typename = "MEDIUMBLOB"
	case mySQLLongBlob:
		typename = "LONGBLOB"
	case mySQLBlob:
		typename = "BLOB"
	case mySQLVarString:
		fallthrough
	case mySQLString:
		typename = "STRING"
	case mySQLGeometry:
		typename = "GEOMETRY"
	default:
		err = fmt.Errorf("no such SQL type %d", t)
	}
	return typename, err
}

// parseDataType takes a SQL type string i.e. INT(11) and returns the SQL type
// enum (mySQLLong), the meta (11), and an error if type is unknown
func parseDataType(s string) (t uint8, meta uint16, err error) {
	s = strings.ToUpper(s)

	b1 := strings.LastIndex(s, "(")
	b2 := strings.LastIndex(s, ")")

	// if pair brackets found, get meta
	if b1 >= 0 && b2 >= 0 {
		var (
			tmp     string
			tmpMeta int
		)
		tmp = s[b1+1 : b2]

		if strings.Contains(s, "DOUBLE") ||
			strings.Contains(s, "FLOAT") ||
			strings.Contains(s, "DECIMAL") {
			// split on comma
			parts := strings.Split(tmp, ",")

			// for each part, convert to int
			for i, p := range parts {
				var val int
				val, err = strconv.Atoi(p)
				if err != nil {
					return
				}

				// add to tmpMeta
				tmpMeta |= val
				if i < len(parts)-1 {
					// if not last, shift 8 bits
					tmpMeta = tmpMeta << 8
				}
			}
		} else {
			tmpMeta, err = strconv.Atoi(tmp)
		}

		if err != nil {
			return
		}

		meta = uint16(tmpMeta)
	}

	// if opening brace found, truncate
	if b1 >= 0 {
		s = s[:b1]
	}

	switch s {
	case "DECIMAL":
		t = mySQLDecimal
	case "TINYINT":
		t = mySQLTiny
	case "SMALLINT":
		t = mySQLShort
	case "INT":
		t = mySQLLong
	case "FLOAT":
		t = mySQLFloat
	case "DOUBLE":
		t = mySQLDouble
	case "NULL":
		t = mySQLNull
	case "TIMESTAMP":
		t = mySQLTimestamp
	case "BIGINT":
		t = mySQLLongLong
	case "DATE":
		t = mySQLDate
	case "TIME":
		t = mySQLTime
	case "DATETIME":
		t = mySQLDateTime
	case "VARCHAR":
		t = mySQLVarChar
	case "BIT":
		t = mySQLBit
	case "ENUM":
		t = mySQLEnum
	case "SET":
		t = mySQLSet
	case "TINYBLOB":
		t = mySQLTinyBlob
	case "MEDIUMBLOB":
		t = mySQLMediumBlob
	case "LONGBLOB":
		t = mySQLLongBlob
	case "LONGTEXT":
		meta = 4
		t = mySQLBlob
	case "TEXT":
		meta = 2
		t = mySQLBlob
	case "BLOB":
		t = mySQLBlob
	case "CHAR":
		meta = 0xfe01
		t = mySQLString
	case "STRING":
		t = mySQLString
	case "GEOMETRY":
		t = mySQLGeometry
	default:
		err = fmt.Errorf("no such SQL type %s", s)
	}

	return
}

// escapeSQLChars escapes any SQL statement specific characters
func escapeSQLChars(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "`", "\\`")
	return s
}

// mapSQLFields takes a map of columns and values and produces a string
// for use in a UPDATE statement
func mapSQLFields(fieldMap map[string]string) string {
	var b strings.Builder
	for k, v := range fieldMap {
		if v == "nil" {
			b.WriteString(
				fmt.Sprintf(
					"`%s` = %s",
					escapeSQLChars(k),
					"NULL",
				),
			)
		} else {
			b.WriteString(
				fmt.Sprintf(
					"`%s` = '%s'",
					escapeSQLChars(k),
					escapeSQLChars(v),
				),
			)
		}
		b.WriteString(", ")
	}
	str := b.String()
	str = str[:len(str)-2]
	return str
}

// mapSQLConditions takes a map of columns and values and produces a string
// for use in WHERE clauses
func mapSQLConditions(conditions map[string]string) string {
	var b strings.Builder
	for k, v := range conditions {
		if v == "nil" {
			b.WriteString(
				fmt.Sprintf(
					"`%s` = %s",
					escapeSQLChars(k),
					"NULL",
				),
			)
		} else {
			b.WriteString(
				fmt.Sprintf(
					"`%s` = '%s'",
					escapeSQLChars(k),
					escapeSQLChars(v),
				),
			)
		}
		b.WriteString(" AND ")
	}
	str := b.String()
	str = str[:len(str)-5]
	return str
}

// stringifyValues takes a slice of values and corresponding columns and
// converts each value into a string equivalent
func stringifyValues(values []interface{}, columns []Column) []string {
	vals := make([]string, len(values))
	for i, col := range values {
		var val string
		if col == nil {
			if !columns[i].IsNullable {
				val = columns[i].DefaultValue
			} else {
				val = "nil"
			}
		} else {
			switch t := col.(type) {
			case string:
				val = t
			case int32:
				val = fmt.Sprintf("%d", t)
			case int64:
				val = fmt.Sprintf("%d", t)
			case int:
				val = fmt.Sprintf("%d", t)
			case float32:
				val = fmt.Sprintf("%f", t)
			case float64:
				val = fmt.Sprintf("%f", t)
			default:
				val = fmt.Sprintf("%v", t)
			}
		}
		vals[i] = toUTF8([]byte(val))
	}
	return vals
}
