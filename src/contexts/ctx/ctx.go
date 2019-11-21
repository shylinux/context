package ctx

import (
	"runtime"
	"toolkit"

	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type CTX struct {
	*Context
}

func (ctx *CTX) Spawn(m *Message, c *Context, arg ...string) Server {
	return &CTX{Context: c}
}
func (ctx *CTX) Begin(m *Message, arg ...string) Server {
	m.Option("ctx.routine", 0)
	m.Option("log.disable", true)
	m.Option("ctx.chain", "aaa", "ssh", "nfs", "cli", "web")

	m.Option("table.limit", 30)
	m.Option("table.offset", 0)
	m.Option("table.format", "object")
	m.Optionv("ctx.form", map[string]int{
		"table.format": 1, "table.offset": 1, "table.limit": 1,
	})

	m.root = m
	m.Sess(m.target.Name, m)
	for _, msg := range m.Search("") {
		if msg.target.root = m.target; msg.target == m.target {
			continue
		}
		msg.target.Begin(msg, arg...)
		m.Sess(msg.target.Name, msg)
	}
	return ctx
}
func (ctx *CTX) Start(m *Message, arg ...string) bool {
	if m.Optionv("bio.ctx", Index); len(arg) == 0 {
		kit.DisableLog = false
		m.Option("log.debug", kit.Right(os.Getenv("ctx_log_debug")))
		m.Option("log.disable", kit.Right(os.Getenv("ctx_log_disable")))
		m.Option("gdb.enable", kit.Right(os.Getenv("ctx_gdb_enable")))
		m.Cmd("log._init")
		m.Cmd("gdb._init")
		m.Cmd("ctx._init")
		m.Option("bio.modal", "active")

		m.Optionv("bio.ctx", m.Target())
		m.Optionv("bio.msg", m)
		m.Cap("stream", "stdio")
		m.Cmd("aaa.role", m.Option("userrole", "root"), "user", m.Option("username", m.Conf("runtime", "boot.username")))
		m.Option("sessid", m.Cmdx("aaa.user", "session", "select"))

		m.Cmd("nfs.source", m.Conf("cli.system", "script.init")).Cmd("nfs.source", "stdio").Cmd("nfs.source", m.Conf("cli.system", "script.exit"))
	} else {
		m.Cmd("ctx._init")
		m.Option("bio.modal", "action")
		for _, v := range m.Sess("cli").Cmd(arg).Meta["result"] {
			fmt.Printf("%s", v)
		}
	}
	m.Cmd("ctx._exit")
	return true
}
func (ctx *CTX) Close(m *Message, arg ...string) bool {
	return true
}

var Pulse = &Message{code: 0, time: time.Now(), source: Index, target: Index, Meta: map[string][]string{}}
var Index = &Context{Name: "ctx", Help: "模块中心", Server: &CTX{},
	Caches: map[string]*Cache{
		"ngo":      &Cache{Name: "ngo", Value: "0", Help: "协程数量"},
		"nserver":  &Cache{Name: "nserver", Value: "0", Help: "服务数量"},
		"ncontext": &Cache{Name: "ncontext", Value: "0", Help: "模块数量"},
		"nmessage": &Cache{Name: "nmessage", Value: "1", Help: "消息数量"},
	},
	Configs: map[string]*Config{
		"help": &Config{Name: "help", Value: map[string]interface{}{
			"index": []interface{}{
				"^_^      欢迎来到云境世界      ^_^",
				"^_^  Welcome to Context world  ^_^",
				"",
				"V2.1: Miss You Forever",
				"Date: 2019.10.29 13:14:21",
				"From: 2017.11.01 00:08:21",
				"",
				"Meet: shylinuxc@gmail.com",
				"More: https://shylinux.com",
				"More: https://github.com/shylinux/context",
				"",
			},
		}, Help: "帮助"},
		"time": &Config{Name: "timer", Value: map[string]interface{}{
			"unit": 1000, "close": "open", "format": "2006-01-02 15:04:05",
		}, Help: "时间参数"},
		"table": &Config{Name: "table", Value: map[string]interface{}{
			"space": " ", "compact": "false", "col_sep": " ", "row_sep": "\n",
			"offset": 0, "limit": 30,
		}, Help: "制表"},
		"call_timeout": &Config{Name: "call_timeout", Value: "60s", Help: "回调超时"},
	},
	Commands: map[string]*Command{
		"_init": &Command{Name: "_init", Help: "启动", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			for _, x := range []string{"lex", "cli", "yac", "nfs", "aaa", "ssh", "web"} {
				m.Cmd(x + "._init")
			}
			return
		}},
		"_exit": &Command{Name: "_exit", Help: "退出", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			for _, x := range []string{"nfs", "cli"} {
				m.Cmd(x + "._exit")
			}
			return
		}},

		"help": &Command{Name: "help topic", Help: "帮助", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Confm("help", "index", func(index int, value string) {
					m.Echo(value).Echo("\n")
				})
				return
			}

			switch arg[0] {
			case "context":
				switch len(arg) {
				case 1:
					keys := []string{}
					values := map[string]*Context{}
					m.Target().root.Travel(m, func(m *Message, i int) bool {
						if _, ok := values[m.Cap("module")]; !ok {
							keys = append(keys, m.Cap("module"))
							values[m.Cap("module")] = m.Target()
						}
						return false
					})

					sort.Strings(keys)
					for _, k := range keys {
						m.Echo("%s: %s %s\n", k, values[k].Name, values[k].Help)
					}
					break
				case 2:
					if msg := m.Find(arg[1]); msg != nil {
						m.Echo("%s: %s %s\n", arg[1], msg.Target().Name, msg.Target().Help)
						m.Echo("commands:\n")
						for k, v := range msg.Target().Commands {
							m.Echo("  %s: %s\n", k, v.Name)
						}
						m.Echo("configs:\n")
						for k, v := range msg.Target().Configs {
							m.Echo("  %s: %s\n", k, v.Name)
						}
						m.Echo("caches:\n")
						for k, v := range msg.Target().Caches {
							m.Echo("  %s: %s\n", k, v.Name)
						}
					}
				default:
					if msg := m.Find(arg[1]); msg != nil {
						m.Echo("%s: %s %s\n", arg[1], msg.Target().Name, msg.Target().Help)
						switch arg[2] {
						case "command":
							for k, v := range msg.Target().Commands {
								if k == arg[3] {
									m.Echo("%s: %s\n%s\n", k, v.Name, v.Help)
								}
							}
						case "config":
							for k, v := range msg.Target().Configs {
								if k == arg[3] {
									m.Echo("%s: %s\n  %s\n", k, v.Name, v.Help)
								}
							}
						case "cache":
							for k, v := range msg.Target().Caches {
								if k == arg[3] {
									m.Echo("%s: %s\n  %s\n", k, v.Name, v.Help)
								}
							}
						}
					}
				}
			case "command":
				keys := []string{}
				values := map[string]*Command{}
				for s := m.Target(); s != nil; s = s.context {
					for k, v := range s.Commands {
						if _, ok := values[k]; ok {
							continue
						}
						if len(arg) > 1 && k == arg[1] {
							switch help := v.Help.(type) {
							case []string:
								m.Echo("%s: %s\n", k, v.Name)
								for _, v := range help {
									m.Echo("  %s\n", v)
								}
							case string:
								m.Echo("%s: %s\n%s\n", k, v.Name, v.Help)
							}
							return
						}
						keys = append(keys, k)
						values[k] = v
					}
				}
				sort.Strings(keys)
				for _, k := range keys {
					m.Echo("%s: %s\n", k, values[k].Name)
				}
			case "config":
				keys := []string{}
				values := map[string]*Config{}
				for s := m.Target(); s != nil; s = s.context {
					for k, v := range s.Configs {
						if _, ok := values[k]; ok {
							continue
						}
						if len(arg) > 1 && k == arg[1] {
							m.Echo("%s(%s): %s %s\n", k, v.Value, v.Name, v.Help)
							return
						}
						keys = append(keys, k)
						values[k] = v
					}
				}
				sort.Strings(keys)
				for _, k := range keys {
					m.Echo("%s(%s): %s\n", k, values[k].Value, values[k].Name)
				}
			case "cache":
				keys := []string{}
				values := map[string]*Cache{}
				for s := m.Target(); s != nil; s = s.context {
					for k, v := range s.Caches {
						if _, ok := values[k]; ok {
							continue
						}
						if len(arg) > 1 && k == arg[1] {
							m.Echo("%s(%s): %s %s\n", k, v.Value, v.Name, v.Help)
							return
						}
						keys = append(keys, k)
						values[k] = v
					}
				}
				sort.Strings(keys)
				for _, k := range keys {
					m.Echo("%s(%s): %s\n", k, values[k].Value, values[k].Name)
				}
			}

			return
		}},
		"cache": &Command{Name: "cache [all] |key [value]|key = value|key name value help|delete key]", Help: "查看、读写、赋值、新建、删除缓存变量", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			all := false
			if len(arg) > 0 && arg[0] == "all" {
				arg, all = arg[1:], true
			}

			switch len(arg) {
			case 0:
				c.BackTrace(m, func(m *Message) bool {
					for k, v := range m.target.Caches {
						m.Add("append", "key", k)
						m.Add("append", "value", m.Cap(k))
						m.Add("append", "name", v.Name)
					}
					return !all
				})
				m.Sort("key", "str").Table()
				return
			case 1:
				m.Echo(m.Cap(arg[0]))
			case 2:
				if arg[0] == "delete" {
					delete(m.target.Caches, arg[1])
					return
				}
				m.Cap(arg[0], arg[1])
			case 3:
				m.Cap(arg[0], arg[0], arg[2], arg[0])
			default:
				m.Echo(m.Cap(arg[0], arg[1:]))
				return
			}
			return
		}},
		"config": &Command{Name: "config [all]", Help: []string{"配置管理",
			"brow: 配置列表",
			"key [chain [value]]: 读写配置",
			"export key...: 导出配置",
			"save|load file key...: 保存、加载配置",
			"create map|list|string key name help: 创建配置",
			"delete key: 删除配置",
		}, Form: map[string]int{"format": 1, "fields": -1}, Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			all := false
			if len(arg) > 0 && arg[0] == "all" {
				arg, all = arg[1:], true
			}
			if len(arg) == 0 {
				arg = append(arg, "brow")
			}

			action, which := "", "-1"
			have := map[string]bool{}
			switch arg[0] {
			case "brow":
			case "export":
				action, arg = arg[0], arg[1:]
				for _, v := range arg {
					have[v] = true
				}
			case "save", "load":
				action, which, arg = arg[0], arg[1], arg[2:]
				for _, v := range arg {
					have[v] = true
				}
			case "create", "delete":
				action, arg = arg[0], arg[1:]

			default:
				var value interface{}
				if len(arg) > 2 && arg[2] == "map" {
					for i := 3; i < len(arg)-1; i += 2 {
						m.Confv(arg[0], []interface{}{arg[1], arg[i]}, arg[i+1])
					}
					value = m.Confv(arg[0], arg[1])
				} else if len(arg) > 2 && arg[2] == "list" {
					for i := 3; i < len(arg); i += 1 {
						m.Confv(arg[0], []interface{}{arg[1], -2}, arg[i])
					}
					return
				} else if len(arg) > 1 && arg[1] == "list" {
					for i := 2; i < len(arg)-1; i += 1 {
						m.Confv(arg[0], -2, arg[i])
					}
					value = m.Confv(arg[0])
				} else if len(arg) > 1 && arg[1] == "map" {
					for i := 2; i < len(arg)-1; i += 2 {
						m.Confv(arg[0], arg[i], arg[i+1])
					}
					value = m.Confv(arg[0])
				} else if len(arg) > 2 {
					value = m.Confv(arg[0], arg[1], arg[2])
				} else if len(arg) > 1 {
					value = m.Confv(arg[0], arg[1])
				} else {
					value = m.Confv(arg[0])
				}

				msg := m.Spawn().Put("option", "_cache", value).Cmd("trans", "_cache")
				m.Copy(msg, "append").Copy(msg, "result")
				return
			}

			save := map[string]interface{}{}
			if action == "load" {
				f, e := os.Open(m.Cmdx("nfs.path", which))
				if e != nil {
					return e
				}
				defer f.Close()

				de := json.NewDecoder(f)
				if e = de.Decode(&save); e != nil {
					m.Log("info", "e: %v", e)
				}
			}

			c.BackTrace(m, func(m *Message) bool {
				for k, v := range m.target.Configs {
					switch action {
					case "export", "save":
						if len(have) == 0 || have[k] {
							save[k] = v.Value
						}
					case "load":
						if x, ok := save[k]; ok && (len(have) == 0 || have[k]) {
							v.Value = x
						}
					case "create":
						m.Assert(k != arg[1], "%s exists", arg[1])
					case "delete":
						if k == arg[0] {
							m.Echo(kit.Formats(v.Value))
							delete(m.target.Configs, k)
						}
					default:
						m.Add("append", "key", k)
						m.Add("append", "value", strings.Replace(strings.Replace(m.Conf(k), "\n", "\\n", -1), "\t", "\\t", -1))
						m.Add("append", "name", v.Name)
					}
				}
				switch action {
				case "create":
					var value interface{}
					switch arg[0] {
					case "map":
						value = map[string]interface{}{}
					case "list":
						value = []interface{}{}
					default:
						value = ""
					}
					m.target.Configs[arg[1]] = &Config{Name: arg[2], Value: value, Help: arg[3]}
					m.Echo(arg[1])
					return true
				}
				return !all
			})
			m.Sort("key", "str").Table()

			switch action {
			case "save":
				buf, e := json.MarshalIndent(save, "", "  ")
				m.Assert(e)
				m.Sess("nfs").Add("option", "data", string(buf)).Cmd("save", which)
			case "export":
				buf, e := json.MarshalIndent(save, "", "  ")
				m.Assert(e)
				m.Echo("%s", string(buf))
			}
			return
		}},
		"command": &Command{Name: "command", Help: []string{"查看或操作命令",
			"brow [all|cmd]: 查看命令",
			"help cmd: 查看帮助",
			"cmd...: 执行命令",
			"list [begin [end]] [prefix] [test [key value]...]: 命令列表",
			"add [list_name name] [list_help help] [cmd|__|_|_val]...: 添加命令, __: 可选参数, _: 必选参数, _val: 默认参数",
		}, Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				arg = append(arg, "brow")
			}

			switch arg[0] {
			case "brow":
				c.BackTrace(m, func(m *Message) bool {
					for k, v := range m.target.Commands {
						if strings.HasPrefix(k, "_") {
							continue
						}
						if len(arg) == 1 || arg[1] == "all" {
							m.Add("append", "cmd", k)
							m.Add("append", "name", v.Name)
						} else if arg[1] == k {
							m.Add("append", "cmd", k)
							m.Add("append", "name", v.Name)
							m.Add("append", "help", v.Name)
						}
					}
					return len(arg) == 1 || arg[1] != "all"
				})
				m.Sort("cmd").Table()

			case "help":
				m.Cmdy("ctx.help", "command", arg[1:])

			case "list":
				arg = arg[1:]
				if m.Cap("list_count") == "" {
					break
				}
				begin, end := 0, m.Capi("list_count")
				if len(arg) > 0 {
					if n, e := strconv.Atoi(arg[0]); e == nil {
						begin, arg = n, arg[1:]
					}
				}
				if len(arg) > 0 {
					if n, e := strconv.Atoi(arg[0]); e == nil {
						end, arg = n, arg[1:]
					}
				}
				prefix := ""
				if len(arg) > 0 && arg[0] != "test" {
					prefix, arg = arg[0], arg[1:]
				}

				test := false
				if len(arg) > 0 && arg[0] == "test" {
					test, arg = true, arg[1:]
					for i := 0; i < len(arg)-1; i += 2 {
						m.Add("option", arg[i], arg[i+1])
					}
				}

				for i := begin; i < end; i++ {
					index := fmt.Sprintf("%d", i)
					if c, ok := m.target.Commands[index]; ok {
						if prefix != "" && !strings.HasPrefix(c.Help.(string), prefix) {
							continue
						}

						if test {
							msg := m.Spawn().Cmd(index)
							m.Add("append", "index", i)
							m.Add("append", "help", c.Help)
							m.Add("append", "msg", msg.messages[0].code)
							m.Add("append", "res", msg.Result(0))
						} else {
							m.Add("append", "index", i)
							m.Add("append", "help", c.Help)
							m.Add("append", "command", fmt.Sprintf("%s", strings.Replace(c.Name, "\n", "\\n", -1)))
						}
					}
				}
				m.Table()

			case "add":
				if m.target.Caches == nil {
					m.target.Caches = map[string]*Cache{}
				}
				if _, ok := m.target.Caches["list_count"]; !ok {
					m.target.Caches["list_count"] = &Cache{Name: "list_count", Value: "0", Help: "list_count"}
				}
				if m.target.Commands == nil {
					m.target.Commands = map[string]*Command{}
				}

				arg = arg[1:]
				list_name, list_help := "", "list_cmd"
				if len(arg) > 1 && arg[0] == "list_name" {
					list_name, arg = arg[1], arg[2:]
				}
				if len(arg) > 1 && arg[0] == "list_help" {
					list_help, arg = arg[1], arg[2:]
				}

				m.target.Commands[m.Cap("list_count")] = &Command{Name: strings.Join(arg, " "), Help: list_help, Hand: func(cmd *Message, c *Context, key string, args ...string) (e error) {
					list := []string{}
					for _, v := range arg {
						if v == "__" {
							if len(args) > 0 {
								v, args = args[0], args[1:]
							} else {
								continue
							}
						} else if strings.HasPrefix(v, "_") {
							if len(args) > 0 {
								v, args = args[0], args[1:]
							} else if len(v) > 1 {
								v = v[1:]
							} else {
								v = "''"
							}
						}
						list = append(list, v)
					}
					list = append(list, args...)

					msg := cmd.Sess("cli").Set("option", "current_ctx", m.target.Name).Cmd("source", strings.Join(list, " "))
					cmd.Copy(msg, "append").Copy(msg, "result").Copy(msg, "target")
					return
				}}

				if list_name != "" {
					m.target.Commands[list_name] = m.target.Commands[m.Cap("list_count")]
				}
				m.Capi("list_count", 1)

			default:
				m.Cmdy(arg)
			}
			return
		}},
		"context": &Command{Name: "context [find|search] [root|back|home] [first|last|rand|magic] [module] [cmd...|switch|list|spawn|start|close]",
			Help: []string{"查找并操作模块",
				"查找方法, find: 精确查找, search: 模糊搜索",
				"查找起点, root: 根模块, back: 父模块, home: 本模块",
				"过滤结果, first: 取第一个, last: 取最后一个, rand: 随机选择, magics: 智能选择",
				"操作方法, cmd...: 执行命令, switch: 切换为当前, list: 查看所有子模块, spwan: 创建子模块并初始化, start: 启动模块, close: 结束模块",
			}, Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
				action := "switch"
				if len(arg) == 0 {
					action = "list"
				}

				method := "search"
				if len(arg) > 0 {
					switch arg[0] {
					case "find", "search":
						method, arg = arg[0], arg[1:]
					}
				}

				root := true
				if len(arg) > 0 {
					switch arg[0] {
					case "root":
						root, arg = true, arg[1:]
					case "home":
						root, arg = false, arg[1:]
					case "back":
						root, arg = false, arg[1:]
						if m.target.context != nil {
							m.target = m.target.context
						}
					}
				}

				ms := []*Message{}
				if len(arg) > 0 {
					switch method {
					case "find":
						if msg := m.Find(arg[0], root); msg != nil {
							ms, arg = append(ms, msg), arg[1:]
						}
					case "search":
						msg := m.Search(arg[0], root)
						if len(msg) > 1 || msg[0] != nil {
							if len(arg) > 1 {
								switch arg[1] {
								case "first":
									ms, arg = append(ms, msg[0]), arg[2:]
								case "last":
									ms, arg = append(ms, msg[len(msg)-1]), arg[2:]
								case "rand":
									ms, arg = append(ms, msg[rand.Intn(len(msg))]), arg[2:]
								case "magics":
									ms, arg = append(ms, msg...), arg[2:]
								default:
									ms, arg = append(ms, msg[0]), arg[1:]
								}
							} else {
								ms, arg = append(ms, msg[0]), arg[1:]
							}
						}

					}
				}

				if len(ms) == 0 {
					ms = append(ms, m)
				}

				if len(arg) > 0 {
					switch arg[0] {
					case "switch", "list", "spawn", "start", "close":
						action, arg = arg[0], arg[1:]
					default:
						action = "cmd"
					}
				}

				for _, msg := range ms {
					if msg == nil {
						continue
					}

					switch action {
					case "cmd":
						if len(arg) == 0 {
							arg = append(arg, "command")
						} else if arg[0] == "command" && len(arg) > 1 {
							arg = arg[1:]
						}
						if msg.Cmd(arg); !msg.Hand {
							msg = msg.Cmd("nfs.cmd", arg)
						}
						msg.CopyTo(m)

					case "switch":
						m.target = msg.target

					case "list":
						cs := []*Context{}
						if msg.target.Name != "ctx" {
							cs = append(cs, msg.target.context)
						}
						msg.target.Travel(msg, func(msg *Message, n int) bool {
							cs = append(cs, msg.target)
							return false
						})
						msg = m.Spawn()

						for _, v := range cs {
							if msg.target = v; v == nil {
								m.Add("append", "names", "")
								m.Add("append", "ctx", "")
								m.Add("append", "msg", "")
								m.Add("append", "status", "")
								m.Add("append", "stream", "")
								m.Add("append", "helps", "")
								continue
							}

							m.Add("append", "names", msg.target.Name)
							if msg.target.context != nil {
								m.Add("append", "ctx", msg.target.context.Name)
							} else {
								m.Add("append", "ctx", "")
							}
							if msg.target.message != nil {
								m.Add("append", "msg", msg.target.message.code)
							} else {
								m.Add("append", "msg", "")
							}
							m.Add("append", "status", msg.Cap("status"))
							m.Add("append", "stream", msg.Cap("stream"))
							m.Add("append", "helps", msg.target.Help)
						}
						m.Table()

					case "spawn":
						msg.target.Spawn(msg, arg[0], arg[1]).Begin(msg, arg[2:]...)
						m.Copy(msg, "append").Copy(msg, "result").Copy(msg, "target")

					case "start":
						msg.target.Start(msg, arg...)
						m.Copy(msg, "append").Copy(msg, "result").Copy(msg, "target")

					case "close":
						msg := m.Spawn()
						m.target = msg.target.context
						msg.target.Close(msg.target.message, arg...)
					}
				}
				return
			}},

		"message": &Command{Name: "message [code] [cmd...]", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			msg := m
			if len(arg) > 0 {
				if code, e := strconv.Atoi(arg[0]); e == nil {
					ms := []*Message{m.root}
					for i := 0; i < len(ms); i++ {
						if ms[i].Code() == code {
							msg = ms[i]
							arg = arg[1:]
							break
						}
						ms = append(ms, ms[i].messages...)
					}
				}
			}
			m.Optionv("bio.msg", msg)

			if len(arg) == 0 {
				ms := []*Message{msg.message, msg}
				for i := 0; i < len(ms); i++ {
					if ms[i] == nil {
						continue
					}
					if ms[i].message != nil {
						m.Push("msg", ms[i].message.code)
					} else {
						m.Push("msg", 0)
					}

					m.Push("code", ms[i].code)
					m.Push("time", ms[i].Time())
					m.Push("source", ms[i].source.Name)
					m.Push("target", ms[i].target.Name)
					m.Push("details", kit.Format(ms[i].Meta["detail"]))
					if i > 0 {
						ms = append(ms, ms[i].messages...)
					}
				}
				m.Table()
				return
			}

			switch arg[0] {
			case "time", "code", "ship", "full", "chain", "stack":
				m.Echo(msg.Format(arg[0]))
			case "spawn":
				sub := msg.Spawn()
				m.Echo("%d", sub.code)
			case "call":
			case "back":
				msg.Back(m)
			case "free":
				msg.Free()
			default:
				m.Cmd(arg)
			}
			return
		}},
		"detail": &Command{Name: "detail [index] [value...]", Help: "查看或添加参数", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			msg := m.Optionv("bio.msg").(*Message)
			if len(arg) == 0 {
				for i, v := range msg.Meta["detail"] {
					m.Push("index", i)
					m.Push("value", v)
				}
				m.Table()
				return
			}

			index := 0
			if i, e := strconv.Atoi(arg[0]); e == nil {
				index, arg = i, arg[1:]
			}
			m.Echo("%s", msg.Detail(index, arg))
			return
		}},
		"copy": &Command{Name: "copy", Help: "查看或添加选项", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			skip := false
			if arg[0] == "skip" {
				skip, arg = true, arg[1:]
			}
			msg := m.Optionv("bio.msg").(*Message)
			// 去年尾参
			for i := len(arg) - 1; i >= 0; i-- {
				if arg[i] == "" {
					arg = arg[:i]
				} else {
					break
				}
			}
			// 默认参数
			args, j := make([]string, 0, len(arg)), 1
			for i := 0; i < len(arg); i++ {
				if strings.HasPrefix(arg[i], "__") {
					if j < len(msg.Meta["detail"]) {
						args = append(args, msg.Meta["detail"][j:]...)
					}
					j = len(msg.Meta["detail"])
				} else if strings.HasPrefix(arg[i], "_") {
					args = append(args, kit.Select(arg[i][1:], msg.Detail(j)))
					j++
				} else {
					args = append(args, arg[i])
				}
			}
			if !skip && j < len(msg.Meta["detail"]) {
				args = append(args, msg.Meta["detail"][j:]...)
			}
			msg.Cmdy(args)
			return
		}},
		"table": &Command{Name: "table", Help: "查看或添加选项", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			msg := m.Optionv("bio.msg").(*Message)
			if len(msg.Meta["append"]) == 0 {
				msg.Meta["append"] = arg
			} else {
				for i, k := range msg.Meta["append"] {
					msg.Push(k, kit.Select("", arg, i))
				}
			}
			return
		}},
		"option": &Command{Name: "option", Help: "查看或添加选项", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			all := false
			if len(arg) > 0 && arg[0] == "all" {
				all, arg = true, arg[1:]
			}

			msg := m.Optionv("bio.msg").(*Message)
			switch len(arg) {
			case 0:
				vals := map[string]interface{}{}
				list := []string{}

				for back := msg; back != nil; back = back.message {
					for _, k := range back.Meta["option"] {
						if _, ok := vals[k]; !ok {
							if _, ok := back.Data[k]; ok {
								vals[k] = back.Data[k]
							} else {
								vals[k] = back.Meta[k]
							}
							list = append(list, k)
						}
					}
					if !all {
						break
					}
				}
				sort.Strings(list)

				for _, k := range list {
					m.Push("key", k)
					m.Push("val", kit.Format(vals[k]))
				}
				m.Table()
			case 1:
				switch v := msg.Optionv(arg[0]).(type) {
				case []string:
					m.Echo(strings.Join(v, ""))
				default:
					m.Echo(kit.Format(v))
				}
			default:
				m.Echo(m.Option(arg[0], arg[1]))
			}
			return
		}},
		"append": &Command{Name: "append", Help: "查看或添加附加值", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			msg := m.Optionv("bio.msg").(*Message)
			if len(arg) == 0 {
				m.Copy(msg, "append")
				m.Table()
				return
			}
			if len(arg) == 1 {
				for i, v := range msg.Meta[arg[0]] {
					m.Push("index", i)
					m.Push("value", v)
				}
				m.Table()
				return
			}
			msg.Push(arg[0], arg[1])
			return
		}},
		"result": &Command{Name: "result [index] [value...]", Help: "查看或添加返回值", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			msg := m.Optionv("bio.msg").(*Message)
			if len(arg) == 0 {
				for i, v := range msg.Meta["result"] {
					m.Push("index", i)
					m.Push("value", strings.Replace(v, "\n", "\\n", -1))
				}
				m.Table()
				return
			}

			index := -2
			if i, e := strconv.Atoi(arg[0]); e == nil {
				index, arg = i, arg[1:]
			}
			m.Echo("%s", msg.Result(index, arg))
			return
		}},

		"trans": &Command{Name: "trans option [type|data|json] limit 10 [index...]", Help: "数据转换",
			Form: map[string]int{"format": 1, "fields": -1},
			Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
				value, arg := m.Optionv(arg[0]), arg[1:]
				if v, ok := value.(string); ok {
					json.Unmarshal([]byte(v), &value)
				}

				view := "data"
				if len(arg) > 0 {
					switch arg[0] {
					case "type", "data", "json":
						view, arg = arg[0], arg[1:]
					}
				}

				limit := kit.Int(kit.Select(m.Conf("table", "limit"), m.Option("table.limit")))
				if len(arg) > 0 && arg[0] == "limit" {
					limit, arg = kit.Int(arg[1]), arg[2:]
				}

				chain := strings.Join(arg, ".")
				if chain != "" {
					value = kit.Chain(value, chain)
				}

				switch view {
				case "type": // 查看数据类型
					switch value := value.(type) {
					case map[string]interface{}:
						for k, v := range value {
							m.Add("append", "key", k)
							m.Add("append", "type", fmt.Sprintf("%T", v))
						}
						m.Sort("key", "str").Table()
					case []interface{}:
						for k, v := range value {
							m.Add("append", "key", k)
							m.Add("append", "type", fmt.Sprintf("%T", v))
						}
						m.Sort("key", "int").Table()
					case nil:
					default:
						m.Add("append", "key", chain)
						m.Add("append", "type", fmt.Sprintf("%T", value))
						m.Sort("key", "str").Table()
					}
					return
				case "data":
				case "json": // 查看文本数据
					b, e := json.MarshalIndent(value, "", " ")
					m.Assert(e)
					m.Echo(string(b))
					return nil
				}

				switch val := value.(type) {
				case map[string]interface{}:
					if m.Option("format") == "table" {
						fields := []string{}
						has := map[string]bool{}
						if m.Options("fields") {
							fields = m.Optionv("fields").([]string)
						} else {
							i := 0
							for _, v := range val {
								if i++; i > kit.Int(kit.Select(m.Conf("table", "limit"), m.Option("table.limit"))) {
									break
								}
								if line, ok := v.(map[string]interface{}); ok {
									for k, _ := range line {
										if h, ok := has[k]; ok && h {
											continue
										}
										has[k], fields = true, append(fields, k)
									}
								}
							}
							sort.Strings(fields)
						}

						i := 0
						for k, v := range val {
							if i++; i > kit.Int(kit.Select(m.Conf("table", "limit"), m.Option("table.limit"))) {
								break
							}
							if line, ok := v.(map[string]interface{}); ok {
								m.Add("append", "key", k)
								for _, field := range fields {
									m.Add("append", field, kit.Format(line[field]))
								}
							}
						}
						m.Table()
						break
					}

					for k, v := range val {
						if m.Option("format") == "object" {
							m.Add("append", k, v)
							continue
						}

						m.Add("append", "key", k)
						switch val := v.(type) {
						case nil:
							m.Add("append", "value", "")
						case string:
							m.Add("append", "value", val)
						case float64:
							m.Add("append", "value", fmt.Sprintf("%d", int(val)))
						default:
							b, _ := json.Marshal(val)
							m.Add("append", "value", fmt.Sprintf("%s", string(b)))
						}
					}
					if m.Option("format") != "object" {
						m.Sort("key", "str")
					}
					m.Table()
				case map[string]string:
					for k, v := range val {
						m.Add("append", "key", k)
						m.Add("append", "value", v)
					}
					m.Sort("key", "str").Table()
				case []interface{}:
					fields := map[string]int{}
					for i, v := range val {
						if i >= limit {
							break
						}
						switch val := v.(type) {
						case map[string]interface{}:
							for k, _ := range val {
								fields[k]++
							}
						}
					}

					if len(fields) > 0 {
						for i, v := range val {
							if i >= limit {
								break
							}
							switch val := v.(type) {
							case map[string]interface{}:
								for k, _ := range fields {
									switch value := val[k].(type) {
									case nil:
										m.Add("append", k, "")
									case string:
										m.Add("append", k, value)
									case float64:
										m.Add("append", k, fmt.Sprintf("%d", int(value)))
									default:
										b, _ := json.Marshal(value)
										m.Add("append", k, fmt.Sprintf("%v", string(b)))
									}
								}
							}
						}
					} else {
						for i, v := range val {
							switch val := v.(type) {
							case nil:
								m.Add("append", "index", i)
								m.Add("append", "value", "")
							case string:
								m.Add("append", "index", i)
								m.Add("append", "value", val)
							case float64:
								m.Add("append", "index", i)
								m.Add("append", "value", fmt.Sprintf("%v", int(val)))
							default:
								m.Add("append", "index", i)
								b, _ := json.Marshal(val)
								m.Add("append", "value", fmt.Sprintf("%v", string(b)))
							}
						}
					}
					m.Table()
				case []string:
					for i, v := range val {
						m.Add("append", "index", i)
						m.Add("append", "value", v)
					}
					m.Table()
				case string:
					m.Echo("%s", val)
				case float64:
					m.Echo("%d", int(val))
				case nil:
				default:
					b, _ := json.Marshal(val)
					m.Echo("%s", string(b))
				}
				return
			}},
		"select": &Command{Name: "select field...",
			Form: map[string]int{"reg": 2, "eq": 2, "expand": 2, "hide": -1, "fields": -1, "group": 1, "order": 2, "limit": 1, "offset": 1, "format": -1, "trans_map": -1, "vertical": 0},
			Help: "选取数据", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
				msg := m.Set("result").Spawn()

				// 解析
				if len(m.Meta["append"]) == 0 {
					return
				}
				nrow := len(m.Meta[m.Meta["append"][0]])
				keys := []string{}
				for i := 0; i < nrow; i++ {
					for j := 0; j < len(m.Meta["expand"]); j += 2 {
						var value interface{}
						json.Unmarshal([]byte(m.Meta[m.Meta["expand"][j]][i]), &value)
						if m.Meta["expand"][j+1] != "" {
							value = kit.Chain(value, m.Meta["expand"][j+1])
						}

						switch val := value.(type) {
						case map[string]interface{}:
							for k, _ := range val {
								keys = append(keys, k)
							}
						default:
							keys = append(keys, m.Meta["expand"][j+1])
						}
					}
				}
				for i := 0; i < nrow; i++ {
					for _, k := range keys {
						m.Add("append", k, "")
					}
				}
				for i := 0; i < nrow; i++ {
					for j := 0; j < len(m.Meta["expand"]); j += 2 {
						var value interface{}
						json.Unmarshal([]byte(m.Meta[m.Meta["expand"][j]][i]), &value)
						if m.Meta["expand"][j+1] != "" {
							value = kit.Chain(value, m.Meta["expand"][j+1])
						}

						switch val := value.(type) {
						case map[string]interface{}:
							for k, v := range val {
								switch val := v.(type) {
								case string:
									m.Meta[k][i] = val
								case float64:
									m.Meta[k][i] = fmt.Sprintf("%d", int(val))
								default:
									b, _ := json.Marshal(val)
									m.Meta[k][i] = string(b)
								}
							}
						case string:
							m.Meta[m.Meta["expand"][j+1]][i] = val
						default:
							m.Meta[m.Meta["expand"][j+1]][i] = kit.Format(val)
						}
					}
				}

				// 隐藏列
				hides := map[string]bool{}
				for _, k := range m.Meta["hide"] {
					hides[k] = true
				}
				if len(arg) == 0 {
					arg = append(arg, m.Meta["append"]...)
				}
				for i := 0; i < nrow; i++ {
					// if len(arg) == 0 || strings.Contains(m.Meta[arg[0]][i], arg[1]) {
					if m.Has("eq") {
						if m.Meta[m.Meta["eq"][0]][i] != m.Meta["eq"][1] {
							continue
						}
					}
					if m.Has("reg") {
						if b, e := regexp.MatchString(m.Meta["reg"][1], m.Meta[m.Meta["reg"][0]][i]); e != nil || !b {
							continue
						}
					}
					for _, k := range arg {
						if hides[k] {
							continue
						}
						msg.Add("append", k, kit.Select("", m.Meta[k], i))
					}
				}
				if len(msg.Meta["append"]) == 0 {
					return
				}
				if len(msg.Meta[msg.Meta["append"][0]]) == 0 {
					return
				}

				// 聚合
				if m.Set("append"); m.Has("group") {
					group := m.Option("group")
					nrow := len(msg.Meta[msg.Meta["append"][0]])

					for i := 0; i < nrow; i++ {
						count := 1

						if group != "" && msg.Meta[group][i] == "" {
							msg.Add("append", "count", 0)
							continue
						}

						for j := i + 1; j < nrow; j++ {
							if group == "" || msg.Meta[group][i] == msg.Meta[group][j] {
								count++
								for _, k := range msg.Meta["append"] {
									if k == "count" {
										continue
									}
									if k == group {
										continue
									}
									m, e := strconv.Atoi(msg.Meta[k][i])
									if e != nil {
										continue
									}
									n, e := strconv.Atoi(msg.Meta[k][j])
									if e != nil {
										continue
									}
									msg.Meta[k][i] = fmt.Sprintf("%d", m+n)

								}

								if group != "" {
									msg.Meta[group][j] = ""
								}
							}
						}

						msg.Add("append", "count", count)
						for _, k := range msg.Meta["append"] {
							m.Add("append", k, msg.Meta[k][i])
						}
						if group == "" {
							break
						}
					}
				} else {
					m.Copy(msg, "append")
				}

				// 排序
				if m.Has("order") {
					m.Sort(m.Option("order"), kit.Select("str", m.Meta["order"], 1))
				}

				// 分页
				offset := kit.Int(kit.Select("0", m.Option("offset")))
				limit := kit.Int(kit.Select(m.Conf("table", "limit"), m.Option("table.limit")))

				nrow = len(m.Meta[m.Meta["append"][0]])
				if offset > nrow {
					offset = nrow
				}
				if limit+offset > nrow {
					limit = nrow - offset
				}
				for _, k := range m.Meta["append"] {
					m.Meta[k] = m.Meta[k][offset : offset+limit]
				}

				// 值转换
				for i := 0; i < len(m.Meta["trans_map"]); i += 3 {
					trans := m.Meta["trans_map"][i:]
					for j := 0; j < len(m.Meta[trans[0]]); j++ {
						if m.Meta[trans[0]][j] == trans[1] {
							m.Meta[trans[0]][j] = trans[2]
						}
					}
				}

				// 格式化
				for i := 0; i < len(m.Meta["format"])-1; i += 2 {
					format := m.Meta["format"]
					for j, v := range m.Meta[format[i]] {
						if v != "" {
							m.Meta[format[i]][j] = fmt.Sprintf(format[i+1], v)
						}
					}
				}

				// 变换列
				if m.Has("vertical") {
					msg := m.Spawn()
					nrow := len(m.Meta[m.Meta["append"][0]])
					sort.Strings(m.Meta["append"])
					msg.Add("append", "field", "")
					msg.Add("append", "value", "")
					for i := 0; i < nrow; i++ {
						for _, k := range m.Meta["append"] {
							msg.Add("append", "field", k)
							msg.Add("append", "value", m.Meta[k][i])
						}
						msg.Add("append", "field", "")
						msg.Add("append", "value", "")
					}
					m.Set("append").Copy(msg, "append")
				}

				// 取单值
				// if len(arg) > 2 {
				// 	if len(m.Meta[arg[2]]) > 0 {
				// 		m.Echo(m.Meta[arg[2]][0])
				// 	}
				// 	return
				// }

				m.Set("result").Table()
				return
			}},
	},
}

func Start(args ...string) bool {
	runtime.GOMAXPROCS(1)

	if len(args) == 0 {
		args = append(args, os.Args[1:]...)
	}
	if len(args) > 0 && args[0] == "start" {
		args = args[1:]
	}
	if len(args) > 0 && args[0] == "daemon" {
		Pulse.Options("bio.modal", "daemon")
		Pulse.Options("daemon", true)
		args = args[1:]
	}

	kit.DisableLog = true
	if Index.Begin(Pulse, args...); Index.Start(Pulse, args...) {
		return Index.Close(Pulse, args...)
	}
	return Index.message.Wait()
}
