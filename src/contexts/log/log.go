package log

import (
	"contexts/ctx"
	"io/ioutil"
	"path"
	"toolkit"

	"fmt"
	"os"
)

type LOG struct {
	file map[string]*os.File
	*ctx.Context
}

func (log *LOG) Value(msg *ctx.Message, arg ...interface{}) map[string]interface{} {
	args := append(kit.Trans(arg...), "value")

	if value, ok := kit.Chain(log.Configs["output"].Value, args).(map[string]interface{}); ok {
		if kit.Right(value["source"]) && kit.Format(value["source"]) != msg.Source().Name {
			return nil
		}

		if kit.Right(value["target"]) && kit.Format(value["target"]) != msg.Target().Name {
			return nil
		}

		// kit.Log("error", "value %v %v", kit.Format(args), kit.Format(value))
		return value
	}
	return nil
}
func (log *LOG) Log(msg *ctx.Message, action string, str string, arg ...interface{}) {
	m := log.Message()
	m.Capi("nlog", 1)

	args := kit.Trans(arg...)
	for _, v := range []string{action, "bench"} {
		for i := len(args); i >= 0; i-- {
			if value := log.Value(m, append([]string{v}, args[:i]...)); kit.Right(value) && kit.Right(value["file"]) {
				name := path.Join(m.Conf("logdir"), kit.Format(value["file"]))
				file, ok := log.file[name]
				if !ok {
					if f, e := os.Create(name); e == nil {
						file, log.file[name] = f, f
						kit.Log("error", "%s log file %s", "open", name)
					} else {
						kit.Log("error", "%s log file %s %s", "open", name, e)
						continue
					}
				}

				fmt.Fprintln(file, fmt.Sprintf("%d %s %s%s %s%s", m.Capi("nout", 1), msg.Format(value["meta"].([]interface{})...),
					kit.Format(value["color_begin"]), action, fmt.Sprintf(str, arg...), kit.Format(value["color_end"])))
				return
			}
		}
	}
}

func (log *LOG) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{}
	c.Configs = map[string]*ctx.Config{}

	s := new(LOG)
	s.Context = c
	return s
}
func (log *LOG) Begin(m *ctx.Message, arg ...string) ctx.Server {
	log.file = map[string]*os.File{}
	os.Mkdir(m.Conf("logdir"), 0770)
	kit.Log("error", "make log dir %s", m.Conf("logdir"))
	ioutil.WriteFile(m.Conf("logpid"), []byte(kit.Format(os.Getpid())), 0666)
	kit.Log("error", "save log file %s", m.Conf("logpid"))

	for _, v := range []string{"error", "bench", "debug"} {
		log.Log(m, v, "hello world\n")
		log.Log(m, v, "hello world")
	}
	return log
}
func (log *LOG) Start(m *ctx.Message, arg ...string) bool {
	return false
}
func (log *LOG) Close(m *ctx.Message, arg ...string) bool {
	switch log.Context {
	case m.Target():
	case m.Source():
	}
	return false
}

var Index = &ctx.Context{Name: "log", Help: "日志中心",
	Caches: map[string]*ctx.Cache{
		"nlog": &ctx.Cache{Name: "nlog", Value: "0", Help: "日志调用数量"},
		"nout": &ctx.Cache{Name: "nout", Value: "0", Help: "日志输出数量"},
	},
	Configs: map[string]*ctx.Config{
		"logdir": &ctx.Config{Name: "logdir", Value: "var/log", Help: ""},
		"logpid": &ctx.Config{Name: "logpid", Value: "var/log/bench.pid", Help: ""},
		"output": &ctx.Config{Name: "output", Value: map[string]interface{}{
			"error":  map[string]interface{}{"value": map[string]interface{}{"file": "error.log", "meta": []interface{}{"time", "ship"}, "color_begin": "\033[31m", "color_end": "\033[0m"}},
			"trace":  map[string]interface{}{"value": map[string]interface{}{"file": "error.log", "meta": []interface{}{"time", "ship"}, "color_begin": "\033[32m", "color_end": "\033[0m"}},
			"debug":  map[string]interface{}{"value": map[string]interface{}{"file": "debug.log", "meta": []interface{}{"time", "ship"}}},
			"search": map[string]interface{}{"value": map[string]interface{}{"file": "debug.log", "meta": []interface{}{"time", "ship"}}},
			"cbs":    map[string]interface{}{"value": map[string]interface{}{"file": "debug.log", "meta": []interface{}{"time", "ship"}}},

			"bench": map[string]interface{}{"value": map[string]interface{}{"file": "bench.log", "meta": []interface{}{"time", "ship"}}},
			"begin": map[string]interface{}{"value": map[string]interface{}{"file": "bench.log", "meta": []interface{}{"time", "ship"}, "color_begin": "\033[31m", "color_end": "\033[0m"}},
			"start": map[string]interface{}{"value": map[string]interface{}{"file": "bench.log", "meta": []interface{}{"time", "ship"}, "color_begin": "\033[31m", "color_end": "\033[0m"}},
			"close": map[string]interface{}{"value": map[string]interface{}{"file": "bench.log", "meta": []interface{}{"time", "ship"}, "color_begin": "\033[31m", "color_end": "\033[0m"}},

			"cmd": map[string]interface{}{"value": map[string]interface{}{"file": "bench.log", "meta": []interface{}{"time", "ship"}, "color_begin": "\033[32m", "color_end": "\033[0m"},
				"lex": map[string]interface{}{"value": map[string]interface{}{"file": "debug.log", "meta": []interface{}{"time", "ship"}, "color_begin": "\033[32m", "color_end": "\033[0m"}},
				"yac": map[string]interface{}{"value": map[string]interface{}{"file": "debug.log", "meta": []interface{}{"time", "ship"}, "color_begin": "\033[32m", "color_end": "\033[0m"}},
				"cli": map[string]interface{}{
					"cmd": map[string]interface{}{"value": map[string]interface{}{"file": "debug.log", "meta": []interface{}{"time", "ship"}, "color_begin": "\033[31m", "color_end": "\033[0m"}},
				},
			},
		}, Help: "日志输出配置"},
	},
	Commands: map[string]*ctx.Command{
		"log": &ctx.Command{Name: "log level string...", Help: "输出日志, level: 日志类型, string: 日志内容", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if log, ok := m.Target().Server.(*LOG); m.Assert(ok) {
				log.Log(m, arg[0], arg[1], arg[2:])
			}
			return
		}},
	},
}

func init() {
	log := &LOG{}
	log.Context = Index
	ctx.Index.Register(Index, log)
}
