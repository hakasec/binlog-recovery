# BinlogRecovery

This is a program for recovering data from row-based binlogs. It takes a binlog file, position, and a connection to a targeted slave server and executes row-based INSERT/UPDATE events.

## Building

To build the project run:

Get dependencies:

```sh
go mod tidy
```

Build project:

```sh
go build
```

To install to the PATH (optional):

```sh
go install
```

## Usage

To run the program, run `binlog-runner` (either from installing or locally) from the command line and use the below options:

```_
Usage:
  -binlog string
        Required. Starting binlog file
  -h string
        Host of the database to execute the binlog on (default "localhost:3306")
  -ignore-dbs string
        Comma delimited list of ignored databases
  -ignore-tables string
        Comma delimited list of ignored tables
  -p    Prompts for password to MySQL server
  -pass string
        Inline password, use -p for prompt
  -pos int
        Binlog starting position
  -u string
        Username for the database (default "root")
```

### Example

```sh
binlog-runner -binlog "<PATH TO BINLOG FOLDER>\myserver-bin.000913" -u myuser -pass mypass -h localhost -ignore-dbs mysql,performance_schema,information_schema -ignore-tables schema1.useless_table -pos 55046688
```

This sets the program to read from the **binlog** at `<PATH TO BINLOG FOLDER>\myserver-bin.000913` from **pos**ition `55046688` and update a server **h**osted at `localhost` with **u**sername `myuser` and **pass**word `mypass`, ignoring the `mysql`, `performance_schema`, and `information_schema` schemas and table `schema1.useless_table`.
