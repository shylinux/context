package ctx

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"math/rand"
	"os"
	"sort"
	"time"
	"toolkit"
)

type CTX struct {
}

func (ctx *CTX) Spawn(m *Message, c *Context, arg ...string) Server {
	s := new(CTX)
	return s
}
func (ctx *CTX) Begin(m *Message, arg ...string) Server {
	m.Sess(m.target.Name, m)
	m.target.root = m.target
	m.root = m
	m.Cap("begin_time", m.Time())
	m.Cap("goos", runtime.GOOS)
	for _, msg := range m.Search("") {
		msg.target.root = m.target
		if msg.target == m.target {
			continue
		}
		msg.target.Begin(msg, arg...)
		m.Sess(msg.target.Name, msg)
	}
	return ctx
}
func (ctx *CTX) Start(m *Message, arg ...string) bool {
	m.Cmd("ctx.init")
	if m.Optionv("ps_target", Index); len(arg) == 0 {
		m.Cap("stream", "shy")
		m.Cmd("cli.source", m.Conf("runtime", "init_shy")).Cmd("cli.source", "stdio").Cmd("cli.source", m.Conf("runtime", "exit_shy"))
	} else {
		m.Cmd("cli.source", arg)
	}

	return true
}
func (ctx *CTX) Close(m *Message, arg ...string) bool {
	return true
}

var Pulse = &Message{code: 0, time: time.Now(), source: Index, target: Index, Meta: map[string][]string{}}
var Index = &Context{Name: "ctx", Help: "模块中心", Server: &CTX{},
	Caches: map[string]*Cache{
		"begin_time": &Cache{Name: "begin_time", Value: "", Help: "启动时间"},
		"goos":       &Cache{Name: "goos", Value: "linux", Help: "启动时间"},
		"ngo":        &Cache{Name: "ngo", Value: "0", Help: "启动时间"},
		"nserver":    &Cache{Name: "nserver", Value: "0", Help: "服务数量"},
		"ncontext":   &Cache{Name: "ncontext", Value: "0", Help: "模块数量"},
		"nmessage":   &Cache{Name: "nmessage", Value: "1", Help: "消息数量"},
	},
	Configs: map[string]*Config{
		"chain":       &Config{Name: "chain", Value: map[string]interface{}{}, Help: "调试模式，on:打印，off:不打印)"},
		"compact_log": &Config{Name: "compact_log(true/false)", Value: "true", Help: "调试模式，on:打印，off:不打印)"},
		"auto_make":   &Config{Name: "auto_make(true/false)", Value: "true", Help: "调试模式，on:打印，off:不打印)"},
		"debug":       &Config{Name: "debug(on/off)", Value: "on", Help: "调试模式，on:打印，off:不打印)"},

		"search_method": &Config{Name: "search_method(find/search)", Value: "search", Help: "搜索方法, find: 模块名精确匹配, search: 模块名或帮助信息模糊匹配"},
		"search_choice": &Config{Name: "search_choice(first/last/rand/magics)", Value: "magics", Help: "搜索匹配, first: 匹配第一个模块, last: 匹配最后一个模块, rand: 随机选择, magics: 加权选择"},
		"search_action": &Config{Name: "search_action(list/switch)", Value: "switch", Help: "搜索操作, list: 输出模块列表, switch: 模块切换"},
		"search_root":   &Config{Name: "search_root(true/false)", Value: "true", Help: "搜索起点, true: 根模块, false: 当前模块"},

		"insert_limit": &Config{Name: "insert_limit(true/false)", Value: "true", Help: "参数的索引"},
		"detail_index": &Config{Name: "detail_index", Value: "0", Help: "参数的索引"},
		"result_index": &Config{Name: "result_index", Value: "-2", Help: "返回值的索引"},

		"list_help":     &Config{Name: "list_help", Value: "list command", Help: "命令列表帮助"},
		"table_compact": &Config{Name: "table_compact", Value: "false", Help: "命令列表帮助"},
		"table_col_sep": &Config{Name: "table_col_sep", Value: "  ", Help: "命令列表帮助"},
		"table_row_sep": &Config{Name: "table_row_sep", Value: "\n", Help: "命令列表帮助"},
		"table_space":   &Config{Name: "table_space", Value: " ", Help: "命令列表帮助"},

		"page_offset": &Config{Name: "page_offset", Value: "0", Help: "列表偏移"},
		"page_limit":  &Config{Name: "page_limit", Value: "10", Help: "列表大小"},

		"time_format":  &Config{Name: "time_format", Value: "2006-01-02 15:04:05", Help: "时间格式"},
		"call_timeout": &Config{Name: "call_timeout", Value: "10s", Help: "回调超时"},
	},
	Commands: map[string]*Command{
		"init": &Command{Name: "init", Help: "启动", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			for _, x := range []string{"cli", "yac", "nfs", "aaa", "log", "ssh", "web", "gdb"} {
				m.Cmd(x + ".init")
			}
			return
		}},
		"help": &Command{Name: "help topic", Help: "帮助", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Echo("usage: help context [module [command|config|cache name]]\n")
				m.Echo("     : 查看模块信息, module: 模块名, command: 模块命令, config: 模块配置, cache: 模块缓存, name: 模块参数\n")
				m.Echo("usage: help command [name]\n")
				m.Echo("     : 查看当前环境下命令, name: 命令名\n")
				m.Echo("usage: help config [name]\n")
				m.Echo("     : 查看当前环境下配置, name: 配置名\n")
				m.Echo("usage: help cache [name]\n")
				m.Echo("     : 查看当前环境下缓存, name: 缓存名\n")
				m.Echo("\n")

				m.Echo("^_^  Welcome to context world  ^_^\n")
				m.Echo("Version: 1.0 A New Language, A New Framework\n")
				m.Echo("More: https://github.com/shylinux/context\n")
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
							for k, v := range v.Form {
								m.Add("append", "arg", k)
								m.Add("append", "len", v)
								m.Echo("  option: %s(%d)\n", k, v)
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

		"message": &Command{Name: "message [code] [cmd...]", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			msg := m
			if ms := m.Find(m.Cap("ps_target")); ms != nil {
				msg = ms
			}

			if len(arg) > 0 {
				if code, e := strconv.Atoi(arg[0]); e == nil {
					if msg = m.root.Tree(code); msg != nil {
						arg = arg[1:]
					}
				}
			}

			if len(arg) == 0 {
				m.Format("summary", msg, "deep")
				msg.CopyTo(m)
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
				msg = msg.Spawn().Cmd(arg)
				m.Copy(msg, "append").Copy(msg, "result")
			}
			return
		}},
		"detail": &Command{Name: "detail [index] [value...]", Help: "查看或添加参数", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			msg := m.message
			if len(arg) == 0 {
				for i, v := range msg.Meta["detail"] {
					m.Add("append", "index", i)
					m.Add("append", "value", v)
				}
				m.Table()
				return
			}

			index := m.Confi("detail_index")
			if i, e := strconv.Atoi(arg[0]); e == nil {
				index, arg = i, arg[1:]
			}
			m.Echo("%s", msg.Detail(index, arg))
			return
		}},
		"option": &Command{Name: "option [all] [key [index] [value...]]", Help: "查看或添加选项", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			all := false
			if len(arg) > 0 && arg[0] == "all" {
				all, arg = true, arg[1:]
			}

			index := -100
			if len(arg) > 1 {
				if i, e := strconv.Atoi(arg[1]); e == nil {
					index = i
					for i := 1; i < len(arg)-1; i++ {
						arg[i] = arg[i+1]
					}
					arg = arg[:len(arg)-1]
				}
			}

			msg := m.message
			for msg = msg; msg != nil; msg = msg.message {
				for _, k := range msg.Meta["option"] {
					if len(arg) == 0 {
						m.Add("append", "key", k)
						m.Add("append", "len", len(msg.Meta[k]))
						m.Add("append", "value", fmt.Sprintf("%v", msg.Meta[k]))
						continue
					}

					if k != arg[0] {
						continue
					}

					if len(arg) > 1 {
						msg.Meta[k] = kit.Array(msg.Meta[k], index, arg[1:])
						m.Echo("%v", msg.Meta[k])
						return
					}

					if index != -100 {
						m.Echo(kit.Array(msg.Meta[k], index)[0])
						return
					}

					for i, v := range msg.Meta[k] {
						m.Add("append", "index", i)
						m.Add("append", "value", v)
					}
					m.Table()
					return
				}

				if !all {
					break
				}
			}
			m.Sort("key", "string").Table()
			return
		}},
		"magic": &Command{Name: "magic", Help: "随机组员", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			switch len(arg) {
			case 0:
				m.Optionv("magic", m.Magic("bench", ""))
			case 1:
				m.Optionv("magic", m.Magic(arg[0], ""))
			case 2:
				m.Optionv("magic", m.Magic(arg[0], arg[1]))
			case 3:
				m.Optionv("magic", m.Magic(arg[0], arg[1], arg[2]))
			}
			m.Cmdy("ctx.trans", "magic")
			return
		}},
		"result": &Command{Name: "result [index] [value...]", Help: "查看或添加返回值", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			msg := m.message
			if len(arg) == 0 {
				for i, v := range msg.Meta["result"] {
					m.Add("append", "index", i)
					m.Add("append", "value", strings.Replace(v, "\n", "\\n", -1))
				}
				m.Table()
				return
			}

			index := m.Confi("result_index")
			if i, e := strconv.Atoi(arg[0]); e == nil {
				index, arg = i, arg[1:]
			}
			m.Echo("%s", msg.Result(index, arg))
			return
		}},
		"append": &Command{Name: "append [all] [key [index] [value...]]", Help: "查看或添加附加值", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			all := false
			if len(arg) > 0 && arg[0] == "all" {
				all, arg = true, arg[1:]
			}

			index := -100
			if len(arg) > 1 {
				if i, e := strconv.Atoi(arg[1]); e == nil {
					index = i
					for i := 1; i < len(arg)-1; i++ {
						arg[i] = arg[i+1]
					}
					arg = arg[:len(arg)-1]
				}
			}

			msg := m.message
			for msg = msg; msg != nil; msg = msg.message {
				for _, k := range msg.Meta["append"] {
					if len(arg) == 0 {
						m.Add("append", "key", k)
						m.Add("append", "value", fmt.Sprintf("%v", msg.Meta[k]))
						continue
					}

					if k != arg[0] {
						continue
					}

					if len(arg) > 1 {
						msg.Meta[k] = kit.Array(msg.Meta[k], index, arg[1:])
						m.Echo("%v", msg.Meta[k])
						return
					}

					if index != -100 {
						m.Echo(kit.Array(msg.Meta[k], index)[0])
						return
					}

					for i, v := range msg.Meta[k] {
						m.Add("append", "index", i)
						m.Add("append", "value", v)
					}
					m.Table()
					return
				}

				if !all {
					break
				}
			}
			m.Table()
			return
		}},
		"session": &Command{Name: "session [all] [key [module]]", Help: "查看或添加会话", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			all := false
			if len(arg) > 0 && arg[0] == "all" {
				all, arg = true, arg[1:]
			}

			msg := m.message
			for msg = msg; msg != nil; msg = msg.message {
				for k, v := range msg.Sessions {
					if len(arg) > 1 {
						msg.Sessions[arg[0]] = msg.Find(arg[1])
						return
					} else if len(arg) > 0 {
						if k == arg[0] {
							m.Echo("%d", v.code)
							return
						}
						continue
					}

					m.Add("append", "key", k)
					m.Add("append", "time", v.time.Format("15:04:05"))
					m.Add("append", "code", v.code)
					m.Add("append", "source", v.source.Name)
					m.Add("append", "target", v.target.Name)
					m.Add("append", "details", fmt.Sprintf("%v", v.Meta["detail"]))
					m.Add("append", "options", fmt.Sprintf("%v", v.Meta["option"]))
				}

				if len(arg) == 0 && !all {
					break
				}
			}
			m.Table()
			return
		}},
		"callback": &Command{Name: "callback", Help: "查看消息", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
			msg := m.message
			for msg := msg; msg != nil; msg = msg.message {
				m.Add("append", "msg", msg.code)
				m.Add("append", "fun", msg.callback)
			}
			m.Table()
			return
		}},

		"context": &Command{Name: "context [find|search] [root|back|home] [first|last|rand|magic] [module] [cmd|switch|list|spawn|start|close]",
			Help: "查找并操作模块;\n查找方法, find: 精确查找, search: 模糊搜索;\n查找起点, root: 根模块, back: 父模块, home: 本模块;\n过滤结果, first: 取第一个, last: 取最后一个, rand: 随机选择, magics: 智能选择;\n操作方法, cmd: 执行命令, switch: 切换为当前, list: 查看所有子模块, spwan: 创建子模块并初始化, start: 启动模块, close: 结束模块",
			Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
				if len(arg) == 1 && arg[0] == "~" && m.target.context != nil {
					m.target = m.target.context
					return
				}

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
						componet := "source"
						if m.Options("bench") && m.Options("username") &&
							!m.Cmds("aaa.work", m.Option("bench"), "right", m.Option("username"), componet, arg[0]) {
							m.Log("info", "check %v: %v failure", componet, arg[0])
							m.Echo("error: ").Echo("no right [%s: %s]", componet, arg[0])
							break
						}

						if msg.Cmd(arg); !msg.Hand {
							msg = msg.Sess("cli").Cmd("cmd", arg)
						}
						msg.CopyTo(m)

					case "switch":
						m.target = msg.target

					case "list":
						cs := []*Context{}
						if msg.target.Name != "ctx" {
							cs = append(cs, msg.target.context)
						}
						msg.Target().Travel(msg, func(msg *Message, n int) bool {
							cs = append(cs, msg.target)
							return false
						})

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
							m.Add("append", "msg", msg.target.message.code)
							m.Add("append", "status", msg.Cap("status"))
							m.Add("append", "stream", msg.Cap("stream"))
							m.Add("append", "helps", msg.target.Help)
						}

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

				if action == "list" {
					m.Table()
				}
				return
			}},
		"command": &Command{Name: "command [all] [show]|[list [begin [end]] [prefix] test [key val]...]|[add [list_name name] [list_help help] cmd...]|[delete cmd]",
			Help: "查看或修改命令, show: 查看命令;\nlist: 查看列表命令, begin: 起始索引, end: 截止索引, prefix: 过滤前缀, test: 执行命令;\nadd: 添加命令, list_name: 命令别名, list_help: 命令帮助;\ndelete: 删除命令",
			Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
				all := false
				if len(arg) > 0 && arg[0] == "all" {
					all, arg = true, arg[1:]
				}

				action := "show"
				if len(arg) > 0 {
					switch arg[0] {
					case "show", "list", "add", "delete":
						action, arg = arg[0], arg[1:]
					}
				}

				switch action {
				case "show":
					c.BackTrace(m, func(m *Message) bool {
						for k, v := range m.target.Commands {
							if len(arg) > 0 {
								if k == arg[0] {
									m.Add("append", "key", k)
									m.Add("append", "name", v.Name)
									m.Add("append", "help", v.Name)
								}
							} else {
								m.Add("append", "key", k)
								m.Add("append", "name", v.Name)
							}
						}

						return !all
					})
					m.Table()
				case "list":
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
								m.Add("append", "help", fmt.Sprintf("%s", c.Help))
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
				case "delete":
					c.BackTrace(m, func(m *Message) bool {
						delete(m.target.Commands, arg[0])
						return !all
					})
				}
				return
			}},
		"config": &Command{Name: "config [all] [export key..] [save|load file key...] [list|map arg...] [create map|list|string key name help] [delete key]",
			Help: "配置管理, export: 导出配置, save: 保存配置到文件, load: 从文件加载配置, create: 创建配置, delete: 删除配置",
			Form: map[string]int{"format": 1, "fields": -1},
			Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
				if len(arg) > 2 && arg[2] == "list" {
					chain := strings.Split(arg[1], ".")
					chain = append(chain, "-2")

					for _, val := range arg[3:] {
						m.Confv(arg[0], chain, val)
					}
					return
				}
				if len(arg) > 2 && arg[2] == "map" {
					chain := strings.Split(arg[1], ".")

					for i := 3; i < len(arg)-1; i += 2 {
						m.Confv(arg[0], append(chain, arg[i]), arg[i+1])
					}
					return
				}

				all := false
				if len(arg) > 0 && arg[0] == "all" {
					arg, all = arg[1:], true
				}

				action, which := "", "-1"
				have := map[string]bool{}
				if len(arg) > 0 {
					switch arg[0] {
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
					}
				}

				if len(arg) == 0 || action != "" {
					save := map[string]interface{}{}
					if action == "load" {
						f, e := os.Open(m.Sess("nfs").Cmd("path", which).Result(0))
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
									delete(m.target.Configs, k)
								}
								fallthrough
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
				}

				var value interface{}
				if len(arg) > 2 {
					value = m.Confv(arg[0], arg[1], arg[2])
				} else if len(arg) > 1 {
					value = m.Confv(arg[0], arg[1])
				} else {
					value = m.Confv(arg[0])
				}

				msg := m.Spawn().Put("option", "_cache", value).Cmd("trans", "_cache")
				m.Copy(msg, "append").Copy(msg, "result")
				return
			}},
		"cache": &Command{Name: "cache [all] |key [value]|key = value|key name value help|delete key]",
			Help: "查看、读写、赋值、新建、删除缓存变量",
			Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
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

				limit := kit.Int(kit.Select(m.Conf("page_limit"), m.Option("limit")))
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
							for _, v := range val {
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

						m.Log("fuck", "what %v", fields)
						m.Log("fuck", "what %v", m.Meta)
						for k, v := range val {
							if line, ok := v.(map[string]interface{}); ok {
								m.Add("append", "key", k)
								for _, field := range fields {
									m.Add("append", field, kit.Format(line[field]))
								}
							}
						}
						m.Log("fuck", "what %v", m.Meta)
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
		"select": &Command{Name: "select key value field",
			Form: map[string]int{"eq": 2, "parse": 2, "hide": -1, "fields": -1, "group": 1, "order": 2, "limit": 1, "offset": 1, "format": -1, "trans_map": -1, "vertical": 0},
			Help: "选取数据", Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
				msg := m.Set("result").Spawn()

				// 解析
				if len(m.Meta["append"]) == 0 {
					return
				}
				nrow := len(m.Meta[m.Meta["append"][0]])
				keys := []string{}
				for i := 0; i < nrow; i++ {
					for j := 0; j < len(m.Meta["parse"]); j += 2 {
						var value interface{}
						json.Unmarshal([]byte(m.Meta[m.Meta["parse"][j]][i]), &value)
						if m.Meta["parse"][j+1] != "" {
							value = kit.Chain(value, m.Meta["parse"][j+1])
						}

						switch val := value.(type) {
						case map[string]interface{}:
							for k, _ := range val {
								keys = append(keys, k)
							}
						default:
							keys = append(keys, m.Meta["parse"][j+1])
						}
					}
				}
				for i := 0; i < nrow; i++ {
					for _, k := range keys {
						m.Add("append", k, "")
					}
				}
				for i := 0; i < nrow; i++ {
					for j := 0; j < len(m.Meta["parse"]); j += 2 {
						var value interface{}
						json.Unmarshal([]byte(m.Meta[m.Meta["parse"][j]][i]), &value)
						if m.Meta["parse"][j+1] != "" {
							value = kit.Chain(value, m.Meta["parse"][j+1])
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
							m.Meta[m.Meta["parse"][j+1]][i] = val
						case float64:
							m.Meta[m.Meta["parse"][j+1]][i] = fmt.Sprintf("%d", int(val))
						default:
							b, _ := json.Marshal(val)
							m.Meta[m.Meta["parse"][j+1]][i] = string(b)
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
					for _, k := range arg {
						if hides[k] {
							continue
						}
						msg.Add("append", k, m.Meta[k][i])
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
				limit := kit.Int(kit.Select(m.Conf("page_limit"), m.Option("limit")))

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
						m.Meta[format[i]][j] = fmt.Sprintf(format[i+1], v)
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
	if len(args) == 0 {
		args = append(args, os.Args[1:]...)
	}
	if len(args) > 0 && args[0] == "daemon" {
		Pulse.Options("daemon", true)
		args = args[1:]
	}

	if Index.Begin(Pulse, args...); Index.Start(Pulse, args...) {
		return Index.Close(Pulse, args...)
	}

	return Index.message.Wait()
}
