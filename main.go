package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/siddontang/go-mysql/replication"
)

var (
	cliFlags *CliFlags
	conn     *sql.DB
	executor *BinlogExecutor
	parser   *replication.BinlogParser

	currentBinlogDetails binlogDetails
)

type binlogDetails struct {
	Name     string
	Position int64
}

func eventRecv(event *replication.BinlogEvent) error {
	fmt.Printf(
		"Executing event (%T) in binlog %s: %d\n",
		event.Event,
		filepath.Base(currentBinlogDetails.Name),
		event.Header.LogPos,
	)

	switch t := event.Event.(type) {
	case *replication.RotateEvent:
		dir := filepath.Dir(cliFlags.BinlogFile)
		file := filepath.Join(dir, string(t.NextLogName))

		currentBinlogDetails = binlogDetails{
			Name:     file,
			Position: int64(t.Position),
		}
		return nil
	case *replication.RowsEvent:
		schema, table := string(t.Table.Schema), string(t.Table.Table)

		if containsStrCaseInsensitive(cliFlags.IgnoreDbs, schema) {
			return nil
		}

		if containsStrCaseInsensitive(cliFlags.IgnoreTables, schema+"."+table) {
			return nil
		}
	case *replication.MariadbGTIDEvent:
		currentBinlogDetails.Position = int64(event.Header.LogPos)
		go logPosition("pos.txt", currentBinlogDetails.Position)
		return nil
	}

	if err := executor.ExecuteEvent(event); err != nil {
		fmt.Printf("%v\n", event.Event)
		return err
	}

	return nil
}

func logPosition(filename string, pos int64) {
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	f.WriteString(fmt.Sprint(pos))
}

func buildConnectionString(username, password, host string) string {
	conf := mysql.NewConfig()
	conf.User = username
	conf.Passwd = password
	conf.Net = "tcp"
	conf.Addr = host
	return conf.FormatDSN()
}

func main() {
	var err error
	cliFlags = NewCliFlags(os.Args[1:])
	conn, err = sql.Open(
		"mysql",
		buildConnectionString(
			cliFlags.Username,
			cliFlags.Password,
			cliFlags.Host,
		),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	executor = NewBinlogExecutor(conn)
	parser = replication.NewBinlogParser()

	currentBinlogDetails = binlogDetails{
		Name:     cliFlags.BinlogFile,
		Position: cliFlags.Position,
	}

	for {
		err = parser.ParseFile(
			currentBinlogDetails.Name,
			currentBinlogDetails.Position,
			eventRecv,
		)
		if err != nil {
			conn.Close()
			panic(err)
		}
		// sleep for 10 seconds
		time.Sleep(time.Second * 10)
	}
}
