package ssh

import (
	"bufio"
	"contexts/ctx"
	"fmt"
	"os/exec"
	"sort"
	"time"
	"toolkit"

	"encoding/hex"
	"io"
	"os"
	"path"
	"strings"
)

type SSH struct {
	relay chan *ctx.Message
	*ctx.Context
}

func (ssh *SSH) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	return &SSH{Context: c}
}
func (ssh *SSH) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return ssh
}
func (ssh *SSH) Start(m *ctx.Message, arg ...string) bool {
	ir, iw := io.Pipe()
	or, ow := io.Pipe()
	er, ew := io.Pipe()
	cmd := exec.Command("ssh", arg[0])
	cmd.Stdin, cmd.Stdout, cmd.Stderr = ir, ow, ew
	cmd.Start()

	relay := m
	done := false
	ssh.relay = make(chan *ctx.Message, 10)
	m.Gos(m.Spawn(), func(msg *ctx.Message) {
		for relay = range ssh.relay {
			done = false
			for _, v := range relay.Meta["detail"][1:] {
				msg.Log("info", "%v", v)
				fmt.Fprint(iw, v, " ")
			}
			fmt.Fprintln(iw)

			ticker, delay := kit.Duration(msg.Conf("ssh.login", "ticker")), kit.Duration(msg.Conf("ssh.login", "ticker"))
			for i := 0; i < msg.Confi("ssh.login", "count"); i++ {
				msg.Log("done", "%d %v", i, done)
				if time.Sleep(ticker); done {
					time.Sleep(delay)
					relay.Back(relay)
					break
				}
			}
		}
	})
	m.Gos(m.Spawn(), func(msg *ctx.Message) {
		for bio := bufio.NewScanner(er); bio.Scan(); {
			msg.Log("warn", "what %v", bio.Text())
			relay.Echo(bio.Text()).Echo("\n")
			done = true
		}
	})
	for bio := bufio.NewScanner(or); bio.Scan(); {
		m.Log("info", "what %v", bio.Text())
		relay.Echo(bio.Text()).Echo("\n")
		done = true
	}
	return false
}
func (ssh *SSH) Close(m *ctx.Message, arg ...string) bool {
	return false
}

var Index = &ctx.Context{Name: "ssh", Help: "集群中心",
	Caches: map[string]*ctx.Cache{
		"nnode": {Name: "nnode", Value: "0", Help: "节点数量"},
	},
	Configs: map[string]*ctx.Config{
		"componet": {Name: "componet", Value: map[string]interface{}{
			"index": []interface{}{
				map[string]interface{}{"name": "ifconfig", "help": "网卡",
					"tmpl": "componet", "view": "", "init": "",
					"type": "protected", "ctx": "ssh", "cmd": "_route",
					"args": []interface{}{"_", "tcp.ifconfig"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "pod", "value": "", "imports": "plugin_pod"},
						map[string]interface{}{"type": "button", "value": "查看", "action": "auto"},
					},
				},
				map[string]interface{}{"name": "proc", "help": "进程",
					"tmpl": "componet", "view": "", "init": "",
					"type": "protected", "ctx": "ssh", "cmd": "_route",
					"args": []interface{}{"_", "cli.proc"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "pod", "value": "", "imports": "plugin_pod"},
						map[string]interface{}{"type": "text", "name": "arg", "value": ""},
						map[string]interface{}{"type": "text", "name": "filter", "view": "long"},
						map[string]interface{}{"type": "button", "value": "执行"},
					},
				},
				map[string]interface{}{"name": "spide", "help": "爬虫",
					"tmpl": "componet", "view": "Context", "init": "",
					"type": "protected", "ctx": "ssh", "cmd": "_route",
					"args": []interface{}{"_", "context", "web", "spide"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "pod", "imports": "plugin_pod"},
						map[string]interface{}{"type": "button", "value": "执行"},
					},
					"exports": []interface{}{"site", "key"},
				},
				map[string]interface{}{"name": "post", "help": "请求",
					"tmpl": "componet", "view": "Context", "init": "",
					"type": "protected", "ctx": "ssh", "cmd": "_route",
					"args": []interface{}{"_", "web.post", "__", "content_type", "application/json", "parse", "json"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "pod", "imports": "plugin_pod"},
						map[string]interface{}{"type": "text", "name": "spide", "value": "dev", "imports": "plugin_site"},
						map[string]interface{}{"type": "text", "name": "url", "value": "/", "view": "long"},
						map[string]interface{}{"type": "button", "value": "执行"},
					},
				},
				map[string]interface{}{"name": "get", "help": "请求",
					"tmpl": "componet", "view": "Context", "init": "",
					"type": "protected", "ctx": "ssh", "cmd": "_route",
					"args": []interface{}{"_", "web.get", "__", "method", "GET", "parse", "json"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "pod", "imports": "plugin_pod"},
						map[string]interface{}{"type": "text", "name": "spide", "value": "dev", "imports": "plugin_site"},
						map[string]interface{}{"type": "text", "name": "url", "value": "/", "view": "long"},
						map[string]interface{}{"type": "button", "value": "执行"},
					},
				},
			},
		}, Help: "组件列表"},

		"data":  {Name: "data", Value: map[string]interface{}{"path": "var/data"}, Help: "聊天数据"},
		"flow":  {Name: "flow", Value: map[string]interface{}{}, Help: "聊天群组"},
		"work":  {Name: "work", Value: map[string]interface{}{}, Help: "工作信息"},
		"node":  {Name: "node", Value: map[string]interface{}{}, Help: "节点信息"},
		"timer": {Name: "timer", Value: map[string]interface{}{"interval": "10s", "timer": ""}, Help: "断线重连"},
		"trust": {Name: "trust", Value: map[string]interface{}{
			"renew": true, "fresh": false, "user": true, "up": true,
		}, Help: "可信节点"},
		"login": {Name: "login", Value: map[string]interface{}{
			"ticker": "10ms",
			"count":  100,
			"delay":  "10ms",
		}, Help: "聊天群组"},
	},
	Commands: map[string]*ctx.Command{
		"_init": {Name: "_init", Help: "启动初始化", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if m.Confs("runtime", "boot.ctx_box") {
				m.Conf("runtime", "node.type", "worker")
				m.Conf("runtime", "node.name", m.Conf("runtime", "boot.pathname"))
			} else {
				m.Conf("runtime", "node.type", "server")
				m.Conf("runtime", "node.name", kit.Key(strings.TrimSuffix(m.Conf("runtime", "boot.hostname"), ".local")))
			}
			m.Conf("runtime", "node.route", m.Conf("runtime", "node.name"))
			m.Conf("runtime", "user.name", m.Conf("runtime", "boot.username"))

			m.Cmd("aaa.role", "tech", "componet", "remote", "command", "tool")
			m.Cmd("aaa.role", "tech", "componet", "source", "command", "tool")
			return
		}},
		"_node": {Name: "_node [init|create name type module|delete name]", Help: "节点", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Confm("ssh.node", func(key string, value map[string]interface{}) {
					m.Push("create_time", value["create_time"])
					m.Push("node", key)
					m.Push("type", value["type"])
				})
				m.Sort("node").Table()
				return
			}

			switch arg[0] {
			// 节点证书
			case "init":
				if !m.Confs("runtime", "node.cert") || !m.Confs("runtime", "node.key") {
					msg := m.Cmd("aaa.rsa", "gen")
					m.Conf("runtime", "node.cert", msg.Append("certificate"))
					m.Conf("runtime", "node.key", msg.Append("private"))
					m.Echo(m.Conf("runtime", "node.cert"))
				}

			// 创建节点
			case "create":
				name := arg[1]
				if arg[2] != "master" {
					for node := m.Confm("node", name); node != nil; node = m.Confm("node", name) {
						name = kit.Format("%s_%s", arg[1], m.Capi("nnode", 1))
					}
				}

				m.Log("info", "create node %s %s", name, arg[2])
				m.Confv("node", name, map[string]interface{}{
					"create_time": m.Time(), "name": name, "type": arg[2], "module": arg[3],
				})
				m.Echo(name)

			// 删除节点
			case "delete":
				m.Log("info", "delete node %s %s", arg[1], kit.Formats(m.Conf("node", arg[1])))
				m.Cmd("aaa.auth", m.Cmdx("aaa.auth", "nodes", arg[1]), "delete", "node")
				delete(m.Confm("node"), arg[1])
			}
			return
		}},
		"user": {Name: "user [init|create|trust [node]]", Help: "用户管理, init: 用户节点, create: 用户创建, trust: 用户授权", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Echo(m.Conf("runtime", "user.route"))
				return
			}
			switch arg[0] {
			// 用户证书
			case "init":
				if m.Confs("runtime", "user.cert") && m.Confs("runtime", "user.key") {
					break
				}
				fallthrough
			// 创建用户
			case "create":
				m.Cmd("aaa.auth", "username", m.Conf("runtime", "user.name"), "delete", "node")

				if len(arg) == 1 { // 本地用户
					msg := m.Cmd("aaa.rsa", "gen")
					m.Conf("runtime", "user.route", m.Conf("runtime", "node.route"))
					m.Conf("runtime", "user.name", m.Conf("runtime", "boot.username"))
					m.Conf("runtime", "user.cert", msg.Append("certificate"))
					m.Conf("runtime", "user.key", msg.Append("private"))

				} else { // 远程用户
					msg := m.Cmd("ssh._route", arg[1], "_check", "user")
					m.Conf("runtime", "user.route", msg.Append("user.route"))
					m.Conf("runtime", "user.name", msg.Append("user.name"))
					m.Conf("runtime", "user.cert", msg.Append("user.cert"))
					m.Conf("runtime", "user.key", "")
				}
				m.Cmd("aaa.user", "root", m.Conf("runtime", "user.name"))
				m.Echo(m.Conf("runtime", "user.cert"))

			case "trust":
				if len(arg) > 1 {
					m.Conf("trust", arg[1], kit.Right(kit.Select("true", arg, 2)))
				}
				if len(arg) > 1 {
					m.Cmdy("ctx.config", "trust", arg[1])
				} else {
					m.Cmdy("ctx.config", "trust")
				}
			}
			m.Append("user.route", m.Conf("runtime", "user.route"))
			return
		}},
		"work": {Name: "work [serve|create|search]", Help: "工作管理, serve: 创建组织, create: 创建员工, search: 搜索员工", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Confm("work", func(key string, value map[string]interface{}) {
					m.Add("append", "key", key)
					m.Add("append", "user", value["user"])
				})
				m.Table()
				return
			}

			switch arg[0] {
			// 创建组织
			case "serve":
				m.Conf("runtime", "work.serve", true)
				m.Conf("runtime", "work.route", m.Conf("runtime", "node.route"))
				m.Conf("runtime", "work.name", m.Conf("runtime", "user.name"))
				m.Conf("work", m.Conf("runtime", "user.name"), map[string]interface{}{
					"create_time": m.Time(), "user": m.Cmd("ssh.user", "init").Append("user.route"),
				})

			// 创建员工
			case "create":
				m.Cmd("ssh.user", "init")
				name := kit.Select(m.Conf("runtime", "user.name"), arg, 1)
				work := kit.Select(m.Conf("runtime", "work.route"), arg, 2)

				if n := m.Cmdx("ssh._route", work, "_check", "work", "create", name, m.Conf("runtime", "user.route")); n != "" {
					m.Conf("runtime", "work.route", work)
					m.Conf("runtime", "work.name", n)
					m.Echo(n)
				} else {
					m.Err("%s from %s", name, work)
				}

			// 共享用户
			case "share":
				name := kit.Select(m.Conf("runtime", "user.name"), arg, 1)
				work := kit.Select(m.Conf("runtime", "work.route"), arg, 2)

				if n := m.Cmdx("ssh._route", work, "_check", "work", name, m.Conf("runtime", "node.route")); n != "" {
					m.Echo(n)
				}

			case "search":
				m.Cmdy("ssh._route", m.Conf("runtime", "work.route"), "_check", "work", "search", arg[1:])
			}
			return
		}},
		"tool": {Name: "tool [group index][run group index chatid arg...]", Help: "工具", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Confm("componet", func(key string, index int, value map[string]interface{}) {
					if kit.Format(value["type"]) != "public" && m.Option("userrole") != "root" {
						return
					}

					m.Push("key", key)
					m.Push("index", index)
					m.Push("name", value["name"])
					m.Push("help", value["help"])
				})
				m.Table()
				return
			}

			switch arg[0] {
			case "run":
				m.Option("plugin", arg[1])
				tool := m.Confm("componet", []string{arg[1], arg[2]})
				if m.Option("userrole") != "root" {
					switch kit.Format(tool["type"]) {
					case "private":
						m.Echo("private componet of %s", m.Conf("runtime", "work.name"))
						return
					case "protected":
						if !m.Confs("flow", []string{arg[3], "user", m.Option("username")}) {
							m.Echo("protected componet of %s", m.Conf("runtime", "work.name"))
							return
						}
					}
				}

				prefix := []string{}
				msg := m.Find(kit.Format(tool["ctx"]))
				if strings.Contains(kit.Format(tool["ctx"]), ":") {
					ps := strings.Split(kit.Format(tool["ctx"]), ":")
					prefix = append(prefix, "_route", ps[0], "context", "find", ps[1])
					msg = m.Sess("ssh")
				}

				if option, ok := tool["options"].(map[string]interface{}); ok {
					for k, v := range option {
						msg.Option(k, v)
					}
				}

				msg.Option("river", arg[3])
				msg.Option("storm", arg[1])
				kit.Map(tool["inputs"], "", func(index int, value map[string]interface{}) {
					if name := kit.Format(value["name"]); name != "" && m.Option(name) != "" {
						m.Log("info", "%v: %v", name, m.Option(name))
						msg.Option(name, m.Option(name))
					}
				})

				arg = arg[4:]
				args := []string{}
				for _, v := range kit.Trans(tool["args"]) {
					if strings.HasPrefix(v, "__") {
						if len(arg) > 0 {
							args, arg = append(args, arg...), nil
						}
					} else if strings.HasPrefix(v, "_") {
						if len(arg) > 0 {
							args, arg = append(args, arg[0]), arg[1:]
						} else {
							args = append(args, "")
						}
					} else {
						args = append(args, msg.Parse(v))
					}
				}

				if len(prefix) > 0 && prefix[1] == "_" {
					if len(args) > 0 {
						prefix[1], args = args[0], args[1:]
					} else if len(arg) > 0 {
						prefix[1], arg = arg[0], arg[1:]
					} else {
						prefix[1] = ""
					}
				}

				msg.Cmd(prefix, tool["cmd"], args, arg)
				msg.CopyTo(m)

			default:
				m.Confm("componet", arg[0:], func(value map[string]interface{}) {
					if kit.Format(value["type"]) == "private" && m.Option("userrole") != "root" {
						m.Log("warn", "%v private", arg)
						return
					}

					m.Push("name", value["name"])
					m.Push("help", value["help"])
					m.Push("init", value["init"])
					m.Push("view", value["view"])

					// if kit.Right(value["init"]) {
					// 	script := m.Cmdx("nfs.load", path.Join(m.Conf("cli.project", "plugin.path"), arg[0], kit.Format(value["init"])), -1)
					// 	if script == "" {
					// 		script = m.Cmdx("nfs.load", path.Join("usr/librarys/plugin", kit.Format(value["init"])), -1)
					// 	}
					// 	m.Push("init", script)
					// } else {
					// 	m.Push("init", "")
					// }
					// if kit.Right(value["view"]) {
					// 	script := m.Cmdx("nfs.load", path.Join(m.Conf("cli.project", "plugin.path"), arg[0], kit.Format(value["view"])), -1)
					// 	if script == "" {
					// 		script = m.Cmdx("nfs.load", path.Join("usr/librarys/plugin", kit.Format(value["view"])), -1)
					// 	}
					// 	m.Push("view", script)
					// } else {
					// 	m.Push("view", "")
					// }
					m.Push("inputs", kit.Format(value["inputs"]))
					m.Push("feature", kit.Format(value["feature"]))
					m.Push("exports", kit.Format(value["exports"]))
					m.Push("display", kit.Format(value["display"]))
				})
				m.Table()
			}
			return
		}},
		"data": {Name: "data show|insert|update [table [index] [key value]...]", Help: "数据", Form: map[string]int{
			"format": 1, "fields": -1,
		}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				arg = append(arg, "show")
			}

			if m.Conf("data", "local") == "single" {
				m.Confm("flow", func(key string, value map[string]interface{}) {
					m.Log("info", "river map %v->%v", m.Option("river"), key)
					m.Option("river", key)
				})
			}

			switch arg[0] {
			case "show":
				if len(arg) > 1 && arg[1] == "" {
					arg = arg[:1]
				}
				switch len(arg) {
				case 1: // 数据库
					m.Confm("flow", []string{m.Option("river"), "data"}, func(key string, value map[string]interface{}) {
						m.Push("create_time", kit.Chains(value, "meta.create_time"))
						m.Push("create_user", kit.Chains(value, "meta.create_user"))
						m.Push("table", key)
						m.Push("count", len(value["list"].([]interface{})))
					})

				case 3: // 记录行
					index := kit.Int(arg[2]) - 1 - m.Confi("flow", []string{m.Option("river"), "data", arg[1], "meta", "offset"})
					switch m.Option("format") {
					case "object":
						m.Confm("flow", []string{m.Option("river"), "data", arg[1], "list", kit.Format(index)}, func(key string, value string) {
							if key != "extra" {
								m.Push(key, value)
							}
						})
						m.Confm("flow", []string{m.Option("river"), "data", arg[1], "list", kit.Format(index), "extra"}, func(key string, value string) {
							m.Push("extra."+key, value)
						})
					default:
						m.Confm("flow", []string{m.Option("river"), "data", arg[1], "list", kit.Format(index)}, func(key string, value string) {
							if key != "extra" {
								m.Push("key", key)
								m.Push("value", value)
							}
						})
						m.Confm("flow", []string{m.Option("river"), "data", arg[1], "list", kit.Format(index), "extra"}, func(key string, value string) {
							m.Push("key", "extra."+key)
							m.Push("value", value)
						})
						m.Sort("key")
					}
				default: // 关系表
					m.Option("cache.limit", kit.Select("10", arg, 3))
					m.Option("cache.offend", kit.Select("0", arg, 4))
					m.Option("cache.match", kit.Select("", arg, 5))
					m.Option("cache.value", kit.Select("", arg, 6))

					keys := []string{}
					if m.Meta["append"] = []string{"id", "when"}; m.Has("fields") {
						keys = kit.Trans(m.Optionv("fields"))
					} else {
						// 字段查询
						hide := map[string]bool{"create_time": true, "update_time": true, "extra": true}
						m.Grows("flow", []string{m.Option("river"), "data", arg[1]}, func(meta map[string]interface{}, index int, value map[string]interface{}) {
							for k := range value {
								if !hide[k] {
									hide[k] = false
								}
							}
						})
						// 字段排序
						for k, hide := range hide {
							if !hide {
								keys = append(keys, k)
							}
						}
						sort.Strings(keys)
					}

					// 查询数据
					m.Grows("flow", []string{m.Option("river"), "data", arg[1]}, func(meta map[string]interface{}, index int, value map[string]interface{}) {
						for _, k := range keys {
							m.Push(k, kit.Format(value[k]))
						}
					})
				}
				m.Table()

			case "save":
				if len(arg) == 1 {
					m.Confm("flow", []string{m.Option("river"), "data"}, func(key string, value map[string]interface{}) {
						arg = append(arg, key)
					})
				}

				for _, v := range arg[1:] {
					data := m.Confm("flow", []string{m.Option("river"), "data", v})
					kit.Marshal(data["meta"], path.Join(m.Conf("ssh.data", "path"), m.Option("river"), v, "/meta.json"))
					kit.Marshal(data["list"], path.Join(m.Conf("ssh.data", "path"), m.Option("river"), v, "/list.csv"))

					l := len(data["list"].([]interface{}))
					m.Push("table", v).Push("count", l+kit.Int(kit.Chain(data["meta"], "offset")))
					m.Log("info", "save table %s:%s %d", m.Option("river"), v, l)
				}
				m.Table()

			case "create":
				if m.Confs("flow", []string{m.Option("river"), "data", arg[1]}) {
					break
				}

				tmpl := map[string]interface{}{"create_time": m.Time(), "update_time": m.Time(), "extra": "{}", "id": 0}
				for i := 2; i < len(arg)-1; i += 2 {
					tmpl[arg[i]] = arg[i+1]
				}

				m.Confv("flow", []string{m.Option("river"), "data", arg[1]}, map[string]interface{}{
					"meta": map[string]interface{}{
						"create_user": m.Option("username"),
						"create_time": m.Time(),
						"create_tmpl": tmpl,
						"store":       path.Join(m.Conf("ssh.data", "path"), m.Option("river"), arg[1], "/auto.csv"),
						"limit":       "30",
						"least":       "10",
					},
					"list": []interface{}{},
				})
				m.Log("info", "create table %s:%s", m.Option("river"), arg[1])

			case "insert":
				if len(arg) == 2 {
					m.Cmd("ssh.data", "show", arg[1])
					break
				}
				if !m.Confs("flow", []string{m.Option("river"), "data", arg[1]}) {
					m.Cmd("ssh.data", "create", arg[1:])
				}

				id := m.Confi("flow", []string{m.Option("river"), "data", arg[1], "meta", "count"}) + 1
				m.Confi("flow", []string{m.Option("river"), "data", arg[1], "meta", "count"}, id)

				data := map[string]interface{}{}
				extra := map[string]interface{}{}
				tmpl := m.Confm("flow", []string{m.Option("river"), "data", arg[1], "meta", "create_tmpl"})
				for k, v := range tmpl {
					data[k] = v
				}
				for i := 2; i < len(arg)-1; i += 2 {
					if _, ok := tmpl[arg[i]]; ok {
						data[arg[i]] = arg[i+1]
					} else {
						extra[arg[i]] = arg[i+1]
					}
				}
				data["create_time"] = m.Time()
				data["update_time"] = m.Time()
				data["extra"] = extra
				data["id"] = id

				m.Log("info", "insert %s:%s %s", m.Option("river"), arg[1], kit.Format(data))
				m.Grow("flow", []string{m.Option("river"), "data", arg[1]}, data)
				m.Cmdy("ssh.data", "save", arg[1])

			case "update":
				if arg[2] == "" {
					break
				}
				offset := m.Confi("flow", []string{m.Option("river"), "data", arg[1], "meta", "offset"})
				index := kit.Int(arg[2]) - 1 - offset
				table, prefix, arg := arg[1], "", arg[3:]
				if index >= 0 {
					if arg[0] == "extra" {
						prefix, arg = "extra.", arg[1:]
					}
					for i := 0; i < len(arg)-1; i += 2 {
						m.Confv("flow", []string{m.Option("river"), "data", table, "list", kit.Format(index), prefix + arg[i]}, arg[i+1])
					}
				}
				m.Cmdy("ssh.data", "show", table, index+1+offset)

			case "import":
				if len(arg) < 3 {
					m.Cmdy("ssh.data", "show", arg)
					return
				}

				id := m.Confi("flow", []string{m.Option("river"), "data", arg[1], "meta", "count"})
				m.Cmd("nfs.import", arg[2:]).Table(func(maps map[string]string) {
					data := map[string]interface{}{}
					for k, v := range maps {
						m.Push(k, v)
						data[k] = v
					}

					if id++; id == 1 {
						m.Cmd("ssh.data", "create", arg[1], data)
					}

					data["id"] = id
					data["extra"] = kit.UnMarshalm(maps["extra"])
					m.Confv("flow", []string{m.Option("river"), "data", arg[1], "list", "-2"}, data)
				})
				m.Confi("flow", []string{m.Option("river"), "data", arg[1], "meta", "count"}, id)
				m.Cmd("ssh.data", "save", arg[1])
			case "export":
				m.Append("directory", path.Join(m.Conf("ssh.data", "path"), m.Option("river"), arg[1], "/list.csv"))
				m.Table()
			}
			return
		}},

		"remote": {Name: "remote auto|dial|listen args...", Help: "连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Cmd("_node")
				return
			}

			switch arg[0] {
			// 自动连接
			case "auto":
				switch m.Conf("runtime", "boot.ctx_type") {
				case "work":
					m.Cmd("ssh.work", "serve")
				case "user":
					m.Cmd("ssh.work", "create")
				case "node":
				}

				if m.Confs("runtime", "boot.ctx_ups") {
					m.Cmd("ssh.remote", "dial", m.Conf("runtime", "boot.ctx_ups"))
					m.Cmd("ssh.remote", "listen", m.Conf("runtime", "boot.ssh_port"))
					m.Cmd("web.serve", "usr", m.Conf("runtime", "boot.web_port"))

				} else if m.Cmd("ssh.remote", "dial", "dev", "/shadow"); !m.Confs("runtime", "boot.ctx_box") {
					m.Cmd("ssh.remote", "listen", m.Conf("runtime", "boot.ssh_port"))
					m.Cmd("web.serve", "usr", m.Conf("runtime", "boot.web_port"))
				}

			// 监听连接
			case "listen":
				m.Cmd("ssh._node", "init")
				m.Call(func(nfs *ctx.Message) *ctx.Message {
					if nfs.Has("node.port") {
						m.Log("info", "node.port %v", nfs.Optionv("node.port"))
						m.Conf("runtime", "node.port", nfs.Optionv("node.port"))
					}
					return nil
				}, "nfs.socket", arg)

			// 添加节点
			case "_add":
				name := m.Cmdx("ssh._node", "create", arg[1], arg[2], m.Format("source"), arg[4:])
				m.Sess("ms_source", false).Free(func(msg *ctx.Message) bool {
					m.Cmd("ssh._node", "delete", name)
					return true
				})

				// 下刷信息
				m.Append("node.name", m.Conf("runtime", "node.name"))
				m.Append("user.name", m.Conf("runtime", "user.name"))
				m.Append("work.name", m.Conf("runtime", "work.name"))
				m.Append("work.route", m.Conf("runtime", "work.route"))
				m.Append("user.route", m.Conf("runtime", "user.route"))
				m.Append("node.route", m.Conf("runtime", "node.route"))
				m.Append("work.script", m.Cmdx("nfs.load", path.Join(m.Conf("cli.publish", "path"), kit.Select("hello", arg[3]), "local.shy")))
				m.Echo(name).Back(m)

			// 断线重连
			case "_redial":
				if !m.Caps("stream") {
					m.Cmdx("ssh.remote", "dial", arg[1:])
				}

			// 连接主机
			case "dial":
				m.Cmd("ssh._node", "init")
				m.Call(func(nfs *ctx.Message) *ctx.Message {
					if m.Caps("stream") {
						return nil
					}
					// 删除重连
					if m.Confs("timer", "timer") {
						m.Conf("timer", "timer", m.Cmdx("cli.timer", "delete", m.Conf("timer", "timer")))
					}

					// 注册设备
					m.Spawn(nfs.Target()).Call(func(node *ctx.Message) *ctx.Message {
						if m.Caps("stream") {
							return nil
						}
						// 添加网关
						name := m.Cmd("ssh._node", "create", node.Append("node.name"), "master", m.Cap("stream", nfs.Format("target")))

						// 清理回调
						nfs.Free(func(nfs *ctx.Message) bool {
							m.Cmd("ssh._node", "delete", name)

							// 断线重连
							m.Cap("stream", "")
							m.Conf("timer", "timer", m.Cmdx("cli.timer", "repeat", m.Conf("timer", "interval"), "context", "ssh", "remote", "_redial", arg[1:]))
							return true
						})

						// 节点路由
						m.Cmd("cli.runtime", "node.route", node.Append("node.route")+"."+node.Result(0))

						// 用户路由
						if m.Confs("runtime", "user.cert") && m.Confs("runtime", "user.key") {
							m.Cmd("cli.runtime", "user.route", m.Conf("runtime", "node.route"))

						} else if node.Appends("user.route") && !m.Confs("runtime", "user.route") {
							m.Cmd("ssh.user", "create", node.Append("user.route"))
						}

						// 工作路由
						if node.Appends("work.route") && !m.Confs("runtime", "work.route") {
							m.Cmd("cli.runtime", "work.route", node.Append("work.route"))
						}

						// 注册脚本
						m.Cmd("nfs.source", m.Cmdx("nfs.temp", m.Append("work.script")))
						return nil
					}, "send", "", "_add", m.Conf("runtime", "node.name"), m.Conf("runtime", "node.type"), m.Conf("runtime", "boot.ctx_type"))
					return nil
				}, "nfs.socket", arg)

			default:
				m.Cmd("_route", arg)
			}
			return
		}},
		"_route": {Name: "_route", Help: "路由", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				return
			}

			// 同步异步
			sync := true
			switch arg[0] {
			case "async", "sync":
				sync, arg = arg[0] == "sync", arg[1:]
			}

			// 局域路由
			old := arg[0]
			if arg[0] == m.Conf("runtime", "node.name") || arg[0] == m.Conf("runtime", "node.route") {
				arg[0] = ""
			} else {
				arg[0] = strings.TrimPrefix(arg[0], m.Conf("runtime", "node.route")+".")
				arg[0] = strings.TrimPrefix(arg[0], m.Conf("runtime", "node.name")+".")
			}

			// 拆分路由
			route, names, arg := arg[0], strings.SplitN(arg[0], ".", 2), arg[1:]
			if len(names) > 1 && names[0] == "" && names[1] != "" {
				names[0], names[1] = names[1], names[0]
			}

			if rest := kit.Select("", names, 1); names[0] != "" {
				// 数字签名
				if !m.Options("remote_code") && arg[0] != "_check" {
					for _, k := range []string{"river"} {
						m.Option(k, m.Option(k))
					}

					hash, meta := kit.Hash("rand",
						m.Option("text.time", m.Time("stamp")),
						m.Option("text.cmd", strings.Join(arg, " ")),
						m.Option("text.route", route),
						m.Option("text.username", m.Option("username")),
						m.Option("node.route", m.Conf("runtime", "node.route")),
						m.Option("user.name", m.Conf("runtime", "user.name")),
						m.Option("user.route", m.Conf("runtime", "user.route")),
						m.Option("work.name", m.Conf("runtime", "work.name")),
						m.Option("work.route", m.Conf("runtime", "work.route")),
					)
					m.Option("text.rand", meta[0])
					m.Option("node.sign", m.Cmdx("aaa.rsa", "sign", m.Conf("runtime", "node.key"), m.Option("text.hash", hash)))
				}

				// 查找路由
				ps := []string{}
				if names[0] == "%" || names[0] == "*" {
					// 广播命令
					m.Confm("node", names[0], func(name string, node map[string]interface{}) {
						if kit.Format(node["type"]) != "master" {
							ps = append(ps, kit.Format(node["module"]))
						} else {
							ps = append(ps, "")
						}
					})

				} else if m.Confm("node", names[0], func(node map[string]interface{}) {
					// 单播命令
					ps = append(ps, kit.Format(node["module"]))

				}) == nil && m.Caps("stream") {
					// 上报命令
					rest, ps = old, append(ps, m.Cap("stream"))
				}
				if len(ps) == 0 {
					// 发送前端
					if m.Option("userrole") == "root" {
						if !strings.Contains(old, ".") {
							old = m.Option("username") + "." + old
						}
						m.CallBack(true, func(msg *ctx.Message) *ctx.Message {
							m.Copy(msg, "append")
							return nil
						}, "web.wss", old, "sync", arg)
						m.Table()
						return
					}

					// 查找失败
					m.Echo("error: not found %s", names[0]).Back(m)
					return
				}

				// 路由转发
				for _, p := range ps {
					m.Find(p, true).Copy(m, "option").CallBack(sync, func(sub *ctx.Message) *ctx.Message {
						m.Log("time", "remote: %v", sub.Format("cost"))
						return m.CopyFuck(sub, "append").CopyFuck(sub, "result")
					}, "send", rest, arg)
				}
				return
			}

			defer func() { m.Back(m) }()

			if !m.Options("remote_code") {
				// 本地调用
				m.Cmdy(arg)

			} else if arg[0] == "_check" {
				// 公有命令
				m.Cmd(arg)

			} else if arg[0] == "_add" {
				// 公有命令
				m.Cmd("remote", arg)

			} else if h := kit.Hashs(
				m.Option("text.rand"),
				m.Option("text.time"),
				m.Option("text.cmd"),
				m.Option("text.route"),
				m.Option("text.username"),
				m.Option("node.route"),
				m.Option("user.name"),
				m.Option("user.route"),
				m.Option("work.name"),
				m.Option("work.route"),
			); h != m.Option("text.hash") {
				// 文本验签
				m.Echo("text error %s != %s", h, m.Option("text.hash"))
				m.Log("warn", "text error")

			} else if !m.Cmds("ssh._right", "node", "check", h) {
				// 设备验签
				m.Echo("node error of %s", m.Option("node.route"))

			} else if arg[0] == "tool" {
				m.Cmd("tool", arg[1:])
			} else {
				// 执行命令
				m.Log("time", "check: %v", m.Format("cost"))
				m.Cmd("_exec", arg)
				m.Log("time", "exec: %v", m.Format("cost"))
			}
			return
		}},
		"_right": {Name: "_right [node|user|work]", Help: []string{"认证",
			"node [check text|trust node]",
			"user [create|check text|share role node...|proxy node|trust node]",
			"work [create node name|check node name]",
			"file [import]",
			"tool ",
		}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Add("append", "key", "node.cert")
				m.Add("append", "route", m.Conf("runtime", "node.route"))
				m.Add("append", "value", m.Conf("runtime", "node.cert"))
				m.Add("append", "key", "user.cert")
				m.Add("append", "route", m.Conf("runtime", "user.route"))
				m.Add("append", "value", m.Conf("runtime", "user.cert"))
				m.Add("append", "key", "work.name")
				m.Add("append", "route", m.Conf("runtime", "work.route"))
				m.Add("append", "value", m.Conf("runtime", "work.name"))
				m.Table()
				return
			}

			switch arg[0] {
			// 节点认证
			case "node":
				if len(arg) == 1 {
					m.Echo(m.Conf("runtime", "node.cert"))
					break
				}

				switch arg[1] {
				// 数字验签
				case "check":
					if m.Option("node.cert", m.Cmd("aaa.auth", "nodes", m.Option("node.route"), "cert").Append("meta")); !m.Options("node.cert") {
						m.Option("node.cert", m.Cmdx("ssh._route", m.Option("node.route"), "_check", "node"))
						m.Cmd("aaa.auth", "nodes", m.Option("node.route"), "cert", m.Option("node.cert"))
					}

					if !m.Cmds("aaa.rsa", "verify", m.Option("node.cert"), m.Option("node.sign"), arg[2]) {
						if m.Log("warn", "node error"); m.Confs("trust", "renew") {
							m.Cmd("aaa.auth", "nodes", m.Option("node.route"), "node", "delete")
						}
						m.Echo("false")
					} else {
						m.Echo("true")
					}

				// 可信节点
				case "trust":
					if m.Confs("trust", arg[2]) {
						m.Echo("true")

					} else if m.Confs("trust", "user") && m.Conf("runtime", "user.route") == arg[2] {
						m.Echo("true")

					} else if m.Confs("trust", "up") && strings.HasPrefix(m.Conf("runtime", "node.route"), arg[2]) {
						m.Echo("true")
					}
				}

			// 用户认证
			case "user":
				if len(arg) == 1 {
					m.Echo(m.Conf("runtime", "user.cert"))
					break
				}

				switch arg[1] {
				// 数字验签
				case "check":
					if m.Option("user.cert", m.Cmd("aaa.auth", "username", m.Option("username"), "cert").Append("meta")); !m.Options("user.cert") {
						m.Option("user.cert", m.Cmd("ssh._route", m.Option("user.route"), "_check", "user").Append("user.cert"))
						m.Cmd("aaa.auth", "username", m.Option("username"), "cert", m.Option("user.cert"))
					}

					if !m.Options("user.cert") || !m.Cmds("aaa.rsa", "verify", m.Option("user.cert"), m.Option("user.sign"), arg[2]) {
						m.Log("warn", "user error")
						m.Echo("false")
					} else {
						m.Echo("true")
					}

				// 共享用户
				case "share":
					for _, route := range arg[3:] {
						user := m.Cmd("ssh._route", route, "_check", "user")
						if m.Cmd("aaa.role", arg[2], "user", user.Append("user.name"), "cert", user.Append("user.cert")); arg[2] == "root" && route != m.Conf("runtime", "node.route") {
							m.Conf("runtime", "user.route", user.Append("user.route"))
							m.Conf("runtime", "user.name", user.Append("user.name"))
							m.Conf("runtime", "user.cert", user.Append("user.cert"))
							m.Conf("runtime", "user.key", "")
						}
					}

				// 代理用户
				case "proxy":
					if len(arg) == 2 {
						m.Cmdy("aaa.auth", "proxy")
						break
					}
					if !m.Cmds("aaa.auth", "proxy", arg[2], "session") {
						m.Cmdy("aaa.sess", "proxy", "proxy", arg[2])
					}

				// 可信代理
				case "trust":
					hash := kit.Hashs("rand", m.Option("text.time", m.Time("stamp")), arg[2])
					m.Option("user.sign", m.Cmdx("ssh._route", m.Option("user.route"), "_check", "user", arg[2], hash))
					m.Echo("%v", m.Options("user.sign") && m.Cmds("ssh._right", "user", "check", hash))
				}

			// 公有认证
			case "work":
				if len(arg) == 1 {
					m.Echo(m.Conf("runtime", "work.name"))
					break
				}

				switch arg[1] {
				// 数字验签
				case "check":
					if arg[3] != m.Cmdx("ssh._route", kit.Select(m.Conf("runtime", "work.route"), m.Option("work.route")), "_check", "work", arg[2]) {
						m.Log("warn", "work error")
						m.Echo("false")
					} else {
						m.Echo("true")
					}
				}

			case "file":
				switch arg[1] {
				case "import":
					msg := m.Cmd("nfs.hash", arg[2])
					h := msg.Result(0)
					m.Conf("file", kit.Hashs(h, msg.Append("name")), map[string]interface{}{
						"create_time": m.Time(),
						"create_user": m.Option("username"),
						"name":        msg.Append("name"),
						"type":        msg.Append("type"),
						"size":        msg.Append("size"),
						"hash":        h,
					})

					m.Cmdy("nfs.copy", path.Join("var/file/hash", h[:2], h), arg[2])

				case "fetch":
					if m.Confs("file", arg[2]) {
						m.Echo(arg[2])
						break
					}

					msg := m.Cmd("ssh._route", arg[3], "_check", "file", arg[2])
					h := msg.Append("hash")
					m.Conf("file", arg[2], map[string]interface{}{
						"create_time": m.Time(),
						"create_user": m.Option("username"),
						"name":        msg.Append("name"),
						"type":        msg.Append("type"),
						"size":        msg.Append("size"),
						"hash":        h,
					})

					p := path.Join("var/file/hash", h[:2], h)
					if m.Cmds("nfs.path", p) {
						m.Echo(arg[2])
						break
					}
					m.Cmdy("nfs.copy", p)
					f, e := os.Create(p)
					m.Assert(e)
					for i := 0; int64(i) < msg.Appendi("size"); i += 1024 {
						msg := m.Cmd("ssh._route", arg[3], "_check", "file", arg[2], 1, 1024, i)
						for _, d := range msg.Meta["data"] {
							b, e := hex.DecodeString(d)
							m.Assert(e)
							_, e = f.Write(b)
							m.Assert(e)
						}
					}

				default:
					m.Confm("file", arg[1], func(file map[string]interface{}) {
						m.Append("hash", file["hash"])
						m.Append("size", file["size"])
						m.Append("type", file["type"])
						m.Append("name", file["name"])
					})
					m.Table()
				}
			}
			return
		}},
		"_check": {Name: "_check node|user|work", Help: []string{"验签",
			"node", "user [node text]", "work name [node cert]",
		}, Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			switch arg[0] {
			// 节点验签
			case "node":
				m.Echo(m.Conf("runtime", "node.cert"))

			// 用户验签
			case "user":
				if len(arg) == 1 {
					m.Append("user.cert", m.Conf("runtime", "user.cert"))
					m.Append("user.name", m.Conf("runtime", "user.name"))
					m.Append("user.route", kit.Select(m.Conf("runtime", "node.route"), m.Conf("runtime", "user.route")))
				} else {
					// 代理签证
					if m.Confs("trust", arg[1]) {
						m.Echo(m.Cmdx("aaa.rsa", "sign", m.Conf("runtime", "user.key"), arg[2]))
					}
				}

			// 工作验签
			case "work":
				switch arg[1] {
				case "search":
					m.Confm("work", func(key string, value map[string]interface{}) {
						m.Add("append", "key", key)
						m.Add("append", "user.route", value["user"])
						m.Add("append", "create_time", value["create_time"])
					})
					m.Table()

				case "create":
					name := arg[2]
					if len(arg) == 3 {
						arg = append(arg, m.Conf("runtime", "node.route"))
					}
					if user := m.Conf("work", []string{name, "user"}); user != "" && user != arg[3] {
						for i := 1; i < 100; i++ {
							name = fmt.Sprintf("%s%02d", arg[2], i)
							if user := m.Conf("work", []string{name, "user"}); user == "" || user == arg[3] {
								break
							}
							name = ""
						}
					}
					if name != "" {
						m.Conf("work", name, map[string]interface{}{"create_time": m.Time(), "user": arg[3]})
						m.Echo(name)
					}

				default:
					if cert := m.Confm("work", arg[1]); len(arg) == 2 {
						if cert != nil {
							m.Echo("%s", cert["user"])
						}
					} else {
						// 工作签证
						if cert == nil {
							m.Conf("work", arg[1], map[string]interface{}{"create_time": m.Time(), "user": arg[2]})
						} else if cert["user"] != arg[2] {
							// 签证失败
							return
						}
						m.Echo(arg[1])
					}
				}

			case "file":
				if len(arg) == 2 {
					m.Confm("file", arg[1], func(file map[string]interface{}) {
						m.Append("hash", file["hash"])
						m.Append("size", file["size"])
						m.Append("type", file["type"])
						m.Append("name", file["name"])
					})
					m.Table()
					break
				}

				h := m.Conf("file", []string{arg[1], "hash"})

				if f, e := os.Open(path.Join("var/file/hash", h[:2], h)); e == nil {
					defer f.Close()

					pos := kit.Int(kit.Select("0", arg, 4))
					size := kit.Int(kit.Select("1024", arg, 3))
					count := kit.Int(kit.Select("3", arg, 2))

					buf := make([]byte, count*size)

					if l, e := f.ReadAt(buf, int64(pos)); e == io.EOF || m.Assert(e) {
						for i := 0; i < count; i++ {
							if l < (i+1)*size {
								m.Add("append", "data", hex.EncodeToString(buf[i*size:l]))
								break
							}
							m.Add("append", "data", hex.EncodeToString(buf[i*size:(i+1)*size]))
						}
					}
				}
			}
			return
		}},
		"_exec": {Name: "_exec", Help: "命令", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			access := m.Option("access", kit.Hashs(
				m.Option("node.route"),
				m.Option("user.route"),
				m.Option("work.route"),
				m.Option("work.name"),
				m.Option("user.name"),
				m.Option("text.username"),
			))

			if m.Option("sessid", m.Cmd("aaa.auth", "access", access, "session").Append("key")) != "" &&
				m.Option("username", m.Cmd("aaa.sess", "user").Append("meta")) != "" {
				// 历史会话
				m.Log("warn", "access: %s", access)
				m.Log("info", "sessid: %s", m.Option("sessid"))
				m.Option("trust", m.Cmdx("aaa.auth", "access", access, "data", "trust"))
				m.Option("userrole", m.Cmdx("aaa.auth", "access", access, "data", "userrole"))

			} else if m.Option("username", m.Option("text.username")); !m.Confs("runtime", "user.route") && m.Confs("trust", "fresh") {
				// 免签节点
				m.Log("warn", "trust fresh %s", m.Option("node.route"))
				m.Option("trust", "fresh")

			} else if m.Cmds("ssh._right", "node", "trust", m.Option("node.route")) {
				// 可信节点
				m.Log("warn", "trust node %s", m.Option("node.route"))
				m.Option("trust", "node")

			} else if m.Options("user.route") && m.Cmds("ssh._right", "node", "trust", m.Option("user.route")) &&
				m.Cmds("ssh._right", "user", "trust", m.Option("node.route")) {
				// 可信用户
				m.Log("warn", "trust user %s", m.Option("user.route"))
				m.Option("trust", "user")

			} else if m.Options("work.route") && m.Cmds("ssh._right", "node", "trust", m.Option("work.route")) &&
				m.Cmds("ssh._right", "work", "check", m.Option("username"), m.Option("node.route")) {
				// 可信工作
				m.Log("warn", "trust work %s", m.Option("work.route"))
				m.Option("userrole", "tech")
				m.Option("trust", "work")

			} else if m.Option("userrole", "void"); m.Confs("trust", "none") {
				// 免签用户
				m.Log("warn", "trust none")
				m.Option("trust", "none")

			} else {
				// 验证失败
				m.Log("warn", "user error of %s", m.Option("node.route"))
				m.Echo("user error")
				return
			}

			// 用户角色
			if m.Log("info", "username: %s", m.Option("username")); !m.Options("userrole") {
				m.Option("userrole", kit.Select("void", m.Cmd("aaa.user", "role").Append("meta")))
			}
			m.Log("info", "userrole: %s", m.Option("userrole"))

			// 创建会话
			if !m.Options("sessid") {
				m.Option("sessid", m.Cmdx("aaa.user", "session", "select"))
				m.Cmd("aaa.auth", m.Cmdx("aaa.auth", m.Option("sessid"), "access", access),
					"data", "trust", m.Option("trust"), "userrole", m.Option("userrole"))
			}

			// 权限检查
			if !m.Cmds("aaa.role", m.Option("userrole"), "check", arg) {
				m.Echo("no right %s %s", "remote", arg[0])
				return
			}

			// 执行命令
			m.Log("time", "right: %v", m.Format("cost"))
			m.Cmdy(arg)
			return
		}},

		"login": {Name: "login address", Help: "网络连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				arg = append(arg, "shy@shylinux.com")
			}
			m.Start(arg[0], arg[0], arg...)
			time.Sleep(1000 * time.Millisecond)
			return
		}},
		"run": {Name: "run cmd", Help: "网络连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if ssh, ok := m.Target().Server.(*SSH); ok && ssh.relay != nil {
				ssh.relay <- m
				m.CallBack(true, func(msg *ctx.Message) (res *ctx.Message) {
					return nil
				}, append([]string{"_"}, arg...))
			}
			return
		}},
	},
}

func init() {
	ctx.Index.Register(Index, &SSH{Context: Index})
}
