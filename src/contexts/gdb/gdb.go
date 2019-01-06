package gdb

import (
	"contexts/ctx"
	"os"
	"os/signal"
	"syscall"
	"toolkit"
)

type GDB struct {
	wait chan interface{}
	goon chan os.Signal

	*ctx.Context
}

func (gdb *GDB) Value(m *ctx.Message, arg ...interface{}) bool {
	if value, ok := kit.Chain(gdb.Configs["debug"].Value, append(arg, "value")).(map[string]interface{}); ok {
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
func (gdb *GDB) Wait(m *ctx.Message, arg ...interface{}) interface{} {
	for i := len(arg); i > 0; i-- {
		if gdb.Value(m, arg[:i]...) {
			if result := kit.Chain(kit.Chain(gdb.Configs["debug"].Value, arg[:i]), []string{"value", "result"}); result != nil {
				m.Log("error", "done %d %v", len(arg[:i]), arg)
				return result
			}
			m.Log("error", "wait %d %v", len(arg[:i]), arg)
			result := <-gdb.wait
			m.Log("error", "done %d %v %v", len(arg[:i]), arg, result)
			return result
		}
	}
	return <-gdb.wait

	return nil
}
func (gdb *GDB) Goon(result interface{}, arg ...interface{}) {
	m := gdb.Message()
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
	gdb.goon = make(chan os.Signal, 10)
	gdb.wait = make(chan interface{}, 10)
	m.Log("debug", "pid %d", os.Getpid())

	signal.Notify(gdb.goon, syscall.Signal(19))
	go func() {
		for {
			select {
			case sig := <-gdb.goon:
				m.Log("error", "signal %v", sig)
				gdb.Goon("hello", "cache", "read", "value")
			}
		}
	}()

	return gdb
}
func (gdb *GDB) Start(m *ctx.Message, arg ...string) bool {
	return false
}
func (gdb *GDB) Close(m *ctx.Message, arg ...string) bool {
	switch gdb.Context {
	case m.Target():
	case m.Source():
	}
	return true
}

var Index = &ctx.Context{Name: "gdb", Help: "调试中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"debug": &ctx.Config{Name: "debug", Value: map[string]interface{}{
			"value": map[string]interface{}{
				"enable": true,
			},
			"command": map[string]interface{}{
				"value": map[string]interface{}{
					"enable": true,
				},
			},
			"config": map[string]interface{}{
				"value": map[string]interface{}{
					"enable": true,
				},
			},
			"cache": map[string]interface{}{
				"value": map[string]interface{}{
					"enable": true,
				},
				"read": map[string]interface{}{
					"value": map[string]interface{}{
						"enable": true,
					},
					"ncontext": map[string]interface{}{
						"value": map[string]interface{}{
							"enable": true,
							"result": "hello",
						},
					},
				},
			},
		}},
	},
	Commands: map[string]*ctx.Command{},
}

func init() {
	gdb := &GDB{}
	gdb.Context = Index
	ctx.Index.Register(Index, gdb)
}
