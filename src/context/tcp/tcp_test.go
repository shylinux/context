package tcp

import (
	"context"
	"flag"
	"testing"
	"time"
)

func TestOpen(t *testing.T) {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		t.Fatal("usages: -args address")
	}

	address := ":9393"
	address = args[0]

	//mysql -u root -p;
	//create database book;
	//grant all on book.* to user identified by 'word'

	ctx.Start()
	m := ctx.Pulse.Spawn(Index)

	m.Meta = nil
	Index.Cmd(m, "listen", address)
}
