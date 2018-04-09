package mdb

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"
)

func TestOpen(t *testing.T) {
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		t.Fatal("usages: -args source driver [table]")
	}

	source := "user:word@/book"
	driver := "mysql"
	source = args[0]
	driver = args[1]

	//mysql -u root -p;
	//create database book;
	//grant all on book.* to user identified by 'word'

	ctx.Start()
	ctx.Index.Conf("debug", "off")
	log.SetOutput(os.Stdout)
	m := ctx.Pulse.Spawn(Index)

	m.Meta = nil
	m.Cmd("open", source, driver)

	m.Meta = nil
	m.Cmd("exec", "insert into program(time, hash, name) values(?, ?, ?)", "1", "2", "3")

	m.Meta = nil
	m.Cmd("exec", "insert into program(time, hash, name) values(?, ?, ?)", "1", "2", "3")

	m.Meta = nil
	m.Cmd("exec", "insert into program(time, hash, name) values(?, ?, ?)", "2", "3", "4")

	m.Meta = nil
	m.Cmd("query", "select time, hash, name from program")

	t.Log()
	for i, rows := 0, len(m.Meta[m.Meta["append"][0]]); i < rows; i++ {
		for _, k := range m.Meta["append"] {
			t.Log(k, m.Meta[k][i])
		}
		t.Log()
	}

	if len(m.Meta["append"]) != 3 || len(m.Meta[m.Meta["append"][0]]) != 2 {
		t.Error()
	}

	m.Meta = nil
	// Index.Exit(m)
}
