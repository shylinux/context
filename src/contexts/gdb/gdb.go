package gdb

import (
	"contexts/ctx"
	"toolkit"

	"os"
	"os/signal"
	"syscall"
)

type GDB struct {
	wait chan interface{}
	goon chan os.Signal

	*ctx.Context
}

func (gdb *GDB) Value(m *ctx.Message, arg ...interface{}) bool {
	if value, ok := kit.Chain(gdb.Configs["debug"].Value, kit.Trans(arg, "value")).(map[string]interface{}); ok {
		if !kit.Right(value["enable"]) {
			return false
		}

		if kit.Right(value["source"]) && kit.Format(value["source"]) != m.Source().Name {
			return false
		}

		if kit.Right(value["target"]) && kit.Format(value["target"]) != m.Target().Name {
			return false
		}

		m.Log("error", "value %v %v", arg, kit.Formats(value))
		return true
	}
	return false
}
func (gdb *GDB) Wait(msg *ctx.Message, arg ...interface{}) interface{} {
	m := gdb.Message()
	if m.Cap("status") != "start" {
		return nil
	}

	for i := len(arg); i > 0; i-- {
		if gdb.Value(m, arg[:i]...) {
			if result := kit.Chain(kit.Chain(gdb.Configs["debug"].Value, arg[:i]), []string{"value", "result"}); result != nil {
				m.Log("error", "done %d %v", len(arg[:i]), arg)
				return result
			}
			if arg[0] == "trace" {
				m.Log("error", "%s", m.Format("full"))
			}
			m.Log("error", "wait %d %v", len(arg[:i]), arg)
			result := <-gdb.wait
			m.Log("error", "done %d %v %v", len(arg[:i]), arg, result)
			return result
		}
	}
	return nil
}
func (gdb *GDB) Goon(result interface{}, arg ...interface{}) {
	m := gdb.Message()
	if m.Cap("status") != "start" {
		return
	}

	m.Log("error", "goon %v %v", arg, result)
	gdb.wait <- result
}

func (gdb *GDB) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(GDB)
	s.Context = c
	return s
}
func (gdb *GDB) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return gdb
}
func (gdb *GDB) Start(m *ctx.Message, arg ...string) bool {
	gdb.goon = make(chan os.Signal, 10)
	gdb.wait = make(chan interface{}, 10)
	signal.Notify(gdb.goon, syscall.Signal(19))
	for {
		select {
		case sig := <-gdb.goon:
			m.Log("error", "signal %v", sig)
			gdb.Goon(nil, "cache", "read", "value")
		}
	}
	return true
}
func (gdb *GDB) Close(m *ctx.Message, arg ...string) bool {
	switch gdb.Context {
	case m.Target():
	case m.Source():
	}
	return false
}

var Index = &ctx.Context{Name: "gdb", Help: "调试中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"debug": &ctx.Config{Name: "debug", Value: map[string]interface{}{"value": map[string]interface{}{"enable": false},
			"trace": map[string]interface{}{"value": map[string]interface{}{"enable": true}},
			"context": map[string]interface{}{"value": map[string]interface{}{"enable": false},
				"begin": map[string]interface{}{"value": map[string]interface{}{"enable": false}},
				"start": map[string]interface{}{"value": map[string]interface{}{"enable": false}},
			},
			"command": map[string]interface{}{"value": map[string]interface{}{"enable": false},
				"demo": map[string]interface{}{"value": map[string]interface{}{"enable": true}},
			},
			"config": map[string]interface{}{"value": map[string]interface{}{"enable": true}},
			"cache": map[string]interface{}{"value": map[string]interface{}{"enable": false},
				"read": map[string]interface{}{"value": map[string]interface{}{"enable": false},
					"ncontext": map[string]interface{}{"value": map[string]interface{}{"enable": false}},
				},
			},
		}},
	},
	Commands: map[string]*ctx.Command{
		"demo": &ctx.Command{Name: "wait arg...", Help: "等待调试", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Echo("hello world")
			return
		}},
		"wait": &ctx.Command{Name: "wait arg...", Help: "等待调试", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if gdb, ok := m.Target().Server.(*GDB); m.Assert(ok) {
				switch v := gdb.Wait(m, arg).(type) {
				case string:
					m.Echo(v)
				case nil:
				}
			}
			return
		}},
		"goon": &ctx.Command{Name: "goon arg...", Help: "继续运行", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if gdb, ok := m.Target().Server.(*GDB); m.Assert(ok) {
				gdb.Goon(arg)
			}
			return
		}},
	},
}

func init() {
	gdb := &GDB{}
	gdb.Context = Index
	ctx.Index.Register(Index, gdb)
}
