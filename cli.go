package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/howeyc/gopass"
)

// CliFlags contains flags captured from the command line
type CliFlags struct {
	rawArgs []string

	BinlogFile     string
	Position       int64
	Host           string
	Username       string
	Password       string
	PromptPassword bool

	IgnoreDbs    []string
	IgnoreTables []string
}

// NewCliFlags creates a new CliFlags object
func NewCliFlags(args []string) *CliFlags {
	cli := &CliFlags{
		rawArgs: args,

		IgnoreDbs:    make([]string, 0),
		IgnoreTables: make([]string, 0),
	}
	cli.populateFlagsFromCli()
	return cli
}

func (c *CliFlags) populateFlagsFromCli() {
	f := flag.NewFlagSet("", flag.ExitOnError)
	f.StringVar(&c.BinlogFile, "binlog", "", "Required. Starting binlog file")
	f.Int64Var(&c.Position, "pos", 0, "Binlog starting position")
	f.StringVar(
		&c.Host,
		"h",
		"localhost:3306",
		"Host of the database to execute the binlog on",
	)
	f.StringVar(&c.Username, "u", "root", "Username for the database")
	f.StringVar(&c.Password, "pass", "", "Inline password, use -p for prompt")
	f.BoolVar(
		&c.PromptPassword,
		"p",
		false,
		"Prompts for password to MySQL server",
	)

	var (
		ignoreDbs    string
		ignoreTables string
	)
	f.StringVar(
		&ignoreDbs,
		"ignore-dbs",
		"",
		"Comma delimited list of ignored databases",
	)
	f.StringVar(
		&ignoreTables,
		"ignore-tables",
		"",
		"Comma delimited list of ignored tables",
	)
	f.Parse(c.rawArgs)

	// check for required flags
	if c.BinlogFile == "" {
		fmt.Fprintln(f.Output(), "binlog is a required argument.")
		f.Usage()
		os.Exit(1)
	}

	// parse ignore lists
	if ignoreDbs != "" {
		c.IgnoreDbs = strings.Split(ignoreDbs, ",")
	}
	if ignoreTables != "" {
		c.IgnoreTables = strings.Split(ignoreTables, ",")
	}

	// prompt for password
	if c.PromptPassword {
		pass, _ := gopass.GetPasswdPrompt(
			"Password: ",
			true,
			os.Stdin,
			f.Output(),
		)
		c.Password = string(pass)
	}
}
