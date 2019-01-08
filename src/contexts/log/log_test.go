package log

import (
	"contexts/ctx"
	"testing"
)

func TestLog(t *testing.T) {
	m := ctx.Pulse.Spawn(Index)
	m.Target().Begin(m)
	m.Sess("log", m)
	m.Log("error", "what %v", 123)
	m.Log("debug", "what %v", 123)
	m.Log("info", "what %v", 123)
	m.Log("cmd", "what %v", 123)
	m.Log("begin", "what %v", 123)
	m.Log("search", "what %v", 123)
	m.Cmd("log", "search", "what %v", 123)
}
