package log

import (
	"contexts/ctx"
	"path"
	"toolkit"

	"fmt"
	"os"
)

type LOG struct {
	queue chan map[string]interface{}
	file  map[string]*os.File
	*ctx.Context
}

func (log *LOG) Value(msg *ctx.Message, arg ...interface{}) []string {
	args := append(kit.Trans(arg...))

	if value, ok := kit.Chain(log.Configs["output"].Value, args).(map[string]interface{}); ok {
		if kit.Right(value["source"]) && kit.Format(value["source"]) != msg.Source().Name {
			return nil
		}

		if kit.Right(value["target"]) && kit.Format(value["target"]) != msg.Target().Name {
			return nil
		}
		return kit.Trans(value["value"])
	}
	return nil
}
func (log *LOG) Log(msg *ctx.Message, action string, str string, arg ...interface{}) {
	if log.queue != nil {
		log.queue <- map[string]interface{}{
			"action": action,
			"str":    str,
			"arg":    arg,
			"msg":    msg,
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
	return log
}
func (log *LOG) Start(m *ctx.Message, arg ...string) bool {
	log.file = map[string]*os.File{}

	os.MkdirAll(m.Conf("logdir"), 0770)
	kit.Log("error", "make log dir %s", m.Conf("logdir"))

	log.queue = make(chan map[string]interface{}, 1024)
	// for _, v := range []string{"error", "bench", "debug"} {
	// 	log.Log(m, v, "hello world\n")
	// 	log.Log(m, v, "hello world")
	// }
	m.Cap("stream", m.Conf("output", []string{"bench", "value", "file"}))

	for {
		select {
		case l := <-log.queue:
			m.Capi("nlog", 1)
			msg := l["msg"].(*ctx.Message)

			args := kit.Trans(l["arg"].([]interface{})...)
		loop:
			for _, v := range []string{kit.Format(l["action"]), "bench"} {
				for i := len(args); i >= 0; i-- {
					value := log.Value(msg, append([]string{v}, args[:i]...))
					if !kit.Right(value) {
						continue
					}
					p := kit.Chains(log.Configs["output"].Value, []string{"file", value[0]})
					if p == "" {
						continue
					}

					file, ok := os.Stdout, true
					if p != "stdout" {
						name := path.Join(m.Conf("logdir"), p)
						file, ok = log.file[name]
						if !ok {
							if f, e := os.Create(name); e == nil {
								file, log.file[name] = f, f
								kit.Log("error", "%s log file %s", "open", name)
							} else {
								kit.Log("error", "%s log file %s %s", "open", name, e)
								continue
							}
						}
					}

					font := m.Conf("output", []string{"font", kit.Select("", value, 1)})
					meta := msg.Format(m.Confv("output", []string{"meta", kit.Select("short", value, 2)}).([]interface{})...)

					fmt.Fprintln(file, fmt.Sprintf("%d %s %s%s %s%s", m.Capi("nout", 1), meta, font,
						kit.Format(l["action"]), fmt.Sprintf(kit.Format(l["str"]), l["arg"].([]interface{})...),
						kit.Select("", "\033[0m", font != "")))
					break loop
				}
			}
		}
	}
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
		"output": &ctx.Config{Name: "output", Value: map[string]interface{}{
			"file": map[string]interface{}{
				"bench": "bench.log",
				"debug": "debug.log",
				"error": "error.log",
				"right": "right.log",
			},
			"font": map[string]interface{}{
				"red":    "\033[31m",
				"green":  "\033[32m",
				"blue":   "\033[34m",
				"yellow": "\033[33m",
			},
			"meta": map[string]interface{}{
				"short": []interface{}{"time", "ship"},
				"long":  []interface{}{"time", "ship"},
			},

			"error":  map[string]interface{}{"value": []interface{}{"error", "red"}},
			"trace":  map[string]interface{}{"value": []interface{}{"error", "red"}},
			"debug":  map[string]interface{}{"value": []interface{}{"debug"}},
			"search": map[string]interface{}{"value": []interface{}{"debug"}},
			"call":   map[string]interface{}{"value": []interface{}{"debug"}},
			"back":   map[string]interface{}{"value": []interface{}{"debug"}},

			"send": map[string]interface{}{"value": []interface{}{"debug"}},
			"recv": map[string]interface{}{"value": []interface{}{"debug"}},

			"bench": map[string]interface{}{"value": []interface{}{"bench"}},
			"begin": map[string]interface{}{"value": []interface{}{"bench", "red"}},
			"start": map[string]interface{}{"value": []interface{}{"bench", "red"}},
			"close": map[string]interface{}{"value": []interface{}{"bench", "red"}},
			"warn":  map[string]interface{}{"value": []interface{}{"bench", "yellow"}},
			"stack": map[string]interface{}{"value": []interface{}{"bench", "yellow"}},

			"cmd": map[string]interface{}{"value": []interface{}{"bench", "green"},
				"lex": map[string]interface{}{"value": []interface{}{"debug", "green"}},
				"yac": map[string]interface{}{"value": []interface{}{"debug", "green"}},
				"cli": map[string]interface{}{
					"cmd": map[string]interface{}{"value": []interface{}{"debug", "red"}},
				},
				"mdb": map[string]interface{}{
					"note": map[string]interface{}{"value": []interface{}{"debug", "red"}},
				},
				"aaa": map[string]interface{}{
					"auth": map[string]interface{}{"value": []interface{}{"debug", "red"}},
					"hash": map[string]interface{}{"value": []interface{}{"debug", "red"}},
					"rsa":  map[string]interface{}{"value": []interface{}{"debug", "red"}},
				},
			},
			"right": map[string]interface{}{"value": []interface{}{"right"}},
		}, Help: "日志输出配置"},
	},
	Commands: map[string]*ctx.Command{
		"_init": &ctx.Command{Name: "_init", Help: "启动日志", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			m.Target().Start(m)
			return
		}},
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
