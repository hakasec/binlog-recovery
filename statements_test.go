package main

import "testing"

func TestUseStatement(t *testing.T) {
	us := NewUseStatement("hello")
	if us.String() != "USE `hello`;" {
		t.Fatalf("Use statement incorrectly formatted: %s\n", us.String())
	}
}

func TestInsertStatement(t *testing.T) {
	is := NewInsertStatement("hello", "world", []string{"world"})
	if is.String() != "INSERT INTO `hello`.`world` VALUES ('world');" {
		t.Fatalf("Insert statement incorrectly formatted: %s\n", is.String())
	}
	is.Values = append(is.Values, "world1")
	if is.String() != "INSERT INTO `hello`.`world` VALUES ('world', 'world1');" {
		t.Fatalf("Insert statement incorrectly formatted: %s\n", is.String())
	}
	is.Values = []string{}
	if is.String() != "INSERT INTO `hello`.`world` VALUES ();" {
		t.Fatalf("Insert statement incorrectly formatted: %s\n", is.String())
	}
	is.Values = []string{"'hello'"}
	if is.String() != "INSERT INTO `hello`.`world` VALUES ('\\'hello\\'');" {
		t.Fatalf("Insert statement incorrectly formatted: %s\n", is.String())
	}
	is.Values = []string{"\\hello"}
	if is.String() != "INSERT INTO `hello`.`world` VALUES ('\\\\hello');" {
		t.Fatalf("Insert statement incorrectly formatted: %s\n", is.String())
	}
	is.Values = []string{"`hello"}
	if is.String() != "INSERT INTO `hello`.`world` VALUES ('\\`hello');" {
		t.Fatalf("Insert statement incorrectly formatted: %s\n", is.String())
	}
	is.Values = []string{"nil"}
	if is.String() != "INSERT INTO `hello`.`world` VALUES (NULL);" {
		t.Fatalf("Insert statement incorrectly formatted: %s\n", is.String())
	}
}

func TestUpdateStatement(t *testing.T) {
	us := NewUpdateStatement(
		"hello",
		"world",
		map[string]string{"world": "hello"},
		map[string]string{"id": "4001"},
	)
	if us.String() != "UPDATE `hello`.`world` SET `world` = 'hello' WHERE `id` = '4001';" {
		t.Fatalf("Update statement incorrectly formatted: %s\n", us.String())
	}
	us.Values["hello"] = "hello1"
	str := us.String()
	if str != "UPDATE `hello`.`world` SET `world` = 'hello', `hello` = 'hello1' WHERE `id` = '4001';" &&
		str != "UPDATE `hello`.`world` SET `hello` = 'hello1', `world` = 'hello' WHERE `id` = '4001';" {
		t.Fatalf("Update statement incorrectly formatted: %s\n", str)
	}
	delete(us.Values, "hello")
	us.Conditions["name"] = "Declan' Soper"
	str = us.String()
	if str != "UPDATE `hello`.`world` SET `world` = 'hello' WHERE `id` = '4001' AND `name` = 'Declan\\' Soper';" &&
		str != "UPDATE `hello`.`world` SET `world` = 'hello' WHERE `name` = 'Declan\\' Soper' AND `id` = '4001';" {
		t.Fatalf("Update statement incorrectly formatted: %s\n", str)
	}
	us.Values["world"] = "nil"
	str = us.String()
	if str != "UPDATE `hello`.`world` SET `world` = NULL WHERE `id` = '4001' AND `name` = 'Declan\\' Soper';" &&
		str != "UPDATE `hello`.`world` SET `world` = NULL WHERE `name` = 'Declan \\' Soper' AND `id` = '4001';" {
		t.Fatalf("Update statement incorrectly formatted: %s\n", str)
	}
}

func TestDeleteStatement(t *testing.T) {
	ds := NewDeleteStatement("hello", "world", map[string]string{"id": "4001"})
	if ds.String() != "DELETE FROM `hello`.`world` WHERE `id` = '4001';" {
		t.Fatalf("Delete statement incorrectly formatted: %s\n", ds.String())
	}
	ds.Conditions["hello"] = "world"
	str := ds.String()
	if str != "DELETE FROM `hello`.`world` WHERE `id` = '4001' AND `hello` = 'world';" &&
		str != "DELETE FROM `hello`.`world` WHERE `hello` = 'world' AND `id` = '4001';" {
		t.Fatalf("Delete statement incorrectly formatted: %s\n", str)
	}
}

func TestSelectStatement(t *testing.T) {
	ss := NewSelectStatement("hello", "world", map[string]string{"id": "4001"})
	str := ss.String()
	if str != "SELECT * FROM `hello`.`world` WHERE `id` = '4001';" {
		t.Fatalf("Select statement incorrectly formatted: %s\n", str)
	}
}

func TestModifyColumnPart(t *testing.T) {
	cd := NewColumnDef("name", mySQLVarChar)
	mc := ModifyColumnPart{ColumnDef: cd}

	str := mc.String()
	if str != "MODIFY COLUMN `name` VARCHAR" {
		t.Fatalf("Modify part incorrectly formatted: %s\n", str)
	}

	mc.Meta = 3
	str = mc.String()
	if str != "MODIFY COLUMN `name` VARCHAR(3)" {
		t.Fatalf("Modify part incorrectly formatted: %s\n", str)
	}

	mc.SetFirst(true)
	str = mc.String()
	if str != "MODIFY COLUMN `name` VARCHAR(3) FIRST" {
		t.Fatalf("Modify part incorrectly formatted: %s\n", str)
	}

	other := NewColumnDef("age", mySQLTiny)
	mc.SetAfter(other)
	str = mc.String()
	if str != "MODIFY COLUMN `name` VARCHAR(3) AFTER `age`" {
		t.Fatalf("Modify part incorrectly formatted: %s\n", str)
	}

	mc.DefaultValue = "1"
	str = mc.String()
	if str != "MODIFY COLUMN `name` VARCHAR(3) DEFAULT '1' AFTER `age`" {
		t.Fatalf("Modify part incorrectly formatted: %s\n", str)
	}

	mc.IsAutoIncrement = true
	str = mc.String()
	if str != "MODIFY COLUMN `name` VARCHAR(3) DEFAULT '1' AUTO_INCREMENT AFTER `age`" {
		t.Fatalf("Modify part incorrectly formatted: %s\n", str)
	}
}

func TestNewColumnPart(t *testing.T) {
	cd := NewColumnDef("name", mySQLVarChar)
	nc := NewColumnPart{ColumnDef: cd}

	str := nc.String()
	if str != "ADD COLUMN `name` VARCHAR" {
		t.Fatalf("Modify part incorrectly formatted: %s\n", str)
	}

	nc.Meta = 3
	str = nc.String()
	if str != "ADD COLUMN `name` VARCHAR(3)" {
		t.Fatalf("Modify part incorrectly formatted: %s\n", str)
	}

	nc.SetFirst(true)
	str = nc.String()
	if str != "ADD COLUMN `name` VARCHAR(3) FIRST" {
		t.Fatalf("Modify part incorrectly formatted: %s\n", str)
	}

	other := NewColumnDef("age", mySQLTiny)
	nc.SetAfter(other)
	str = nc.String()
	if str != "ADD COLUMN `name` VARCHAR(3) AFTER `age`" {
		t.Fatalf("Modify part incorrectly formatted: %s\n", str)
	}

	nc.DefaultValue = "1"
	str = nc.String()
	if str != "ADD COLUMN `name` VARCHAR(3) DEFAULT '1' AFTER `age`" {
		t.Fatalf("Modify part incorrectly formatted: %s\n", str)
	}
}

func TestDropColumnPart(t *testing.T) {
	dc := DropColumnPart{ColumnName: "test"}

	str := dc.String()
	if str != "DROP COLUMN `test`" {
		t.Fatalf("Drop part incorrectly formatted: %s\n", str)
	}
}

func TestAlterTableStatement(t *testing.T) {
	cd1 := NewColumnDef("id", mySQLInt24)

	modify := ModifyColumnPart{ColumnDef: cd1}
	as := NewAlterTableStatement("test", "test", modify)

	str := as.String()
	if str != "ALTER TABLE `test`.`test` MODIFY COLUMN `id` INT;" {
		t.Fatalf("Alter statement incorrectly formatted: %s\n", str)
	}

	cd2 := NewColumnDef("name", mySQLVarChar)
	cd2.Meta = 16
	new := NewColumnPart{ColumnDef: cd2}
	as = NewAlterTableStatement("test", "test", modify, new)

	str = as.String()
	if str != "ALTER TABLE `test`.`test` MODIFY COLUMN `id` INT, ADD COLUMN `name` VARCHAR(16);" {
		t.Fatalf("Alter statement incorrectly formatted: %s\n", str)
	}

	dc := DropColumnPart{ColumnName: "test"}
	as = NewAlterTableStatement("test", "test", new, dc)

	str = as.String()
	if str != "ALTER TABLE `test`.`test` ADD COLUMN `name` VARCHAR(16), DROP COLUMN `test`;" {
		t.Fatalf("Alter statement incorrectly formatted: %s\n", str)
	}
}

func TestGetColumnNullFromBitmask(t *testing.T) {
	bitmask := []byte{
		0,
		0,
		192,
		6,
	}

	if !getColumnIsNullFromBitmask(22, bitmask) {
		t.Fatal("22nd bit should be true\n")
	}
	if getColumnIsNullFromBitmask(0, bitmask) {
		t.Fatal("0th bit should be false\n")
	}
	if !getColumnIsNullFromBitmask(26, bitmask) {
		t.Fatal("26th bit should be true\n")
	}
	if getColumnIsNullFromBitmask(27, bitmask) {
		t.Fatal("27th bit should be false\n")
	}
}

func TestParseDataType(t *testing.T) {
	ty, meta, _ := parseDataType("varchar(50)")
	if ty != mySQLVarChar || meta != 50 {
		t.Fatalf(
			"type or meta incorrect for 'varchar(50)': %d, %d\n",
			ty,
			meta,
		)
	}

	ty, meta, _ = parseDataType("INT(11)")
	if ty != mySQLLong || meta != 11 {
		t.Fatalf("type or meta incorrect for 'INT(11)': %d, %d\n", ty, meta)
	}

	ty, meta, _ = parseDataType("DateTime")
	if ty != mySQLDateTime || meta != 0 {
		t.Fatalf("type or meta incorrect for 'DateTime': %d, %d\n", ty, meta)
	}

	ty, meta, _ = parseDataType("CHAR")
	if ty != mySQLString || meta != 0xfe01 {
		t.Fatalf("type or meta incorrect for 'CHAR': %d, %d\n", ty, meta)
	}
}

// func TestStringifyValues(t *testing.T) {
// 	vals := []interface{}{"\xA3", 0}
// 	valStr := stringifyValues(vals)
// 	if valStr[0] != "£" {
// 		t.Fatalf("Unicode conversion failed: %v\n", valStr[0])
// 	}
// 	if valStr[1] != "0" {
// 		t.Fatalf("Conversion to string failed: %v\n", valStr[1])
// 	}
// }
