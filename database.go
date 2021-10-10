package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-mysql-org/go-mysql/replication"
	_ "github.com/go-sql-driver/mysql"
)

// BinlogExecutor reads in a MySQL binlog and executes INSERT/UPDATES on
// a target slave database
type BinlogExecutor struct {
	conn *sql.DB

	currentSchema string
	currentTable  string
}

// NewBinlogExecutor creates a new BinlogExecutor with a connection
// to the target slave
func NewBinlogExecutor(conn *sql.DB) *BinlogExecutor {
	return &BinlogExecutor{
		conn: conn,
	}
}

// ExecuteEvent executes a binlog event
func (e *BinlogExecutor) ExecuteEvent(event *replication.BinlogEvent) error {
	switch event.Header.EventType {
	case replication.TABLE_MAP_EVENT:
		tme := event.Event.(*replication.TableMapEvent)
		err := e.changeSchema(tme)
		if err != nil {
			return err
		}
		return e.mapTable(tme)
	case replication.WRITE_ROWS_EVENTv0:
		fallthrough
	case replication.WRITE_ROWS_EVENTv1:
		fallthrough
	case replication.WRITE_ROWS_EVENTv2:
		return e.executeInsert(event.Event.(*replication.RowsEvent))
	case replication.UPDATE_ROWS_EVENTv0:
		fallthrough
	case replication.UPDATE_ROWS_EVENTv1:
		fallthrough
	case replication.UPDATE_ROWS_EVENTv2:
		return e.executeUpdate(event.Event.(*replication.RowsEvent))
	case replication.DELETE_ROWS_EVENTv0:
		fallthrough
	case replication.DELETE_ROWS_EVENTv1:
		fallthrough
	case replication.DELETE_ROWS_EVENTv2:
		return e.executeDelete(event.Event.(*replication.RowsEvent))
	}
	return nil
}

func (e *BinlogExecutor) getTableColumns(
	schemaName,
	tableName string,
) ([]Column, error) {
	query := "SELECT `TABLE_SCHEMA`, `TABLE_NAME`, `COLUMN_TYPE`, `COLUMN_NAME`, " +
		"`ORDINAL_POSITION`, `COLUMN_KEY`, `COLUMN_DEFAULT`, `IS_NULLABLE`, `EXTRA` " +
		"FROM `information_schema`.`columns` " +
		"WHERE `table_name` = ? AND `table_schema` = ?"

	rows, err := e.conn.Query(query, tableName, schemaName)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var columnInfo []Column
	for rows.Next() {
		inter := struct {
			tableSchema  string
			tableName    string
			ttype        string
			name         string
			position     int
			keys         string
			nullable     string
			defaultValue interface{}
			extra        string
		}{}

		err := rows.Scan(
			&inter.tableSchema,
			&inter.tableName,
			&inter.ttype,
			&inter.name,
			&inter.position,
			&inter.keys,
			&inter.defaultValue,
			&inter.nullable,
			&inter.extra,
		)
		if err != nil {
			return nil, err
		}

		var defaultValue string
		if inter.defaultValue == nil {
			defaultValue = "NULL"
		} else {
			defaultValue = string(inter.defaultValue.([]uint8))
			if defaultValue[0] == '\'' || defaultValue[0] == '"' {
				defaultValue, _ = strconv.Unquote(defaultValue)
			}
		}

		col := Column{
			TableSchema: inter.tableSchema,
			TableName:   inter.tableName,
			Type:        inter.ttype,
			Name:        inter.name,
			Position:    inter.position,
			IsAutoIncrement: strings.Contains(
				strings.ToUpper(inter.extra),
				"AUTO_INCREMENT",
			),
			IsPrimary:    strings.Contains(inter.keys, "PRI"),
			IsNullable:   strings.Contains(inter.nullable, "YES"),
			DefaultValue: defaultValue,
		}
		columnInfo = append(columnInfo, col)
	}

	us := NewUseStatement(e.currentSchema)

	_, err = e.conn.Exec(us.String())
	if err != nil {
		return nil, err
	}

	return columnInfo, nil
}

func (e *BinlogExecutor) changeSchema(
	event *replication.TableMapEvent,
) error {
	schema, table := string(event.Schema), string(event.Table)
	if schema == e.currentSchema && table == e.currentTable {
		return nil
	}
	e.currentSchema = schema
	e.currentTable = table
	us := NewUseStatement(schema)
	_, err := e.conn.Exec(us.String())
	return err
}

func (e *BinlogExecutor) doesRowExist(
	conditions map[string]string,
) (bool, error) {
	ss := NewSelectStatement(e.currentSchema, e.currentTable, conditions)

	rows, err := e.conn.Query(ss.String())
	if err != nil {
		return false, err
	}

	defer rows.Close()

	us := NewUseStatement(e.currentSchema)
	_, err = e.conn.Exec(us.String())
	if err != nil {
		return false, err
	}

	return rows.Next(), nil
}

func (e *BinlogExecutor) mapTable(event *replication.TableMapEvent) error {
	table, schema := string(event.Table), string(event.Schema)

	current, err := e.getTableColumns(schema, table)
	if err != nil {
		return err
	}

	cols := make([]Part, 0)

	for i := uint64(0); i < event.ColumnCount; i++ {
		t, meta, err := parseDataType(current[i].Type)
		if err != nil {
			return fmt.Errorf("error mapping table: %v", err)
		}

		cd := NewColumnDef(current[i].Name, event.ColumnType[i])

		cd.IsNullable = getColumnIsNullFromBitmask(uint(i), event.NullBitmap)

		// carry across values that cannot be got from a map event
		cd.IsAutoIncrement = current[i].IsAutoIncrement
		cd.DefaultValue = current[i].DefaultValue

		var isDifferent bool
		if current[i].IsNullable != cd.IsNullable {
			isDifferent = true
		}

		tn, _ := getSQLTypeName(t)
		cdtn, _ := getSQLTypeName(cd.DataType)
		if tn != cdtn {
			// if the type changes, use new meta
			cd.Meta = event.ColumnMeta[i]
			isDifferent = true
		} else {
			// if the type is the same, check size
			if event.ColumnMeta[i] != 0 {
				// if meta isn't 0, change it
				cd.Meta = event.ColumnMeta[i]
			} else {
				// else keep it
				cd.Meta = meta
			}
		}

		if meta != cd.Meta {
			isDifferent = true
		}

		if isDifferent {
			cols = append(cols, ModifyColumnPart{ColumnDef: cd})
		}
	}

	if len(cols) > 0 {
		return e.alterTable(schema, table, cols)
	}

	return nil
}

func (e *BinlogExecutor) alterTable(schemaName, tableName string, parts []Part) error {
	as := NewAlterTableStatement(schemaName, tableName, parts...)

	_, err := e.conn.Exec(as.String())

	return err
}

func (e *BinlogExecutor) executeInsert(event *replication.RowsEvent) error {
	e.changeSchema(event.Table)

	cols, err := e.getTableColumns(e.currentSchema, e.currentTable)
	if err != nil {
		return fmt.Errorf("error getting column info: %v", err)
	}

	for _, row := range event.Rows {
		err := e.executeInsertSingle(stringifyValues(row, cols))
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *BinlogExecutor) executeInsertSingle(vals []string) error {
	is := NewInsertStatement(e.currentSchema, e.currentTable, vals)

	result, err := e.conn.Exec(is.String())
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil
		}
		return fmt.Errorf("error inserting: %v", err)
	}

	if r, _ := result.RowsAffected(); r <= 0 {
		return fmt.Errorf("failed to insert record: %v", is.String())
	}

	return nil
}

func (e *BinlogExecutor) executeUpdate(event *replication.RowsEvent) error {
	e.changeSchema(event.Table)

	cols, err := e.getTableColumns(e.currentSchema, e.currentTable)
	if err != nil {
		return fmt.Errorf("error getting column info: %v", err)
	}

	for _, row := range event.Rows {
		err := e.executeUpdateSingle(stringifyValues(row, cols))
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *BinlogExecutor) executeUpdateSingle(vals []string) error {
	cols, err := e.getTableColumns(e.currentSchema, e.currentTable)
	if err != nil {
		return fmt.Errorf("error getting column info: %v", err)
	}
	condMap := make(map[string]string)
	valMap := make(map[string]string)
	for _, col := range cols {
		if col.IsPrimary {
			condMap[col.Name] = vals[col.Position-1]
		}
		valMap[col.Name] = vals[col.Position-1]
	}
	if ok, _ := e.doesRowExist(condMap); !ok {
		return e.executeInsertSingle(vals)
	}
	us := NewUpdateStatement(
		e.currentSchema,
		e.currentTable,
		valMap,
		condMap,
	)
	_, err = e.conn.Exec(us.String())
	if err != nil {
		return fmt.Errorf("error updating: %v", err)
	}
	return nil
}

func (e *BinlogExecutor) executeDelete(event *replication.RowsEvent) error {
	e.changeSchema(event.Table)

	cols, err := e.getTableColumns(e.currentSchema, e.currentTable)
	if err != nil {
		return fmt.Errorf("error getting column info: %v", err)
	}

	for _, row := range event.Rows {
		err := e.executeDeleteSingle(stringifyValues(row, cols))
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *BinlogExecutor) executeDeleteSingle(vals []string) error {
	cols, err := e.getTableColumns(e.currentSchema, e.currentTable)
	if err != nil {
		return fmt.Errorf("error getting table info: %v", err)
	}
	condMap := make(map[string]string)
	for _, col := range cols {
		if col.IsPrimary {
			condMap[col.Name] = vals[col.Position-1]
		}
	}
	ds := NewDeleteStatement(
		e.currentSchema,
		e.currentTable,
		condMap,
	)
	_, err = e.conn.Exec(ds.String())
	if err != nil {
		return fmt.Errorf("error deleting: %v", err)
	}
	return nil
}
