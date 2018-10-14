package gdb

import (
	"contexts/ctx"
)

type GDB struct {
	*ctx.Context
}

func (gdb *GDB) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server { // {{{
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(GDB)
	s.Context = c
	return s
}

// }}}
func (gdb *GDB) Begin(m *ctx.Message, arg ...string) ctx.Server { // {{{
	return gdb
}

// }}}
func (gdb *GDB) Start(m *ctx.Message, arg ...string) bool { // {{{
	return false
}

// }}}
func (gdb *GDB) Close(m *ctx.Message, arg ...string) bool { // {{{
	switch gdb.Context {
	case m.Target():
	case m.Source():
	}
	return true
}

var Index = &ctx.Context{Name: "gdb", Help: "调试中心",
	Caches:   map[string]*ctx.Cache{},
	Configs:  map[string]*ctx.Config{},
	Commands: map[string]*ctx.Command{},
}

func init() {
	gdb := &GDB{}
	gdb.Context = Index
	ctx.Index.Register(Index, gdb)
}
