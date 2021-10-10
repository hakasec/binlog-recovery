package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"database/sql"

	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/go-sql-driver/mysql"
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

// eventRecv is a binlog parser callback function
// events are sent to this function, which then passes
// relevant events to the binlog executor
func eventRecv(event *replication.BinlogEvent) error {
	fmt.Printf(
		"Executing event (%T) in binlog %s: %d\n",
		event.Event,
		filepath.Base(currentBinlogDetails.Name),
		event.Header.LogPos,
	)

	switch t := event.Event.(type) {
	// if rotate, change current binlog details
	case *replication.RotateEvent:
		dir := filepath.Dir(cliFlags.BinlogFile)
		file := filepath.Join(dir, string(t.NextLogName))

		currentBinlogDetails = binlogDetails{
			Name:     file,
			Position: int64(t.Position),
		}
		return nil
	// if row event, check for ignored dbs and tables
	case *replication.RowsEvent:
		schema, table := string(t.Table.Schema), string(t.Table.Table)

		if containsStrCaseInsensitive(cliFlags.IgnoreDbs, schema) {
			return nil
		}

		if containsStrCaseInsensitive(cliFlags.IgnoreTables, schema+"."+table) {
			return nil
		}
	// if GTID event, track the current binlog position
	// this is used in case of an error
	case *replication.MariadbGTIDEvent:
		currentBinlogDetails.Position = int64(event.Header.LogPos)
		go logPosition("pos.txt", currentBinlogDetails.Position)
		return nil
	}

	// execute event
	if err := executor.ExecuteEvent(event); err != nil {
		fmt.Printf("%v\n", event.Event)
		return err
	}

	return nil
}

// logPosition writes a given pos to a file at filename
func logPosition(filename string, pos int64) {
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	f.WriteString(fmt.Sprint(pos))
}

// buildConnectionString builds a connection string with a username,
// password, and host
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
	// parse command line args
	cliFlags = NewCliFlags(os.Args[1:])
	// open database connection
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

	// create executor and aim at db via connection
	executor = NewBinlogExecutor(conn)
	// create binlog parser
	parser = replication.NewBinlogParser()

	// initialise current binlog details
	currentBinlogDetails = binlogDetails{
		Name:     cliFlags.BinlogFile,
		Position: cliFlags.Position,
	}

	// for loop to restart parsing when end of
	// active binlog is reached
	for {
		// start parsing, sends parsed events to eventRecv func
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
