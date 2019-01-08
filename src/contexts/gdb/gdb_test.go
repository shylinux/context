package gdb

import (
	"contexts/ctx"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	m := ctx.Pulse.Spawn(Index)
	m.Sess("gdb", m)
	m.Target().Begin(m).Start(m)

	go func() {
		t.Logf("%s", m.Cmd("wait", "config", "demo").Result(0))

		// t.Logf("%s", m.Cmd("demo", "what").Result(0))
	}()

	time.Sleep(time.Second * 3)

	m.Spawn().Cmd("goon", "nice")

	// gdb := m.Target().Server.(*GDB)
	// gdb.Goon("yes", "command", "demo")
	//
	time.Sleep(time.Second * 3)
}
