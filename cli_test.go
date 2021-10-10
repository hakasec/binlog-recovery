package main

import (
	"testing"
)

func TestNewCliFlags(t *testing.T) {
	args := []string{
		"-binlog",
		"hello",
		"-h",
		"world",
		"-ignore-dbs",
		"hello,world,123",
		"-ignore-tables",
		"hello.world,paul.john",
		"-pass",
		"hello",
	}

	cli := NewCliFlags(args)
	if cli.BinlogFile != "hello" {
		t.Fatalf("Expected 'hello' got '%s'\n", cli.BinlogFile)
	}
	if cli.Host != "world" {
		t.Fatalf("Expected 'world' got '%s'\n", cli.Host)
	}
	if cli.IgnoreDbs[0] != "hello" ||
		cli.IgnoreDbs[1] != "world" ||
		cli.IgnoreDbs[2] != "123" {
		t.Fatalf("Expected '[hello world 123]' got '%v'\n", cli.IgnoreDbs)
	}
	if cli.IgnoreTables[0] != "hello.world" ||
		cli.IgnoreTables[1] != "paul.john" {
		t.Fatalf(
			"Expected '[hello.world paul.john]' got '%v'\n",
			cli.IgnoreTables,
		)
	}
	if cli.Password != "hello" {
		t.Fatalf("Expect 'hello' got '%s'\n", cli.Password)
	}
}
