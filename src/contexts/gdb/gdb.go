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
	m.Cmd("nfs.save", m.Conf("logpid"), os.Getpid())
	gdb.wait = make(chan interface{}, 10)
	gdb.goon = make(chan os.Signal, 10)

	m.Confm("signal", func(sig string, action string) {
		m.Log("signal", "add %s: %s", action, sig)
		signal.Notify(gdb.goon, syscall.Signal(kit.Int(sig)))
	})

	for {
		select {
		case sig := <-gdb.goon:
			action := m.Conf("signal", sig)
			m.Log("signal", "%v: %v", action, sig)
			break
			switch action {
			case "segv":
			case "quit":
				m.Cmd("cli.exit", 0)
			case "restart":
				m.Cmd("cli.exit", 1)
			case "upgrade":
				m.Find("web.code").Cmd("upgrade", "system")
			default:
				// gdb.Goon(nil, "cache", "read", "value")
			}
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
		"logpid": &ctx.Config{Name: "logpid", Value: "var/run/bench.pid", Help: ""},
		"signal": &ctx.Config{Name: "signal", Value: map[string]interface{}{
			"1":  "HUP",
			"2":  "INT",
			"3":  "QUIT",
			"15": "TERM",
			"28": "WINCH",
			"30": "USR1",
			"31": "USR2",

			// "9":  "KILL",
			// "10": "BUS",
			// "11": "SEGV",
			// "17": "STOP",

			"5":  "TRAP",
			"6":  "ABRT",
			"14": "ALRM",
			"20": "CHLD",
			"19": "CONT",
			"18": "TSTP",
			"21": "TTIN",
			"22": "TTOUT",

			"13": "PIPE",
			"16": "URG",
			"23": "IO",

			"4":  "ILL",
			"7":  "EMT",
			"8":  "FPE",
			"12": "SYS",
			"24": "XCPU",
			"25": "XFSZ",
			"26": "VTALRM",
			"27": "PROF",
			"29": "INFO",
		}, Help: "信号"},
		"debug": &ctx.Config{Name: "debug", Value: map[string]interface{}{"value": map[string]interface{}{"enable": false},
			"trace": map[string]interface{}{"value": map[string]interface{}{"enable": true}},
			"context": map[string]interface{}{"value": map[string]interface{}{"enable": false},
				"begin": map[string]interface{}{"value": map[string]interface{}{"enable": false}},
				"start": map[string]interface{}{"value": map[string]interface{}{"enable": false}},
			},
			"command": map[string]interface{}{"value": map[string]interface{}{"enable": false},
				"shit": map[string]interface{}{"value": map[string]interface{}{"enable": true}},
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
		"_init": &ctx.Command{Name: "_init", Help: "等待调试", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Target().Start(m)
			return
		}},
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
