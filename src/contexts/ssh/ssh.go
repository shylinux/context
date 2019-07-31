package ssh

import (
	"contexts/ctx"
	"toolkit"

	"encoding/hex"
	"io"
	"os"
	"path"
	"strings"
)

type SSH struct {
	*ctx.Context
}

func (ssh *SSH) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	return &SSH{Context: c}
}
func (ssh *SSH) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return ssh
}
func (ssh *SSH) Start(m *ctx.Message, arg ...string) bool {
	return true
}
func (ssh *SSH) Close(m *ctx.Message, arg ...string) bool {
	return false
}

var Index = &ctx.Context{Name: "ssh", Help: "集群中心",
	Caches: map[string]*ctx.Cache{
		"nnode": &ctx.Cache{Name: "nnode", Value: "0", Help: "节点数量"},
	},
	Configs: map[string]*ctx.Config{
		"componet": &ctx.Config{Name: "componet", Value: map[string]interface{}{
			"favor": []interface{}{
				map[string]interface{}{"componet_name": "clip", "componet_help": "粘贴板",
					"componet_tmpl": "componet", "componet_view": "Context", "componet_init": "",
					"componet_type": "public", "componet_ctx": "ssh", "componet_cmd": "_route",
					"componet_args": []interface{}{"$$", "context", "aaa", "clip"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "you", "imports": "plugin_you", "action": "auto"},
						map[string]interface{}{"type": "text", "name": "txt", "view": "long"},
						map[string]interface{}{"type": "button", "value": "存储", "action": "auto"},
					},
				},
				map[string]interface{}{"componet_name": "qrcode", "componet_help": "二维码",
					"componet_tmpl": "componet", "componet_view": "QRCode", "componet_init": "initQRCode.js",
					"componet_type": "public", "componet_ctx": "web.chat", "componet_cmd": "login",
					"componet_args": []interface{}{"qrcode"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "txt", "view": "long"},
						map[string]interface{}{"type": "button", "value": "生成"},
					},
				},
				map[string]interface{}{"componet_name": "salary", "componet_help": "工资单",
					"componet_tmpl": "componet", "componet_view": "Salary", "componet_init": "",
					"componet_type": "public", "componet_ctx": "web.chat", "componet_cmd": "salary",
					"componet_args": []interface{}{}, "inputs": []interface{}{
						map[string]interface{}{"label": "total", "type": "text", "name": "text"},
						map[string]interface{}{"label": "base", "type": "text", "name": "total", "value": "9000"},
						map[string]interface{}{"type": "button", "value": "计算"},
					},
				},
			},
			"index": []interface{}{
				map[string]interface{}{"componet_name": "dir", "componet_help": "目录",
					"componet_tmpl": "componet", "componet_view": "Context", "componet_init": "",
					"componet_type": "private", "componet_ctx": "ssh", "componet_cmd": "_route",
					"componet_args": []interface{}{"$$", "context", "nfs", "dir", "$$", "time", "size", "line", "path"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "pod", "imports": []interface{}{"plugin_you", "plugin_pod"}, "action": "auto"},
						map[string]interface{}{"type": "text", "name": "dir", "value": "usr/script", "imports": "plugin_dir", "action": "auto", "view": "long"},
						map[string]interface{}{"type": "button", "value": "查看", "action": "auto"},
						map[string]interface{}{"type": "button", "value": "回退", "click": "Back"},
					},
					"display":  map[string]interface{}{"deal": "editor"},
					"exports":  []interface{}{"dir", "", "dir"},
					"dir_root": []interface{}{"/"},
				},
				map[string]interface{}{"componet_name": "commit", "componet_help": "提交",
					"componet_tmpl": "componet", "componet_view": "Context", "componet_init": "",
					"componet_type": "private", "componet_ctx": "nfs", "componet_cmd": "git",
					"componet_args": []interface{}{}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "dir", "view": "long"},
						map[string]interface{}{"type": "select", "name": "cmd", "values": []interface{}{
							"add", "commit", "checkout", "merge", "init",
						}},
						map[string]interface{}{"type": "text", "name": "commit", "view": "long"},
						map[string]interface{}{"type": "button", "value": "执行"},
					},
				},
				map[string]interface{}{"componet_name": "status", "componet_help": "记录",
					"componet_tmpl": "componet", "componet_view": "Context", "componet_init": "",
					"componet_type": "private", "componet_ctx": "nfs", "componet_cmd": "git",
					"componet_args": []interface{}{}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "dir", "view": "long"},
						map[string]interface{}{"type": "select", "name": "cmd", "values": []interface{}{
							"branch", "status", "diff", "log", "sum", "push", "update",
						}},
						map[string]interface{}{"type": "button", "value": "执行"},
					},
				},
				map[string]interface{}{"componet_name": "spide", "componet_help": "爬虫",
					"componet_tmpl": "componet", "componet_view": "Context", "componet_init": "",
					"componet_type": "private", "componet_ctx": "ssh", "componet_cmd": "_route",
					"componet_args": []interface{}{"$$", "context", "web", "spide"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "pod", "imports": "plugin_pod"},
						map[string]interface{}{"type": "button", "value": "执行"},
					},
					"exports": []interface{}{"site", "key"},
				},
				map[string]interface{}{"componet_name": "post", "componet_help": "请求",
					"componet_tmpl": "componet", "componet_view": "Context", "componet_init": "",
					"componet_type": "private", "componet_ctx": "ssh", "componet_cmd": "_route",
					"componet_args": []interface{}{"$$", "context", "web", "post", "$$", "content_type", "application/json", "parse", "json"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "pod", "imports": "plugin_pod"},
						map[string]interface{}{"type": "text", "name": "spide", "value": "zuo", "imports": "plugin_site"},
						map[string]interface{}{"type": "text", "name": "url", "value": "/", "view": "long"},
						map[string]interface{}{"type": "button", "value": "执行"},
					},
				},
				map[string]interface{}{"componet_name": "get", "componet_help": "请求",
					"componet_tmpl": "componet", "componet_view": "Context", "componet_init": "",
					"componet_type": "private", "componet_ctx": "ssh", "componet_cmd": "_route",
					"componet_args": []interface{}{"$$", "context", "web", "get", "method", "GET", "parse", "json"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "pod", "imports": "plugin_pod"},
						map[string]interface{}{"type": "text", "name": "spide", "imports": "plugin_site"},
						map[string]interface{}{"type": "text", "name": "url", "value": "/", "view": "long"},
						map[string]interface{}{"type": "button", "value": "执行"},
					},
				},
				map[string]interface{}{"componet_name": "ifconfig", "componet_help": "ifconfig",
					"componet_tmpl": "componet", "componet_view": "Context", "componet_init": "",
					"componet_type": "private", "componet_ctx": "tcp", "componet_cmd": "ifconfig",
					"componet_args": []interface{}{}, "inputs": []interface{}{
						map[string]interface{}{"type": "button", "value": "网卡"},
					},
				},
				map[string]interface{}{"componet_name": "proc", "componet_help": "proc",
					"componet_tmpl": "componet", "componet_view": "Company", "componet_init": "",
					"componet_type": "private", "componet_ctx": "ssh", "componet_cmd": "_route",
					"componet_args": []interface{}{"$$", "context", "cli", "proc"}, "inputs": []interface{}{
						map[string]interface{}{"type": "text", "name": "pod", "value": "", "imports": "plugin_pod"},
						map[string]interface{}{"type": "text", "name": "cmd", "value": ""},
						map[string]interface{}{"type": "text", "name": "cmd", "view": "long"},
						map[string]interface{}{"type": "button", "value": "执行"},
					},
				},
			},
		}, Help: "组件列表"},

		"flow":  &ctx.Config{Name: "flow", Value: map[string]interface{}{}, Help: "聊天群组"},
		"work":  &ctx.Config{Name: "work", Value: map[string]interface{}{}, Help: "工作信息"},
		"node":  &ctx.Config{Name: "node", Value: map[string]interface{}{}, Help: "节点信息"},
		"timer": &ctx.Config{Name: "timer", Value: map[string]interface{}{"interval": "10s", "timer": ""}, Help: "断线重连"},
		"trust": &ctx.Config{Name: "trust", Value: map[string]interface{}{
			"renew": true, "fresh": false, "user": true, "up": true,
		}, Help: "可信节点"},
	},
	Commands: map[string]*ctx.Command{
		"_init": &ctx.Command{Name: "_init", Help: "启动初始化", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
		"_node": &ctx.Command{Name: "_node [init|create name type module|delete name]", Help: "节点", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Confm("ssh.node", func(key string, value map[string]interface{}) {
					m.Push("create_time", value["create_time"])
					m.Push("pod", key)
					m.Push("type", value["type"])
				})
				m.Sort("pod").Table()
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
		"user": &ctx.Command{Name: "user [init|create|trust [node]]", Help: "用户管理, init: 用户节点, create: 用户创建, trust: 用户授权", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
					m.Conf("trust", arg[1], kit.Right(arg[2]))
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
		"work": &ctx.Command{Name: "work [serve|create|search]", Help: "工作管理, serve: 创建组织, create: 创建员工, search: 搜索员工", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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

				if n := m.Cmdx("ssh._route", work, "_check", "work", name, m.Conf("runtime", "user.route")); n != "" {
					m.Conf("runtime", "work.route", work)
					m.Conf("runtime", "work.name", n)
					m.Echo(n)
				} else {
					m.Echo("error: %s from %s", name, work)
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
		"tool": &ctx.Command{Name: "tool [group index][run group index chatid arg...]", Help: "工具", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
			if len(arg) == 0 {
				m.Confm("componet", func(key string, index int, value map[string]interface{}) {
					if kit.Format(value["componet_type"]) != "public" && m.Option("userrole") != "root" {
						return
					}

					m.Push("key", key)
					m.Push("index", index)
					m.Push("name", value["componet_name"])
					m.Push("help", value["componet_help"])
				})
				m.Table()
				return
			}

			switch arg[0] {
			case "run":
				tool := m.Confm("componet", []string{arg[1], arg[2]})
				if m.Option("userrole") != "root" {
					switch kit.Format(tool["componet_type"]) {
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

				msg := m.Find(kit.Format(tool["componet_ctx"]))
				if option, ok := tool["options"].(map[string]interface{}); ok {
					for k, v := range option {
						msg.Option(k, v)
					}
				}

				arg = arg[4:]
				args := []string{}
				for _, v := range kit.Trans(tool["componet_args"]) {
					if v == "$$" {
						if len(arg) > 0 {
							args = append(args, arg[0])
							arg = arg[1:]
						} else {
							args = append(args, "")
						}
					} else {
						args = append(args, msg.Parse(v))
					}
				}
				msg.Cmd(tool["componet_cmd"], args, arg).CopyTo(m)

			default:
				m.Confm("componet", arg[0:], func(value map[string]interface{}) {
					if kit.Format(value["componet_type"]) == "private" && m.Option("userrole") != "root" {
						return
					}

					m.Push("name", value["componet_name"])
					m.Push("help", value["componet_help"])
					m.Push("view", value["componet_view"])
					if kit.Right(value["componet_init"]) {
						script := m.Cmdx("nfs.load", path.Join(m.Conf("cli.publish", "path"), arg[0], kit.Format(value["componet_init"])), -1)
						if script == "" {
							script = m.Cmdx("nfs.load", path.Join("usr/librarys/plugin", kit.Format(value["componet_init"])), -1)
						}
						m.Push("init", script)
					} else {
						m.Push("init", "")
					}
					m.Push("inputs", kit.Format(value["inputs"]))
					m.Push("exports", kit.Format(value["exports"]))
					m.Push("display", kit.Format(value["display"]))
				})
				m.Table()
			}
			return
		}},

		"remote": &ctx.Command{Name: "remote auto|dial|listen args...", Help: "连接", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
				m.Append("node.route", m.Conf("runtime", "node.route"))
				m.Append("user.route", m.Conf("runtime", "user.route"))
				m.Append("work.route", m.Conf("runtime", "work.route"))
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
						m.Cmd("nfs.source", m.Cmdx("nfs.hash", m.Append("work.script")))
						return nil
					}, "send", "", "_add", m.Conf("runtime", "node.name"), m.Conf("runtime", "node.type"), m.Conf("runtime", "boot.ctx_type"))
					return nil
				}, "nfs.socket", arg)

			default:
				m.Cmd("_route", arg)
			}
			return
		}},
		"_route": &ctx.Command{Name: "_route", Help: "路由", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
					// 查找失败
					m.Echo("error: not found %s", names[0]).Back(m)
					return
				}

				// 路由转发
				for _, p := range ps {
					m.Find(p, true).Copy(m, "option").CallBack(sync, func(sub *ctx.Message) *ctx.Message {
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

			} else {
				// 执行命令
				m.Cmd("_exec", arg)
			}
			return
		}},
		"_right": &ctx.Command{Name: "_right [node|user|work]", Help: []string{"认证",
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
						m.Cmd("aaa.auth", "username", m.Option("username"), "userrole", "void")
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
		"_check": &ctx.Command{Name: "_check node|user|work", Help: []string{"验签",
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
		"_exec": &ctx.Command{Name: "_exec", Help: "命令", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) (e error) {
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
			if !m.Cmds("aaa.role", m.Option("userrole"), "check", "remote", arg[0]) {
				m.Echo("no right %s %s", "remote", arg[0])
				return
			}

			// 执行命令
			m.Cmdy(arg)
			return
		}},
	},
}

func init() {
	ctx.Index.Register(Index, &SSH{Context: Index})
}
